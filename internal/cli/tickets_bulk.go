package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/tickets"
)

// bulkCandidate is a minimal ticket projection used for selection + preview.
type bulkCandidate struct {
	ID         int    `json:"id"`
	Type       int    `json:"type"`
	Subject    string `json:"subject"`
	Status     string `json:"status"`
	BaseStatus string `json:"baseStatus"`
	OwnerTeam  string `json:"ownerTeam"`
	Owner      *struct {
		BusinessName string `json:"businessName"`
	} `json:"owner,omitempty"`
	CreatedDate   string                `json:"createdDate"`
	LastUpdate    string                `json:"lastUpdate"`
	Justification string                `json:"justification"`
	Clients       []bulkCandidateClient `json:"clients,omitempty"`
}

// visibility renders Movidesk ticket.type as "público"/"interno". Type 1 is
// internal and type 2 is public; missing/unknown returns "—".
func (c bulkCandidate) visibility() string {
	switch c.Type {
	case 1:
		return "interno"
	case 2:
		return "público"
	default:
		return "—"
	}
}

type bulkCandidateClient struct {
	BusinessName string `json:"businessName"`
	Organization *struct {
		BusinessName string `json:"businessName"`
	} `json:"organization,omitempty"`
}

// clientLabel renders "Org — Cliente" or whatever pieces are present, in a
// compact form suitable for the picker/preview.
func (c bulkCandidate) clientLabel() string {
	if len(c.Clients) == 0 {
		return ""
	}
	cl := c.Clients[0]
	org := ""
	if cl.Organization != nil {
		org = cl.Organization.BusinessName
	}
	switch {
	case org != "" && cl.BusinessName != "":
		return org + " — " + cl.BusinessName
	case org != "":
		return org
	default:
		return cl.BusinessName
	}
}

// daysSince parses a Movidesk timestamp and returns a "Nd" string. Empty when
// the timestamp is missing or unparseable.
func daysSince(ts string) string {
	if ts == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999Z", "2006-01-02T15:04:05Z"} {
		t, err := time.Parse(layout, ts)
		if err == nil {
			d := int(time.Since(t).Hours() / 24)
			if d < 0 {
				d = 0
			}
			return fmt.Sprintf("%dd", d)
		}
	}
	return ""
}

// bulkSelection captures how the user picks tickets to act on.
type bulkSelection struct {
	of            odataFlags
	ids           []int
	idsFile       string
	pick          bool
	allFromFilter bool
	source        string
}

func (s *bulkSelection) bind(cmd *cobra.Command) {
	s.of.bind(cmd)
	cmd.Flags().IntSliceVar(&s.ids, "ids", nil, "lista de ids separados por vírgula (pula a listagem)")
	cmd.Flags().StringVar(&s.idsFile, "ids-file", "", "arquivo com um id por linha (# inicia comentário); aceita também IDs separados por vírgula")
	cmd.Flags().BoolVar(&s.pick, "pick", false, "abre seletor TUI mesmo quando --ids/--ids-file forem informados")
	cmd.Flags().BoolVar(&s.allFromFilter, "all-from-filter", false, "usa todos os resultados do --filter sem abrir o seletor (necessário em ambientes sem TTY)")
	cmd.Flags().StringVar(&s.source, "source", "live", "fonte da listagem: live (/tickets, últimos 90d), past (/tickets/past, arquivados), both (mescla as duas)")
}

// bulkExec captures runtime/safety knobs.
type bulkExec struct {
	dryRun          bool
	force           bool
	continueOnError bool
	reportPath      string
}

func (e *bulkExec) bind(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&e.dryRun, "dry-run", false, "mostra o que seria feito sem chamar a API")
	cmd.Flags().BoolVar(&e.force, "force", false, "pula o prompt de confirmação (obrigatório fora de TTY)")
	cmd.Flags().BoolVar(&e.continueOnError, "continue-on-error", false, "continua o lote mesmo após uma falha individual")
	cmd.Flags().StringVar(&e.reportPath, "report", "", "grava resultado por ticket em arquivo JSONL")
}

func newTicketsBulkUpdateCmd() *cobra.Command {
	var (
		sel          bulkSelection
		exec         bulkExec
		file         string
		template     string
		templateFile string
		sets         []string
	)
	cmd := &cobra.Command{
		Use:   "bulk-update",
		Short: "Aplica o mesmo PATCH a vários chamados (com seletor interativo)",
		Long: `Atualiza vários chamados de uma só vez. Os alvos podem vir de --ids,
--ids-file ou de um --filter OData (com seletor TUI por padrão quando em TTY).

O corpo do PATCH é montado por --file, --from-template[-file] ou --set chave=valor
(mesmas regras de 'tickets update').

Respeita o limite de 10 req/min: lotes grandes podem demorar. Use --report para
gravar o status de cada chamado em um JSONL e --continue-on-error para não abortar.`,
		Example: `  # encerra em lote por filtro com confirmação interativa
  movidesk-cli tickets bulk-update \
    --filter "baseStatus eq 'Stopped' and ownerTeam eq 'Qlik'" --top 100 \
    --set status=Resolvido --set justification=Resolvido

  # reaproveita IDs de uma listagem prévia
  movidesk-cli tickets list --filter "..." --output json | jq -r '.[].id' > /tmp/ids
  movidesk-cli tickets bulk-update --ids-file /tmp/ids --set ownerTeam=Suporte`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := loadBody(file, template, templateFile, sets)
			if err != nil {
				return err
			}
			if len(body) == 0 {
				return errors.New("nenhum campo para atualizar; passe --file, --from-template[-file] ou --set chave=valor")
			}
			return runBulk(cmd, &sel, &exec, body, "tickets bulk-update")
		},
	}
	sel.bind(cmd)
	exec.bind(cmd)
	cmd.Flags().StringVarP(&file, "file", "f", "", "caminho do corpo JSON de patch")
	cmd.Flags().StringVar(&template, "from-template", "", "carrega ~/.movidesk/templates/<nome>.json")
	cmd.Flags().StringVar(&templateFile, "from-template-file", "", "carrega template de um caminho específico")
	cmd.Flags().StringSliceVar(&sets, "set", nil, "sobrescreve campos inline, ex.: --set status=Resolvido")
	return cmd
}

func newTicketsBulkCloseCmd() *cobra.Command {
	var (
		sel           bulkSelection
		exec          bulkExec
		message       string
		justification string
		status        string
		actionType    int
		public        bool
	)
	cmd := &cobra.Command{
		Use:   "bulk-close",
		Short: "Encerra vários chamados em lote, registrando uma ação com a mensagem",
		Long: `Atualiza o status para 'Resolvido' (configurável) e adiciona uma ação
com a mensagem informada em cada chamado selecionado. Equivale a um
'tickets bulk-update' que monta o corpo automaticamente.

Use --public para registrar a ação como pública (visível pelo cliente). Sem --public
a ação é interna (type=1). O nome exato do status deve bater com o configurado
no tenant (padrão: "Resolvido"). O campo justification é sempre enviado (o Movidesk
exige ao mudar Status); fica vazio quando --justification não é informado, o que
funciona pra status sem motivos cadastrados.`,
		Example: `  movidesk-cli tickets bulk-close \
    --filter "baseStatus eq 'Stopped' and lastUpdate lt 2026-05-01T00:00:00.000Z" \
    --message "Fechado por inatividade após 30 dias sem retorno do cliente"

  # variação pública e com justificativa customizada
  movidesk-cli tickets bulk-close --ids 12,34,56 \
    --message "Concluído conforme alinhamento" --public --justification "Concluído"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(message) == "" {
				return errors.New("--message é obrigatório (texto da ação de fechamento)")
			}
			if public {
				actionType = 2
			}
			if actionType == 0 {
				actionType = 1
			}
			if actionType != 1 && actionType != 2 {
				return fmt.Errorf("--action-type inválido %d (1=interna, 2=pública)", actionType)
			}
			if status == "" {
				status = "Resolvido"
			}
			body := map[string]any{
				"status":        status,
				"justification": justification,
				"actions": []map[string]any{{
					"type":        actionType,
					"origin":      9,
					"description": message,
				}},
			}
			return runBulk(cmd, &sel, &exec, body, "tickets bulk-close")
		},
	}
	sel.bind(cmd)
	exec.bind(cmd)
	cmd.Flags().StringVar(&message, "message", "", "texto da ação de fechamento (obrigatório)")
	cmd.Flags().StringVar(&justification, "justification", "", "justificativa do ticket (padrão: igual a --status)")
	cmd.Flags().StringVar(&status, "status", "", "nome do status final (padrão: Resolvido)")
	cmd.Flags().IntVar(&actionType, "action-type", 0, "tipo da ação: 1=interna, 2=pública (padrão: 1)")
	cmd.Flags().BoolVar(&public, "public", false, "atalho para --action-type=2")
	_ = cmd.MarkFlagRequired("message")
	return cmd
}

func runBulk(cmd *cobra.Command, sel *bulkSelection, exec *bulkExec, body map[string]any, action string) error {
	r, err := resolveClient(cmd)
	if err != nil {
		return err
	}
	svc := tickets.New(r.client)

	injectActionsCreatedBy(body, r.userID)

	candidates, err := resolveCandidates(cmd, svc, sel)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return errors.New("nenhum chamado selecionado")
	}

	stderr := cmd.ErrOrStderr()
	printPreview(stderr, candidates, body, action)

	if exec.dryRun {
		fmt.Fprintf(stderr, "\n[dry-run] nada foi enviado.\n")
		return nil
	}

	if !exec.force {
		if err := confirm(cmd, fmt.Sprintf("Confirmar %s em %d chamado(s)?", action, len(candidates))); err != nil {
			return err
		}
	}

	report, closeReport, err := openReport(exec.reportPath)
	if err != nil {
		return err
	}
	defer closeReport()

	var okCount, failCount int
	for i, c := range candidates {
		fmt.Fprintf(stderr, "[%d/%d] PATCH ticket %d…", i+1, len(candidates), c.ID)
		_, perr := svc.Update(cmd.Context(), c.ID, body)
		if perr != nil {
			failCount++
			fmt.Fprintf(stderr, " falhou: %v\n", perr)
			writeReport(report, c.ID, false, perr.Error())
			if !exec.continueOnError {
				return fmt.Errorf("ticket %d: %w (use --continue-on-error pra seguir adiante)", c.ID, perr)
			}
			continue
		}
		okCount++
		fmt.Fprintln(stderr, " ok")
		writeReport(report, c.ID, true, "")
	}

	fmt.Fprintf(stderr, "\nResumo: %d ok, %d falha(s) de %d total.\n", okCount, failCount, len(candidates))
	if exec.reportPath != "" {
		fmt.Fprintf(stderr, "Relatório: %s\n", exec.reportPath)
	}
	if failCount > 0 && !exec.continueOnError {
		// Won't happen because we returned earlier, but keeps exit semantics explicit.
		return fmt.Errorf("%d falha(s) durante o lote", failCount)
	}
	if failCount > 0 {
		return fmt.Errorf("%d de %d chamado(s) falharam", failCount, len(candidates))
	}
	return nil
}

// resolveCandidates picks tickets from --ids, --ids-file, or --filter (with TUI).
func resolveCandidates(cmd *cobra.Command, svc *tickets.Service, sel *bulkSelection) ([]bulkCandidate, error) {
	idsProvided := len(sel.ids) > 0 || sel.idsFile != ""
	if idsProvided && (sel.of.filter != "" || sel.of.all || sel.of.top > 0) && !sel.pick {
		return nil, errors.New("use --ids/--ids-file OU --filter (combine com --pick pra abrir o seletor mesmo com IDs)")
	}

	if idsProvided && !sel.pick {
		return candidatesFromIDs(sel)
	}

	if sel.of.filter == "" {
		return nil, errors.New("informe --filter (com OData) ou --ids/--ids-file")
	}

	listed, err := listCandidates(cmd, svc, sel)
	if err != nil {
		return nil, err
	}
	if len(listed) == 0 {
		return nil, errors.New("o filtro não retornou nenhum chamado")
	}

	if sel.allFromFilter {
		return listed, nil
	}
	if !isStdinTTY() {
		return nil, errors.New("ambiente não-TTY: passe --all-from-filter pra usar todos os resultados, ou rode em terminal")
	}
	return pickInteractive(listed)
}

func candidatesFromIDs(sel *bulkSelection) ([]bulkCandidate, error) {
	idSet := map[int]struct{}{}
	add := func(id int) {
		if id <= 0 {
			return
		}
		idSet[id] = struct{}{}
	}
	for _, id := range sel.ids {
		add(id)
	}
	if sel.idsFile != "" {
		raw, err := os.ReadFile(sel.idsFile)
		if err != nil {
			return nil, fmt.Errorf("ler --ids-file: %w", err)
		}
		for _, tok := range tokenizeIDFile(string(raw)) {
			id, err := strconv.Atoi(tok)
			if err != nil {
				return nil, fmt.Errorf("--ids-file: valor %q não é um id válido", tok)
			}
			add(id)
		}
	}
	if len(idSet) == 0 {
		return nil, errors.New("nenhum id válido informado")
	}
	out := make([]bulkCandidate, 0, len(idSet))
	for id := range idSet {
		out = append(out, bulkCandidate{ID: id})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func tokenizeIDFile(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		for _, part := range strings.FieldsFunc(line, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\r' }) {
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func listCandidates(cmd *cobra.Command, svc *tickets.Service, sel *bulkSelection) ([]bulkCandidate, error) {
	source := strings.ToLower(strings.TrimSpace(sel.source))
	if source == "" {
		source = "live"
	}
	switch source {
	case "live":
		return fetchCandidates(cmd, svc, sel, false)
	case "past":
		return fetchCandidates(cmd, svc, sel, true)
	case "both":
		live, err := fetchCandidates(cmd, svc, sel, false)
		if err != nil {
			return nil, fmt.Errorf("source=both: live: %w", err)
		}
		past, err := fetchCandidates(cmd, svc, sel, true)
		if err != nil {
			return nil, fmt.Errorf("source=both: past: %w", err)
		}
		return dedupCandidates(append(live, past...)), nil
	default:
		return nil, fmt.Errorf("--source inválido %q (use live, past ou both)", sel.source)
	}
}

// fetchCandidates lists from /tickets or /tickets/past using the same OData
// flags. Forces a lean projection so the TUI selector stays snappy.
func fetchCandidates(cmd *cobra.Command, svc *tickets.Service, sel *bulkSelection, past bool) ([]bulkCandidate, error) {
	q := sel.of.query()
	q.Select = []string{"id", "type", "subject", "status", "baseStatus", "owner", "ownerTeam", "createdDate", "lastUpdate", "justification"}
	if len(q.Expand) == 0 {
		q.Expand = []string{"clients($expand=organization)"}
	}

	var raw []byte
	var err error
	if sel.of.all {
		if q.Top == 0 {
			q.Top = 100
		}
		var pages []json.RawMessage
		if past {
			pages, err = svc.PaginatePast(cmd.Context(), q, q.Top, sel.of.max)
		} else {
			pages, err = svc.Paginate(cmd.Context(), q, false, q.Top, sel.of.max)
		}
		if err != nil {
			return nil, err
		}
		raw, err = json.Marshal(pages)
		if err != nil {
			return nil, err
		}
	} else {
		if past {
			raw, err = svc.Past(cmd.Context(), q)
		} else {
			raw, err = svc.List(cmd.Context(), q, false)
		}
		if err != nil {
			return nil, err
		}
	}

	var rows []bulkCandidate
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("decodificar listagem: %w", err)
	}
	return rows, nil
}

func dedupCandidates(in []bulkCandidate) []bulkCandidate {
	seen := map[int]struct{}{}
	out := make([]bulkCandidate, 0, len(in))
	for _, c := range in {
		if _, ok := seen[c.ID]; ok {
			continue
		}
		seen[c.ID] = struct{}{}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func printPreview(w io.Writer, rows []bulkCandidate, body map[string]any, action string) {
	fmt.Fprintf(w, "%s — %d chamado(s) selecionado(s):\n", action, len(rows))
	limit := len(rows)
	if limit > 10 {
		limit = 10
	}
	for i := 0; i < limit; i++ {
		c := rows[i]
		age := daysSince(c.LastUpdate)
		if age != "" {
			age = " · " + age + " desde última alt."
		}
		client := c.clientLabel()
		if client != "" {
			client = "  cliente: " + truncate(client, 50)
		}
		just := c.Justification
		if just != "" {
			just = "  motivo: " + truncate(just, 30)
		}
		fmt.Fprintf(w, "  #%d  [%s · %s]%s%s%s\n      %s\n", c.ID, c.Status, c.visibility(), age, client, just, truncate(c.Subject, 90))
	}
	if len(rows) > limit {
		fmt.Fprintf(w, "  … e mais %d.\n", len(rows)-limit)
	}
	pretty, _ := json.MarshalIndent(body, "  ", "  ")
	fmt.Fprintf(w, "\nPatch:\n  %s\n", pretty)
	eta := estimateETA(len(rows))
	if eta > 0 {
		fmt.Fprintf(w, "\nETA (10 req/min): ~%s\n", eta.Round(time.Second))
	}
}

// injectActionsCreatedBy sets {"createdBy":{"id": userID}} on each entry of
// body["actions"] that doesn't already carry one. Matches the convention used
// by `tickets actions add`. Top-level ticket fields are left untouched.
func injectActionsCreatedBy(body map[string]any, userID string) {
	if userID == "" || body == nil {
		return
	}
	raw, ok := body["actions"]
	if !ok {
		return
	}
	arr, ok := raw.([]map[string]any)
	if ok {
		for _, a := range arr {
			if _, present := a["createdBy"]; !present {
				a["createdBy"] = map[string]any{"id": userID}
			}
		}
		return
	}
	// fallback for []any (e.g., bodies loaded from --file/--from-template)
	if anyArr, ok := raw.([]any); ok {
		for _, item := range anyArr {
			a, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if _, present := a["createdBy"]; !present {
				a["createdBy"] = map[string]any{"id": userID}
			}
		}
	}
}

func estimateETA(n int) time.Duration {
	if n <= 0 {
		return 0
	}
	// First request is immediate; remaining wait ~6s each at 10 req/min.
	return time.Duration(n-1) * 6 * time.Second
}

func isStdinTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func openReport(path string) (*bufio.Writer, func(), error) {
	if path == "" {
		return nil, func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, func() {}, fmt.Errorf("criar --report: %w", err)
	}
	bw := bufio.NewWriter(f)
	closer := func() {
		_ = bw.Flush()
		_ = f.Close()
	}
	return bw, closer, nil
}

func writeReport(w *bufio.Writer, id int, ok bool, errMsg string) {
	if w == nil {
		return
	}
	line, _ := json.Marshal(struct {
		ID    int    `json:"id"`
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
		At    string `json:"at"`
	}{ID: id, OK: ok, Error: errMsg, At: time.Now().UTC().Format(time.RFC3339)})
	_, _ = w.Write(line)
	_, _ = w.WriteString("\n")
}

