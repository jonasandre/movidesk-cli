package movidesk

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(srv *httptest.Server) *Client {
	c := New(srv.URL, "test-token")
	// Disable rate limiter for unit tests (use generous bucket).
	c.Limiter = NewLimiter(1000, time.Minute)
	return c
}

func TestDo_InjectsToken(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query().Get("token")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	body, err := c.Do(context.Background(), "GET", "/tickets", nil, nil)
	require.NoError(t, err)
	assert.Contains(t, string(body), "ok")
	assert.Equal(t, "test-token", seen)
}

func TestDo_AppliesODataParams(t *testing.T) {
	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	v := url.Values{}
	v.Set("$filter", "id eq 1")
	v.Set("$top", "5")
	_, err := c.Do(context.Background(), "GET", "/tickets", v, nil)
	require.NoError(t, err)
	assert.Equal(t, "id eq 1", query.Get("$filter"))
	assert.Equal(t, "5", query.Get("$top"))
}

func TestDo_PostJSONBody(t *testing.T) {
	var got string
	var ct string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":42}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	body, err := c.Post(context.Background(), "/tickets", nil, map[string]any{"subject": "x"})
	require.NoError(t, err)
	assert.Equal(t, "application/json", ct)
	assert.Contains(t, got, `"subject":"x"`)
	assert.Contains(t, string(body), `"id":42`)
}

func TestDo_RetriesOn429WithRetryAfter(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"err":"rate"}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.Retry.BaseBackoff = time.Millisecond
	body, err := c.Do(context.Background(), "GET", "/tickets", nil, nil)
	require.NoError(t, err)
	assert.Contains(t, string(body), "ok")
	assert.Equal(t, int32(2), calls)
}

func TestDo_ReturnsAPIErrorOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"err":"unauthorized"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.Retry.MaxAttempts = 1
	_, err := c.Do(context.Background(), "GET", "/tickets", nil, nil)
	require.Error(t, err)
	var ae *APIError
	require.True(t, errors.As(err, &ae))
	assert.Equal(t, 401, ae.Status)
	assert.True(t, IsUnauthorized(err))
}

func TestDo_DoesNotRetry4xxOtherThan429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(400)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Do(context.Background(), "GET", "/tickets", nil, nil)
	require.Error(t, err)
	assert.Equal(t, int32(1), calls)
}

func TestRedact_TokenStripped(t *testing.T) {
	got := redact("https://api.movidesk.com/public/v1/tickets?token=secret&id=1")
	assert.Contains(t, got, "token=REDACTED")
	assert.NotContains(t, got, "secret")
}

func TestParseRetryAfter_Numeric(t *testing.T) {
	assert.Equal(t, 5*time.Second, parseRetryAfter("5"))
	assert.Equal(t, time.Duration(0), parseRetryAfter(""))
}

func TestLimiter_AllowsCapacity(t *testing.T) {
	l := NewLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		require.NoError(t, l.Wait(context.Background()))
	}
}

func TestLimiter_BlocksThenReleases(t *testing.T) {
	l := NewLimiter(2, 50*time.Millisecond)
	require.NoError(t, l.Wait(context.Background()))
	require.NoError(t, l.Wait(context.Background()))

	start := time.Now()
	require.NoError(t, l.Wait(context.Background()))
	assert.GreaterOrEqual(t, time.Since(start), 30*time.Millisecond)
}

func TestRetryPolicy_ShouldRetry(t *testing.T) {
	p := DefaultRetry()
	assert.True(t, p.ShouldRetry(&APIError{Status: 429}, 1))
	assert.True(t, p.ShouldRetry(&APIError{Status: 503}, 1))
	assert.False(t, p.ShouldRetry(&APIError{Status: 401}, 1))
	assert.False(t, p.ShouldRetry(&APIError{Status: 429}, 99))
	// Network error retries.
	assert.True(t, p.ShouldRetry(errors.New("dial: connection refused"), 1))
}

func TestRetryPolicy_BackoffHonorsRetryAfter(t *testing.T) {
	p := DefaultRetry()
	assert.Equal(t, 7*time.Second, p.Backoff(1, 7*time.Second))
	assert.Equal(t, p.MaxBackoff, p.Backoff(1, 999*time.Hour))
	assert.Equal(t, time.Second, p.Backoff(1, 0))
	assert.Equal(t, 2*time.Second, p.Backoff(2, 0))
}

func TestUnauthorized_NotApiError(t *testing.T) {
	assert.False(t, IsUnauthorized(errors.New("boom")))
}

func TestDo_PostDoesNotRetryOn5xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"err":"unavailable"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.Retry.BaseBackoff = time.Millisecond
	_, err := c.Post(context.Background(), "/tickets", nil, map[string]any{"subject": "x"})
	require.Error(t, err)
	// POST must not retry even on 5xx; exactly one attempt expected.
	assert.Equal(t, int32(1), calls)
}

func TestDo_PatchDoesNotRetryOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(429)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.Retry.BaseBackoff = time.Millisecond
	_, err := c.Patch(context.Background(), "/tickets/1", nil, map[string]any{"status": 4})
	require.Error(t, err)
	// PATCH is not idempotent; must not retry.
	assert.Equal(t, int32(1), calls)
}

func TestDo_DeleteDoesNotRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(503)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.Retry.BaseBackoff = time.Millisecond
	_, err := c.Delete(context.Background(), "/tickets/1", nil)
	require.Error(t, err)
	assert.Equal(t, int32(1), calls)
}

func TestDo_GetRetriesTransportError(t *testing.T) {
	// Simulate a server that closes the connection on the first attempt.
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			// Hijack to force a transport-level error on the first attempt.
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Error("server does not support hijacking")
				return
			}
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	c.Retry.BaseBackoff = time.Millisecond
	body, err := c.Do(context.Background(), "GET", "/tickets", nil, nil)
	require.NoError(t, err)
	assert.Contains(t, string(body), "ok")
	assert.Equal(t, int32(2), calls)
}

func TestLimiter_CancelledContextDoesNotCorruptStamps(t *testing.T) {
	// Fill capacity to 2, then cancel a waiter. Existing stamps must not be removed.
	l := NewLimiter(2, time.Minute)
	require.NoError(t, l.Wait(context.Background()))
	require.NoError(t, l.Wait(context.Background()))

	// Now the limiter is full. A waiter with a cancelled context should fail
	// and leave the two existing stamps intact.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	err := l.Wait(ctx)
	require.ErrorIs(t, err, context.Canceled)

	// The two original stamps must still be there; any further immediate wait
	// should still need to block (not succeed instantly).
	wait := l.reserve()
	assert.Greater(t, wait, time.Duration(0), "stamps should still be present after cancelled wait")
}

func TestIsSafeMethod(t *testing.T) {
	assert.True(t, isSafeMethod("GET"))
	assert.True(t, isSafeMethod("get"))
	assert.True(t, isSafeMethod("HEAD"))
	assert.True(t, isSafeMethod("OPTIONS"))
	assert.False(t, isSafeMethod("POST"))
	assert.False(t, isSafeMethod("PATCH"))
	assert.False(t, isSafeMethod("DELETE"))
	assert.False(t, isSafeMethod("PUT"))
}

// Compile-time guard so unused imports don't break.
var _ = strings.Contains
