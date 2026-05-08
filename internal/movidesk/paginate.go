package movidesk

import (
	"context"
	"encoding/json"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

// PageFetcher fetches one page given a base query (with $skip set by Paginate).
// It must return the raw JSON body for the page so Paginate can detect EOF.
type PageFetcher func(ctx context.Context, q odata.Query) ([]byte, error)

// Paginate keeps issuing pages until an empty array is returned or the maximum
// is reached. Pages are decoded into []json.RawMessage and concatenated.
//
// max <= 0 means no upper bound. pageSize <= 0 defaults to 100.
func Paginate(ctx context.Context, base odata.Query, fetch PageFetcher, pageSize, max int) ([]json.RawMessage, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	out := []json.RawMessage{}
	q := base
	q.Top = pageSize
	q.Skip = base.Skip

	for {
		body, err := fetch(ctx, q)
		if err != nil {
			return nil, err
		}
		var page []json.RawMessage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		if len(page) == 0 {
			return out, nil
		}
		out = append(out, page...)
		if max > 0 && len(out) >= max {
			return out[:max], nil
		}
		if len(page) < pageSize {
			return out, nil
		}
		q.Skip += pageSize
	}
}
