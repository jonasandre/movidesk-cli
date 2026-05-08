package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectCreatedBy_AddsWhenAbsent(t *testing.T) {
	body := map[string]any{"subject": "x"}
	injectCreatedBy(body, "u-bot")
	got, ok := body["createdBy"].(map[string]any)
	if assert.True(t, ok) {
		assert.Equal(t, "u-bot", got["id"])
	}
}

func TestInjectCreatedBy_PreservesExplicit(t *testing.T) {
	body := map[string]any{
		"subject":   "x",
		"createdBy": map[string]any{"id": "explicit"},
	}
	injectCreatedBy(body, "u-bot")
	got := body["createdBy"].(map[string]any)
	assert.Equal(t, "explicit", got["id"])
}

func TestInjectCreatedBy_NoOpWhenUserEmpty(t *testing.T) {
	body := map[string]any{"subject": "x"}
	injectCreatedBy(body, "")
	_, ok := body["createdBy"]
	assert.False(t, ok)
}

func TestInjectCreatedBy_NoOpWhenBodyNil(t *testing.T) {
	injectCreatedBy(nil, "u-bot")
	// no panic, no return value to check
}
