package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestSelfPasswdGet 覆盖 ppp_passwd.show TYPE=data 的成功路径与参数断言。
func TestSelfPasswdGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ppp_passwd", "show", map[string]any{"data": []any{map[string]any{
		"enabled": "no", "allow_web": 1, "allow_pppoe": 1, "allow_pptp": 1,
		"allow_l2tp": 1, "allow_ovpn": 1, "allow_quit": 1, "id": 1,
	}}})
	cfg, err := New(ts.Client()).SelfPasswd.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.IsEnabled() || cfg.AllowWeb != 1 ||
		cfg.AllowPPPoE != 1 || cfg.AllowPPTP != 1 || cfg.AllowL2TP != 1 ||
		cfg.AllowOVPN != 1 || cfg.AllowQuit != 1 {
		t.Fatalf("字段解码不符: %+v", cfg)
	}
	rec := ts.CallsFor("ppp_passwd")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": cfgTypeData})
}

// TestSelfPasswdGetEmpty 覆盖配置未落库路径。
func TestSelfPasswdGetEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ppp_passwd", "show", map[string]any{"data": []any{}})
	cfg, err := New(ts.Client()).SelfPasswd.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg != nil {
		t.Fatalf("空配置应返回 nil, got %+v", cfg)
	}
}

// TestSelfPasswdSet 覆盖 save 写动作的参数断言（仅离线）。
func TestSelfPasswdSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ppp_passwd", "save", map[string]any{})
	cfg := SelfPasswdConfig{ID: 1, Enabled: EnabledYes, AllowWeb: 1, AllowQuit: 0}
	if err := New(ts.Client()).SelfPasswd.Set(context.Background(), cfg); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("ppp_passwd")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"] != float64(1) || p["enabled"] != EnabledYes ||
		p["allow_web"] != float64(1) || p["allow_quit"] != float64(0) {
		t.Fatalf("save param 字段不符: %v", p)
	}
}

// TestSelfPasswdError 覆盖固件错误路径。
func TestSelfPasswdError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ppp_passwd", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (x)", FuncName: "ppp_passwd", Action: action}
	})
	if _, err := New(ts.Client()).SelfPasswd.Get(context.Background()); err == nil {
		t.Fatal("Get 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := New(ts.Client()).SelfPasswd.Set(context.Background(), SelfPasswdConfig{ID: 1}); err == nil {
		t.Fatal("Set 应失败")
	}
}
