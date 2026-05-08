package tickets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

func newSvcWith(t *testing.T, h http.HandlerFunc) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c := movidesk.New(srv.URL, "tok")
	c.Limiter = movidesk.NewLimiter(1000, 0)
	return New(c), srv
}

func TestListActions(t *testing.T) {
	var seen url.Values
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query()
		w.Write([]byte(`{"id":1,"actions":[{"id":1,"type":1,"description":"hello"},{"id":2,"type":2,"description":"world"}]}`))
	})
	defer srv.Close()

	got, err := svc.ListActions(context.Background(), 1)
	require.NoError(t, err)
	assert.Contains(t, seen.Get("$expand"), "actions")
	require.Len(t, got, 2)
	assert.Equal(t, "world", got[1].Description)
}

func TestGetAction_FoundAndMissing(t *testing.T) {
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"actions":[{"id":3,"description":"x"},{"id":4,"description":"y"}]}`))
	})
	defer srv.Close()

	a, err := svc.GetAction(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.Equal(t, "y", a.Description)

	_, err = svc.GetAction(context.Background(), 1, 999)
	require.Error(t, err)
}

func TestRelations(t *testing.T) {
	var seen url.Values
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query()
		w.Write([]byte(`{"parentTickets":[{"id":10}],"childrenTickets":[{"id":20},{"id":21}]}`))
	})
	defer srv.Close()

	parents, children, err := svc.Relations(context.Background(), 1)
	require.NoError(t, err)
	expand := seen.Get("$expand")
	assert.Contains(t, expand, "parentTickets")
	assert.Contains(t, expand, "childrenTickets")
	require.Len(t, parents, 1)
	assert.Equal(t, 10, parents[0].ID)
	require.Len(t, children, 2)
}

func TestTimeline_OrdersChronologically(t *testing.T) {
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"actions":[{"id":1,"type":2,"description":"reply","createdDate":"2026-04-01T13:30:00Z","createdBy":{"businessName":"Joe"}}],
			"statusHistories":[{"status":"Novo","changedDate":"2026-04-01T13:00:00Z"},{"status":"Em atendimento","changedDate":"2026-04-01T13:05:00Z"}],
			"ownerHistories":[{"ownerTeam":"Tier 1","owner":{"businessName":"Mike"},"changedDate":"2026-04-01T13:10:00Z"}]
		}`))
	})
	defer srv.Close()

	got, err := svc.Timeline(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, got, 4)
	assert.Equal(t, "status", got[0].Kind)
	assert.Equal(t, "Novo", got[0].Summary)
	assert.Equal(t, "Em atendimento", got[1].Summary)
	assert.Equal(t, "owner", got[2].Kind)
	assert.Contains(t, got[2].Summary, "Mike")
	assert.True(t, strings.HasPrefix(got[3].Kind, "action:"))
}

func TestAddAction_PostsActionsArray(t *testing.T) {
	var captured string
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		captured = string(buf)
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := svc.AddAction(context.Background(), 1, Action{Type: 1, Description: "internal note"})
	require.NoError(t, err)
	assert.Contains(t, captured, `"actions"`)
	assert.Contains(t, captured, `"type":1`)
	assert.Contains(t, captured, `"description":"internal note"`)
}

func TestAddAction_RejectsExplicitID(t *testing.T) {
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()
	_, err := svc.AddAction(context.Background(), 1, Action{ID: 5, Type: 1, Description: "x"})
	require.Error(t, err)
}

func TestUpdateAction(t *testing.T) {
	var captured string
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		captured = string(buf)
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := svc.UpdateAction(context.Background(), 1, Action{ID: 7, Description: "edited"})
	require.NoError(t, err)
	assert.Contains(t, captured, `"id":7`)
	assert.Contains(t, captured, `"description":"edited"`)
}

func TestUpdateAction_RequiresID(t *testing.T) {
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()
	_, err := svc.UpdateAction(context.Background(), 1, Action{Description: "x"})
	require.Error(t, err)
}

func TestDeleteAction(t *testing.T) {
	var captured string
	svc, srv := newSvcWith(t, func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		captured = string(buf)
		w.WriteHeader(200)
	})
	defer srv.Close()

	_, err := svc.DeleteAction(context.Background(), 1, 9)
	require.NoError(t, err)
	assert.Contains(t, captured, `"id":9`)
	assert.Contains(t, captured, `"isDeleted":true`)
}
