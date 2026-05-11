## movidesk-cli auth login

Adiciona ou atualiza um tenant e armazena seu token de API

### Synopsis

Adiciona ou atualiza um tenant do Movidesk. O token é lido de um prompt
oculto (ou da stdin quando não há TTY) e salvo no chaveiro do sistema operacional
quando disponível; caso contrário, em arquivo criptografado em ~/.movidesk.

Por padrão, login valida o token fazendo GET /persons?$top=1 contra a base URL
configurada. Use --skip-verify para pular essa verificação.

Após validar o token, login pode solicitar interativamente um usuário padrão
(Cod. Ref.) que será injetado automaticamente como createdBy nas escritas que
exigem atribuição. Passe --user <id> para definir sem prompt, ou deixe vazio
para pular. Use --skip-verify-user para pular a checagem de existência (útil
quando o token não tem permissão de ler a API de pessoas).

```
movidesk-cli auth login [flags]
```

### Options

```
      --base-url string    sobrepõe a base URL da API (sandbox)
  -h, --help               help for login
      --label string       rótulo legível, ex.: "Acme Prod"
      --make-default       define este tenant como o atual
      --skip-verify        não valida o token contra a API
      --skip-verify-user   pula a checagem de existência do usuário padrão
      --tenant string      nome do tenant (obrigatório)
```

### Options inherited from parent commands

```
      --compact         JSON compacto (sem indentação)
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli auth](movidesk-cli_auth.md)	 - Gerencia tokens e tenants do Movidesk

