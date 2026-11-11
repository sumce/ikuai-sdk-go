package security

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// firewallSample 构造 firewall.show TYPE=all 的响应 Data（真机实测形态）。
func firewallSample() map[string]any {
	return map[string]any{"cpu_cores": 20, "status": 0}
}

// TestFirewallGet 覆盖 firewall.show TYPE=all 的成功路径与参数断言。
func TestFirewallGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("firewall", "show", firewallSample())
	st, err := New(ts.Client()).Firewall.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if st == nil || st.CPUCores != 20 || st.Status != 0 {
		t.Fatalf("状态解码不符: %+v", st)
	}
	wantParam(t, paramOf(t, ts.CallsFor("firewall")[0]), map[string]any{"TYPE": FirewallTypeAll})
}

// TestFirewallStatus 覆盖 firewall.show TYPE=status 的成功路径与参数断言。
func TestFirewallStatus(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("firewall", "show", map[string]any{"status": 0})
	n, err := New(ts.Client()).Firewall.Status(context.Background())
	if err != nil {
		t.Fatalf("Status 失败: %v", err)
	}
	if n != 0 {
		t.Fatalf("状态值不符: %d", n)
	}
	wantParam(t, paramOf(t, ts.CallsFor("firewall")[0]), map[string]any{"TYPE": FirewallTypeStatus})
}

// TestFirewallSave 覆盖 firewall.save（写动作，仅离线；action/字段待真机验证）。
func TestFirewallSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("firewall", "save", map[string]any{})
	st := FirewallStatus{CPUCores: 20, Status: 0}
	if err := New(ts.Client()).Firewall.Save(context.Background(), st); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("firewall")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["cpu_cores"].(float64) != 20 || p["status"].(float64) != 0 {
		t.Fatalf("save 参数不符: %v", p)
	}
}

// TestFirewallError 覆盖固件错误路径（真机实测未知 TYPE 回 30006）。
func TestFirewallError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("firewall", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (data)", FuncName: "firewall", Action: action}
	})
	svc := New(ts.Client()).Firewall
	if _, err := svc.Get(context.Background()); err == nil {
		t.Fatal("Get 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.Status(context.Background()); err == nil {
		t.Fatal("Status 应失败")
	}
	if err := svc.Save(context.Background(), FirewallStatus{}); err == nil {
		t.Fatal("Save 应失败")
	}
}
