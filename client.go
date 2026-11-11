package ikuai

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// defaultTimeout bounds each API request when the caller's context has
// no deadline of its own.
const defaultTimeout = 30 * time.Second

// userAgent mirrors the official web console closely enough that the
// firmware treats the SDK like a browser session.
const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// Client talks to one iKuaiOS router. It is safe for concurrent use;
// the login session is shared and refreshed automatically.
type Client struct {
	base             *url.URL
	username         string
	password         string
	loginPass        string
	rememberPassword bool

	http          *http.Client
	timeout       time.Duration
	userAgent     string
	tlsSkipVerify bool
	autoRelogin   bool
	logger        Logger

	mu       sync.Mutex
	sessKey  string
	loggedIn bool
}

// NewClient creates a Client for the router at rawBase (e.g.
// "http://192.168.1.1", "https://router.local:443", trailing slash
// optional). Credentials are those of a web console account. Call
// Login before issuing requests, or rely on auto re-login of the first
// Call.
func NewClient(rawBase, username, password string, opts ...Option) (*Client, error) {
	rawBase = strings.TrimRight(rawBase, "/")
	u, err := url.Parse(rawBase)
	if err != nil {
		return nil, fmt.Errorf("ikuai: invalid base URL %q: %w", rawBase, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("ikuai: base URL must be http(s), got %q", rawBase)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("ikuai: base URL %q has no host", rawBase)
	}
	c := &Client{
		base:             u,
		username:         username,
		password:         password,
		http:             &http.Client{Timeout: defaultTimeout},
		timeout:          defaultTimeout,
		userAgent:        defaultUserAgent,
		autoRelogin:      true,
		rememberPassword: false,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.tlsSkipVerify {
		tr := c.http.Transport
		if tr == nil {
			tr = &http.Transport{}
		}
		if trt, ok := tr.(*http.Transport); ok {
			clone := trt.Clone()
			if clone.TLSClientConfig == nil {
				clone.TLSClientConfig = &tls.Config{}
			}
			clone.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // opt-in option
			c.http.Transport = clone
		}
	}
	return c, nil
}

// BaseURL returns the normalized management base URL.
func (c *Client) BaseURL() string { return strings.TrimRight(c.base.String(), "/") }

// LoggedIn reports whether the client currently holds a session key.
func (c *Client) LoggedIn() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loggedIn && c.sessKey != ""
}

func (c *Client) debugf(format string, v ...any) {
	if c.logger != nil {
		c.logger.Printf(format, v...)
	}
}

// sessionCookie renders the Cookie header exactly like the web console:
// "sess_key=...; username=...; login=1".
func (c *Client) sessionCookie() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sessKey == "" {
		return ""
	}
	return "sess_key=" + c.sessKey + "; username=" + url.QueryEscape(c.username) + "; login=1"
}

func (c *Client) setSession(key string) {
	c.mu.Lock()
	c.sessKey = key
	c.loggedIn = key != ""
	c.mu.Unlock()
}

// ensureSession logs in when no session is held yet. Concurrent callers
// are coalesced by the mutex (single-flight).
func (c *Client) ensureSession(ctx context.Context) error {
	c.mu.Lock()
	ok := c.sessKey != ""
	c.mu.Unlock()
	if ok {
		return nil
	}
	return c.Login(ctx)
}

// httpResult carries everything the SDK needs from an HTTP exchange.
type httpResult struct {
	Body   []byte
	Status int
	Header http.Header
}

// do sends an HTTP request to the router and returns the response.
// body may be nil, []byte, a string or any JSON-marshalable value
// (marshalled with Content-Type application/json).
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any) (*httpResult, error) {
	u := *c.base
	u.Path = strings.TrimRight(u.Path, "/") + path
	if query != nil {
		u.RawQuery = query.Encode()
	}
	var rdr io.Reader
	isJSON := false
	switch b := body.(type) {
	case nil:
	case []byte:
		rdr = bytes.NewReader(b)
	case string:
		rdr = strings.NewReader(b)
	default:
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("ikuai: encode request body: %w", err)
		}
		rdr = bytes.NewReader(raw)
		isJSON = true
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rdr)
	if err != nil {
		return nil, fmt.Errorf("ikuai: build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	if isJSON {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	if ck := c.sessionCookie(); ck != "" {
		req.Header.Set("Cookie", ck)
	}
	if ref := req.URL; c.base.Host == ref.Host {
		req.Header.Set("Origin", c.base.Scheme+"://"+c.base.Host)
		req.Header.Set("Referer", c.base.Scheme+"://"+c.base.Host+"/")
	}
	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ikuai: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ikuai: read response: %w", err)
	}
	c.debugf("%s %s -> %d (%d bytes, %s)", method, path, resp.StatusCode, len(raw), time.Since(start).Round(time.Millisecond))
	return &httpResult{Body: raw, Status: resp.StatusCode, Header: resp.Header}, nil
}

// callRequest is the /Action/call payload.
type callRequest struct {
	FuncName string `json:"func_name"`
	Action   string `json:"action"`
	Param    any    `json:"param"`
}

// Call invokes func_name.action through /Action/call and returns the
// decoded envelope. param may be nil, a map[string]any, a struct or any
// JSON-marshalable value; nil is sent as an empty JSON object, matching
// the web console. Non-success results are returned as *APIError.
//
// Call is the escape hatch that gives access to every func_name of the
// router, including ones without a typed wrapper in this SDK.
func (c *Client) Call(ctx context.Context, funcName, action string, param any) (*Response, error) {
	resp, err := c.callOnce(ctx, funcName, action, param)
	if err == nil {
		return resp, nil
	}
	// Transparent re-login + one retry on session expiry.
	if ae, ok := err.(*APIError); ok && ae.IsAuthError() && c.autoRelogin {
		c.debugf("session expired (Result=%d); re-login and retry %s.%s", ae.Code, funcName, action)
		c.setSession("")
		if lerr := c.Login(ctx); lerr != nil {
			return nil, lerr
		}
		return c.callOnce(ctx, funcName, action, param)
	}
	return nil, err
}

// CallData is Call plus decoding of the Data member into out.
func (c *Client) CallData(ctx context.Context, funcName, action string, param, out any) error {
	resp, err := c.Call(ctx, funcName, action, param)
	if err != nil {
		return err
	}
	return resp.decodeData(out)
}

func (c *Client) callOnce(ctx context.Context, funcName, action string, param any) (*Response, error) {
	if funcName == "" {
		return nil, fmt.Errorf("ikuai: func_name must not be empty")
	}
	if action == "" {
		action = "show"
	}
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	if param == nil {
		param = map[string]any{}
	}
	res, err := c.do(ctx, http.MethodPost, "/Action/call", nil, callRequest{
		FuncName: funcName,
		Action:   action,
		Param:    param,
	})
	if err != nil {
		return nil, err
	}
	resp, derr := decodeEnvelope(res.Body)
	if derr != nil {
		return nil, derr
	}
	resp.Result = normalizeResult(resp.Result, res.Status)
	if !resp.OK() {
		return resp, &APIError{
			Code:       resp.Result,
			ErrMsg:     resp.ErrMsg,
			FuncName:   funcName,
			Action:     action,
			HTTPStatus: res.Status,
			Body:       string(res.Body),
		}
	}
	return resp, nil
}

// normalizeResult maps odd HTTP statuses onto synthetic result codes so
// that callers always see an *APIError with a meaningful ErrMsg.
func normalizeResult(code, httpStatus int) int {
	if code != 0 {
		return code
	}
	if httpStatus == http.StatusUnauthorized {
		return CodeNoLoginAuth
	}
	return code
}

// CallRaw performs a raw HTTP request against the router (relative to
// the management base URL) with the session cookie attached, returning
// the raw body. Useful for endpoints outside /Action/call such as
// /json/menu.json or /static/js/first.json.
func (c *Client) CallRaw(ctx context.Context, method, path string, query url.Values) ([]byte, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, err
	}
	res, err := c.do(ctx, method, path, query, nil)
	if err != nil {
		return nil, err
	}
	if res.Status >= 400 {
		return res.Body, &APIError{HTTPStatus: res.Status, ErrMsg: http.StatusText(res.Status), Body: string(res.Body)}
	}
	return res.Body, nil
}
