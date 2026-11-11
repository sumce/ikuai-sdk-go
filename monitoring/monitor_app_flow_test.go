package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// appFlowHistorySample 构造抓包形态的 app_history 响应 Data。
func appFlowHistorySample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"Total_up": 27396288, "Total_down": 27894229, "Total": 55290517,
			"id": 65981, "timestamp": 1794239520,
			"HTTP_up": 2921126, "HTTP_down": 1669029,
			"Download_up": 2979668, "Download_down": 134524,
			"Transport_up": 0, "Transport_down": 0,
			"IM_up": 834, "IM_down": 4226,
			"Video_up": 26280, "Video_down": 25783,
			"Common_up": 24197, "Common_down": 23336,
			"Game_up": 11063, "Game_down": 13903,
			"Others_up": 8784, "Others_down": 26145,
			"Test_up": 0, "Test_down": 0,
			"Unknown_up": 21424334, "Unknown_down": 259,
		}},
	}
}

// TestAppFlowHistory 覆盖 monitor_app_flow.show TYPE=app_history。
func TestAppFlowHistory(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 抓包默认参数",
			opts: nil,
			want: map[string]any{
				"TYPE":       "app_history",
				"interface":  "all",
				"datetype":   "day",
				"time_range": "all",
			},
		},
		{
			name: "opts.Extra 覆盖 datetype/time_range",
			opts: &ikuai.ListOptions{Extra: map[string]any{
				"datetype": "hour", "time_range": "day",
			}},
			want: map[string]any{
				"TYPE":       "app_history",
				"interface":  "all",
				"datetype":   "hour",
				"time_range": "day",
			},
		},
		{
			name: "opts.Extra 覆盖 interface",
			opts: &ikuai.ListOptions{Extra: map[string]any{"interface": "adsl01"}},
			want: map[string]any{
				"TYPE":       "app_history",
				"interface":  "adsl01",
				"datetype":   "day",
				"time_range": "all",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("monitor_app_flow", "show", appFlowHistorySample())
			page, err := New(ts.Client()).AppFlow.History(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("History 失败: %v", err)
			}
			if len(page.Rows) != 1 {
				t.Fatalf("行数不符: %d", len(page.Rows))
			}
			r := page.Rows[0]
			if r.Total != 55290517 || r.HTTPUp != 2921126 || r.UnknownUp != 21424334 ||
				r.ID != 65981 || r.Timestamp != 1794239520 {
				t.Fatalf("历史行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("monitor_app_flow")[0]), tt.want)
		})
	}
}

// TestAppFlowFlow 覆盖 TYPE=flow 协议排行。
func TestAppFlowFlow(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_app_flow", "show", map[string]any{
		"data": []any{map[string]any{
			"id": 1, "app_name": "Unknown", "connect_num": 6932,
			"upload": 14401992, "download": 15826743, "total": 30228735,
			"total_up": 29582373621781, "total_down": 27821102153207,
		}},
	})
	page, err := New(ts.Client()).AppFlow.Flow(context.Background(), nil)
	if err != nil {
		t.Fatalf("Flow 失败: %v", err)
	}
	if len(page.Rows) != 1 || page.Rows[0].AppName != "Unknown" ||
		page.Rows[0].TotalUp != 29582373621781 || page.Rows[0].Total != 30228735 {
		t.Fatalf("协议行解码不符: %+v", page.Rows)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_app_flow")[0]),
		map[string]any{"TYPE": AppFlowTypeFlow})
}

// TestAppFlowInterfaces 覆盖 TYPE=interface,app_history_cleantime。
func TestAppFlowInterfaces(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_app_flow", "show", map[string]any{
		"interface": []any{
			[]any{"all"},
			[]any{"adsl01", "移动私网"},
		},
		"app_history_cleantime": 0,
	})
	out, err := New(ts.Client()).AppFlow.Interfaces(context.Background())
	if err != nil {
		t.Fatalf("Interfaces 失败: %v", err)
	}
	if out.CleanTime != 0 || len(out.Interface) != 2 || out.Interface[1][1] != "移动私网" {
		t.Fatalf("接口/清零时间解码不符: %+v", out)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_app_flow")[0]),
		map[string]any{"TYPE": "interface,app_history_cleantime"})
}

// TestAppFlowError 覆盖固件错误路径。
func TestAppFlowError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("monitor_app_flow", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95004, ErrMsg: "失败", FuncName: "monitor_app_flow", Action: action}
	})
	_, err := New(ts.Client()).AppFlow.History(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95004 {
		t.Fatalf("期望 APIError(95004), got %v", err)
	}
}
