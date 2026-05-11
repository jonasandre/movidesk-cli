package persons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

const path = "/persons"

// Service binds the /persons endpoint to a Movidesk client.
type Service struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *Service { return &Service{C: c} }

// List queries /persons with the supplied OData query.
func (s *Service) List(ctx context.Context, q odata.Query) ([]byte, error) {
	return s.C.Get(ctx, path, q, nil)
}

// Get fetches a single person by id (Cod. Ref.).
func (s *Service) Get(ctx context.Context, id string) ([]byte, error) {
	v := url.Values{}
	v.Set("id", id)
	return s.C.Do(ctx, "GET", path, v, nil)
}

// Create posts a new person. returnAllProperties asks the server to echo back
// the full record (otherwise only the id is returned).
func (s *Service) Create(ctx context.Context, body any, returnAllProperties bool) ([]byte, error) {
	v := url.Values{}
	if returnAllProperties {
		v.Set("returnAllProperties", "true")
	}
	return s.C.Post(ctx, path, v, body)
}

// Update patches a person by id. As with tickets, array fields are replaced;
// use ReadMergePatchCustomField for safe partial customFieldValues edits.
func (s *Service) Update(ctx context.Context, id string, body any) ([]byte, error) {
	v := url.Values{}
	v.Set("id", id)
	return s.C.Patch(ctx, path, v, body)
}

// Delete removes a person by id.
func (s *Service) Delete(ctx context.Context, id string) ([]byte, error) {
	v := url.Values{}
	v.Set("id", id)
	return s.C.Delete(ctx, path, v)
}

// Paginate walks /persons pages with the given base query.
func (s *Service) Paginate(ctx context.Context, q odata.Query, pageSize, max int) ([]json.RawMessage, error) {
	return movidesk.Paginate(ctx, q, func(ctx context.Context, q odata.Query) ([]byte, error) {
		return s.List(ctx, q)
	}, pageSize, max)
}

// ListCustomFieldValues returns the customFieldValues[] of a person.
func (s *Service) ListCustomFieldValues(ctx context.Context, id string) ([]CustomFieldValue, error) {
	v := url.Values{}
	v.Set("id", id)
	q := odata.Query{Expand: []string{"customFieldValues"}}
	q.Apply(v)
	body, err := s.C.Do(ctx, "GET", path, v, nil)
	if err != nil {
		return nil, err
	}
	var p Person
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("decodificar pessoa: %w", err)
	}
	return p.CustomFieldValues, nil
}

// SetCustomFieldValue applies one CustomFieldValue change with read-merge-patch
// semantics. Movidesk's PATCH on persons.customFieldValues mirrors the ticket
// behavior — entries omitted from the body are deleted server-side.
func (s *Service) SetCustomFieldValue(ctx context.Context, id string, change CustomFieldValue) ([]byte, error) {
	if change.CustomFieldID == 0 {
		return nil, errors.New("CustomFieldID é obrigatório")
	}
	if change.CustomFieldRuleID == 0 {
		return nil, errors.New("CustomFieldRuleID é obrigatório")
	}
	if change.Line == 0 {
		change.Line = 1
	}
	current, err := s.ListCustomFieldValues(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ler customFieldValues: %w", err)
	}
	merged := mergeCustomFields(current, change)
	body := map[string]any{"customFieldValues": stripExtra(merged)}
	return s.Update(ctx, id, body)
}

// ClearCustomFieldValue removes (fieldID, ruleID, line) from a person.
// If line == 0, every line for that (fieldID, ruleID) is removed.
func (s *Service) ClearCustomFieldValue(ctx context.Context, id string, fieldID, ruleID, line int) ([]byte, error) {
	if fieldID == 0 {
		return nil, errors.New("fieldID é obrigatório")
	}
	current, err := s.ListCustomFieldValues(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ler customFieldValues: %w", err)
	}
	kept := make([]CustomFieldValue, 0, len(current))
	for _, v := range current {
		if v.CustomFieldID == fieldID {
			if ruleID != 0 && v.CustomFieldRuleID != ruleID {
				kept = append(kept, v)
				continue
			}
			if line != 0 && v.Line != line {
				kept = append(kept, v)
				continue
			}
			continue
		}
		kept = append(kept, v)
	}
	body := map[string]any{"customFieldValues": stripExtra(kept)}
	return s.Update(ctx, id, body)
}

func mergeCustomFields(current []CustomFieldValue, change CustomFieldValue) []CustomFieldValue {
	out := make([]CustomFieldValue, 0, len(current)+1)
	replaced := false
	for _, v := range current {
		if v.CustomFieldID == change.CustomFieldID &&
			v.CustomFieldRuleID == change.CustomFieldRuleID &&
			v.Line == change.Line {
			out = append(out, change)
			replaced = true
			continue
		}
		out = append(out, v)
	}
	if !replaced {
		out = append(out, change)
	}
	return out
}

func stripExtra(in []CustomFieldValue) []CustomFieldValue {
	out := make([]CustomFieldValue, len(in))
	for i, v := range in {
		v.Extra = nil
		out[i] = v
	}
	return out
}
