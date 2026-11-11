package ikuai

import (
	"net/http"
	"time"
)

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient supplies a custom *http.Client (timeouts, proxies,
// transports). A per-request context timeout is still honoured.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.http = hc
		}
	}
}

// WithTimeout sets the default per-request timeout used when the
// caller's context carries none. Defaults to 30s.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithUserAgent overrides the User-Agent header. Defaults to a value
// that mirrors the web console.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithTLSInsecureSkipVerify skips TLS certificate verification (useful
// for self-signed certificates on https management addresses).
func WithTLSInsecureSkipVerify(skip bool) Option {
	return func(c *Client) { c.tlsSkipVerify = skip }
}

// WithAutoRelogin enables (default) or disables transparent re-login +
// single retry when a call fails with the session-expired error.
func WithAutoRelogin(enabled bool) Option {
	return func(c *Client) { c.autoRelogin = enabled }
}

// WithLoginPass supplies the optional "pass" field of the login
// request. Real firmware accepts logins without it; the field only
// exists for exotic deployments. The value is sent verbatim, so callers
// needing base64("salt_"+challenge) must encode it themselves.
func WithLoginPass(pass string) Option {
	return func(c *Client) { c.loginPass = pass }
}

// WithRememberPassword sets the remember_password flag of the login
// request. It only affects server-side session persistence behaviour.
func WithRememberPassword(v bool) Option {
	return func(c *Client) { c.rememberPassword = v }
}

// WithLogger installs a debug logger receiving one line per request.
type Logger interface {
	Printf(format string, v ...any)
}

// WithLogger enables request logging.
func WithLogger(l Logger) Option {
	return func(c *Client) { c.logger = l }
}
