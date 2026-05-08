package tickets

import (
	"context"
	"errors"
	"fmt"
)

// ListCustomFieldValues returns the customFieldValues[] of a ticket.
func (s *Service) ListCustomFieldValues(ctx context.Context, ticketID int) ([]CustomFieldValue, error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"customFieldValues"})
	if err != nil {
		return nil, err
	}
	return tk.CustomFieldValues, nil
}

// SetCustomFieldValue applies one CustomFieldValue change with read-merge-patch
// semantics. This is required because Movidesk deletes any customFieldValues
// entry not present in the PATCH body. We GET the ticket, mutate, and PATCH
// the merged list back so other entries survive.
//
// "Same key" means same (CustomFieldID, CustomFieldRuleID, Line). When a match
// exists it is replaced; otherwise the new entry is appended.
func (s *Service) SetCustomFieldValue(ctx context.Context, ticketID int, change CustomFieldValue) ([]byte, error) {
	if change.CustomFieldID == 0 {
		return nil, errors.New("CustomFieldID is required")
	}
	if change.CustomFieldRuleID == 0 {
		return nil, errors.New("CustomFieldRuleID is required")
	}
	if change.Line == 0 {
		change.Line = 1
	}
	current, err := s.ListCustomFieldValues(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("read customFieldValues: %w", err)
	}
	merged := mergeCustomFields(current, change)
	body := map[string]any{"customFieldValues": stripExtra(merged)}
	return s.Update(ctx, ticketID, body)
}

// ClearCustomFieldValue removes a (fieldID, ruleID, line) entry by omission.
// If line == 0, every line for that (fieldID, ruleID) is removed.
func (s *Service) ClearCustomFieldValue(ctx context.Context, ticketID, fieldID, ruleID, line int) ([]byte, error) {
	if fieldID == 0 {
		return nil, errors.New("fieldID is required")
	}
	current, err := s.ListCustomFieldValues(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("read customFieldValues: %w", err)
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
			// match → drop
			continue
		}
		kept = append(kept, v)
	}
	body := map[string]any{"customFieldValues": stripExtra(kept)}
	return s.Update(ctx, ticketID, body)
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

// stripExtra removes the Extra field before re-marshaling so PATCH bodies
// don't carry stale raw bytes from a previous decode.
func stripExtra(in []CustomFieldValue) []CustomFieldValue {
	out := make([]CustomFieldValue, len(in))
	for i, v := range in {
		v.Extra = nil
		out[i] = v
	}
	return out
}
