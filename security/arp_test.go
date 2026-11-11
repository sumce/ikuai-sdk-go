package security

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// arpSample 构造 arp.show 的分页响应 Data（行字段为真机实测形态）。
func arpSample() map[string]any {
	return map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{
				"id": 1, "ip_addr": "172.22.2.3", "ip_addr_int": 2887123459,
				"mac": "00:0c:29:57:23:25", "interface": "vlan1",
				"bind_type": "0", "bind_state": "0", "comment": "",
			},
			map[string]any{
				"id": 2, "ip_addr": "172.23.0.2", "ip_addr_int": 2887188482,
				"mac": "00:0c:29:a1:ff:c3", "interface": "lan5",
				"bind_type": "1", "bind_state": "1", "comment": "服务器",
				"offline_time": "2026-11-10 08:00:00",
			},
		},
	}
}

// arpSelectorsSample 构造 arp.show TYPE=all 的选择器响应 Data（真机实测）。
func arpSelectorsSample() map[string]any {
	return map[string]any{
		"total":            21,
		"data":             []any{},
		"interface":        []any{[]any{"lan1"}, []any{"adsl07", "联通静公网"}},
		"lan_interface":    []any{[]any{"lan1"}, []any{"vlan1"}},
		"arp_filter":       0,
		"arp_filter_iface": "any",
		"dhcpd_arp":        0,
	}
}

// TestARPList 覆盖 arp.show 的成功路径与参数断言。
func TestARPList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
		},
		{
			name: "自定义分页",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 50, OrderBy: "ip_addr_int", Order: "desc"},
			want: map[string]any{"TYPE": "total,data", "limit": "0,50", "ORDER_BY": "ip_addr_int", "ORDER": "desc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("arp", "show", arpSample())
			page, err := New(ts.Client()).ARP.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 2 || len(page.Rows) != 2 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			e := page.Rows[0]
			if e.ID != 1 || e.IPAddr != "172.22.2.3" || e.IPAddrInt != 2887123459 ||
				e.MAC != "00:0c:29:57:23:25" || e.Interface != "vlan1" ||
				e.BindType != ARPBindTypeNormal || e.BindState != ARPBindStateUnbound ||
				e.Comment != "" {
				t.Fatalf("ARP 行解码不符: %+v", e)
			}
			if e.IsBound() || e.IsUnique() {
				t.Fatal("行 1 应为未绑定/普通")
			}
			if e.OfflineTime != nil {
				t.Fatalf("行 1 不应有 offline_time: %v", e.OfflineTime)
			}
			b := page.Rows[1]
			if !b.IsBound() || !b.IsUnique() || b.Comment != "服务器" ||
				b.OfflineTime == nil || *b.OfflineTime != "2026-11-10 08:00:00" {
				t.Fatalf("ARP 行 2 解码不符: %+v", b)
			}
			wantParam(t, paramOf(t, ts.CallsFor("arp")[0]), tt.want)
		})
	}
}

// TestARPSearch 覆盖 arp.show 的 FINDS/KEYWORDS 检索参数（真机实测形态）。
func TestARPSearch(t *testing.T) {
	tests := []struct {
		name string
		key  string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 关键字检索",
			key:  "172.22",
			opts: nil,
			want: map[string]any{
				"TYPE": "total,data", "limit": "0,20",
				"FINDS": "ip_addr,mac,comment,interface", "KEYWORDS": "172.22",
			},
		},
		{
			name: "opts 覆盖 FINDS",
			key:  "00:0c:29",
			opts: &ikuai.ListOptions{Limit: 5, Extra: map[string]any{"FINDS": "mac"}},
			want: map[string]any{
				"TYPE": "total,data", "limit": "0,5",
				"FINDS": "mac", "KEYWORDS": "00:0c:29",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("arp", "show", arpSample())
			page, err := New(ts.Client()).ARP.Search(context.Background(), tt.key, tt.opts)
			if err != nil {
				t.Fatalf("Search 失败: %v", err)
			}
			if page.Total != 2 || len(page.Rows) != 2 {
				t.Fatalf("检索解码不符: %+v", page)
			}
			wantParam(t, paramOf(t, ts.CallsFor("arp")[0]), tt.want)
		})
	}
}

// TestARPSelectors 覆盖 arp.show TYPE=all 选择器与全局开关解码。
func TestARPSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("arp", "show", arpSelectorsSample())
	sel, err := New(ts.Client()).ARP.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 2 || sel.Interface[1][1] != "联通静公网" ||
		len(sel.LANInterface) != 2 || sel.LANInterface[1][0] != "vlan1" {
		t.Fatalf("接口选择器解码不符: %+v", sel)
	}
	if sel.ARPFilter != 0 || sel.ARPFilterIface != "any" || sel.DHCPDARP != 0 {
		t.Fatalf("全局开关解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("arp")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestARPError 覆盖固件错误路径（arp 为只读子服务）。
func TestARPError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("arp", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "arp", Action: action}
	})
	svc := New(ts.Client()).ARP
	if _, err := svc.List(context.Background(), nil); err == nil {
		t.Fatal("List 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.Search(context.Background(), "x", nil); err == nil {
		t.Fatal("Search 应失败")
	}
	if _, err := svc.Selectors(context.Background()); err == nil {
		t.Fatal("Selectors 应失败")
	}
}
