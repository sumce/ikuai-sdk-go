package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// updownSample 构造真机探测形态的 stream_updown.show 响应 Data
// （探测真机列表为空，行字段为推断形态，见 UpDownRule 注释）。
func updownSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "no", "comment": "备用",
			"interface": "adsl10", "protocol": "tcp+udp",
			"src_addr": "", "dst_addr": "", "src_port": "", "dst_port": "",
			"week": "1234567", "time": "00:00-23:59",
		}},
	}
}

// TestUpDownList 覆盖 stream_updown.show 分页列表的成功路径与参数断言。
func TestUpDownList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_updown", "show", updownSample())
	svc := New(ts.Client()).UpDown

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || r.IsEnabled() || r.Protocol != "tcp+udp" ||
		r.Interface != "adsl10" {
		t.Fatalf("规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_updown")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestUpDownSelectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestUpDownSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_updown", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface": []any{[]any{"adsl01", "移动私网"}},
		"ipgroup":   []any{"监控"},
		"macgroup":  []any{},
		"dtgroup":   []any{},
		"protocol":  []any{"tcp", "udp", "tcp+udp", "icmp", "any"},
	})
	sel, err := New(ts.Client()).UpDown.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 1 || sel.Interface[0][1] != "移动私网" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.Protocol) != 5 || sel.Protocol[4] != "any" {
		t.Fatalf("协议选择器解码不符: %v", sel.Protocol)
	}
	if len(sel.IPGroup) != 1 || sel.IPGroup[0] != "监控" ||
		sel.MACGroup == nil || sel.DTGroup == nil {
		t.Fatalf("其余区块解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_updown")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestUpDownWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestUpDownWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_updown", "add", map[string]any{})
	ts.SetResponse("stream_updown", "edit", map[string]any{})
	ts.SetResponse("stream_updown", "del", map[string]any{})
	svc := New(ts.Client()).UpDown
	ctx := context.Background()

	rowID, err := svc.Add(ctx, UpDownRule{Enabled: EnabledYes,
		Interface: "wan1", Protocol: "any", Comment: "测试"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("stream_updown")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "interface", "protocol", "comment"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, UpDownRule{ID: 6, Enabled: EnabledNo}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("stream_updown")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}

	if err := svc.Delete(ctx, 6); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if rec = ts.CallsFor("stream_updown")[2]; rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 6})
}

// TestUpDownError 覆盖固件错误路径。
func TestUpDownError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("stream_updown", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "stream_updown", Action: action}
	})
	svc := New(ts.Client()).UpDown
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if err := svc.Edit(ctx, UpDownRule{ID: 1}); err == nil {
		t.Fatal("edit 错误应上抛")
	}
}
