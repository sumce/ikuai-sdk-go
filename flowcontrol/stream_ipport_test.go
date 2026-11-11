package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// ipportSample 构造真机探测形态的 stream_ipport.show 响应 Data
// （TYPE=total,data，行字段为真机 3.7.24 实测）。
func ipportSample() map[string]any {
	return map[string]any{
		"total": 9,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "comment": "出租上网", "nexthop": "",
			"interface": "adsl07", "mode": 0, "type": 0, "dst_addr": "",
			"protocol": "any", "src_port": "", "dst_port": "", "dst_type": 0,
			"area_code": "", "iface_band": 1, "week": "1234567",
			"time": "00:00-23:59", "src_addr": "172.22.1.2,172.22.1.3",
		}},
	}
}

// TestIPPortList 覆盖 stream_ipport.show 分页列表的成功路径与参数断言。
func TestIPPortList(t *testing.T) {
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
			name: "取 5 行",
			opts: &ikuai.ListOptions{Limit: 5},
			want: map[string]any{"TYPE": "total,data", "limit": "0,5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("stream_ipport", "show", ipportSample())
			page, err := New(ts.Client()).IPPort.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 9 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || !r.IsEnabled() || r.Interface != "adsl07" ||
				r.Protocol != "any" || r.IfaceBand != 1 ||
				r.SrcAddr != "172.22.1.2,172.22.1.3" || r.Week != "1234567" {
				t.Fatalf("规则行解码不符: %+v", r)
			}
			if r.Comment != "出租上网" || r.Time != "00:00-23:59" {
				t.Fatalf("规则行文本字段解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("stream_ipport")[0]), tt.want)
		})
	}
}

// TestIPPortSelectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestIPPortSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_ipport", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface":    []any{[]any{"adsl01", "移动私网"}},
		"ipgroup":      []any{"公共无线", "前台"},
		"macgroup":     []any{},
		"dtgroup":      []any{},
		"protocol":     []any{"tcp", "udp", "tcp+udp", "icmp", "any"},
		"iplib_status": 0,
	})
	sel, err := New(ts.Client()).IPPort.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 1 || sel.Interface[0][1] != "移动私网" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.IPGroup) != 2 || sel.IPGroup[0] != "公共无线" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	if len(sel.Protocol) != 5 || sel.Protocol[2] != "tcp+udp" {
		t.Fatalf("协议选择器解码不符: %v", sel.Protocol)
	}
	if sel.IPLibStatus != 0 || sel.MACGroup == nil || sel.DTGroup == nil {
		t.Fatalf("其余区块解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_ipport")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestIPPortWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestIPPortWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_ipport", "add", map[string]any{})
	ts.SetResponse("stream_ipport", "edit", map[string]any{})
	ts.SetResponse("stream_ipport", "del", map[string]any{})
	svc := New(ts.Client()).IPPort
	ctx := context.Background()

	rule := IPPortRule{Enabled: EnabledYes, Comment: "测试分流",
		Interface: "wan1", Protocol: "tcp", SrcAddr: "192.168.1.0/24",
		Week: "1234567", Time: "00:00-23:59", IfaceBand: 1}
	rowID, err := svc.Add(ctx, rule)
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("stream_ipport")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "comment", "interface", "protocol",
		"src_addr", "week", "time", "iface_band"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, IPPortRule{ID: 5, Enabled: EnabledNo}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	rec = ts.CallsFor("stream_ipport")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(5) {
		t.Fatalf("edit 参数应携带 id=5: %v", p)
	}

	if err := svc.Delete(ctx, 5); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_ipport")[2]),
		map[string]any{"id": 5})
	if err := svc.Delete(ctx, 5, 6); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_ipport")[3]),
		map[string]any{"ids": []any{5, 6}})
}

// TestIPPortError 覆盖固件错误路径。
func TestIPPortError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("stream_ipport", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "stream_ipport", Action: action}
	})
	svc := New(ts.Client()).IPPort
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Selectors(ctx); err == nil {
		t.Fatal("TYPE=all show 错误应上抛")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}
