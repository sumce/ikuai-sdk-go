package ikuai

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrInvalidParam 是业务包参数校验失败时用于 errors.Is 判断的哨兵。
var ErrInvalidParam = errors.New("ikuai: 参数非法")

// iKuaiOS result codes observed on the wire (verified against a 3.7.x
// device and Burp captures of the official web UI).
const (
	// CodeLoginSuccess is returned by /Action/login on success.
	CodeLoginSuccess = 10000
	// CodeBadCredentials is "账号或密码错误" from /Action/login.
	CodeBadCredentials = 10001
	// CodeNoLoginAuth is "no login authentication": missing, stale or
	// expired session on /Action/call.
	CodeNoLoginAuth = 10014
	// CodeSuccess is returned by /Action/call on success.
	CodeSuccess = 30000
)

// APIError is returned for every non-successful iKuai API result and
// for failed HTTP exchanges.
type APIError struct {
	// Code is the iKuai "Result" field (e.g. 30000, 10014).
	Code int
	// ErrMsg is the iKuai "ErrMsg" field verbatim.
	ErrMsg string
	// FuncName and Action identify the call that failed (may be empty
	// for login errors).
	FuncName string
	Action   string
	// HTTPStatus is the underlying HTTP status code.
	HTTPStatus int
	// Body holds the raw response body when it could not be decoded.
	Body string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.FuncName != "" {
		return fmt.Sprintf("ikuai: %s.%s failed: Result=%d ErrMsg=%q (HTTP %d)",
			e.FuncName, e.Action, e.Code, e.ErrMsg, e.HTTPStatus)
	}
	return fmt.Sprintf("ikuai: Result=%d ErrMsg=%q (HTTP %d)", e.Code, e.ErrMsg, e.HTTPStatus)
}

// IsCode reports whether the error carries the given iKuai result code.
func (e *APIError) IsCode(code int) bool { return e != nil && e.Code == code }

// IsAuthError reports whether the error means "not authenticated /
// session expired". Such errors are retried once after re-login when
// WithAutoRelogin is enabled (default).
func (e *APIError) IsAuthError() bool {
	if e == nil {
		return false
	}
	if e.Code == CodeNoLoginAuth {
		return true
	}
	return e.HTTPStatus == http.StatusUnauthorized
}

// IsBadCredentials reports whether a login failed because of a wrong
// username or password.
func (e *APIError) IsBadCredentials() bool { return e.IsCode(CodeBadCredentials) }

// Is lets errors.Is match a wrapped *APIError against a target carrying
// the expected result code, e.g.
//
//	errors.Is(err, &ikuai.APIError{Code: ikuai.CodeNoLoginAuth})
//
// Non-zero FuncName/Action fields of the target must match too; zero
// fields act as wildcards so coarse targets stay useful.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	if e.Code != t.Code {
		return false
	}
	if t.FuncName != "" && e.FuncName != t.FuncName {
		return false
	}
	if t.Action != "" && e.Action != t.Action {
		return false
	}
	return true
}

// AsAPIError unwraps err into *APIError when possible, following the
// full error chain via errors.As (single Unwrap, Unwrap() []error and
// any combination thereof).
func AsAPIError(err error) (*APIError, bool) {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
