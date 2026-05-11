package mcp

import (
	"errors"
	"fmt"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

// wrapAPIError translates *movidesk.APIError into messages tuned for an LLM
// consumer. Rate-limit errors are the most common failure mode under
// chat-driven usage so they get a dedicated explanation.
//
// The returned error is meant to be handed back to the MCP SDK, which marks
// the tool result with IsError=true and surfaces the message to the model.
func wrapAPIError(err error) error {
	if err == nil {
		return nil
	}
	var ae *movidesk.APIError
	if !errors.As(err, &ae) {
		return err
	}
	switch ae.Status {
	case 429:
		return fmt.Errorf("movidesk rate limit atingido (10 req/min). Aguarde ~1 minuto antes de retentar ou restrinja a consulta com $filter/$top. (path %s)", ae.Path)
	case 401, 403:
		return fmt.Errorf("movidesk negou a chamada (HTTP %d). Verifique o token do tenant: %s", ae.Status, ae.Body)
	case 400:
		return fmt.Errorf("movidesk recusou os parâmetros (HTTP 400). Detalhe: %s", ae.Body)
	default:
		return err
	}
}
