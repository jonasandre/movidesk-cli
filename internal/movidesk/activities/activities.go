// Package activities covers Movidesk's /activity endpoint (note the singular).
//
// Pagination here is cursor-based (limit/startingAfter), not OData.
package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

const (
	pathActivity     = "/activity"
	pathAddTeams     = "/addTeamsToActivity"
	defaultPageLimit = 100
	maxPageLimit     = 100
)

// Team is an activity.teams[n] entry.
type Team struct {
	Name string `json:"name,omitempty"`
}

// Activity models /activity.
type Activity struct {
	ID               int    `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	IsActive         bool   `json:"isActive"`
	IsAllowsAllTeams bool   `json:"isAllowsAllTeams"`
	Teams            []Team `json:"teams,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Activity and stores raw bytes in Extra.
func (a *Activity) UnmarshalJSON(data []byte) error {
	type alias Activity
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*a = Activity(x)
	a.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// Page is a cursor-paginated response page.
type Page struct {
	HasMore bool              `json:"hasMore"`
	Items   []json.RawMessage `json:"items"`
}

// Service binds /activity to a Movidesk client.
type Service struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *Service { return &Service{C: c} }

// Get fetches a single activity by id.
func (s *Service) Get(ctx context.Context, id int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return s.C.Do(ctx, "GET", pathActivity, v, nil)
}

// ListPage queries one page of activities. limit must be in [1..100]; <=0
// uses the default of 100. nameFilter optionally filters by substring on name.
// startingAfter is the cursor returned via the previous page (last item id).
func (s *Service) ListPage(ctx context.Context, limit int, startingAfter, nameFilter string) (*Page, error) {
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	v := url.Values{}
	v.Set("limit", strconv.Itoa(limit))
	if startingAfter != "" {
		v.Set("startingAfter", startingAfter)
	}
	if nameFilter != "" {
		v.Set("name", nameFilter)
	}
	body, err := s.C.Do(ctx, "GET", pathActivity, v, nil)
	if err != nil {
		return nil, err
	}
	var p Page
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("decode activity page: %w", err)
	}
	return &p, nil
}

// ListAll walks every page of activities. max <= 0 means no upper bound.
func (s *Service) ListAll(ctx context.Context, nameFilter string, max int) ([]json.RawMessage, error) {
	var (
		out      []json.RawMessage
		cursor   string
		pageSize = defaultPageLimit
	)
	for {
		page, err := s.ListPage(ctx, pageSize, cursor, nameFilter)
		if err != nil {
			return nil, err
		}
		out = append(out, page.Items...)
		if max > 0 && len(out) >= max {
			return out[:max], nil
		}
		if !page.HasMore || len(page.Items) == 0 {
			return out, nil
		}
		// Cursor = id of the last item in this page.
		var last struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(page.Items[len(page.Items)-1], &last); err != nil {
			return nil, err
		}
		cursor = strconv.Itoa(last.ID)
	}
}

// Create posts a new activity.
func (s *Service) Create(ctx context.Context, body any) ([]byte, error) {
	return s.C.Post(ctx, pathActivity, nil, body)
}

// Update patches an activity by id.
func (s *Service) Update(ctx context.Context, id int, body any) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return s.C.Patch(ctx, pathActivity, v, body)
}

// Delete removes an activity.
func (s *Service) Delete(ctx context.Context, id int) ([]byte, error) {
	v := url.Values{}
	v.Set("id", strconv.Itoa(id))
	return s.C.Delete(ctx, pathActivity, v)
}

// AddTeams appends teams to an existing activity. Returns the full updated
// teams list per Movidesk docs.
func (s *Service) AddTeams(ctx context.Context, activityID int, teams []string) ([]byte, error) {
	v := url.Values{}
	v.Set("activityId", strconv.Itoa(activityID))
	return s.C.Post(ctx, pathAddTeams, v, teams)
}
