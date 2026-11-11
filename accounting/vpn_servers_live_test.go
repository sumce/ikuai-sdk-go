package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveVPNServers 真机只读验证五个 VPN 服务端配置（*_server.show
// TYPE=data）。写动作（save）禁止对真机调用。
func TestLiveVPNServers(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).VPNServers

	pppoe, err := svc.PPPoE(ctx)
	if err != nil {
		t.Fatalf("pppoe_server.show 失败: %v", err)
	}
	if pppoe != nil {
		t.Logf("PPPoE: enabled=%s server_ip=%s interface=%s authmode=%d mtu=%d",
			pppoe.Enabled, pppoe.ServerIP, pppoe.Interface, pppoe.Authmode, pppoe.MTU)
	} else {
		t.Log("PPPoE 配置未落库")
	}

	pptp, err := svc.PPTP(ctx)
	if err != nil {
		t.Fatalf("pptp_server.show 失败: %v", err)
	}
	if pptp != nil {
		t.Logf("PPTP: enabled=%s server_ip=%s port=%d mppe=%d",
			pptp.Enabled, pptp.ServerIP, pptp.ServerPort, pptp.OpenMPPE)
	} else {
		t.Log("PPTP 配置未落库")
	}

	l2tp, err := svc.L2TP(ctx)
	if err != nil {
		t.Fatalf("l2tp_server.show 失败: %v", err)
	}
	if l2tp != nil {
		t.Logf("L2TP: enabled=%s server_ip=%s port=%d force_ipsec=%d",
			l2tp.Enabled, l2tp.ServerIP, l2tp.ServerPort, l2tp.ForceIPsec)
	} else {
		t.Log("L2TP 配置未落库")
	}

	ovpn, err := svc.OpenVPN(ctx)
	if err != nil {
		t.Fatalf("openvpn-server.show 失败: %v", err)
	}
	if ovpn != nil {
		t.Logf("OpenVPN: enabled=%s port=%s subnet=%s proto=%s cipher=%s topology=%s",
			ovpn.Enabled, ovpn.Port, ovpn.Subnet, ovpn.Proto, ovpn.Cipher, ovpn.Topology)
	} else {
		t.Log("OpenVPN 配置未落库")
	}
}

// TestLiveIKEv2 真机只读验证 IPSec IKEv2 服务端（ike_server.show）。
// 真机 3.7.24 实证：未授权时返回 30002 Not found funcname(ike_server)，
// 本测试对两种形态（成功/30002）均视为通过并记录实际行为。
func TestLiveIKEv2(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).VPNServers

	row, err := svc.IKEv2(ctx)
	if err != nil {
		ae, ok := ikuai.AsAPIError(err)
		if !ok {
			t.Fatalf("ike_server.show 非预期错误: %v", err)
		}
		if ae.Code == 30002 {
			t.Logf("真机行为（已记录）: ike_server 不可用: %v", err)
			return
		}
		t.Fatalf("ike_server.show 非预期结果码: %v", err)
	}
	if row == nil {
		t.Log("IKEv2 配置未落库")
		return
	}
	m, err := row.Map()
	if err != nil {
		t.Fatalf("解码 ike_server 行失败: %v", err)
	}
	t.Logf("IKEv2 行: %v", m)
}
