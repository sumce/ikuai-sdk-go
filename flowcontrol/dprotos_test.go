package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// dprotosSample 构造真机探测形态的 dprotos.show 响应 Data
// （探测真机列表为空，行字段为推断形态，见 CustomProto 注释）。
func dprotosSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "name": "自定义游戏", "protocol": "tcp",
			"ports": "7000-8000,9000", "comment": "测试协议",
		}},
	}
}

// TestCustomProtosList 覆盖 dprotos.show 分页列表的成功路径与参数断言。
func TestCustomProtosList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dprotos", "show", dprotosSample())
	svc := New(ts.Client()).CustomProtos

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Name != "自定义游戏" || r.Protocol != "tcp" ||
		r.Ports != "7000-8000,9000" || r.Comment != "测试协议" {
		t.Fatalf("协议行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dprotos")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestCustomProtosSelectors 覆盖 TYPE=all 表单选择器路径
// （真机实测仅 ipgroup 区块）。
func TestCustomProtosSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dprotos", "show", map[string]any{
		"total": 0, "data": []any{},
		"ipgroup": []any{"公共无线", "前台", "监控", "运营", "销售"},
	})
	sel, err := New(ts.Client()).CustomProtos.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 5 || sel.IPGroup[0] != "公共无线" || sel.IPGroup[4] != "销售" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dprotos")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestCustomProtosWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestCustomProtosWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dprotos", "add", map[string]any{})
	ts.SetResponse("dprotos", "edit", map[string]any{})
	ts.SetResponse("dprotos", "del", map[string]any{})
	svc := New(ts.Client()).CustomProtos
	ctx := context.Background()

	rowID, err := svc.Add(ctx, CustomProto{Name: "测试协议",
		Protocol: "udp", Ports: "53", Comment: "测试"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("dprotos")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"name", "protocol", "ports", "comment"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, CustomProto{ID: 3, Ports: "54"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("dprotos")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(3) {
		t.Fatalf("edit 参数应携带 id=3: %v", p)
	}

	if err := svc.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if rec = ts.CallsFor("dprotos")[2]; rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 3})
}

// TestCustomProtosError 覆盖固件错误路径。
func TestCustomProtosError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dprotos", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "dprotos", Action: action}
	})
	svc := New(ts.Client()).CustomProtos
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

// TestAdvancedProtosList 覆盖 dprotos_l7.show 分页列表的成功路径与参数断言
// （探测真机列表为空，行字段为推断形态，见 AdvancedProto 注释）。
func TestAdvancedProtosList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dprotos_l7", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "name": "高级协议", "comment": "测试",
		}},
	})
	svc := New(ts.Client()).AdvancedProtos

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Name != "高级协议" || r.Comment != "测试" {
		t.Fatalf("协议行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dprotos_l7")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestAdvancedProtosWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestAdvancedProtosWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dprotos_l7", "add", map[string]any{})
	ts.SetResponse("dprotos_l7", "edit", map[string]any{})
	ts.SetResponse("dprotos_l7", "del", map[string]any{})
	svc := New(ts.Client()).AdvancedProtos
	ctx := context.Background()

	rowID, err := svc.Add(ctx, AdvancedProto{Name: "高级协议", Comment: "测试"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("dprotos_l7")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"name", "comment"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, AdvancedProto{ID: 2, Name: "改名"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("dprotos_l7")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}

	if err := svc.Delete(ctx, 2); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if rec = ts.CallsFor("dprotos_l7")[2]; rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 2})
}

// TestAdvancedProtosError 覆盖固件错误路径。
func TestAdvancedProtosError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dprotos_l7", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "dprotos_l7", Action: action}
	})
	svc := New(ts.Client()).AdvancedProtos
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if err := svc.Edit(ctx, AdvancedProto{ID: 1}); err == nil {
		t.Fatal("edit 错误应上抛")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}
