package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// layer7Sample 构造真机探测形态的 stream_layer7.show 响应 Data
// （探测真机列表为空，行字段为推断形态，见 Layer7Rule 注释）。
func layer7Sample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "comment": "游戏走电信",
			"interface": "adsl01", "app_proto": "game",
			"src_addr": "192.168.2.0/24", "dst_addr": "",
			"week": "1234567", "time": "00:00-23:59",
		}},
	}
}

// TestLayer7List 覆盖 stream_layer7.show 分页列表的成功路径与参数断言。
func TestLayer7List(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_layer7", "show", layer7Sample())
	svc := New(ts.Client()).Layer7

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.AppProto != "game" ||
		r.Interface != "adsl01" || r.SrcAddr != "192.168.2.0/24" {
		t.Fatalf("规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_layer7")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	// 自定义分页。
	if _, err := svc.List(context.Background(), &ikuai.ListOptions{Offset: 10, Limit: 10}); err != nil {
		t.Fatalf("List 第二页失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_layer7")[1]),
		map[string]any{"TYPE": "total,data", "limit": "10,10"})
}

// TestLayer7Selectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestLayer7Selectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_layer7", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface": []any{[]any{"wan1", "电信外网"}},
		"ipgroup":   []any{"前台"},
		"macgroup":  []any{},
		"dtgroup":   []any{},
		"syn_proxy": 0,
	})
	sel, err := New(ts.Client()).Layer7.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 1 || sel.Interface[0][0] != "wan1" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.IPGroup) != 1 || sel.IPGroup[0] != "前台" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	if sel.SynProxy != 0 || sel.MACGroup == nil || sel.DTGroup == nil {
		t.Fatalf("其余区块解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_layer7")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestLayer7Writes 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestLayer7Writes(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_layer7", "add", map[string]any{})
	ts.SetResponse("stream_layer7", "edit", map[string]any{})
	ts.SetResponse("stream_layer7", "del", map[string]any{})
	svc := New(ts.Client()).Layer7
	ctx := context.Background()

	rowID, err := svc.Add(ctx, Layer7Rule{Enabled: EnabledYes,
		AppProto: "game", Interface: "wan1", Comment: "测试"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("stream_layer7")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "app_proto", "interface", "comment"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, Layer7Rule{ID: 2, Enabled: EnabledNo}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("stream_layer7")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}

	if err := svc.Delete(ctx, 2); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if rec = ts.CallsFor("stream_layer7")[2]; rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 2})
}

// TestLayer7Error 覆盖固件错误路径。
func TestLayer7Error(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("stream_layer7", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "stream_layer7", Action: action}
	})
	svc := New(ts.Client()).Layer7
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if err := svc.Edit(ctx, Layer7Rule{ID: 1}); err == nil {
		t.Fatal("edit 错误应上抛")
	}
}
