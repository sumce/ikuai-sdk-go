package ikuai_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	ikuai "github.com/sumce/ikuai-sdk-go"
)

// TestAPIErrorIsViaErrorsIs 验证包装后的 *APIError 可用 errors.Is 按
// 结果码匹配（目标中的零值 FuncName/Action 为通配）。
func TestAPIErrorIsViaErrorsIs(t *testing.T) {
	raw := &ikuai.APIError{Code: ikuai.CodeNoLoginAuth, ErrMsg: "no login authentication", FuncName: "dnat", Action: "show"}
	wrapped := fmt.Errorf("查询端口映射: %w", raw)

	if !errors.Is(wrapped, &ikuai.APIError{Code: ikuai.CodeNoLoginAuth}) {
		t.Fatal("errors.Is 应按 Code=10014 匹配（通配 FuncName/Action）")
	}
	if !errors.Is(wrapped, &ikuai.APIError{Code: ikuai.CodeNoLoginAuth, FuncName: "dnat"}) {
		t.Fatal("errors.Is 应在 FuncName 一致时匹配")
	}
	if errors.Is(wrapped, &ikuai.APIError{Code: ikuai.CodeNoLoginAuth, FuncName: "firewall"}) {
		t.Fatal("FuncName 不一致时 errors.Is 不应匹配")
	}
	if errors.Is(wrapped, &ikuai.APIError{Code: ikuai.CodeSuccess}) {
		t.Fatal("Code 不一致时 errors.Is 不应匹配")
	}
	if errors.Is(wrapped, errors.New("其他错误")) {
		t.Fatal("非 *APIError 目标不应匹配")
	}
}

// TestAPIErrorMethods 验证错误类型上的判定方法。
func TestAPIErrorMethods(t *testing.T) {
	ae := &ikuai.APIError{Code: ikuai.CodeBadCredentials, ErrMsg: "账号或密码错误"}
	if !ae.IsBadCredentials() {
		t.Fatal("10001 应判定为凭据错误")
	}

	auth := &ikuai.APIError{Code: ikuai.CodeNoLoginAuth}
	if !auth.IsAuthError() {
		t.Fatal("10014 应判定为会话失效")
	}
	auth401 := &ikuai.APIError{HTTPStatus: http.StatusUnauthorized}
	if !auth401.IsAuthError() {
		t.Fatal("HTTP 401 应判定为会话失效")
	}
	var nilErr *ikuai.APIError
	if nilErr.IsAuthError() || nilErr.IsCode(1) {
		t.Fatal("nil 接收者应安全返回 false")
	}
}

// TestAsAPIError 验证 AsAPIError 走标准 errors.As 链路，
// 支持任意深度与多重解包形式。
func TestAsAPIError(t *testing.T) {
	raw := &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname", FuncName: "wireguard", Action: "show"}

	// 单层 %w。
	one := fmt.Errorf("调用失败: %w", raw)
	got, ok := ikuai.AsAPIError(one)
	if !ok || got.Code != 30002 || got.FuncName != "wireguard" {
		t.Fatalf("单层包装应还原 APIError: ok=%v got=%+v", ok, got)
	}

	// 多层 %w。
	two := fmt.Errorf("外层: %w", one)
	if _, ok := ikuai.AsAPIError(two); !ok {
		t.Fatal("多层包装也应还原 APIError")
	}

	// 非 APIError。
	if _, ok := ikuai.AsAPIError(errors.New("boom")); ok {
		t.Fatal("普通错误不应还原出 APIError")
	}
	if _, ok := ikuai.AsAPIError(nil); ok {
		t.Fatal("nil 不应还原出 APIError")
	}
}
