// Package movidesk is the internal SDK that wraps the Movidesk public REST API.
//
// Movidesk authenticates via the ?token query parameter, applies OData
// conventions for query, rate-limits to 10 requests/min, and returns 429 with
// a retry-after header on exhaustion. This client centralizes those concerns.
package movidesk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

const (
	DefaultTimeout = 30 * time.Second
	UserAgent      = "movidesk-cli"
)

// Client talks to the Movidesk REST API.
type Client struct {
	BaseURL    string
	Token      string
	HTTP       *http.Client
	Limiter    *Limiter
	Retry      RetryPolicy
	UserAgent  string
	OnRequest  func(method, url string)  // optional verbose hook
	OnResponse func(status int, ms int64) // optional verbose hook
}

// New creates a Client with sane defaults: 30s timeout, 10 req/min limiter,
// retry on 429/5xx.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Token:     token,
		HTTP:      &http.Client{Timeout: DefaultTimeout},
		Limiter:   NewLimiter(10, time.Minute),
		Retry:     DefaultRetry(),
		UserAgent: UserAgent,
	}
}

// APIError is returned when Movidesk responds with a non-2xx status.
type APIError struct {
	Status int
	Body   string
	Method string
	Path   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("movidesk %s %s: HTTP %d: %s", e.Method, e.Path, e.Status, truncate(e.Body, 500))
}

// IsUnauthorized reports whether the error is a 401/403.
func IsUnauthorized(err error) bool {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.Status == http.StatusUnauthorized || ae.Status == http.StatusForbidden
	}
	return false
}

// Do executes a request with rate limiting and retry. The token query param is
// injected automatically; callers must NOT include it in path/params.
func (c *Client) Do(ctx context.Context, method, path string, params url.Values, body any) ([]byte, error) {
	var attempt int
	for {
		attempt++
		if err := c.Limiter.Wait(ctx); err != nil {
			return nil, err
		}
		data, retryAfter, err := c.do(ctx, method, path, params, body)
		if err == nil {
			return data, nil
		}
		if c.Retry.Disabled || !c.Retry.ShouldRetry(err, attempt) {
			return nil, err
		}
		wait := c.Retry.Backoff(attempt, retryAfter)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) do(ctx context.Context, method, path string, params url.Values, body any) ([]byte, time.Duration, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("token", c.Token)
	full := c.BaseURL + path
	if enc := params.Encode(); enc != "" {
		full += "?" + enc
	}

	var reqBody io.Reader
	contentType := ""
	if body != nil {
		switch b := body.(type) {
		case []byte:
			reqBody = bytes.NewReader(b)
			contentType = "application/json"
		case io.Reader:
			reqBody = b
			if ct, ok := body.(interface{ ContentType() string }); ok {
				contentType = ct.ContentType()
			}
		default:
			buf, err := json.Marshal(body)
			if err != nil {
				return nil, 0, fmt.Errorf("marshal body: %w", err)
			}
			reqBody = bytes.NewReader(buf)
			contentType = "application/json"
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, full, reqBody)
	if err != nil {
		return nil, 0, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	if c.OnRequest != nil {
		c.OnRequest(method, redact(full))
	}
	start := time.Now()
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if c.OnResponse != nil {
		c.OnResponse(resp.StatusCode, time.Since(start).Milliseconds())
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, 0, nil
	}
	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
	return nil, retryAfter, &APIError{
		Status: resp.StatusCode,
		Body:   string(respBody),
		Method: method,
		Path:   path,
	}
}

// Get is a convenience for GET with OData params.
func (c *Client) Get(ctx context.Context, path string, q odata.Query, extra url.Values) ([]byte, error) {
	v := url.Values{}
	q.Apply(v)
	for k, vs := range extra {
		for _, val := range vs {
			v.Add(k, val)
		}
	}
	return c.Do(ctx, http.MethodGet, path, v, nil)
}

// Post is a convenience for POST with a JSON body.
func (c *Client) Post(ctx context.Context, path string, params url.Values, body any) ([]byte, error) {
	return c.Do(ctx, http.MethodPost, path, params, body)
}

// Patch is a convenience for PATCH with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, params url.Values, body any) ([]byte, error) {
	return c.Do(ctx, http.MethodPatch, path, params, body)
}

// Delete is a convenience for DELETE.
func (c *Client) Delete(ctx context.Context, path string, params url.Values) ([]byte, error) {
	return c.Do(ctx, http.MethodDelete, path, params, nil)
}

// DecodeJSON unmarshals a successful response body into v.
func DecodeJSON(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}

func redact(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	q := parsed.Query()
	if q.Get("token") != "" {
		q.Set("token", "REDACTED")
		parsed.RawQuery = q.Encode()
	}
	return parsed.String()
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	// Numeric seconds.
	var secs int
	if _, err := fmt.Sscanf(v, "%d", &secs); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	// HTTP-date.
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
