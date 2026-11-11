package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// qosSample 构造真机探测形态的 simple_qos/mac_qos.show 响应 Data
// （行字段为真机 3.7.24 实测）。
func qosSample() map[string]any {
	return map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{
				"id": 1, "enabled": "yes", "comment": "",
				"interface": "adsl01,adsl02", "ip_addr": "172.22.1.2",
				"upload": 18560, "download": 18560, "type": 1, "attr": 0,
				"protocol": "any", "src_port": "", "dst_port": "",
				"week": "1234567", "time": "00:00-23:59",
			},
			map[string]any{
				"id": 2, "enabled": "no", "comment": "备用",
				"interface": "adsl03", "ip_addr": "172.22.1.3",
				"upload": 23600, "download": 23600, "type": 1, "attr": 0,
				"protocol": "any", "src_port": "", "dst_port": "",
				"week": "1234567", "time": "00:00-23:59",
			},
		},
	}
}

// macQoSSample 构造真机探测形态的 mac_qos.show 响应 Data
// （行字段为真机 3.7.24 实测）。
func macQoSSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "comment": "快速添加",
			"interface": "", "mac_addr": "00:0c:29:0e:10:5f",
			"ip_type": "4", "upload": 25000, "download": 25000,
			"type": 0, "attr": 3, "week": "1234567", "time": "00:00-23:59",
		}},
	}
}

// TestSimpleQoSList 覆盖 simple_qos.show 分页列表的成功路径与参数断言。
func TestSimpleQoSList(t *testing.T) {
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
			name: "第二页",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 20},
			want: map[string]any{"TYPE": "total,data", "limit": "20,20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("simple_qos", "show", qosSample())
			page, err := New(ts.Client()).SimpleQoS.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 2 || len(page.Rows) != 2 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || !r.IsEnabled() || r.Upload != 18560 ||
				r.Download != 18560 || r.IPAddr != "172.22.1.2" ||
				r.Type != 1 || r.Attr != 0 || r.Protocol != "any" {
				t.Fatalf("规则行解码不符: %+v", r)
			}
			if page.Rows[1].IsEnabled() {
				t.Fatalf("第 2 行 enabled=no, IsEnabled 应为 false")
			}
			wantParam(t, paramOf(t, ts.CallsFor("simple_qos")[0]), tt.want)
		})
	}
}

// TestSimpleQoSSelectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestSimpleQoSSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("simple_qos", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface": []any{[]any{"adsl01", "移动私网"}},
		"ipgroup":   []any{"公共无线", "前台"},
	})
	sel, err := New(ts.Client()).SimpleQoS.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 1 || sel.Interface[0][1] != "移动私网" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.IPGroup) != 2 || sel.IPGroup[1] != "前台" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("simple_qos")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestSimpleQoSWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestSimpleQoSWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("simple_qos", "add", map[string]any{})
	ts.SetResponse("simple_qos", "edit", map[string]any{})
	ts.SetResponse("simple_qos", "del", map[string]any{})
	svc := New(ts.Client()).SimpleQoS
	ctx := context.Background()

	rule := SimpleQoSRule{Enabled: EnabledYes, IPAddr: "172.22.1.2",
		Upload: 10240, Download: 20480, Interface: "wan1",
		Protocol: "any", Week: "1234567", Time: "00:00-23:59"}
	rowID, err := svc.Add(ctx, rule)
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("simple_qos")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "ip_addr", "upload", "download",
		"interface", "protocol", "week", "time"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}
	if p["upload"] != float64(10240) || p["download"] != float64(20480) {
		t.Fatalf("add 限速值不符: %v", p)
	}

	if err := svc.Edit(ctx, SimpleQoSRule{ID: 1, Upload: 4096, Download: 4096}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("simple_qos")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if rec = ts.CallsFor("simple_qos")[2]; rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 1})
}

// TestSimpleQoSError 覆盖固件错误路径。
func TestSimpleQoSError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("simple_qos", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "simple_qos", Action: action}
	})
	svc := New(ts.Client()).SimpleQoS
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}

// TestMacQoSList 覆盖 mac_qos.show 分页列表的成功路径与参数断言。
func TestMacQoSList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("mac_qos", "show", macQoSSample())
	svc := New(ts.Client()).MacQoS

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.MACAddr != "00:0c:29:0e:10:5f" ||
		r.IPType != "4" || r.Upload != 25000 || r.Download != 25000 ||
		r.Type != 0 || r.Attr != 3 || r.Comment != "快速添加" {
		t.Fatalf("规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("mac_qos")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestMacQoSSelectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestMacQoSSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("mac_qos", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface": []any{[]any{"adsl01", "移动私网"}},
		"macgroup":  []any{},
		"dtgroup":   []any{},
	})
	sel, err := New(ts.Client()).MacQoS.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 1 || sel.Interface[0][0] != "adsl01" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if sel.MACGroup == nil || sel.DTGroup == nil {
		t.Fatalf("空区块应解码为空切片: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("mac_qos")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestMacQoSWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestMacQoSWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("mac_qos", "add", map[string]any{})
	ts.SetResponse("mac_qos", "edit", map[string]any{})
	ts.SetResponse("mac_qos", "del", map[string]any{})
	svc := New(ts.Client()).MacQoS
	ctx := context.Background()

	rule := MacQoSRule{Enabled: EnabledYes, MACAddr: "00:11:22:33:44:55",
		IPType: "4", Upload: 5120, Download: 10240,
		Week: "1234567", Time: "00:00-23:59"}
	rowID, err := svc.Add(ctx, rule)
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("mac_qos")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "mac_addr", "ip_type", "upload",
		"download", "week", "time"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, MacQoSRule{ID: 9, Upload: 2048, Download: 2048}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("mac_qos")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(9) {
		t.Fatalf("edit 参数应携带 id=9: %v", p)
	}

	if err := svc.Delete(ctx, 9); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("mac_qos")[2]), map[string]any{"id": 9})
	if err := svc.Delete(ctx, 9, 10); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("mac_qos")[3]),
		map[string]any{"ids": []any{9, 10}})
}

// TestMacQoSError 覆盖固件错误路径。
func TestMacQoSError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("mac_qos", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "mac_qos", Action: action}
	})
	svc := New(ts.Client()).MacQoS
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Selectors(ctx); err == nil {
		t.Fatal("TYPE=all show 错误应上抛")
	}
	if _, err := svc.Add(ctx, MacQoSRule{}); err == nil {
		t.Fatal("add 错误应上抛")
	}
}
