package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// dnatSample 构造抓包形态的 dnat.show 响应 Data。
func dnatSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"wan_port": "7422,30501-45000", "lan_port": "22,30501-45000",
			"lan_addr_int": 2887123202, "id": 1, "enabled": "yes",
			"comment": "出租300", "interface": "adsl07", "lan_addr": "172.22.1.2",
			"src_addr": "", "protocol": "tcp+udp",
		}},
	}
}

// TestDNATList 覆盖 dnat.show 的成功路径与参数断言（抓包实证）。
func TestDNATList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页（抓包形态）",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
		},
		{
			name: "自定义分页",
			opts: &ikuai.ListOptions{Offset: 40, Limit: 20, OrderBy: "id", Order: "desc"},
			want: map[string]any{"TYPE": "total,data", "limit": "40,20", "ORDER_BY": "id", "ORDER": "desc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("dnat", "show", dnatSample())
			page, err := New(ts.Client()).DNAT.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.WANPort != "7422,30501-45000" || r.Interface != "adsl07" ||
				r.LANAddr != "172.22.1.2" || r.Protocol != DNATProtocolTCPUDP ||
				!r.IsEnabled() {
				t.Fatalf("规则行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("dnat")[0]), tt.want)
		})
	}
}

// TestDNATSelectors 覆盖 TYPE=interface,protocol 候选路径（抓包实证）。
func TestDNATSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dnat", "show", map[string]any{
		"interface": []any{[]any{"adsl07", "联通静公网"}, []any{"lan1"}},
		"protocol":  []any{"tcp", "udp", "tcp+udp"},
	})
	sel, err := New(ts.Client()).DNAT.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 2 || sel.Interface[0][0] != "adsl07" ||
		sel.Interface[0][1] != "联通静公网" || len(sel.Protocol) != 3 {
		t.Fatalf("候选解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dnat")[0]),
		map[string]any{"TYPE": DNATTypeInterfaceProtocol})
}

// TestDNATIPGroups 覆盖 TYPE=ipgroup 候选路径（抓包实证）。
func TestDNATIPGroups(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dnat", "show", map[string]any{"ipgroup": []any{"出租300", "办公"}})
	groups, err := New(ts.Client()).DNAT.IPGroups(context.Background())
	if err != nil {
		t.Fatalf("IPGroups 失败: %v", err)
	}
	if len(groups) != 2 || groups[0] != "出租300" {
		t.Fatalf("分组候选不符: %v", groups)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dnat")[0]),
		map[string]any{"TYPE": DNATTypeIPGroup})
}

// TestDNATAddEditDelete 覆盖 add/edit/del（add/del 形态抓包实证）。
func TestDNATAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dnat", "add", map[string]any{})
	ts.SetResponse("dnat", "edit", map[string]any{})
	ts.SetResponse("dnat", "del", map[string]any{})
	svc := New(ts.Client()).DNAT
	ctx := context.Background()

	// add 参数形态与抓包一致。
	id, err := svc.Add(ctx, DNATRule{
		Enabled: EnabledYes, Comment: "test", Interface: "adsl01",
		LANAddr: "172.22.2.1", Protocol: DNATProtocolTCP,
		WANPort: "22112", LANPort: "22", SrcAddr: "",
	})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	_ = id
	p := paramOf(t, ts.CallsFor("dnat")[0])
	for k, want := range map[string]any{
		"enabled": "yes", "comment": "test", "interface": "adsl01",
		"lan_addr": "172.22.2.1", "protocol": "tcp",
		"wan_port": "22112", "lan_port": "22", "src_addr": "",
	} {
		if p[k] != want {
			t.Fatalf("add param[%q] = %v, 期望 %v", k, p[k], want)
		}
	}
	if _, hasID := p["id"]; hasID {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, DNATRule{ID: 29, Comment: "x"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 29); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	recs := ts.CallsFor("dnat")
	if recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s", recs[1].Action, recs[2].Action)
	}
	if p := paramOf(t, recs[2]); p["id"] != float64(29) {
		t.Fatalf("del 参数不符（抓包实证 {\"id\":N}）: %v", p)
	}
}

// TestDNATError 覆盖固件错误路径。
func TestDNATError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dnat", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "dnat", Action: action}
	})
	_, err := New(ts.Client()).DNAT.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
