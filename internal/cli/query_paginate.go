package cli

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

// paginateGeneric is the engine behind `query --all`. It uses the OData
// builder for $top/$skip and forwards extra params. Each page is decoded into
// json.RawMessage; auto-pagination stops on an empty page or short page.
func paginateGeneric(ctx context.Context, c *movidesk.Client, path string, base odata.Query, max int, extraParams []string) ([]json.RawMessage, error) {
	fetch := func(ctx context.Context, q odata.Query) ([]byte, error) {
		v := url.Values{}
		q.Apply(v)
		for _, kv := range extraParams {
			k, val, ok := strings.Cut(kv, "=")
			if !ok {
				continue
			}
			v.Set(k, val)
		}
		return c.Do(ctx, "GET", path, v, nil)
	}
	return movidesk.Paginate(ctx, base, fetch, base.Top, max)
}
