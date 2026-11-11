package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestVPNServersGet 覆盖五个服务端 show TYPE=data 的成功路径与参数断言。
func TestVPNServersGet(t *testing.T) {
	fixtures := []struct {
		funcName string
		data     map[string]any
		assert   func(t *testing.T, s *VPNServersService)
	}{
		{
			funcName: "pppoe_server",
			data: map[string]any{
				"id": 1, "enabled": "no", "server_name": "iKuai", "server_ip": "10.1.1.1",
				"interface": "lan1", "addr_pool": "10.1.1.2-10.1.1.254",
				"authmode": 0, "nas_identifier": "iKuai", "authport": 1812,
				"accountport": 1813, "mtu": 1480, "mru": 1480,
				"lcp_echo_interval": 10, "lcp_echo_failure": 3, "restart_week": 1234567,
			},
			assert: func(t *testing.T, s *VPNServersService) {
				cfg, err := s.PPPoE(context.Background())
				if err != nil {
					t.Fatalf("PPPoE 失败: %v", err)
				}
				if cfg == nil || cfg.ID != 1 || cfg.IsEnabled() ||
					cfg.ServerIP != "10.1.1.1" || cfg.Authmode != AuthmodeLocalAccount ||
					cfg.MTU != 1480 || cfg.LCPEchoInterval != 10 ||
					cfg.RestartWeek != 1234567 {
					t.Fatalf("PPPoE 字段解码不符: %+v", cfg)
				}
			},
		},
		{
			funcName: "pptp_server",
			data: map[string]any{
				"id": 1, "enabled": "yes", "server_ip": "10.0.0.1", "server_port": 1723,
				"addr_pool": "10.0.0.2-10.0.0.254", "open_mppe": 2, "mtu": 1400, "mru": 1400,
			},
			assert: func(t *testing.T, s *VPNServersService) {
				cfg, err := s.PPTP(context.Background())
				if err != nil {
					t.Fatalf("PPTP 失败: %v", err)
				}
				if cfg == nil || !cfg.IsEnabled() || cfg.ServerPort != 1723 ||
					cfg.OpenMPPE != 2 || cfg.MTU != 1400 {
					t.Fatalf("PPTP 字段解码不符: %+v", cfg)
				}
			},
		},
		{
			funcName: "l2tp_server",
			data: map[string]any{
				"id": 1, "enabled": "yes", "server_ip": "10.1.0.1", "server_port": 1701,
				"force_ipsec": 0, "mtu": 1400, "mru": 1400,
			},
			assert: func(t *testing.T, s *VPNServersService) {
				cfg, err := s.L2TP(context.Background())
				if err != nil {
					t.Fatalf("L2TP 失败: %v", err)
				}
				if cfg == nil || !cfg.IsEnabled() || cfg.ServerPort != 1701 ||
					cfg.ForceIPsec != 0 || cfg.MTU != 1400 {
					t.Fatalf("L2TP 字段解码不符: %+v", cfg)
				}
			},
		},
		{
			funcName: "openvpn-server",
			data: map[string]any{
				"port": "1194", "subnet": "10.7.7.0", "mask": "255.255.255.0",
				"tun_mtu": "1400", "proto": "udp", "cipher": "BF-CBC",
				"comp_lzo": "1", "dev_type": "tun", "topology": "subnet",
				"method": 0, "push_gateway": "0", "enabled": "no",
			},
			assert: func(t *testing.T, s *VPNServersService) {
				cfg, err := s.OpenVPN(context.Background())
				if err != nil {
					t.Fatalf("OpenVPN 失败: %v", err)
				}
				if cfg == nil || cfg.IsEnabled() || cfg.Port != "1194" ||
					cfg.Subnet != "10.7.7.0" || cfg.TunMTU != "1400" ||
					cfg.CompLZO != "1" || cfg.Topology != "subnet" {
					t.Fatalf("OpenVPN 字段解码不符: %+v", cfg)
				}
			},
		},
		{
			funcName: "ike_server",
			data:     map[string]any{"id": 1, "enabled": "no"},
			assert: func(t *testing.T, s *VPNServersService) {
				row, err := s.IKEv2(context.Background())
				if err != nil {
					t.Fatalf("IKEv2 失败: %v", err)
				}
				if row == nil {
					t.Fatal("IKEv2 行不应为 nil")
				}
				m, err := row.Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["id"] != float64(1) {
					t.Fatalf("行内容不符: %v", m)
				}
			},
		},
	}
	for _, fx := range fixtures {
		t.Run(fx.funcName, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse(fx.funcName, "show", map[string]any{"data": []any{fx.data}})
			svc := New(ts.Client()).VPNServers
			fx.assert(t, svc)
			rec := ts.CallsFor(fx.funcName)[0]
			if rec.Action != "show" {
				t.Fatalf("action = %q, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), map[string]any{"TYPE": cfgTypeData})
		})
	}
}

// TestVPNServersSet 覆盖 save 写动作的参数断言（仅离线）。
func TestVPNServersSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"pppoe_server", "pptp_server", "l2tp_server", "openvpn-server"} {
		ts.SetResponse(fn, "save", map[string]any{})
	}
	svc := New(ts.Client()).VPNServers
	ctx := context.Background()
	if err := svc.SetPPPoE(ctx, PPPoEServerConfig{ID: 1, Enabled: EnabledYes, ServerIP: "10.1.1.1"}); err != nil {
		t.Fatalf("SetPPPoE 失败: %v", err)
	}
	if err := svc.SetPPTP(ctx, PPTPServerConfig{ID: 1, Enabled: EnabledYes}); err != nil {
		t.Fatalf("SetPPTP 失败: %v", err)
	}
	if err := svc.SetL2TP(ctx, L2TPServerConfig{ID: 1, Enabled: EnabledNo}); err != nil {
		t.Fatalf("SetL2TP 失败: %v", err)
	}
	if err := svc.SetOpenVPN(ctx, OpenVPNServerConfig{Port: "1195", Enabled: EnabledYes}); err != nil {
		t.Fatalf("SetOpenVPN 失败: %v", err)
	}
	for _, fn := range []string{"pppoe_server", "pptp_server", "l2tp_server", "openvpn-server"} {
		recs := ts.CallsFor(fn)
		if len(recs) != 1 || recs[0].Action != "save" {
			t.Fatalf("%s 应有一次 save 调用, got %+v", fn, recs)
		}
		p := paramOf(t, recs[0])
		// openvpn-server 行无 id 字段（真机实证）；其余三个 save 随行携带 id。
		if fn == "openvpn-server" {
			if _, ok := p["id"]; ok {
				t.Fatalf("openvpn-server save 不应携带 id: %v", p)
			}
			if p["enabled"] != EnabledYes {
				t.Fatalf("openvpn-server save param 字段不符: %v", p)
			}
			continue
		}
		if p["id"] != float64(1) {
			t.Fatalf("%s save param 缺少 id: %v", fn, p)
		}
	}
}

// TestVPNServersError 覆盖固件错误路径（含 ike_server 未授权 30002 形态）。
func TestVPNServersError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ike_server", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname(ike_server)", FuncName: "ike_server", Action: action}
	})
	_, err := New(ts.Client()).VPNServers.IKEv2(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
}

// TestDecodePEMField 覆盖固件证书编码还原（真机实证 @/# 规则）。
func TestDecodePEMField(t *testing.T) {
	if got := DecodePEMField(""); got != "" {
		t.Fatalf("空串应返回空串, got %q", got)
	}
	got := DecodePEMField("-----BEGIN#X-----@abc#def@-----END#X-----")
	want := "-----BEGIN X-----\nabc def\n-----END X-----"
	if got != want {
		t.Fatalf("DecodePEMField = %q, 期望 %q", got, want)
	}
}
