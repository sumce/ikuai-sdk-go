package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestPayLogList 覆盖 ppp_paylog.show 的成功路径、真机行解码与参数断言。
func TestPayLogList(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		wantP  map[string]any
		assert func(t *testing.T, page *ikuai.Page[PayLogEntry])
	}{
		{
			name:  "nil-opts 默认分页(真机行形态)",
			opts:  nil,
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,20"},
			assert: func(t *testing.T, page *ikuai.Page[PayLogEntry]) {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				e := page.Rows[0]
				if e.ID != 1 || e.Username != "admin" || e.Admin != "admin" ||
					e.Action != PayLogActionOpen || e.Packname != "自定义" ||
					e.FeeMoney != 0 || e.Time().Unix() != 1791741677 ||
					e.Comment != "自定义,账号到期时间:2029-12-21," {
					t.Fatalf("真机行字段解码不符: %+v", e)
				}
			},
		},
		{
			name: "关键字搜索",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 10,
				Finds: "username,name,admin,comment", Keywords: "开户"},
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,10",
				"FINDS": "username,name,admin,comment", "KEYWORDS": "开户"},
			assert: func(t *testing.T, page *ikuai.Page[PayLogEntry]) {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ppp_paylog", "show", map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"packname": "自定义", "feemoney": 0, "id": 1,
					"timestamp": 1791741677, "admin": "admin", "name": "",
					"username": "admin", "comment": "自定义,账号到期时间:2029-12-21,",
					"action": "开户",
				}},
			})
			page, err := New(ts.Client()).PayLog.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			tt.assert(t, page)
			wantParam(t, paramOf(t, ts.CallsFor("ppp_paylog")[0]), tt.wantP)
		})
	}
}

// TestPayLogError 覆盖固件错误路径。
func TestPayLogError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ppp_paylog", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95000, ErrMsg: "失败", FuncName: "ppp_paylog", Action: action}
	})
	_, err := New(ts.Client()).PayLog.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95000 {
		t.Fatalf("期望 APIError(95000), got %v", err)
	}
}
