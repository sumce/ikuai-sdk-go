// 核心包冒烟测试（外部测试包，同时是各业务包测试的参考范例）。
package ikuai_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestLoginSuccess 验证登录成功后持有会话，并携带标准 Cookie。
func TestLoginSuccess(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	c := ts.Client()
	ctx := context.Background()
	if err := c.Login(ctx); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if !c.LoggedIn() {
		t.Fatal("登录后 LoggedIn() 应为 true")
	}
	// 触发一次调用，断言 Cookie 形态与浏览器一致。
	ts.SetResponse("homepage", "show", map[string]any{"sysstat": map[string]any{}})
	if _, err := c.Call(ctx, "homepage", "show", map[string]any{"TYPE": "sysstat"}); err != nil {
		t.Fatalf("Call 失败: %v", err)
	}
	rec := ts.CallsFor("homepage")[0]
	cookie := rec.Cookie
	if !contains(cookie, "sess_key=") || !contains(cookie, "username=admin") || !contains(cookie, "login=1") {
		t.Fatalf("Cookie 形态不符: %q", cookie)
	}
}

// TestLoginBadCredentials 验证错误凭据返回 *APIError(10001)。
func TestLoginBadCredentials(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	c, err := ikuai.NewClient(ts.URL, "admin", "wrong-password")
	if err != nil {
		t.Fatalf("构造失败: %v", err)
	}
	err = c.Login(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || !ae.IsBadCredentials() || ae.Code != 10001 {
		t.Fatalf("期望 APIError(10001), got %v", err)
	}
}

// TestAutoRelogin 验证会话失效(10014)时自动重登并重试一次。
func TestAutoRelogin(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	c := ts.Client()
	ctx := context.Background()
	if err := c.Login(ctx); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	ts.SetResponse("dnat", "show", map[string]any{"total": 0, "data": []any{}})
	ts.ExpireSessions() // 使现有会话失效 → 下一次调用先得 10014
	if _, err := c.Call(ctx, "dnat", "show", nil); err != nil {
		t.Fatalf("自动重登后重试失败: %v", err)
	}
	// 第一次调用(10014 失败) + 重登后重试(成功) = 2 条记录。
	if n := len(ts.CallsFor("dnat")); n != 2 {
		t.Fatalf("dnat 应被调用 2 次(1 次 10014 + 1 次重试), got %d", n)
	}
}

// TestCallError 验证非 30000 结果包装为 *APIError。
func TestCallError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	c := ts.Client()
	ctx := context.Background()
	_ = c.Login(ctx)
	ts.HandleFuncName("dnat", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95001, ErrMsg: "参数缺失", FuncName: "dnat", Action: action}
	})
	_, err := c.Call(ctx, "dnat", "add", nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95001 || ae.ErrMsg != "参数缺失" {
		t.Fatalf("期望 APIError(95001), got %v", err)
	}
}

// TestQuirkEmptyData 验证固件非法 JSON 怪癖（"Data":}）被核心层修复。
func TestQuirkEmptyData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Action/login":
			w.Header().Add("Set-Cookie", "sess_key=abcdef0123456789abcdef0123456789; path=/;")
			_, _ = w.Write([]byte(`{"ErrMsg":"Success","Result":10000}`))
		default:
			_, _ = w.Write([]byte(`{"Result":30000,"ErrMsg":"Success","Data":}`))
		}
	}))
	defer srv.Close()
	c, err := ikuai.NewClient(srv.URL, "u", "p")
	if err != nil {
		t.Fatalf("构造失败: %v", err)
	}
	resp, err := c.Call(context.Background(), "auth_code", "show", nil)
	if err != nil {
		t.Fatalf("非法 Data 应被修复: %v", err)
	}
	if !resp.OK() {
		t.Fatal("Result 应为 30000")
	}
}

// TestCrudDeleteParams 验证通用 CRUD 的 del 参数形态。
func TestCrudDeleteParams(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	c := ts.Client()
	ctx := context.Background()
	_ = c.Login(ctx)
	ts.HandleFuncName("dnat", func(action string, param json.RawMessage) (any, error) {
		if action == "del" {
			var p map[string]any
			_ = json.Unmarshal(param, &p)
			if _, single := p["id"]; !single {
				if _, multi := p["ids"]; !multi {
					t.Errorf("del 参数应含 id 或 ids: %s", param)
				}
			}
		}
		return nil, nil
	})
	if err := c.CRUD("dnat").Delete(ctx, int64(7)); err != nil {
		t.Fatalf("Delete 单个失败: %v", err)
	}
	if err := c.CRUD("dnat").Delete(ctx, int64(1), int64(2), int64(3)); err != nil {
		t.Fatalf("Delete 多个失败: %v", err)
	}
}

// TestIDParam 验证 del 参数渲染规则。
func TestIDParam(t *testing.T) {
	if p := ikuai.IDParam([]int64{7}); p["id"] != int64(7) {
		t.Fatalf("单个 id: %v", p)
	}
	p := ikuai.IDParam([]int64{1, 2})
	ids, ok := p["ids"].([]any)
	if !ok || len(ids) != 2 {
		t.Fatalf("多个 ids: %v", p)
	}
	if p := ikuai.IDParam([]int64{}); len(p["ids"].([]any)) != 0 {
		t.Fatalf("空 ids: %v", p)
	}
}

// TestListOptionsBuildParam 验证 show 参数构造。
func TestListOptionsBuildParam(t *testing.T) {
	opts := &ikuai.ListOptions{Offset: 20, Limit: 20, OrderBy: "ip_addr_int",
		Finds: "ip_addr,mac", Keywords: "172.22",
		Filters: []ikuai.Filter{{Field: "read", Op: "==", Value: 0}}}
	p := opts.BuildParam()
	if p["TYPE"] != "total,data" || p["limit"] != "20,20" || p["ORDER_BY"] != "ip_addr_int" {
		t.Fatalf("BuildParam: %v", p)
	}
	if p["FILTER1"] != "read,==,0" || p["FINDS"] != "ip_addr,mac" || p["KEYWORDS"] != "172.22" {
		t.Fatalf("BuildParam 过滤项: %v", p)
	}
	if p := ikuai.DefaultListOptions().BuildParam(); len(p) == 0 {
		t.Fatalf("默认选项应有参数: %v", p)
	}
	if !errors.Is(ikuai.ErrInvalidParam, ikuai.ErrInvalidParam) {
		t.Fatal("哨兵不可用")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
