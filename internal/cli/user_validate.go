package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

// validateUser confirms a person id exists in the tenant. Returns the
// businessName so the caller can echo "Default user: u-123 (Joe Doe)".
func validateUser(ctx context.Context, baseURL, token, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("id de usuário vazio")
	}
	c := movidesk.New(baseURL, token)
	c.HTTP.Timeout = 15 * time.Second
	c.Retry.MaxAttempts = 1

	v := url.Values{}
	v.Set("id", userID)
	v.Set("$select", "id,businessName")

	body, err := c.Do(ctx, "GET", "/persons", v, nil)
	if err != nil {
		return "", err
	}
	var p struct {
		ID           string `json:"id"`
		BusinessName string `json:"businessName"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return "", fmt.Errorf("decodificar pessoa: %w", err)
	}
	if p.ID == "" {
		return "", fmt.Errorf("usuário %q não encontrado", userID)
	}
	return p.BusinessName, nil
}
