package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// lbpcSample 构造真机探测形态的 lb_pcc.show 响应 Data（TYPE=total,data；
// 探测真机列表为空，行字段为推断形态，见 MultiWanRule 注释）。
func lbpcSample() map[string]any {
	return map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{
				"id": 1, "enabled": "yes", "comment": "电信主线路",
				"interface": "adsl01", "isp": "chinatelecom", "weight": 20,
			},
			map[string]any{
				"id": 2, "enabled": "no", "comment": "联通备线",
				"interface": "adsl02", "isp": "chinaunicom", "weight": 10,
			},
		},
	}
}

// TestMultiWanList 覆盖 lb_pcc.show 分页列表的成功路径与参数断言。
func TestMultiWanList(t *testing.T) {
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
			ts.SetResponse("lb_pcc", "show", lbpcSample())
			page, err := New(ts.Client()).MultiWan.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 2 || len(page.Rows) != 2 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || r.Enabled != EnabledYes || r.Interface != "adsl01" ||
				r.ISP != "chinatelecom" || r.Weight != 20 {
				t.Fatalf("规则行解码不符: %+v", r)
			}
			if !r.IsEnabled() {
				t.Fatalf("IsEnabled 应为 true")
			}
			if page.Rows[1].IsEnabled() {
				t.Fatalf("第 2 行 enabled=no, IsEnabled 应为 false")
			}
			wantParam(t, paramOf(t, ts.CallsFor("lb_pcc")[0]), tt.want)
		})
	}
}

// TestMultiWanSelectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestMultiWanSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("lb_pcc", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface": []any{
			[]any{"adsl01", "移动私网"},
			[]any{"wan1", "电信外网"},
		},
		"isplist": []any{"all", "chinatelecom", "chinaunicom", "chinamobile", "chinacernet"},
	})
	sel, err := New(ts.Client()).MultiWan.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 2 || sel.Interface[0][0] != "adsl01" ||
		sel.Interface[0][1] != "移动私网" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.ISPList) != 5 || sel.ISPList[0] != "all" || sel.ISPList[1] != "chinatelecom" {
		t.Fatalf("运营商选择器解码不符: %v", sel.ISPList)
	}
	wantParam(t, paramOf(t, ts.CallsFor("lb_pcc")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestMultiWanWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestMultiWanWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("lb_pcc", "add", map[string]any{})
	ts.SetResponse("lb_pcc", "edit", map[string]any{})
	ts.SetResponse("lb_pcc", "del", map[string]any{})
	svc := New(ts.Client()).MultiWan
	ctx := context.Background()

	rule := MultiWanRule{Enabled: EnabledYes, Comment: "测试",
		Interface: "adsl01", ISP: "all", Weight: 5}
	rowID, err := svc.Add(ctx, rule)
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("lb_pcc")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "comment", "interface", "isp", "weight"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}
	if len(p) != 5 {
		t.Fatalf("add 参数不应含 id: %v", p)
	}

	if err := svc.Edit(ctx, MultiWanRule{ID: 3, Enabled: EnabledNo}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	rec = ts.CallsFor("lb_pcc")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(3) {
		t.Fatalf("edit 参数应携带 id=3: %v", p)
	}

	if err := svc.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	rec = ts.CallsFor("lb_pcc")[2]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 3})

	if err := svc.Delete(ctx, 3, 4, 5); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	rec = ts.CallsFor("lb_pcc")[3]
	wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{3, 4, 5}})
}

// TestMultiWanError 覆盖固件错误路径（show/add → *ikuai.APIError）。
func TestMultiWanError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("lb_pcc", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "lb_pcc", Action: action}
	})
	svc := New(ts.Client()).MultiWan
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Add(ctx, MultiWanRule{}); err == nil {
		t.Fatal("add 错误应上抛")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}
