package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// macCommentSample 构造真机探测形态的 mac_comment.show 响应 Data
// （TYPE=total,data；真机 3.7.24 实测行结构 mac/comment/id）。
func macCommentSample() map[string]any {
	return map[string]any{
		"total": 9,
		"data": []any{
			map[string]any{"mac": "e0:d5:5e:cd:5b:20", "comment": "R", "id": 2},
			map[string]any{"mac": "00:0c:29:3b:11:af", "comment": "虚拟机", "id": 3},
		},
	}
}

// TestMACCommentList 覆盖 mac_comment.show 分页的成功路径与参数断言。
func TestMACCommentList(t *testing.T) {
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
		{
			name: "关键字搜索",
			opts: &ikuai.ListOptions{Finds: "mac,comment", Keywords: "虚拟机"},
			want: map[string]any{"TYPE": "total,data",
				"FINDS": "mac,comment", "KEYWORDS": "虚拟机"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("mac_comment", "show", macCommentSample())
			page, err := New(ts.Client()).MACComment.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 9 || len(page.Rows) != 2 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 2 || r.MAC != "e0:d5:5e:cd:5b:20" || r.Comment != "R" {
				t.Fatalf("备注行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("mac_comment")[0]), tt.want)
		})
	}
}

// TestMACCommentWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestMACCommentWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("mac_comment", "add", map[string]any{})
	ts.SetResponse("mac_comment", "edit", map[string]any{})
	ts.SetResponse("mac_comment", "del", map[string]any{})
	svc := New(ts.Client()).MACComment
	ctx := context.Background()

	rowID, err := svc.Add(ctx, MACComment{MAC: "aa:bb:cc:dd:ee:ff", Comment: "测试终端"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("mac_comment")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if len(p) != 2 || p["mac"] != "aa:bb:cc:dd:ee:ff" || p["comment"] != "测试终端" {
		t.Fatalf("add 参数不符: %v", p)
	}

	if err := svc.Edit(ctx, MACComment{ID: 3, MAC: "00:0c:29:3b:11:af", Comment: "改名"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	rec = ts.CallsFor("mac_comment")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(3) || p["comment"] != "改名" {
		t.Fatalf("edit 参数不符: %v", p)
	}

	if err := svc.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	rec = ts.CallsFor("mac_comment")[2]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": 3})

	if err := svc.Delete(ctx, 3, 4, 5); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	rec = ts.CallsFor("mac_comment")[3]
	wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{3, 4, 5}})
}

// TestMACCommentError 覆盖固件错误路径（show/add → *ikuai.APIError）。
func TestMACCommentError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("mac_comment", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "mac_comment", Action: action}
	})
	svc := New(ts.Client()).MACComment
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Add(ctx, MACComment{}); err == nil {
		t.Fatal("add 错误应上抛")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}
