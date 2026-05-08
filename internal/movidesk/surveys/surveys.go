// Package surveys covers Movidesk's /survey/questions and /survey/responses
// endpoints. Both are read-only for API consumers.
package surveys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

const (
	pathQuestions = "/survey/questions"
	pathResponses = "/survey/responses"
)

// QuestionLanguage is question.languages[n].
type QuestionLanguage struct {
	CultureID   string `json:"cultureId,omitempty"`
	Description string `json:"description,omitempty"`
}

// Question is /survey/questions row.
//
// Type: 1=Satisfeito/Insatisfeito, 2=Faces Sorridentes, 3=NPS, 4=Sim/Não.
type Question struct {
	ID        string             `json:"id,omitempty"`
	Languages []QuestionLanguage `json:"languages,omitempty"`
	IsActive  bool               `json:"isActive"`
	Type      int                `json:"type,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Question and stores raw bytes in Extra.
func (q *Question) UnmarshalJSON(data []byte) error {
	type alias Question
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*q = Question(x)
	q.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// Response is /survey/responses row.
//
// Type matches Question.Type. Value semantics depend on Type.
type Response struct {
	ID           string `json:"id,omitempty"`
	QuestionID   string `json:"questionId,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	Type         int    `json:"type,omitempty"`
	TicketID     int    `json:"ticketId,omitempty"`
	ResponseDate string `json:"responseDate,omitempty"`
	Commentary   string `json:"commentary,omitempty"`
	Value        int    `json:"value,omitempty"`

	Extra json.RawMessage `json:"-"`
}

// UnmarshalJSON populates Response and stores raw bytes in Extra.
func (r *Response) UnmarshalJSON(data []byte) error {
	type alias Response
	var x alias
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	*r = Response(x)
	r.Extra = append(json.RawMessage(nil), data...)
	return nil
}

// ResponsePage is the cursor-paginated /survey/responses payload.
type ResponsePage struct {
	HasMore bool              `json:"hasMore"`
	Items   []json.RawMessage `json:"items"`
}

// Service binds /survey/* endpoints to a Movidesk client.
type Service struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *Service { return &Service{C: c} }

// ListQuestions returns the full set of survey questions.
//
// Optional typeFilter narrows the result to a specific question type.
func (s *Service) ListQuestions(ctx context.Context, typeFilter int) ([]byte, error) {
	v := url.Values{}
	if typeFilter > 0 {
		v.Set("type", strconv.Itoa(typeFilter))
	}
	return s.C.Do(ctx, "GET", pathQuestions, v, nil)
}

// GetQuestion fetches a single question by id.
func (s *Service) GetQuestion(ctx context.Context, id string) ([]byte, error) {
	return s.C.Do(ctx, "GET", pathQuestions+"/"+url.PathEscape(id), nil, nil)
}

// ListResponsesPage queries one cursor-paginated page of responses.
func (s *Service) ListResponsesPage(ctx context.Context, limit int, startingAfter string) (*ResponsePage, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	v := url.Values{}
	v.Set("limit", strconv.Itoa(limit))
	if startingAfter != "" {
		v.Set("startingAfter", startingAfter)
	}
	body, err := s.C.Do(ctx, "GET", pathResponses, v, nil)
	if err != nil {
		return nil, err
	}
	var p ResponsePage
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("decode response page: %w", err)
	}
	return &p, nil
}

// ListAllResponses walks every page. max <= 0 means no upper bound.
func (s *Service) ListAllResponses(ctx context.Context, max int) ([]json.RawMessage, error) {
	var (
		out    []json.RawMessage
		cursor string
	)
	for {
		page, err := s.ListResponsesPage(ctx, 100, cursor)
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
		var last struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(page.Items[len(page.Items)-1], &last); err != nil {
			return nil, err
		}
		cursor = last.ID
	}
}
