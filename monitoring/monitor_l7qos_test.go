package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestL7QoSStatus 覆盖 monitor_l7qos.show 的成功路径与参数断言。
// 真机与抓包中 data/interface 均可能为空数组，此处同时覆盖空与非空形态。
func TestL7QoSStatus(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		data   map[string]any
		want   map[string]any
		assert func(t *testing.T, st *L7QoSStatus)
	}{
		{
			name: "nil-opts 抓包默认参数(空数据)",
			opts: nil,
			data: map[string]any{"data": []any{}, "interface": []any{}},
			want: map[string]any{"TYPE": L7QoSTypeDefault, "interface": ""},
			assert: func(t *testing.T, st *L7QoSStatus) {
				if len(st.Data) != 0 || len(st.Interface) != 0 {
					t.Fatalf("空数据应解码为空: %+v", st)
				}
			},
		},
		{
			name: "opts.Extra 指定接口",
			opts: &ikuai.ListOptions{Extra: map[string]any{"interface": "all"}},
			data: map[string]any{"data": []any{}, "interface": []any{}},
			want: map[string]any{"TYPE": L7QoSTypeDefault, "interface": "all"},
			assert: func(t *testing.T, st *L7QoSStatus) {
				if st == nil {
					t.Fatal("结果不应为 nil")
				}
			},
		},
		{
			name: "仅 data 区块",
			opts: &ikuai.ListOptions{Type: L7QoSTypeData},
			data: map[string]any{"data": []any{map[string]any{"id": 1}}},
			want: map[string]any{"TYPE": "data", "interface": ""},
			assert: func(t *testing.T, st *L7QoSStatus) {
				if len(st.Data) != 1 {
					t.Fatalf("data 行数不符: %s", st.Data)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("monitor_l7qos", "show", tt.data)
			st, err := New(ts.Client()).L7QoS.Status(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Status 失败: %v", err)
			}
			tt.assert(t, st)
			wantParam(t, paramOf(t, ts.CallsFor("monitor_l7qos")[0]), tt.want)
		})
	}
}

// TestL7QoSError 覆盖固件错误路径。
func TestL7QoSError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("monitor_l7qos", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95005, ErrMsg: "失败", FuncName: "monitor_l7qos", Action: action}
	})
	_, err := New(ts.Client()).L7QoS.Status(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95005 {
		t.Fatalf("期望 APIError(95005), got %v", err)
	}
}
