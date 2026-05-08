package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----- Activities -----

func TestE2E_Activities_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/activity", r.URL.Path)
		assert.Equal(t, "needle", r.URL.Query().Get("name"))
		w.Write([]byte(`{"hasMore":false,"items":[{"id":1,"name":"x","isActive":true,"isAllowsAllTeams":true,"teams":[]}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "activities", "list", "--name", "needle", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":1`)
}

func TestE2E_Activities_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("id"))
		w.Write([]byte(`{"id":5,"name":"x","isActive":true,"isAllowsAllTeams":true}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "activities", "get", "5", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":5`)
}

func TestE2E_Activities_Create(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		buf, _ := io.ReadAll(r.Body)
		got = string(buf)
		w.Write([]byte(`12`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "activities", "create", "--set", "name=Triage", "--set", "isActive=true", "--set", "isAllowsAllTeams=true")
	require.NoError(t, err)
	assert.Contains(t, got, `"name":"Triage"`)
}

func TestE2E_Activities_AddTeams(t *testing.T) {
	var seen, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.Write([]byte(`["Suporte"]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "activities", "add-teams", "1", "--team", "Suporte")
	require.NoError(t, err)
	assert.Contains(t, seen, "activityId=1")
	assert.Contains(t, body, "Suporte")
}

func TestE2E_Activities_Delete_Force(t *testing.T) {
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "activities", "delete", "1", "--force")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", method)
}

// ----- Contracts -----

func TestE2E_Contracts_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/timeAgreement", r.URL.Path)
		w.Write([]byte(`[{"id":1,"name":"Default","isActive":true}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "contracts", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "Default")
}

func TestE2E_Contracts_Consumption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/timeAgreementConsumption", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "%24filter=name+eq")
		w.Write([]byte(`[{"id":"abc","name":"Default","period":"2026-04-01T00:00:00Z","consumedHours":10}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "contracts", "consumption", "list", "--filter", "name eq 'Default'", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "consumedHours")
}

// ----- Surveys -----

func TestE2E_Surveys_QuestionsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/survey/questions", r.URL.Path)
		w.Write([]byte(`[{"id":"a","isActive":true,"type":3}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "surveys", "questions", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":"a"`)
}

func TestE2E_Surveys_QuestionsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/survey/questions/QWMv", r.URL.Path)
		w.Write([]byte(`{"id":"QWMv","isActive":true,"type":3}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "surveys", "questions", "get", "QWMv", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "QWMv")
}

func TestE2E_Surveys_ResponsesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/survey/responses", r.URL.Path)
		w.Write([]byte(`{"hasMore":false,"items":[{"id":"x","value":1}]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "surveys", "responses", "list", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":"x"`)
}

// ----- KB -----

func TestE2E_KB_ArticlesGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/article/19040", r.URL.Path)
		w.Write([]byte(`{"id":19040,"title":"My article","articleStatus":1}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "kb", "articles", "get", "19040", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, "My article")
}

// ----- Telephony -----

func TestE2E_Telephony_Queue(t *testing.T) {
	var seen, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"telephony", "queue",
		"--event", "receivedCall",
		"--set", "id=abc",
		"--set", "queueId=1",
		"--set", "clientNumber=555",
		"--set", "callDate=2026-04-01T13:00",
	)
	require.NoError(t, err)
	assert.Equal(t, "/asterisk_receivedCall", seen)
	assert.Contains(t, body, `"queueId":1`)
}

func TestE2E_Telephony_NonQueue(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.RawQuery
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t,
		"telephony", "nonqueue",
		"--event", "startTransferedCall",
		"--param", "id=abc",
		"--param", "branchLine=1001",
	)
	require.NoError(t, err)
	assert.Contains(t, seen, "id=abc")
	assert.Contains(t, seen, "branchLine=1001")
}

func TestE2E_Telephony_MadeCallLink(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		w.WriteHeader(200)
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "telephony", "made-call-link", "--set", "id=abc", "--set", "link=https://x")
	require.NoError(t, err)
	assert.Equal(t, "/setMadeCallLink", seen)
}

// ----- Custom fields option pool -----

func TestE2E_CustomFields_OptionsAdd(t *testing.T) {
	var seen, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "customfields", "options", "add", "--field", "125529", "--value", "A", "--value", "B")
	require.NoError(t, err)
	assert.Equal(t, "/ticketCustomFieldValue/InsertValues", seen)
	assert.Contains(t, body, `"customfieldid":"125529"`)
	assert.Contains(t, body, `"customfieldvalues":["A","B"]`)
}

func TestE2E_CustomFields_OptionsRename(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		body = string(buf)
		w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "customfields", "options", "rename", "--field", "125529", "--pair", "OLD=NEW")
	require.NoError(t, err)
	assert.Contains(t, body, `"oldname":"OLD"`)
	assert.Contains(t, body, `"newname":"NEW"`)
}

func TestE2E_CustomFields_OptionsRemove(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Path
		w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "customfields", "options", "remove", "--field", "125529", "--value", "X")
	require.NoError(t, err)
	assert.Equal(t, "/ticketCustomFieldValue/DeleteValues", seen)
}

// ----- Query escape hatch -----

func TestE2E_Query_GET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tickets", r.URL.Path)
		assert.Equal(t, "id eq 1", r.URL.Query().Get("$filter"))
		w.Write([]byte(`[{"id":1}]`))
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "query", "tickets", "--filter", "id eq 1", "--output", "json", "--compact")
	require.NoError(t, err)
	assert.Contains(t, out, `"id":1`)
}

func TestE2E_Query_AllPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("$skip") {
		case "", "0":
			w.Write([]byte(`[{"id":1},{"id":2}]`))
		case "2":
			w.Write([]byte(`[{"id":3}]`))
		default:
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	out, _, err := runCmd(t, "query", "/tickets", "--all", "--top", "2", "--output", "json", "--compact")
	require.NoError(t, err)
	low := strings.ToLower(out)
	assert.Contains(t, low, `"id":1`)
	assert.Contains(t, low, `"id":3`)
}

func TestE2E_Query_RejectsWriteMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server should not be hit")
	}))
	defer srv.Close()
	setupTenant(t, srv.URL)

	_, _, err := runCmd(t, "query", "/tickets", "--method", "POST")
	require.Error(t, err)
}
