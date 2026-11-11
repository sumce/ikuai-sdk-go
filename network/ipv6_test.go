package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// ipv6Sample 构造真机形态的 ipv6.show 响应 Data。
func ipv6Sample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"link_addr":        "fe80::c9e7:5619:ed0a:cf15",
			"dhcp6_ip_gateway": "fe80::6a8f:84ff:fef3:3056",
			"ipv6_addr":        "", "dhcp6_ip_addr": "2408:8220:302:b440:792b:a629:3bd2:e5ce/64",
			"dhcp6_dns1": "2408:8888::8", "preferred_lft": "96184",
			"id": 1, "enabled": "yes", "dhcp6_dns2": "2408:8899::8",
			"interface": "adsl07", "internet": "dhcp",
			"dhcp6_prefix1": "2408:8220:320:88e0::/60",
			"prefix":        "auto", "prefix_hint": "", "force_prefix": 0,
			"force_gen_duid": 1, "valid_lft": "182584", "ipv6_gateway": "",
		}},
	}
}

// TestIPv6ListSet 覆盖 ipv6 的 show 与 set（写动作仅离线；行形态真机实证）。
func TestIPv6ListSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ipv6", "show", ipv6Sample())
	ts.SetResponse("ipv6", "set", map[string]any{})
	svc := New(ts.Client()).IPv6
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	r := page.Rows[0]
	if page.Total != 1 || r.Interface != "adsl07" || r.Internet != "dhcp" ||
		r.Prefix != "auto" || r.ForceGenDUID != 1 || r.DHCP6DNS1 != "2408:8888::8" ||
		r.PreferredLft != "96184" || !r.IsEnabled() {
		t.Fatalf("IPv6 配置解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ipv6")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	if err := svc.Set(ctx, IPv6Config{ID: 1, Enabled: EnabledYes, Internet: "dhcp", Prefix: "auto"}); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("ipv6")[1]
	if rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
	if p := paramOf(t, rec); p["internet"] != "dhcp" {
		t.Fatalf("set 参数不符: %v", p)
	}
}

// TestIPv6StreamList 覆盖 ipv6_stream.show（运行状态，行形态真机实证）。
func TestIPv6StreamList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ipv6_stream", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"total_download": "7734675835797", "upload": 1012689,
			"interface": "adsl07", "conn": 2059, "id": 1,
			"download": 13979212, "total_upload": "994198463506",
		}},
	})
	page, err := New(ts.Client()).IPv6.StreamList(context.Background(), nil)
	if err != nil {
		t.Fatalf("StreamList 失败: %v", err)
	}
	r := page.Rows[0]
	if page.Total != 1 || r.Interface != "adsl07" || r.Conn != 2059 ||
		r.Download != 13979212 || r.TotalDownload != "7734675835797" {
		t.Fatalf("IPv6 分流解码不符: %+v", r)
	}
	rec := ts.CallsFor("ipv6_stream")[0]
	if rec.FuncName != "ipv6_stream" || rec.Action != "show" {
		t.Fatalf("调用记录不符: %s.%s", rec.FuncName, rec.Action)
	}
}

// TestDHCP6LeaseList 覆盖 dhcp6_lease.show TYPE=all（真机实证形态）。
func TestDHCP6LeaseList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dhcp6_lease", "show", map[string]any{
		"client_total": 1,
		"client_data": []any{map[string]any{
			"start_time": 1794331084, "id": 1272, "interface": "lan4",
			"link_addr": "fe80::18dc:c29:5f2b:68bd",
			"duid":      "00010001322b3bc1d8bbc19a8c8f",
			"ipv6_addr": "2408:8220:320:88e2::ff3 ", "mac": "4c:10:d5:42:05:b6",
			"comment": "", "timeout": 5936, "hostname": "computer",
			"expires": 1794338284,
		}},
	})
	page, err := New(ts.Client()).IPv6.LeaseList(context.Background(), nil)
	if err != nil {
		t.Fatalf("LeaseList 失败: %v", err)
	}
	if page.ClientTotal != 1 || len(page.Rows) != 1 {
		t.Fatalf("DHCPv6 租约解码不符: %+v", page)
	}
	row := page.Rows[0]
	if row.DUID != "00010001322b3bc1d8bbc19a8c8f" || row.Hostname != "computer" ||
		row.Expires != 1794338284 {
		t.Fatalf("租约行解码不符: %+v", row)
	}
	// 固件 ipv6_addr 带尾随空格（真机观测），辅助方法应去除。
	if row.IPv6AddrTrimmed() != "2408:8220:320:88e2::ff3" {
		t.Fatalf("IPv6AddrTrimmed = %q", row.IPv6AddrTrimmed())
	}
	wantParam(t, paramOf(t, ts.CallsFor("dhcp6_lease")[0]),
		map[string]any{"TYPE": DHCP6LeaseTypeAll, "limit": "0,20"})
}

// TestIPv6StaticACLNeighbor 覆盖 ipv6_static/dhcp6_acl_mac CRUD 与
// ipv6_neighbor.show（邻居行形态真机实证）。
func TestIPv6StaticACLNeighbor(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"ipv6_static", "dhcp6_acl_mac"} {
		for _, a := range []string{"show", "add", "edit", "del"} {
			ts.SetResponse(fn, a, map[string]any{})
		}
		ts.SetResponse(fn, "show", map[string]any{"total": 0, "data": []any{}})
	}
	ts.SetResponse("ipv6_neighbor", "show", map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{"comment": "", "interface": "vlan1",
				"ipv6_addr": "fe80::20c:29ff:fe42:d0f7", "status": "STALE",
				"id": 1, "mac": "00:0c:29:42:d0:f7"},
			map[string]any{"comment": "", "interface": "lan4",
				"ipv6_addr": "2408:8220:320:88e2:3150:52ae:7f4:3df8", "status": "DELAY",
				"id": 10, "mac": "4c:10:d5:42:05:b6"},
		},
	})
	svc := New(ts.Client()).IPv6
	ctx := context.Background()

	if _, err := svc.StaticList(ctx, nil); err != nil {
		t.Fatalf("StaticList 失败: %v", err)
	}
	if _, err := svc.StaticAdd(ctx, IPv6Static{Interface: "lan4", DUID: "00010001322b3bc1", IPv6Addr: "2408::100"}); err != nil {
		t.Fatalf("StaticAdd 失败: %v", err)
	}
	if err := svc.StaticEdit(ctx, IPv6Static{ID: 1}); err != nil {
		t.Fatalf("StaticEdit 失败: %v", err)
	}
	if err := svc.StaticDelete(ctx, 1); err != nil {
		t.Fatalf("StaticDelete 失败: %v", err)
	}

	if _, err := svc.ACLList(ctx, nil); err != nil {
		t.Fatalf("ACLList 失败: %v", err)
	}
	if _, err := svc.ACLAdd(ctx, DHCP6ACLMac{MAC: "aa:bb:cc:dd:ee:ff"}); err != nil {
		t.Fatalf("ACLAdd 失败: %v", err)
	}
	if err := svc.ACLEdit(ctx, DHCP6ACLMac{ID: 1}); err != nil {
		t.Fatalf("ACLEdit 失败: %v", err)
	}
	if err := svc.ACLDelete(ctx, 1); err != nil {
		t.Fatalf("ACLDelete 失败: %v", err)
	}

	nb, err := svc.NeighborList(ctx, nil)
	if err != nil {
		t.Fatalf("NeighborList 失败: %v", err)
	}
	if nb.Total != 2 || nb.Rows[0].Status != IPv6NeighborStatusStale ||
		nb.Rows[1].Status != IPv6NeighborStatusDelay ||
		nb.Rows[1].Interface != "lan4" {
		t.Fatalf("邻居表解码不符: %+v", nb)
	}
	for _, fn := range []string{"ipv6_static", "dhcp6_acl_mac"} {
		assertActionSeq(t, ts.CallsFor(fn), "show", "add", "edit", "del")
	}
}

// TestIPv6Error 覆盖固件错误路径（模拟 dhcp6_lease 传入错误 TYPE 的
// 真机错误形态）。
func TestIPv6Error(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dhcp6_lease", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (data)", FuncName: "dhcp6_lease", Action: action}
	})
	_, err := New(ts.Client()).IPv6.LeaseList(context.Background(),
		&ikuai.ListOptions{Type: "total,data"})
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
