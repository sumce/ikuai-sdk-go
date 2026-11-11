package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestOnlineUsersList 覆盖 ppp_online.show 的成功路径与参数断言。
func TestOnlineUsersList(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		data   map[string]any
		wantP  map[string]any
		assert func(t *testing.T, page *ikuai.Page[OnlineUser])
	}{
		{
			name:  "nil-opts 默认分页(真机空表形态)",
			opts:  nil,
			data:  map[string]any{"total": 0, "data": []any{}},
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,20"},
			assert: func(t *testing.T, page *ikuai.Page[OnlineUser]) {
				if page.Total != 0 || len(page.Rows) != 0 {
					t.Fatalf("空列表解码不符: %+v", page)
				}
			},
		},
		{
			name: "带行数据解码(Raw 原样保留)",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 5},
			data: map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"username": "u1", "ip": "192.168.1.66", "upload": 1024,
				}},
			},
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,5"},
			assert: func(t *testing.T, page *ikuai.Page[OnlineUser]) {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["username"] != "u1" || m["ip"] != "192.168.1.66" {
					t.Fatalf("行内容不符: %v", m)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ppp_online", "show", tt.data)
			page, err := New(ts.Client()).OnlineUsers.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			tt.assert(t, page)
			wantParam(t, paramOf(t, ts.CallsFor("ppp_online")[0]), tt.wantP)
		})
	}
}

// TestOnlineUsersListError 覆盖固件错误路径。
func TestOnlineUsersListError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ppp_online", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (x)", FuncName: "ppp_online", Action: action}
	})
	_, err := New(ts.Client()).OnlineUsers.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
