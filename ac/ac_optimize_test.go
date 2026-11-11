package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestNetOptimizeGet 覆盖 ac_net_optimize.show 的成功路径与参数断言
// （真机探测实证返回 {"total","data","mode","frequencies"}）。
func TestNetOptimizeGet(t *testing.T) {
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
			name: "仅取 data 区块",
			opts: &ikuai.ListOptions{Type: "data"},
			want: map[string]any{"TYPE": "data"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ac_net_optimize", "show", map[string]any{
				"total": 1, "data": []any{map[string]any{"bogus_row": 1}},
				"mode": 0, "frequencies": 0,
			})
			info, err := New(ts.Client()).NetOptimize.Get(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Get 失败: %v", err)
			}
			if info.Total != 1 || info.Mode != 0 || info.Frequencies != 0 || len(info.Data) != 1 {
				t.Fatalf("解码不符: %+v", info)
			}
			m, err := info.Data[0].Map()
			if err != nil || m["bogus_row"] != float64(1) {
				t.Fatalf("ACOptimizeItem.Map 解码不符: %v %v", m, err)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ac_net_optimize")[0]), tt.want)
		})
	}
}

// TestNetOptimizeSet 覆盖 set 写动作参数形态（⚠️ 写动作，仅离线；
// 字段语义待真机验证）。
func TestNetOptimizeSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_net_optimize", "set", map[string]any{})
	if err := New(ts.Client()).NetOptimize.Set(context.Background(),
		NetOptimizeConfig{Mode: 1, Frequencies: 0}); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("ac_net_optimize")[0]
	if rec.Action != "set" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	wantParam(t, paramOf(t, rec),
		map[string]any{"mode": float64(1), "frequencies": float64(0)})
}

// TestNetOptimizeError 覆盖固件错误路径。
func TestNetOptimizeError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ac_net_optimize", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "ac_net_optimize", Action: action}
	})
	if _, err := New(ts.Client()).NetOptimize.Get(context.Background(), nil); err == nil {
		t.Fatal("Get 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if err := New(ts.Client()).NetOptimize.Set(context.Background(), NetOptimizeConfig{}); err == nil {
		t.Fatal("Set 应失败")
	}
}
