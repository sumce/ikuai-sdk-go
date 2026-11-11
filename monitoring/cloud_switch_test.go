package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestCloudSwitchList 覆盖 cloud_switch.show 的成功路径与参数断言
// （默认按 id 降序，抓包实证）。
func TestCloudSwitchList(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		want   map[string]any
		assert func(t *testing.T, page *ikuai.Page[CloudSwitchDevice])
	}{
		{
			name: "nil-opts 默认分页+id 降序(无交换机)",
			opts: nil,
			want: map[string]any{
				"TYPE": "total,data", "limit": "0,20",
				"ORDER_BY": "id", "ORDER": "desc",
			},
			assert: func(t *testing.T, page *ikuai.Page[CloudSwitchDevice]) {
				if page.Total != 0 || len(page.Rows) != 0 {
					t.Fatalf("空列表解码不符: %+v", page)
				}
			},
		},
		{
			name: "带行数据解码(Raw 原样保留)",
			opts: nil,
			want: map[string]any{
				"TYPE": "total,data", "limit": "0,20",
				"ORDER_BY": "id", "ORDER": "desc",
			},
			assert: func(t *testing.T, page *ikuai.Page[CloudSwitchDevice]) {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["name"] != "sw-01" || m["ip"] != "192.168.2.2" {
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
					"data":  []any{map[string]any{"name": "sw-01", "ip": "192.168.2.2"}},
				}
			}
			ts.SetResponse("cloud_switch", "show", data)
			page, err := New(ts.Client()).CloudSwitch.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			tt.assert(t, page)
			wantParam(t, paramOf(t, ts.CallsFor("cloud_switch")[0]), tt.want)
		})
	}
}

// TestCloudSwitchError 覆盖固件错误路径。
func TestCloudSwitchError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("cloud_switch", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95009, ErrMsg: "失败", FuncName: "cloud_switch", Action: action}
	})
	_, err := New(ts.Client()).CloudSwitch.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95009 {
		t.Fatalf("期望 APIError(95009), got %v", err)
	}
}
