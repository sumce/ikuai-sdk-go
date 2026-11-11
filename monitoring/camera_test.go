package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestCameraList 覆盖 camera.show 的成功路径与参数断言。
func TestCameraList(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		want   map[string]any
		assert func(t *testing.T, page *ikuai.Page[CameraDevice])
	}{
		{
			name: "nil-opts 默认分页(抓包形态,无摄像头)",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
			assert: func(t *testing.T, page *ikuai.Page[CameraDevice]) {
				if page.Total != 0 || len(page.Rows) != 0 {
					t.Fatalf("空列表解码不符: %+v", page)
				}
			},
		},
		{
			name: "带行数据解码(Raw 原样保留)",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 5},
			want: map[string]any{"TYPE": "total,data", "limit": "0,5"},
			assert: func(t *testing.T, page *ikuai.Page[CameraDevice]) {
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["ip"] != "192.168.1.100" || m["mac"] != "aa:bb:cc:dd:ee:ff" {
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
						"ip": "192.168.1.100", "mac": "aa:bb:cc:dd:ee:ff",
					}},
				}
			}
			ts.SetResponse("camera", "show", data)
			page, err := New(ts.Client()).Camera.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			tt.assert(t, page)
			wantParam(t, paramOf(t, ts.CallsFor("camera")[0]), tt.want)
		})
	}
}

// TestCameraError 覆盖固件错误路径。
func TestCameraError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("camera", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95008, ErrMsg: "失败", FuncName: "camera", Action: action}
	})
	_, err := New(ts.Client()).Camera.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95008 {
		t.Fatalf("期望 APIError(95008), got %v", err)
	}
}
