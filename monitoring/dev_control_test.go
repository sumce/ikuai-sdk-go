package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestDevControlList 覆盖 dev_control.show 的成功路径与参数断言。
func TestDevControlList(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		want   map[string]any
		assert func(t *testing.T, page *ikuai.Page[DevControlDevice])
	}{
		{
			name: "nil-opts 默认分页(无智能设备)",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
			assert: func(t *testing.T, page *ikuai.Page[DevControlDevice]) {
				if page.Total != 0 || len(page.Rows) != 0 {
					t.Fatalf("空列表解码不符: %+v", page)
				}
			},
		},
		{
			name: "带行数据解码(Raw 原样保留)",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 10},
			want: map[string]any{"TYPE": "total,data", "limit": "0,10"},
			assert: func(t *testing.T, page *ikuai.Page[DevControlDevice]) {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["mac"] != "11:22:33:44:55:66" || m["online"] != float64(1) {
					t.Fatalf("行内容不符: %v", m)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			data := map[string]any{"total": 0, "data": []any{}}
			if tt.name == "带行数据解码(Raw 原样保留)" {
				data = map[string]any{
					"total": 1,
					"data": []any{map[string]any{
						"mac": "11:22:33:44:55:66", "online": 1,
					}},
				}
			}
			ts.SetResponse("dev_control", "show", data)
			page, err := New(ts.Client()).DevControl.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			tt.assert(t, page)
			wantParam(t, paramOf(t, ts.CallsFor("dev_control")[0]), tt.want)
		})
	}
}

// TestDevControlCheckOnlineStatus 覆盖 dev_control.check_online_status
// 的成功路径与参数断言（抓包实证 param 为空对象）。
func TestDevControlCheckOnlineStatus(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dev_control", "check_online_status", nil)
	if err := New(ts.Client()).DevControl.CheckOnlineStatus(context.Background()); err != nil {
		t.Fatalf("CheckOnlineStatus 失败: %v", err)
	}
	rec := ts.CallsFor("dev_control")[0]
	if rec.FuncName != "dev_control" || rec.Action != "check_online_status" {
		t.Fatalf("func/action = %s/%s, 期望 dev_control/check_online_status",
			rec.FuncName, rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestDevControlError 覆盖固件错误路径。
func TestDevControlError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dev_control", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95010, ErrMsg: "失败", FuncName: "dev_control", Action: action}
	})
	_, err := New(ts.Client()).DevControl.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95010 {
		t.Fatalf("期望 APIError(95010), got %v", err)
	}
	if err := New(ts.Client()).DevControl.CheckOnlineStatus(context.Background()); err == nil {
		t.Fatal("check_online_status 错误应上抛")
	}
}
