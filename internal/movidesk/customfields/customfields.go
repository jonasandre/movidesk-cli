// Package customfields covers Movidesk's /ticketCustomFieldValue option-pool
// endpoints. These manage the pool of allowed values on list-type custom
// fields — they do NOT manage the value of a custom field on a given ticket
// or person (use the tickets / persons packages for that).
package customfields

import (
	"context"
	"errors"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

const (
	pathInsert = "/ticketCustomFieldValue/InsertValues"
	pathUpdate = "/ticketCustomFieldValue/UpdateValues"
	pathDelete = "/ticketCustomFieldValue/DeleteValues"
)

// InsertBody is the body for InsertValues / DeleteValues — a flat list of names.
type InsertBody struct {
	CustomFieldID    string   `json:"customfieldid"`
	CustomFieldValues []string `json:"customfieldvalues"`
}

// UpdateBody is the body for UpdateValues — pairs of {oldname, newname}.
type UpdateBody struct {
	CustomFieldID    string         `json:"customfieldid"`
	CustomFieldValues []UpdatePair  `json:"customfieldvalues"`
}

// UpdatePair maps an old option name to its new name.
type UpdatePair struct {
	OldName string `json:"oldname"`
	NewName string `json:"newname"`
}

// API binds the option-pool endpoints to a Movidesk client.
type API struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *API { return &API{C: c} }

// AddOptions inserts new option names into a list-type custom field's pool.
func (a *API) AddOptions(ctx context.Context, customFieldID string, names []string) ([]byte, error) {
	if customFieldID == "" {
		return nil, errors.New("customFieldID is required")
	}
	if len(names) == 0 {
		return nil, errors.New("at least one option name is required")
	}
	return a.C.Post(ctx, pathInsert, nil, InsertBody{CustomFieldID: customFieldID, CustomFieldValues: names})
}

// RenameOptions renames existing option names. Each pair: {oldname, newname}.
func (a *API) RenameOptions(ctx context.Context, customFieldID string, pairs []UpdatePair) ([]byte, error) {
	if customFieldID == "" {
		return nil, errors.New("customFieldID is required")
	}
	if len(pairs) == 0 {
		return nil, errors.New("at least one pair is required")
	}
	return a.C.Post(ctx, pathUpdate, nil, UpdateBody{CustomFieldID: customFieldID, CustomFieldValues: pairs})
}

// RemoveOptions deletes option names from a list-type custom field's pool.
func (a *API) RemoveOptions(ctx context.Context, customFieldID string, names []string) ([]byte, error) {
	if customFieldID == "" {
		return nil, errors.New("customFieldID is required")
	}
	if len(names) == 0 {
		return nil, errors.New("at least one option name is required")
	}
	return a.C.Post(ctx, pathDelete, nil, InsertBody{CustomFieldID: customFieldID, CustomFieldValues: names})
}
