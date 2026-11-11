package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestVPNClientPPTP 覆盖 pptp_client 全方法族（行字段待真机验证，
// 离线用例仅保证解码通路与 func/action 断言）。
func TestVPNClientPPTP(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, a := range []string{"show", "add", "edit", "del"} {
		ts.SetResponse("pptp_client", a, map[string]any{})
	}
	// show 样本字段待真机验证，仅保证解码通路。
	ts.SetResponse("pptp_client", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "interface": "adsl01",
			"server_addr": "vpn.example.com", "username": "u", "password": "p", "comment": "",
		}},
	})
	svc := New(ts.Client()).VPNClient
	ctx := context.Background()

	page, err := svc.PPTPList(ctx, nil)
	if err != nil {
		t.Fatalf("PPTPList 失败: %v", err)
	}
	if page.Total != 1 || page.Rows[0].ServerAddr != "vpn.example.com" || !page.Rows[0].IsEnabled() {
		t.Fatalf("PPTP 行解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("pptp_client")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	if _, err := svc.PPTPAdd(ctx, PPTPClient{Enabled: EnabledYes, Interface: "adsl01", ServerAddr: "vpn.example.com"}); err != nil {
		t.Fatalf("PPTPAdd 失败: %v", err)
	}
	if err := svc.PPTPEdit(ctx, PPTPClient{ID: 1}); err != nil {
		t.Fatalf("PPTPEdit 失败: %v", err)
	}
	if err := svc.PPTPDelete(ctx, 1); err != nil {
		t.Fatalf("PPTPDelete 失败: %v", err)
	}
	assertActionSeq(t, ts.CallsFor("pptp_client"), "show", "add", "edit", "del")
}

// TestVPNClientL2TPOpenVPNIPSec 覆盖 l2tp/openvpn/ipsec 三族的方法通路。
func TestVPNClientL2TPOpenVPNIPSec(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"l2tp_client", "openvpn-client", "ipsec-vpn"} {
		for _, a := range []string{"show", "add", "edit", "del"} {
			ts.SetResponse(fn, a, map[string]any{})
		}
		ts.SetResponse(fn, "show", map[string]any{"total": 0, "data": []any{}})
	}
	svc := New(ts.Client()).VPNClient
	ctx := context.Background()

	if _, err := svc.L2TPList(ctx, nil); err != nil {
		t.Fatalf("L2TPList 失败: %v", err)
	}
	if _, err := svc.L2TPAdd(ctx, L2TPClient{Enabled: EnabledYes, ServerAddr: "l2tp.example.com"}); err != nil {
		t.Fatalf("L2TPAdd 失败: %v", err)
	}
	if err := svc.L2TPEdit(ctx, L2TPClient{ID: 1}); err != nil {
		t.Fatalf("L2TPEdit 失败: %v", err)
	}
	if err := svc.L2TPDelete(ctx, 1); err != nil {
		t.Fatalf("L2TPDelete 失败: %v", err)
	}

	if _, err := svc.OpenVPNList(ctx, nil); err != nil {
		t.Fatalf("OpenVPNList 失败: %v", err)
	}
	if _, err := svc.OpenVPNAdd(ctx, OpenVPNClient{Enabled: EnabledYes, ServerAddr: "ovpn.example.com", ServerPort: 1194}); err != nil {
		t.Fatalf("OpenVPNAdd 失败: %v", err)
	}
	if err := svc.OpenVPNEdit(ctx, OpenVPNClient{ID: 1}); err != nil {
		t.Fatalf("OpenVPNEdit 失败: %v", err)
	}
	if err := svc.OpenVPNDelete(ctx, 1); err != nil {
		t.Fatalf("OpenVPNDelete 失败: %v", err)
	}

	if _, err := svc.IPSecList(ctx, nil); err != nil {
		t.Fatalf("IPSecList 失败: %v", err)
	}
	if _, err := svc.IPSecAdd(ctx, IPSecClient{Enabled: EnabledYes, ServerAddr: "10.0.0.1"}); err != nil {
		t.Fatalf("IPSecAdd 失败: %v", err)
	}
	if err := svc.IPSecEdit(ctx, IPSecClient{ID: 1}); err != nil {
		t.Fatalf("IPSecEdit 失败: %v", err)
	}
	if err := svc.IPSecDelete(ctx, 1); err != nil {
		t.Fatalf("IPSecDelete 失败: %v", err)
	}

	for _, fn := range []string{"l2tp_client", "openvpn-client", "ipsec-vpn"} {
		assertActionSeq(t, ts.CallsFor(fn), "show", "add", "edit", "del")
	}
}

// TestVPNClientWireGuardIKE 覆盖 wireguard/ike_client 的方法通路
// （真机无该模块，show 回 30002，由 live test 记录错误形态）。
func TestVPNClientWireGuardIKE(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"wireguard", "ike_client"} {
		for _, a := range []string{"show", "add", "edit", "del"} {
			ts.SetResponse(fn, a, map[string]any{})
		}
		ts.SetResponse(fn, "show", map[string]any{"total": 0, "data": []any{}})
	}
	svc := New(ts.Client()).VPNClient
	ctx := context.Background()

	if _, err := svc.WireGuardList(ctx, nil); err != nil {
		t.Fatalf("WireGuardList 失败: %v", err)
	}
	if _, err := svc.WireGuardAdd(ctx, WireGuardClient{Enabled: EnabledYes, ServerAddr: "wg.example.com", ServerPort: 51820}); err != nil {
		t.Fatalf("WireGuardAdd 失败: %v", err)
	}
	if err := svc.WireGuardEdit(ctx, WireGuardClient{ID: 1}); err != nil {
		t.Fatalf("WireGuardEdit 失败: %v", err)
	}
	if err := svc.WireGuardDelete(ctx, 1); err != nil {
		t.Fatalf("WireGuardDelete 失败: %v", err)
	}

	if _, err := svc.IKEList(ctx, nil); err != nil {
		t.Fatalf("IKEList 失败: %v", err)
	}
	if _, err := svc.IKEAdd(ctx, IKEClient{Enabled: EnabledYes, ServerAddr: "ike.example.com"}); err != nil {
		t.Fatalf("IKEAdd 失败: %v", err)
	}
	if err := svc.IKEEdit(ctx, IKEClient{ID: 1}); err != nil {
		t.Fatalf("IKEEdit 失败: %v", err)
	}
	if err := svc.IKEDelete(ctx, 1); err != nil {
		t.Fatalf("IKEDelete 失败: %v", err)
	}

	for _, fn := range []string{"wireguard", "ike_client"} {
		assertActionSeq(t, ts.CallsFor(fn), "show", "add", "edit", "del")
	}
}

// TestVPNClientSDWANWuying 覆盖 ik_web_sdwan/wuying（真机实证形态）。
func TestVPNClientSDWANWuying(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_web_sdwan", "show", map[string]any{
		"data": map[string]any{"code": "2020", "message": "connect server failed"},
	})
	ts.SetResponse("wuying", "show", map[string]any{
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "tocloud": 1,
			"wan_addr": "10.30.20.0/24", "lan_addr": "10.31.20.0/24",
			"bind_wan_addr": "10.32.20.0/24", "bind_lan_addr": "10.33.20.0/24",
			"interface": "", "veth_mac": "", "node": 0, "status": 0,
			"message": "", "timestamp": 0, "api_token": "", "libver": "", "comment": "",
		}},
		"routeid": "d072dadeaacc4feb98a22dd0878b867a",
	})
	ts.SetResponse("wuying", "set", map[string]any{})
	svc := New(ts.Client()).VPNClient
	ctx := context.Background()

	st, err := svc.SDWANStatus(ctx)
	if err != nil {
		t.Fatalf("SDWANStatus 失败: %v", err)
	}
	if st.Code != "2020" || st.Message != "connect server failed" {
		t.Fatalf("SD-WAN 状态解码不符: %+v", st)
	}
	// 真机实证：ik_web_sdwan.show 不带 TYPE/limit。
	wantParam(t, paramOf(t, ts.CallsFor("ik_web_sdwan")[0]), map[string]any{})

	page, err := svc.WuyingGet(ctx)
	if err != nil {
		t.Fatalf("WuyingGet 失败: %v", err)
	}
	if page.Routeid != "d072dadeaacc4feb98a22dd0878b867a" || len(page.Rows) != 1 ||
		page.Rows[0].WANAddr != "10.30.20.0/24" || page.Rows[0].ToCloud != 1 ||
		!page.Rows[0].IsEnabled() {
		t.Fatalf("无影配置解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("wuying")[0]), map[string]any{})

	if err := svc.WuyingSet(ctx, WuyingConfig{ID: 1, Enabled: EnabledYes}); err != nil {
		t.Fatalf("WuyingSet 失败: %v", err)
	}
	if rec := ts.CallsFor("wuying")[1]; rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
}

// TestVPNClientError 覆盖固件错误路径。
func TestVPNClientError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	// 模拟 wireguard 模块缺失的真机错误形态。
	ts.HandleFuncName("wireguard", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname(wireguard)", FuncName: "wireguard", Action: action}
	})
	_, err := New(ts.Client()).VPNClient.WireGuardList(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
}

// assertActionSeq 断言 func 的调用动作序列。
func assertActionSeq(t *testing.T, recs []testserver.CallRecord, want ...string) {
	t.Helper()
	if len(recs) != len(want) {
		t.Fatalf("应有 %d 次调用, got %d (%v)", len(want), len(recs), recs)
	}
	for i, a := range want {
		if recs[i].Action != a {
			t.Fatalf("第 %d 次调用 action = %q, 期望 %q", i+1, recs[i].Action, a)
		}
	}
}
