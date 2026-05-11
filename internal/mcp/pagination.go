package mcp

import (
	"encoding/json"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// defaultMaxRows caps `all=true` pagination when the caller did not supply
	// a `max` value. Without a cap, an LLM driving a tenant with tens of
	// thousands of rows can easily exhaust the 10 req/min budget mid-session.
	defaultMaxRows = 500

	// maxResultBytes caps the size of a single tool response. Anything beyond
	// this is truncated with a marker so the LLM can narrow $select / $top
	// instead of silently swallowing partial data.
	maxResultBytes = 256 * 1024
)

// applyDefaultMax returns max unchanged when an explicit value is set, or the
// safe default when all=true and max=0. With all=false the value is irrelevant
// (the service returns a single page).
func applyDefaultMax(all bool, max int) int {
	if all && max == 0 {
		return defaultMaxRows
	}
	return max
}

// rawResult wraps a raw JSON body from Movidesk in a CallToolResult with size
// capping. The body is returned as a single TextContent block so models can
// parse it with their normal JSON heuristics.
func rawResult(raw []byte) *mcpsdk.CallToolResult {
	body, truncated := capBytes(raw)
	content := []mcpsdk.Content{&mcpsdk.TextContent{Text: string(body)}}
	if truncated {
		content = append(content, &mcpsdk.TextContent{
			Text: fmt.Sprintf("\n[truncado: resposta excedeu %d bytes; refine $select ou $top]", maxResultBytes),
		})
	}
	return &mcpsdk.CallToolResult{Content: content}
}

// rowsResult marshals a slice of json.RawMessage (returned by every Paginate
// method) into a single JSON array, then routes through rawResult for capping.
func rowsResult(rows []json.RawMessage) *mcpsdk.CallToolResult {
	buf, err := json.Marshal(rows)
	if err != nil {
		return &mcpsdk.CallToolResult{
			IsError: true,
			Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: "falha ao serializar resultado: " + err.Error()}},
		}
	}
	return rawResult(buf)
}

func capBytes(raw []byte) ([]byte, bool) {
	if len(raw) <= maxResultBytes {
		return raw, false
	}
	return raw[:maxResultBytes], true
}
