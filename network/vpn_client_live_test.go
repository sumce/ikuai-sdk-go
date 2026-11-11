package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveVPNClient 真机只读验证 VPN 客户端族 8 个 func 的 show 类方法。
// wireguard/ike_client 在真机 3.7.24 无模块（30002 Not found funcname），
// 该错误形态视为通过。
func TestLiveVPNClient(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).VPNClient

	// pptp_client / l2tp_client / openvpn-client / ipsec-vpn：真机实证空表。
	type listCall struct {
		name string
		fn   func() (total int64, err error)
	}
	lists := []listCall{
		{"pptp_client", func() (int64, error) {
			p, err := svc.PPTPList(ctx, nil)
			if err != nil {
				return 0, err
			}
			return p.Total, nil
		}},
		{"l2tp_client", func() (int64, error) {
			p, err := svc.L2TPList(ctx, nil)
			if err != nil {
				return 0, err
			}
			return p.Total, nil
		}},
		{"openvpn-client", func() (int64, error) {
			p, err := svc.OpenVPNList(ctx, nil)
			if err != nil {
				return 0, err
			}
			return p.Total, nil
		}},
		{"ipsec-vpn", func() (int64, error) {
			p, err := svc.IPSecList(ctx, nil)
			if err != nil {
				return 0, err
			}
			return p.Total, nil
		}},
	}
	for _, lc := range lists {
		total, err := lc.fn()
		if err != nil {
			t.Fatalf("%s.show 失败: %v", lc.name, err)
		}
		t.Logf("%s total=%d", lc.name, total)
	}

	// wireguard / ike_client：真机无模块回 30002，记录错误形态照常通过。
	for _, tc := range []struct {
		name string
		fn   func() (int64, error)
	}{
		{"wireguard", func() (int64, error) {
			p, err := svc.WireGuardList(ctx, nil)
			if err != nil {
				return 0, err
			}
			return p.Total, nil
		}},
		{"ike_client", func() (int64, error) {
			p, err := svc.IKEList(ctx, nil)
			if err != nil {
				return 0, err
			}
			return p.Total, nil
		}},
	} {
		total, err := tc.fn()
		if err != nil {
			ae, ok := ikuai.AsAPIError(err)
			if !ok || ae.Code != 30002 {
				t.Fatalf("%s.show 应为 30002 Not found funcname 或成功, got %v", tc.name, err)
			}
			t.Logf("%s: 真机无该模块 (Result=30002)", tc.name)
			continue
		}
		t.Logf("%s total=%d", tc.name, total)
	}

	// ik_web_sdwan.show（真机实证状态对象）。
	st, err := svc.SDWANStatus(ctx)
	if err != nil {
		t.Fatalf("ik_web_sdwan.show 失败: %v", err)
	}
	t.Logf("SD-WAN 状态: code=%s message=%s", st.Code, st.Message)

	// wuying.show（真机实证单行配置 + routeid）。
	wy, err := svc.WuyingGet(ctx)
	if err != nil {
		t.Fatalf("wuying.show 失败: %v", err)
	}
	t.Logf("无影 routeid=%s rows=%d", wy.Routeid, len(wy.Rows))
	if len(wy.Rows) > 0 {
		r := wy.Rows[0]
		t.Logf("无影配置: enabled=%s wan=%s lan=%s", r.Enabled, r.WANAddr, r.LANAddr)
	}
}
