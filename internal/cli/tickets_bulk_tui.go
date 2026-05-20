package cli

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// pickInteractive opens a Bubble Tea selector for bulk operations. Returns
// the subset the user confirmed, or an error if the user aborted.
func pickInteractive(rows []bulkCandidate) ([]bulkCandidate, error) {
	m := newPickerModel(rows)
	prog := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	final, err := prog.Run()
	if err != nil {
		return nil, fmt.Errorf("seletor TUI: %w", err)
	}
	pm := final.(*pickerModel)
	if pm.aborted {
		return nil, errors.New("seleção cancelada")
	}
	out := make([]bulkCandidate, 0, len(pm.selected))
	for i, r := range pm.rows {
		if pm.selected[i] {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("nenhum chamado selecionado")
	}
	return out, nil
}

type pickerModel struct {
	rows     []bulkCandidate
	visible  []int
	selected map[int]bool
	cursor   int
	filter   string
	search   bool
	aborted  bool
	width    int
	height   int
}

func newPickerModel(rows []bulkCandidate) *pickerModel {
	m := &pickerModel{
		rows:     rows,
		selected: map[int]bool{},
	}
	m.applyFilter()
	return m
}

func (m *pickerModel) Init() tea.Cmd { return nil }

func (m *pickerModel) applyFilter() {
	m.visible = m.visible[:0]
	needle := strings.ToLower(strings.TrimSpace(m.filter))
	for i, r := range m.rows {
		if needle == "" || matchesNeedle(r, needle) {
			m.visible = append(m.visible, i)
		}
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func matchesNeedle(r bulkCandidate, needle string) bool {
	if strings.Contains(strings.ToLower(r.Subject), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(r.Status), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(r.OwnerTeam), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(r.Justification), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(r.clientLabel()), needle) {
		return true
	}
	if strings.Contains(strings.ToLower(r.visibility()), needle) {
		return true
	}
	if strings.Contains(fmt.Sprintf("%d", r.ID), needle) {
		return true
	}
	if r.Owner != nil && strings.Contains(strings.ToLower(r.Owner.BusinessName), needle) {
		return true
	}
	return false
}

func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.search {
			return m.updateSearch(msg)
		}
		return m.updateBrowse(msg)
	}
	return m, nil
}

func (m *pickerModel) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.aborted = true
		return m, tea.Quit
	case "enter":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = len(m.visible) - 1
	case "pgup":
		m.cursor -= m.pageSize()
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "pgdown":
		m.cursor += m.pageSize()
		if m.cursor >= len(m.visible) {
			m.cursor = len(m.visible) - 1
		}
	case " ", "x":
		if len(m.visible) == 0 {
			return m, nil
		}
		idx := m.visible[m.cursor]
		m.selected[idx] = !m.selected[idx]
	case "a":
		anyOff := false
		for _, i := range m.visible {
			if !m.selected[i] {
				anyOff = true
				break
			}
		}
		for _, i := range m.visible {
			m.selected[i] = anyOff
		}
	case "n":
		for _, i := range m.visible {
			m.selected[i] = false
		}
	case "/":
		m.search = true
	}
	return m, nil
}

func (m *pickerModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.search = false
	case tea.KeyEnter:
		m.search = false
	case tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
	case tea.KeyRunes, tea.KeySpace:
		m.filter += string(msg.Runes)
		m.applyFilter()
	}
	return m, nil
}

func (m *pickerModel) pageSize() int {
	if m.height < 8 {
		return 5
	}
	return m.height - 6
}

var (
	stylTitle    = lipgloss.NewStyle().Bold(true)
	stylHint     = lipgloss.NewStyle().Faint(true)
	stylSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	stylCursor   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
)

func (m *pickerModel) View() string {
	var b strings.Builder
	count := 0
	for _, v := range m.selected {
		if v {
			count++
		}
	}
	b.WriteString(stylTitle.Render(fmt.Sprintf("Seletor de chamados — %d/%d marcados", count, len(m.rows))))
	b.WriteString("\n")
	hint := "↑/↓ move • espaço marca • a marca todos visíveis • n limpa • / busca • enter confirma • esc/q cancela"
	if m.search {
		hint = fmt.Sprintf("buscar: %s_   (esc/enter encerra)", m.filter)
	} else if m.filter != "" {
		hint = fmt.Sprintf("filtro: %q  •  /  edita  •  enter confirma", m.filter)
	}
	b.WriteString(stylHint.Render(hint))
	b.WriteString("\n\n")

	if len(m.visible) == 0 {
		b.WriteString(stylHint.Render("  (nenhum chamado bate com o filtro)"))
		b.WriteString("\n")
		return b.String()
	}

	header := fmt.Sprintf("      %-7s %-8s %-12s %-5s %-18s %-28s %s",
		"id", "visib.", "status", "idade", "motivo", "cliente / organização", "assunto")
	b.WriteString(stylHint.Render(header))
	b.WriteString("\n")

	page := m.pageSize()
	start := m.cursor - page/2
	if start < 0 {
		start = 0
	}
	end := start + page
	if end > len(m.visible) {
		end = len(m.visible)
		start = end - page
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		idx := m.visible[i]
		r := m.rows[idx]
		mark := "[ ]"
		if m.selected[idx] {
			mark = "[x]"
		}
		age := daysSince(r.LastUpdate)
		ageCol := truncate(age, 5)
		just := truncate(r.Justification, 18)
		client := truncate(r.clientLabel(), 28)
		line := fmt.Sprintf("  %s #%-6d %-8s %-12s %-5s %-18s %-28s %s",
			mark, r.ID, r.visibility(), truncate(r.Status, 12), ageCol, just, client, truncate(r.Subject, 50))
		if i == m.cursor {
			line = "›" + line[1:]
			line = stylCursor.Render(line)
		} else if m.selected[idx] {
			line = stylSelected.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	if end < len(m.visible) {
		b.WriteString(stylHint.Render(fmt.Sprintf("  … %d a mais abaixo", len(m.visible)-end)))
		b.WriteString("\n")
	}
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
