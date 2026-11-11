package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// webauthFixture 是真机 webauth.show TYPE=data 行的字段子集（真机实证）。
func webauthFixture() map[string]any {
	return map[string]any{
		"id": 1, "enabled": "no", "max_time": 0, "idle_time": 60,
		"weixin": 1, "interface": "lan1", "group_key": "testing123",
		"nasname": "iKuai", "radius_ip": "127.0.0.1", "radius_key": "123",
		"radius_authport": 1812, "radius_accountport": 1813, "radius_interval": 60,
		"tryout_time": 30, "popups": 1,
	}
}

// TestWebAuthGet 覆盖 webauth.show TYPE=data 的成功路径与参数断言。
func TestWebAuthGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webauth", "show", map[string]any{"data": []any{webauthFixture()}})
	cfg, err := New(ts.Client()).WebAuth.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("配置不应为 nil")
	}
	if cfg.ID != 1 || cfg.IsEnabled() || cfg.IdleTime != 60 || cfg.Weixin != 1 ||
		cfg.Interface != "lan1" || cfg.GroupKey != "testing123" ||
		cfg.NasName != "iKuai" || cfg.RadiusAuthport != 1812 ||
		cfg.RadiusAccountport != 1813 || cfg.RadiusInterval != 60 ||
		cfg.TryoutTime != 30 || cfg.Popups != 1 {
		t.Fatalf("字段解码不符: %+v", cfg)
	}
	rec := ts.CallsFor("webauth")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": WebAuthTypeData})
}

// TestWebAuthGetEmpty 覆盖配置未落库（data 为空）路径。
func TestWebAuthGetEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webauth", "show", map[string]any{"data": []any{}})
	cfg, err := New(ts.Client()).WebAuth.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg != nil {
		t.Fatalf("空配置应返回 nil, got %+v", cfg)
	}
}

// TestWebAuthSet 覆盖 webauth.save 的参数断言（写动作，仅离线）。
func TestWebAuthSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webauth", "save", map[string]any{})
	cfg := WebAuthConfig{ID: 1, Enabled: EnabledYes, IdleTime: 120, Interface: "lan1", Weixin: 1}
	if err := New(ts.Client()).WebAuth.Set(context.Background(), cfg); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("webauth")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"] != float64(1) || p["enabled"] != EnabledYes || p["idle_time"] != float64(120) {
		t.Fatalf("save param 字段不符: %v", p)
	}
}

// TestWebAuthError 覆盖固件错误路径。
func TestWebAuthError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("webauth", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total,data)", FuncName: "webauth", Action: action}
	})
	if _, err := New(ts.Client()).WebAuth.Get(context.Background()); err == nil {
		t.Fatal("Get 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := New(ts.Client()).WebAuth.Set(context.Background(), WebAuthConfig{ID: 1}); err == nil {
		t.Fatal("Set 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
