// Package telephony covers Movidesk's call event hooks.
//
// Two flavors:
//   - With queue control: POST /asterisk_receivedCall, /asterisk_transferedCall,
//     /asterisk_completedCall, /asterisk_lostCall, /asterisk_canceledCall.
//   - Without queue control: GET /asterisk_startTransferedCall,
//     /asterisk_completedCall, /asterisk_startCanceledCall.
//
// Plus /setMadeCallLink (POST) — attach a recording link to an outbound call.
//
// Movidesk treats these as side-channel events from a phone system, not as a
// REST resource family. The CLI exposes them as low-level "fire event"
// commands so integrators can replay events from a script.
package telephony

import (
	"context"
	"net/url"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
)

const (
	// With queue (POST).
	pathReceivedCall   = "/asterisk_receivedCall"
	pathTransferedCall = "/asterisk_transferedCall"
	pathCompletedCall  = "/asterisk_completedCall"
	pathLostCall       = "/asterisk_lostCall"
	pathCanceledCall   = "/asterisk_canceledCall"

	// Without queue (GET with query params).
	pathStartTransferedCall = "/asterisk_startTransferedCall"
	pathStartCanceledCall   = "/asterisk_startCanceledCall"

	// Made-call link (POST), shared across both flavors.
	pathSetMadeCallLink = "/setMadeCallLink"
)

// Service binds telephony endpoints to a Movidesk client.
type Service struct {
	C *movidesk.Client
}

func New(c *movidesk.Client) *Service { return &Service{C: c} }

// QueuePost dispatches a POST /asterisk_<event> call (queue-controlled flavor).
// event must be one of: "receivedCall", "transferedCall", "completedCall",
// "lostCall", "canceledCall". body is forwarded as-is.
func (s *Service) QueuePost(ctx context.Context, event string, body any) ([]byte, error) {
	p := pathFor(event)
	if p == "" {
		return nil, errUnknownEvent(event)
	}
	return s.C.Post(ctx, p, nil, body)
}

// NonQueueGet dispatches a GET /asterisk_<event> call (no-queue flavor) with
// query parameters. event must be one of: "startTransferedCall",
// "completedCall", "startCanceledCall".
func (s *Service) NonQueueGet(ctx context.Context, event string, params url.Values) ([]byte, error) {
	var p string
	switch event {
	case "startTransferedCall":
		p = pathStartTransferedCall
	case "completedCall":
		p = pathCompletedCall
	case "startCanceledCall":
		p = pathStartCanceledCall
	default:
		return nil, errUnknownEvent(event)
	}
	return s.C.Do(ctx, "GET", p, params, nil)
}

// SetMadeCallLink attaches a recording URL to an outbound (made) call.
func (s *Service) SetMadeCallLink(ctx context.Context, body any) ([]byte, error) {
	return s.C.Post(ctx, pathSetMadeCallLink, nil, body)
}

func pathFor(event string) string {
	switch event {
	case "receivedCall":
		return pathReceivedCall
	case "transferedCall":
		return pathTransferedCall
	case "completedCall":
		return pathCompletedCall
	case "lostCall":
		return pathLostCall
	case "canceledCall":
		return pathCanceledCall
	}
	return ""
}

type unknownEventError string

func (e unknownEventError) Error() string { return "telephony: unknown event " + string(e) }

func errUnknownEvent(event string) error { return unknownEventError(event) }
