package ikuai

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

// loginRequest is the /Action/login payload, byte-identical to the one
// the web console sends:
//
//	{"username":"admin","passwd":"<md5(password)>","pass":"...","remember_password":false}
//
// "pass" is optional on current firmware (verified: login succeeds with
// username+passwd only); it is included when configured via
// WithLoginPass.
type loginRequest struct {
	Username         string `json:"username"`
	Passwd           string `json:"passwd"`
	Pass             string `json:"pass,omitempty"`
	RememberPassword *bool  `json:"remember_password"`
}

// sessKeyPattern extracts sess_key from a Set-Cookie header value.
var sessKeyPattern = regexp.MustCompile(`sess_key=([0-9a-fA-F]+)`)

// Login authenticates against POST /Action/login and stores the
// session key returned in the Set-Cookie header. It is safe to call
// concurrently and to call again to replace an existing session
// (previous sessions on the router remain valid until they time out).
func (c *Client) Login(ctx context.Context) error {
	sum := md5.Sum([]byte(c.password))
	req := loginRequest{
		Username:         c.username,
		Passwd:           hex.EncodeToString(sum[:]),
		Pass:             c.loginPass,
		RememberPassword: &c.rememberPassword,
	}
	res, err := c.do(ctx, http.MethodPost, "/Action/login", nil, req)
	if err != nil {
		return err
	}
	var env struct {
		Result int    `json:"Result"`
		ErrMsg string `json:"ErrMsg"`
	}
	if jerr := json.Unmarshal(sanitizeJSON(res.Body), &env); jerr != nil {
		return fmt.Errorf("ikuai: login: cannot decode response: %w (body: %.200s)", jerr, res.Body)
	}
	if env.Result != CodeLoginSuccess {
		return &APIError{Code: env.Result, ErrMsg: env.ErrMsg, HTTPStatus: res.Status, Body: string(res.Body)}
	}
	key := extractSessKey(res.Header.Values("Set-Cookie"))
	if key == "" {
		return fmt.Errorf("ikuai: login succeeded but no sess_key cookie was returned (HTTP %d)", res.Status)
	}
	c.setSession(key)
	c.debugf("login ok (user=%s, sess_key=%s...)", c.username, key[:min(8, len(key))])
	return nil
}

// extractSessKey pulls the sess_key value out of Set-Cookie header
// values.
func extractSessKey(setCookies []string) string {
	for _, ck := range setCookies {
		if m := sessKeyPattern.FindStringSubmatch(ck); m != nil {
			return m[1]
		}
	}
	return ""
}

// Logout clears the local session state. iKuaiOS does not expose a
// server-side logout endpoint, so the router session simply expires by
// its own idle timeout.
func (c *Client) Logout(_ context.Context) {
	c.setSession("")
}

// KeepAlive refreshes the session with a lightweight homepage:show
// call, the same call the web console repeats to keep sessions alive.
func (c *Client) KeepAlive(ctx context.Context) error {
	_, err := c.Call(ctx, "homepage", "show", map[string]any{"TYPE": "sysstat"})
	return err
}
