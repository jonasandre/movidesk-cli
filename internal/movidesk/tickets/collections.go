package tickets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

// fetchExpanded retrieves a ticket with the requested $expand entries and
// decodes it. Used by every collection helper so we keep one place that
// understands the read-time semantics of /tickets.
func (s *Service) fetchExpanded(ctx context.Context, ticketID int, expand []string) (*Ticket, []byte, error) {
	q := odata.Query{Expand: expand}
	v := url.Values{}
	q.Apply(v)
	v.Set("id", strconv.Itoa(ticketID))
	body, err := s.C.Do(ctx, "GET", pathTickets, v, nil)
	if err != nil {
		return nil, nil, err
	}
	var tk Ticket
	if err := json.Unmarshal(body, &tk); err != nil {
		return nil, nil, fmt.Errorf("decodificar chamado: %w", err)
	}
	return &tk, body, nil
}

// ListActions returns the actions[] of a ticket.
func (s *Service) ListActions(ctx context.Context, ticketID int) ([]Action, error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"actions"})
	if err != nil {
		return nil, err
	}
	return tk.Actions, nil
}

// GetAction returns a single action by id.
func (s *Service) GetAction(ctx context.Context, ticketID, actionID int) (*Action, error) {
	actions, err := s.ListActions(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	for i := range actions {
		if actions[i].ID == actionID {
			return &actions[i], nil
		}
	}
	return nil, fmt.Errorf("ação %d não encontrada no chamado %d", actionID, ticketID)
}

// ListClients returns the clients[] of a ticket.
func (s *Service) ListClients(ctx context.Context, ticketID int) ([]Client, error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"clients"})
	if err != nil {
		return nil, err
	}
	return tk.Clients, nil
}

// Relations returns parent and child tickets.
func (s *Service) Relations(ctx context.Context, ticketID int) (parents, children []ParentChild, err error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"parentTickets", "childrenTickets"})
	if err != nil {
		return nil, nil, err
	}
	return tk.ParentTickets, tk.ChildrenTickets, nil
}

// ListAssets returns the assets[] of a ticket.
func (s *Service) ListAssets(ctx context.Context, ticketID int) ([]Asset, error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"assets"})
	if err != nil {
		return nil, err
	}
	return tk.Assets, nil
}

// Histories returns owner + status histories.
func (s *Service) Histories(ctx context.Context, ticketID int) (owners []OwnerHistory, statuses []StatusHistory, err error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"ownerHistories", "statusHistories"})
	if err != nil {
		return nil, nil, err
	}
	return tk.OwnerHistories, tk.StatusHistories, nil
}

// TimelineEntry is a unified row across actions, status histories and owner
// histories, sorted chronologically. Useful for `tickets timeline`.
type TimelineEntry struct {
	When    string `json:"when"`
	Kind    string `json:"kind"`
	Actor   string `json:"actor,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// Timeline composes a chronological view from actions + status + owner
// histories on a single ticket.
func (s *Service) Timeline(ctx context.Context, ticketID int) ([]TimelineEntry, error) {
	tk, _, err := s.fetchExpanded(ctx, ticketID, []string{"actions", "statusHistories", "ownerHistories"})
	if err != nil {
		return nil, err
	}
	out := make([]TimelineEntry, 0, len(tk.Actions)+len(tk.StatusHistories)+len(tk.OwnerHistories))
	for _, a := range tk.Actions {
		actor := ""
		if a.CreatedBy != nil {
			actor = a.CreatedBy.BusinessName
		}
		out = append(out, TimelineEntry{
			When:    a.CreatedDate,
			Kind:    fmt.Sprintf("action:%s", actionTypeName(a.Type)),
			Actor:   actor,
			Summary: truncate(a.Description, 80),
		})
	}
	for _, h := range tk.StatusHistories {
		actor := ""
		if h.ChangedBy != nil {
			actor = h.ChangedBy.BusinessName
		}
		out = append(out, TimelineEntry{
			When:    h.ChangedDate,
			Kind:    "status",
			Actor:   actor,
			Summary: h.Status,
		})
	}
	for _, h := range tk.OwnerHistories {
		actor := ""
		if h.ChangedBy != nil {
			actor = h.ChangedBy.BusinessName
		}
		owner := ""
		if h.Owner != nil {
			owner = h.Owner.BusinessName
		}
		summary := owner
		if h.OwnerTeam != "" {
			summary = fmt.Sprintf("%s (%s)", owner, h.OwnerTeam)
		}
		out = append(out, TimelineEntry{
			When:    h.ChangedDate,
			Kind:    "owner",
			Actor:   actor,
			Summary: summary,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].When < out[j].When })
	return out, nil
}

func actionTypeName(t int) string {
	switch t {
	case 1:
		return "internal"
	case 2:
		return "public"
	default:
		return "?"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ----- Action writes -----

// AddAction appends a new action to a ticket via PATCH. Movidesk uses
// id-less actions in the array to mean "create new".
func (s *Service) AddAction(ctx context.Context, ticketID int, a Action) ([]byte, error) {
	if a.ID != 0 {
		return nil, errors.New("AddAction: leave Action.ID at zero; use UpdateAction to edit")
	}
	body := map[string]any{"actions": []Action{a}}
	return s.Update(ctx, ticketID, body)
}

// UpdateAction patches an existing action by id.
func (s *Service) UpdateAction(ctx context.Context, ticketID int, a Action) ([]byte, error) {
	if a.ID == 0 {
		return nil, errors.New("UpdateAction: Action.ID is required")
	}
	body := map[string]any{"actions": []Action{a}}
	return s.Update(ctx, ticketID, body)
}

// DeleteAction soft-deletes an action by id (Movidesk uses isDeleted: true).
func (s *Service) DeleteAction(ctx context.Context, ticketID, actionID int) ([]byte, error) {
	if actionID == 0 {
		return nil, errors.New("DeleteAction: actionID is required")
	}
	body := map[string]any{
		"actions": []map[string]any{
			{"id": actionID, "isDeleted": true},
		},
	}
	return s.Update(ctx, ticketID, body)
}
