[CmdletBinding()]
param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\movidesk-cli"
)

$ErrorActionPreference = "Stop"

function Get-ReleaseUrl {
    param([string]$RequestedVersion)

    if ($RequestedVersion -eq "latest") {
        return "https://api.github.com/repos/jonasandre/movidesk-cli/releases/latest"
    }

    $tag = if ($RequestedVersion.StartsWith("v")) { $RequestedVersion } else { "v$RequestedVersion" }
    return "https://api.github.com/repos/jonasandre/movidesk-cli/releases/tags/$tag"
}

function Get-AssetArch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if (-not $arch) {
        throw "Nao foi possivel detectar a arquitetura do Windows."
    }

    switch ($arch.ToUpperInvariant()) {
        "ARM64" { return "arm64" }
        "AMD64" { return "amd64" }
        default { throw "Arquitetura do Windows nao suportada: $arch" }
    }
}

function Add-PathEntry {
    param([string]$Entry)

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $entries = @()
    if ($userPath) {
        $entries = $userPath.Split(";") | Where-Object { $_ }
    }

    if ($entries -notcontains $Entry) {
        $newPath = (@($entries) + $Entry) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    }

    $sessionEntries = $env:Path.Split(";") | Where-Object { $_ }
    if ($sessionEntries -notcontains $Entry) {
        $env:Path = (@($sessionEntries) + $Entry) -join ";"
    }
}

function Install-ClaudeSkillIfPresent {
    param([string]$ReleaseTag)

    $claudeDir = Join-Path $env:USERPROFILE ".claude"
    if (-not (Test-Path $claudeDir)) {
        Write-Host "Claude nao detectado em $claudeDir; pulando instalacao da skill movidesk-mcp."
        return $false
    }

    $skillDir = Join-Path $claudeDir "skills\movidesk-mcp"
    $skillPath = Join-Path $skillDir "SKILL.md"
    $skillUrl = "https://raw.githubusercontent.com/jonasandre/movidesk-cli/$ReleaseTag/.claude/skills/movidesk-mcp/SKILL.md"

    New-Item -ItemType Directory -Force -Path $skillDir | Out-Null
    Invoke-WebRequest -Uri $skillUrl -OutFile $skillPath
    Write-Host "Skill movidesk-mcp instalada em $skillPath"
    return $true
}

$release = Invoke-RestMethod -Uri (Get-ReleaseUrl -RequestedVersion $Version)
$resolvedVersion = if ($release.tag_name.StartsWith("v")) { $release.tag_name.Substring(1) } else { $release.tag_name }
$assetArch = Get-AssetArch
$assetName = "movidesk-cli_${resolvedVersion}_windows_${assetArch}.zip"
$asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1

if (-not $asset) {
    throw "Asset nao encontrado no release $($release.tag_name): $assetName"
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$resolvedInstallDir = (Resolve-Path $InstallDir).Path
$downloadPath = Join-Path ([IO.Path]::GetTempPath()) $asset.name

try {
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $downloadPath
    Expand-Archive -Path $downloadPath -DestinationPath $resolvedInstallDir -Force
}
finally {
    if (Test-Path $downloadPath) {
        Remove-Item -Path $downloadPath -Force
    }
}

Add-PathEntry -Entry $resolvedInstallDir
$null = Install-ClaudeSkillIfPresent -ReleaseTag $release.tag_name

$exePath = Join-Path $resolvedInstallDir "movidesk-cli.exe"
if (-not (Test-Path $exePath)) {
    throw "Instalacao concluida, mas o executavel nao foi encontrado em $exePath"
}

Write-Host "movidesk-cli instalado em $exePath"
Write-Host "Path do usuario atualizado com $resolvedInstallDir"
Write-Host "Abra um novo terminal ou rode 'movidesk-cli --help' neste mesmo PowerShell."
