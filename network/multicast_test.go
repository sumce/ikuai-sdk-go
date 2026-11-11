package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestIGMPProxy 覆盖 igmp_proxy 的 show（行形态真机实证）与 set（仅离线）。
func TestIGMPProxy(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("igmp_proxy", "show", map[string]any{
		"data": []any{map[string]any{
			"id": 1, "enabled": "no", "downstream": "lan4,vlan3",
			"version": "3", "upstream": "adsl07",
		}},
	})
	ts.SetResponse("igmp_proxy", "set", map[string]any{})
	svc := New(ts.Client()).Multicast
	ctx := context.Background()

	cfg, err := svc.IGMPGet(ctx)
	if err != nil {
		t.Fatalf("IGMPGet 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.Upstream != "adsl07" ||
		cfg.Version != "3" || cfg.IsEnabled() {
		t.Fatalf("IGMP 配置解码不符: %+v", cfg)
	}
	if ds := cfg.Downstreams(); len(ds) != 2 || ds[0] != "lan4" || ds[1] != "vlan3" {
		t.Fatalf("Downstreams 解析不符: %v", ds)
	}
	wantParam(t, paramOf(t, ts.CallsFor("igmp_proxy")[0]),
		map[string]any{"TYPE": "total,data"})

	if err := svc.IGMPSet(ctx, IGMPProxy{ID: 1, Enabled: EnabledYes, Upstream: "adsl07", Version: "3"}); err != nil {
		t.Fatalf("IGMPSet 失败: %v", err)
	}
	rec := ts.CallsFor("igmp_proxy")[1]
	if rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
	if p := paramOf(t, rec); p["upstream"] != "adsl07" {
		t.Fatalf("set 参数不符: %v", p)
	}
}

// TestIPTV 覆盖 iptv 的 show（空参数，真机实证）与 set（仅离线）。
func TestIPTV(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("iptv", "show", map[string]any{
		"data": []any{map[string]any{
			"mode": 0, "wan_iface": "eth0", "wan_vlanid": 51,
			"lan_iface": "lan3", "lan_vlanid": 0, "id": 1, "enabled": "no",
		}},
	})
	ts.SetResponse("iptv", "set", map[string]any{})
	svc := New(ts.Client()).Multicast
	ctx := context.Background()

	cfg, err := svc.IPTVGet(ctx)
	if err != nil {
		t.Fatalf("IPTVGet 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.WANIface != "eth0" ||
		cfg.WANVLANID != 51 || cfg.LANIface != "lan3" || cfg.IsEnabled() {
		t.Fatalf("IPTV 配置解码不符: %+v", cfg)
	}
	// 真机实证：iptv.show 不带 TYPE/limit（TYPE=total,data 回 30006）。
	wantParam(t, paramOf(t, ts.CallsFor("iptv")[0]), map[string]any{})

	if err := svc.IPTVSet(ctx, IPTVConfig{ID: 1, Enabled: EnabledYes, Mode: 0}); err != nil {
		t.Fatalf("IPTVSet 失败: %v", err)
	}
	rec := ts.CallsFor("iptv")[1]
	if rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
}

// TestMulticastError 覆盖固件错误路径（模拟 iptv 传入 TYPE 的真机错误形态）。
func TestMulticastError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("iptv", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "iptv", Action: action}
	})
	_, err := New(ts.Client()).Multicast.IPTVGet(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
