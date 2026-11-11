package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// acGroupSample 构造 ac_group.show 的分页响应 Data（行字段空表推导）。
func acGroupSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 2, "name": "办公区", "ap_count": 8,
			"channel": "auto", "max_pc": 32,
		}},
	}
}

// TestGroupList 覆盖 ac_group.show 的成功路径与参数断言。
func TestGroupList(t *testing.T) {
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
			name: "关键字过滤",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 10, OrderBy: "id", Order: "asc"},
			want: map[string]any{"TYPE": "total,data", "limit": "0,10", "ORDER_BY": "id", "ORDER": "asc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ac_group", "show", acGroupSample())
			page, err := New(ts.Client()).Group.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			g := page.Rows[0]
			if g.ID != 2 || g.Name != "办公区" || g.APCount != 8 ||
				g.Channel != "auto" || g.MaxPC != 32 {
				t.Fatalf("分组行解码不符: %+v", g)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ac_group")[0]), tt.want)
		})
	}
}

// TestGroupAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestGroupAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_group", "add", map[string]any{})
	ts.SetResponse("ac_group", "edit", map[string]any{})
	ts.SetResponse("ac_group", "del", map[string]any{})
	svc := New(ts.Client()).Group
	ctx := context.Background()

	id, err := svc.Add(ctx, APGroup{Name: "办公区", MaxPC: 32})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	_ = id
	p := paramOf(t, ts.CallsFor("ac_group")[0])
	if p["name"] != "办公区" || p["max_pc"] != float64(32) {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, hasID := p["id"]; hasID {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, APGroup{ID: 2, Name: "新名称"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 2); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if err := svc.Delete(ctx, 2, 3, 4); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	recs := ts.CallsFor("ac_group")
	if recs[1].Action != "edit" || recs[2].Action != "del" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action)
	}
	if _, ok := paramOf(t, recs[3])["ids"]; !ok {
		t.Fatalf("多 ID del 应使用 ids: %v", paramOf(t, recs[3]))
	}
}

// TestGroupAddValidation 覆盖 Add 的参数校验（name 必填）。
func TestGroupAddValidation(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client()).Group
	if _, err := svc.Add(context.Background(), APGroup{}); !errIsInvalidParam(err) {
		t.Fatalf("空 name 应返回 ErrInvalidParam, got %v", err)
	}
	if _, err := svc.Add(context.Background(), APGroup{Name: "  "}); !errIsInvalidParam(err) {
		t.Fatalf("空白 name 应返回 ErrInvalidParam, got %v", err)
	}
	if len(ts.CallsFor("ac_group")) != 0 {
		t.Fatal("参数校验失败不应发起调用")
	}
}

// TestGroupError 覆盖固件错误路径。
func TestGroupError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ac_group", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "ac_group", Action: action}
	})
	if _, err := New(ts.Client()).Group.Add(context.Background(), APGroup{Name: "x"}); err == nil {
		t.Fatal("Add 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := New(ts.Client()).Group.List(context.Background(), nil); err == nil {
		t.Fatal("List 应失败")
	}
}
