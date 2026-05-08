package tickets

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fullTicketJSON = `{
  "id": 42,
  "protocol": "MOVI202109000042",
  "type": 2,
  "subject": "Outage at the warehouse",
  "category": "Suporte",
  "urgency": "Alta",
  "status": "Em atendimento",
  "baseStatus": "InAttendance",
  "justification": null,
  "origin": 9,
  "createdDate": "2026-04-01T13:00:00.000Z",
  "originEmailAccount": "support@acme.com",
  "owner": {"id": "u-1", "businessName": "Joe", "personType": 1, "profileType": 1},
  "ownerTeam": "Tier 2",
  "createdBy": {"id": "u-2", "businessName": "Jane", "personType": 1, "profileType": 2},
  "serviceFull": ["Suporte", "Infra", "Rede"],
  "serviceFirstLevelId": 100,
  "serviceFirstLevel": "Suporte",
  "serviceSecondLevel": "Infra",
  "serviceThirdLevel": "Rede",
  "tags": ["p1", "warehouse"],
  "cc": "ops@acme.com",
  "resolvedIn": null,
  "closedIn": null,
  "lastActionDate": "2026-04-02T09:30:00.000Z",
  "lastUpdate": "2026-04-02T09:30:00.000Z",
  "actionCount": 3,
  "lifetimeWorkingTime": 480,
  "stoppedTime": 60,
  "stoppedTimeWorkingTime": 30,
  "resolvedInFirstCall": false,
  "chatWidget": null,
  "sequence": 12,
  "slaAgreement": "Default",
  "slaAgreementRule": "Alta + Suporte",
  "slaSolutionTime": 480,
  "slaResponseTime": 60,
  "slaSolutionChangedByUser": false,
  "slaSolutionDate": "2026-04-03T13:00:00.000Z",
  "slaSolutionDateIsPaused": false,
  "slaResponseDate": "2026-04-01T14:00:00.000Z",
  "slaRealResponseDate": "2026-04-01T13:30:00.000Z",
  "jiraIssueKey": "OPS-123",
  "redmineIssueId": 0,
  "clients": [
    {"id": "c-1", "businessName": "Acme", "email": "ops@acme.com", "personType": 2, "profileType": 2}
  ],
  "actions": [
    {
      "id": 1,
      "type": 2,
      "origin": 9,
      "description": "Initial report",
      "createdDate": "2026-04-01T13:00:00.000Z",
      "createdBy": {"id": "u-2", "businessName": "Jane", "personType": 1, "profileType": 2},
      "tags": ["intake"],
      "timeAppointments": [
        {"id": 11, "activity": "Triage", "workTime": "00:30", "accountable": {"id": "u-1", "businessName": "Joe"}}
      ],
      "expenses": [
        {"id": 21, "type": "Reembolso", "value": "120,50"}
      ],
      "attachments": [
        {"fileName": "report.pdf", "createdDate": "2026-04-01T13:01:00.000Z"}
      ]
    }
  ],
  "parentTickets": [{"id": 10, "subject": "Master incident"}],
  "childrenTickets": [{"id": 50, "subject": "Sub-issue"}, {"id": 51, "subject": "Sub-issue 2"}],
  "ownerHistories": [
    {"ownerTeam": "Tier 1", "owner": {"id": "u-3", "businessName": "Mike"}, "changedDate": "2026-04-01T13:05:00.000Z", "permanencyTime": 60}
  ],
  "statusHistories": [
    {"status": "Novo", "changedDate": "2026-04-01T13:00:00.000Z", "permanencyTime": 5},
    {"status": "Em atendimento", "changedDate": "2026-04-01T13:05:00.000Z", "permanencyTime": 100}
  ],
  "customFieldValues": [
    {"customFieldId": 125529, "customFieldRuleId": 1, "line": 1, "value": "high"},
    {"customFieldId": 125530, "customFieldRuleId": 1, "line": 1, "items": [{"customFieldItem": "Alta"}]}
  ],
  "assets": [
    {"id": "asset-1", "name": "Switch core", "label": "SW-01"}
  ]
}`

func TestTicket_FullSchema_Unmarshal(t *testing.T) {
	var tk Ticket
	require.NoError(t, json.Unmarshal([]byte(fullTicketJSON), &tk))

	assert.Equal(t, 42, tk.ID)
	assert.Equal(t, "MOVI202109000042", tk.Protocol)
	assert.Equal(t, "Em atendimento", tk.Status)
	assert.Equal(t, []string{"Suporte", "Infra", "Rede"}, tk.ServiceFull)
	assert.Equal(t, 100, tk.ServiceFirstLevelID)
	assert.Equal(t, "Tier 2", tk.OwnerTeam)
	assert.Equal(t, "ops@acme.com", tk.CC)
	assert.Equal(t, 12, tk.Sequence)

	// SLA cluster
	assert.Equal(t, "Default", tk.SLAAgreement)
	assert.Equal(t, 480, tk.SLASolutionTime)
	assert.Equal(t, "2026-04-03T13:00:00.000Z", tk.SLASolutionDate)

	// Working time + chat
	assert.Equal(t, 480, tk.LifetimeWorkingTime)
	assert.Equal(t, 60, tk.StoppedTime)
	assert.False(t, tk.ResolvedInFirstCall)

	// Integrations
	assert.Equal(t, "OPS-123", tk.JiraIssueKey)

	// Collections
	require.Len(t, tk.Clients, 1)
	assert.Equal(t, "Acme", tk.Clients[0].BusinessName)
	require.Len(t, tk.Actions, 1)
	assert.Equal(t, "Initial report", tk.Actions[0].Description)
	require.Len(t, tk.Actions[0].TimeAppointments, 1)
	assert.Equal(t, "Triage", tk.Actions[0].TimeAppointments[0].Activity)
	require.Len(t, tk.Actions[0].Expenses, 1)
	assert.Equal(t, "120,50", tk.Actions[0].Expenses[0].Value)
	require.Len(t, tk.Actions[0].Attachments, 1)
	assert.Equal(t, "report.pdf", tk.Actions[0].Attachments[0].FileName)
	require.Len(t, tk.ParentTickets, 1)
	assert.Equal(t, 10, tk.ParentTickets[0].ID)
	require.Len(t, tk.ChildrenTickets, 2)
	require.Len(t, tk.OwnerHistories, 1)
	require.Len(t, tk.StatusHistories, 2)
	require.Len(t, tk.CustomFieldValues, 2)
	assert.Equal(t, 125529, tk.CustomFieldValues[0].CustomFieldID)
	assert.Equal(t, "high", tk.CustomFieldValues[0].Value)
	require.Len(t, tk.CustomFieldValues[1].Items, 1)
	assert.Equal(t, "Alta", tk.CustomFieldValues[1].Items[0].CustomFieldItem)
	require.Len(t, tk.Assets, 1)
	assert.Equal(t, "SW-01", tk.Assets[0].Label)

	// Extra preserves raw bytes
	assert.NotEmpty(t, tk.Extra)
	assert.Contains(t, string(tk.Extra), "OPS-123")
}

func TestTicket_RoundTrip_PreservesUnknownFields(t *testing.T) {
	const withFutureField = `{"id":1,"subject":"x","unknownFutureField":{"foo":"bar"}}`
	var tk Ticket
	require.NoError(t, json.Unmarshal([]byte(withFutureField), &tk))
	// Extra carries the raw payload, including unknownFutureField.
	assert.Contains(t, string(tk.Extra), `"unknownFutureField"`)
}

func TestAction_UnmarshalCapturesExtra(t *testing.T) {
	const a = `{"id":7,"type":2,"description":"hi","customExt":42}`
	var act Action
	require.NoError(t, json.Unmarshal([]byte(a), &act))
	assert.Equal(t, 7, act.ID)
	assert.Equal(t, "hi", act.Description)
	assert.Contains(t, string(act.Extra), `"customExt":42`)
}

func TestCustomFieldValue_UnmarshalCapturesExtra(t *testing.T) {
	const c = `{"customFieldId":1,"customFieldRuleId":2,"line":1,"value":"v","futureFlag":true}`
	var cv CustomFieldValue
	require.NoError(t, json.Unmarshal([]byte(c), &cv))
	assert.Equal(t, 1, cv.CustomFieldID)
	assert.Equal(t, 2, cv.CustomFieldRuleID)
	assert.Equal(t, "v", cv.Value)
	assert.Contains(t, string(cv.Extra), `"futureFlag":true`)
}
