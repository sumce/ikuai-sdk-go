package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// sysctlSample 构造真机探测形态的 ik_sysctl.show 响应 Data（TYPE=data）。
func sysctlSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"close_wait_timeout": 10, "time_wait_timeout": 10, "close_timeout": 5,
			"udp_timeout": 10, "udp_stream_timeout": 60, "icmp_timeout": 10,
			"id": 1, "bbr": 0, "syn_recv_timeout": 5, "syn_send_timeout": 5,
			"established_timeout": 1800, "fin_wait_timeout": 10, "last_ack_timeout": 10,
		}},
	}
}

// TestSysctlGet 覆盖 ik_sysctl.show 的成功路径与参数断言。
func TestSysctlGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_sysctl", "show", sysctlSample())
	cfg, err := New(ts.Client()).Sysctl.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.BBR != 0 || cfg.EstablishedTimeout != 1800 ||
		cfg.UDPStreamTimeout != 60 || cfg.SynRecvTimeout != 5 ||
		cfg.TimeWaitTimeout != 10 || cfg.ICMPTimeout != 10 {
		t.Fatalf("内核参数行解码不符: %+v", cfg)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ik_sysctl")[0]),
		map[string]any{"TYPE": SysctlTypeData})
}

// TestSysctlList 覆盖 List 的 TYPE 归一参数。
func TestSysctlList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_sysctl", "show", sysctlSample())
	page, err := New(ts.Client()).Sysctl.List(context.Background(), &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(page.Rows) != 1 || page.Rows[0].ID != 1 {
		t.Fatalf("行解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ik_sysctl")[0]),
		map[string]any{"TYPE": SysctlTypeData, "limit": "0,5"})
}

// TestSysctlSave 覆盖 ik_sysctl.save（写动作，离线）。
func TestSysctlSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_sysctl", "save", map[string]any{})
	cfg := SysctlConfig{ID: 1, BBR: 1, EstablishedTimeout: 3600}
	if err := New(ts.Client()).Sysctl.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("ik_sysctl")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"].(float64) != 1 || p["bbr"].(float64) != 1 || p["established_timeout"].(float64) != 3600 {
		t.Fatalf("save 参数不符: %v", p)
	}
}

// TestSysctlError 覆盖固件错误路径。
func TestSysctlError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ik_sysctl", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "ik_sysctl", Action: action}
	})
	if _, err := New(ts.Client()).Sysctl.Get(context.Background()); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := New(ts.Client()).Sysctl.Save(context.Background(), SysctlConfig{ID: 1}); err == nil {
		t.Fatal("save 错误应上抛")
	}
}
