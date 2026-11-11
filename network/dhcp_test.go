package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// dhcpServerSample 构造真机形态的 dhcp_server.show 响应 Data。
func dhcpServerSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "interface": "lan4", "phy_ifnames": "all",
			"addr_pool": "192.168.2.1-192.168.2.10", "exclude_pool": "",
			"netmask": "255.255.255.0", "gateway": "192.168.2.1",
			"dns1": "192.168.2.1", "dns2": "223.5.5.5", "domain": "",
			"next_server": "", "wins1": "", "wins2": "", "lease": 720,
			"delay": 0, "status": 1, "available": 10,
			"check_addr_valid": 1, "check_relay_only": 0,
			"opt_type121": 2, "opt121": "172.20.0.0/24,192.168.2.1",
		}},
	}
}

// dhcpRowSample 构造 dhcp_static/dhcp_lease 同形的行样本。
func dhcpRowSample() map[string]any {
	return map[string]any{
		"id": 5, "interface": "lan4", "ip_addr": "192.168.2.8",
		"ip_addr_int": 3232236040, "mac": "4c:10:d5:42:05:b6",
		"hostname": "computer", "status": 1, "start_time": 1794331084,
		"end_time": 1794338284, "timeout": 5936, "comment": "ikuai-sdk-test-",
	}
}

// TestDHCPServerCRUD 覆盖 dhcp_server 的 show/add/edit/del。
func TestDHCPServerCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dhcp_server", "show", dhcpServerSample())
	ts.SetResponse("dhcp_server", "add", map[string]any{})
	ts.SetResponse("dhcp_server", "edit", map[string]any{})
	ts.SetResponse("dhcp_server", "del", map[string]any{})
	svc := New(ts.Client()).DHCP

	page, err := svc.ServerList(context.Background(), nil)
	if err != nil {
		t.Fatalf("ServerList 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.Interface != "lan4" || r.AddrPool != "192.168.2.1-192.168.2.10" ||
		r.Lease != 720 || r.Available != 10 || r.OptType121 != 2 ||
		r.Opt121 != "172.20.0.0/24,192.168.2.1" || !r.IsEnabled() {
		t.Fatalf("服务端行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dhcp_server")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	id, err := svc.ServerAdd(context.Background(), DHCPServer{Enabled: EnabledYes, Interface: "lan4", AddrPool: "192.168.2.1-192.168.2.10", Netmask: "255.255.255.0", Gateway: "192.168.2.1"})
	if err != nil {
		t.Fatalf("ServerAdd 失败: %v", err)
	}
	_ = id
	recs := ts.CallsFor("dhcp_server")
	if recs[1].Action != "add" {
		t.Fatalf("action = %q, 期望 add", recs[1].Action)
	}
	p := paramOf(t, recs[1])
	if p["interface"] != "lan4" || p["enabled"] != "yes" {
		t.Fatalf("add 参数不符: %v", p)
	}

	if err := svc.ServerEdit(context.Background(), DHCPServer{ID: 1, Lease: 1440}); err != nil {
		t.Fatalf("ServerEdit 失败: %v", err)
	}
	if ts.CallsFor("dhcp_server")[2].Action != "edit" {
		t.Fatal("第 3 次调用应为 edit")
	}

	if err := svc.ServerDelete(context.Background(), 1, 2); err != nil {
		t.Fatalf("ServerDelete 失败: %v", err)
	}
	if ts.CallsFor("dhcp_server")[3].Action != "del" {
		t.Fatal("第 4 次调用应为 del")
	}
}

// TestDHCPStaticCRUD 覆盖 dhcp_static 的 show/add/edit/del。
func TestDHCPStaticCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dhcp_static", "show", map[string]any{"total": 1, "data": []any{dhcpRowSample()}})
	ts.SetResponse("dhcp_static", "add", map[string]any{})
	ts.SetResponse("dhcp_static", "edit", map[string]any{})
	ts.SetResponse("dhcp_static", "del", map[string]any{})
	svc := New(ts.Client()).DHCP

	page, err := svc.StaticList(context.Background(), nil)
	if err != nil {
		t.Fatalf("StaticList 失败: %v", err)
	}
	if page.Total != 1 || page.Rows[0].IPAddr != "192.168.2.8" ||
		page.Rows[0].Status != DHCPStatusStatic {
		t.Fatalf("静态分配解码不符: %+v", page)
	}
	recs := ts.CallsFor("dhcp_static")
	if len(recs) != 1 || recs[0].FuncName != "dhcp_static" || recs[0].Action != "show" {
		t.Fatalf("调用记录不符: %+v", recs)
	}

	if _, err := svc.StaticAdd(context.Background(), DHCPStatic{Interface: "lan4", MAC: "4c:10:d5:42:05:b6", IPAddr: "192.168.2.8"}); err != nil {
		t.Fatalf("StaticAdd 失败: %v", err)
	}
	if ts.CallsFor("dhcp_static")[1].Action != "add" {
		t.Fatal("第 2 次调用应为 add")
	}
	if err := svc.StaticEdit(context.Background(), DHCPStatic{ID: 5, Comment: "x"}); err != nil {
		t.Fatalf("StaticEdit 失败: %v", err)
	}
	if err := svc.StaticDelete(context.Background(), 5); err != nil {
		t.Fatalf("StaticDelete 失败: %v", err)
	}
	recs = ts.CallsFor("dhcp_static")
	if recs[2].Action != "edit" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s", recs[2].Action, recs[3].Action)
	}
	// 单个 id 形态断言 {"id":5}。
	if p := paramOf(t, recs[3]); p["id"] != float64(5) {
		t.Fatalf("del 参数不符: %v", p)
	}
}

// TestDHCPLeaseList 覆盖 dhcp_lease.show（运行状态表，只读）。
func TestDHCPLeaseList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dhcp_lease", "show", map[string]any{"total": 1, "data": []any{dhcpRowSample()}})
	page, err := New(ts.Client()).DHCP.LeaseList(context.Background(), &ikuai.ListOptions{Offset: 20, Limit: 20})
	if err != nil {
		t.Fatalf("LeaseList 失败: %v", err)
	}
	if page.Total != 1 || page.Rows[0].MAC != "4c:10:d5:42:05:b6" || page.Rows[0].Timeout != 5936 {
		t.Fatalf("租约解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dhcp_lease")[0]),
		map[string]any{"TYPE": "total,data", "limit": "20,20"})
}

// TestDHCPACLMacCRUD 覆盖 dhcp_acl_mac 的 show/add/edit/del。
func TestDHCPACLMacCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dhcp_acl_mac", "show", map[string]any{"total": 0, "data": []any{}})
	ts.SetResponse("dhcp_acl_mac", "add", map[string]any{})
	ts.SetResponse("dhcp_acl_mac", "edit", map[string]any{})
	ts.SetResponse("dhcp_acl_mac", "del", map[string]any{})
	svc := New(ts.Client()).DHCP

	if _, err := svc.ACLList(context.Background(), nil); err != nil {
		t.Fatalf("ACLList 失败: %v", err)
	}
	if _, err := svc.ACLAdd(context.Background(), DHCPACLMac{MAC: "aa:bb:cc:dd:ee:ff"}); err != nil {
		t.Fatalf("ACLAdd 失败: %v", err)
	}
	if err := svc.ACLEdit(context.Background(), DHCPACLMac{ID: 1}); err != nil {
		t.Fatalf("ACLEdit 失败: %v", err)
	}
	if err := svc.ACLDelete(context.Background(), 1, 2, 3); err != nil {
		t.Fatalf("ACLDelete 失败: %v", err)
	}
	recs := ts.CallsFor("dhcp_acl_mac")
	if len(recs) != 4 {
		t.Fatalf("应有 4 次调用, got %d", len(recs))
	}
	for i, want := range []string{"show", "add", "edit", "del"} {
		if recs[i].Action != want {
			t.Fatalf("第 %d 次调用 action = %q, 期望 %q", i+1, recs[i].Action, want)
		}
	}
	// 多 id 形态断言 {"ids":[1,2,3]}。
	p := paramOf(t, recs[3])
	ids, ok := p["ids"].([]any)
	if !ok || len(ids) != 3 || ids[0] != float64(1) {
		t.Fatalf("del 多 id 参数不符: %v", p)
	}
}

// TestDHCPError 覆盖固件错误路径。
func TestDHCPError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dhcp_server", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "dhcp_server", Action: action}
	})
	_, err := New(ts.Client()).DHCP.ServerList(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
