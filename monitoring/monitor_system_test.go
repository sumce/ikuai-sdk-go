package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestSystemCurves 覆盖 monitor_system.show 全曲线成功路径与参数断言。
func TestSystemCurves(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		data   map[string]any
		want   map[string]any
		assert func(t *testing.T, m *SystemMonitor)
	}{
		{
			name: "nil-opts 全部曲线+hour/avg(抓包默认)",
			opts: nil,
			data: map[string]any{
				"cpu":             []any{map[string]any{"cpu": 6, "timestamp": 1794319920}},
				"memory":          []any{map[string]any{"memory": 3212532, "memory_use": 6, "timestamp": 1794319920}},
				"disk_space_used": []any{map[string]any{"disk_space_used": 79, "disk_space_use": 0, "timestamp": 1794319920}},
				"on_terminal":     []any{map[string]any{"on_terminal": 143, "timestamp": 1794319920}},
				"conn_num":        []any{map[string]any{"conn_num": 8777, "timestamp": 1794319920}},
				"cputemp1":        []any{map[string]any{"cputemp1": 39, "timestamp": 1794319920}},
				"cputemp2":        []any{map[string]any{"cputemp2": 0, "timestamp": 1794319920}},
			},
			want: map[string]any{
				"TYPE":     SystemTypeAllCurves,
				"datetype": "hour",
				"math":     "avg",
			},
			assert: func(t *testing.T, m *SystemMonitor) {
				if len(m.CPU) != 1 || m.CPU[0].CPU != 6 {
					t.Fatalf("cpu 曲线解码不符: %+v", m.CPU)
				}
				if m.Memory[0].Memory != 3212532 || m.Memory[0].MemoryUse != 6 {
					t.Fatalf("memory 曲线解码不符: %+v", m.Memory)
				}
				if m.DiskSpaceUsed[0].DiskSpaceUsed != 79 {
					t.Fatalf("disk 曲线解码不符: %+v", m.DiskSpaceUsed)
				}
				if m.OnTerminal[0].OnTerminal != 143 || m.ConnNum[0].ConnNum != 8777 {
					t.Fatalf("终端/连接曲线解码不符: %+v %+v", m.OnTerminal, m.ConnNum)
				}
				if m.CpuTemp1[0].CpuTemp1 != 39 || m.CpuTemp2[0].CpuTemp2 != 0 {
					t.Fatalf("温度曲线解码不符: %+v %+v", m.CpuTemp1, m.CpuTemp2)
				}
			},
		},
		{
			name: "day/max 聚合下 cpu 为小数(真机实证)",
			opts: &ikuai.ListOptions{Extra: map[string]any{"datetype": "day", "math": "max"}},
			data: map[string]any{
				"cpu": []any{map[string]any{"cpu": 11.54, "timestamp": 1794240360}},
			},
			want: map[string]any{
				"TYPE":     SystemTypeAllCurves,
				"datetype": "day",
				"math":     "max",
			},
			assert: func(t *testing.T, m *SystemMonitor) {
				if len(m.CPU) != 1 || m.CPU[0].CPU != 11.54 {
					t.Fatalf("cpu 小数解码不符: %+v", m.CPU)
				}
			},
		},
		{
			name: "opts.Type 单曲线",
			opts: &ikuai.ListOptions{Type: SystemTypeCPU},
			data: map[string]any{
				"cpu": []any{map[string]any{"cpu": 6, "timestamp": 1794319920}},
			},
			want: map[string]any{
				"TYPE":     "cpu",
				"datetype": "hour",
				"math":     "avg",
			},
			assert: func(t *testing.T, m *SystemMonitor) {
				if len(m.CPU) != 1 || m.Memory != nil {
					t.Fatalf("单曲线解码不符: %+v", m)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("monitor_system", "show", tt.data)
			m, err := New(ts.Client()).System.Curves(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Curves 失败: %v", err)
			}
			tt.assert(t, m)
			wantParam(t, paramOf(t, ts.CallsFor("monitor_system")[0]), tt.want)
		})
	}
}

// TestSystemRateStat 覆盖 TYPE=rate_stat 峰值速率曲线。
func TestSystemRateStat(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_system", "show", map[string]any{
		"rate_stat": []any{map[string]any{
			"timestamp": 1794243540, "max_upload": 27483391,
			"max_download": 27942509, "rx_packages": 1754459, "tx_packages": 1768013,
		}},
	})
	out, err := New(ts.Client()).System.RateStat(context.Background(), nil)
	if err != nil {
		t.Fatalf("RateStat 失败: %v", err)
	}
	if len(out.RateStat) != 1 || out.RateStat[0].MaxUpload != 27483391 ||
		out.RateStat[0].RxPackages != 1754459 {
		t.Fatalf("rate_stat 解码不符: %+v", out.RateStat)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_system")[0]), map[string]any{
		"TYPE": "rate_stat", "datetype": "hour", "math": "avg",
	})
}

// TestSystemCpuTempSupport 覆盖 TYPE=cputemp_support。
func TestSystemCpuTempSupport(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_system", "show", map[string]any{"cputemp_support": 1})
	n, err := New(ts.Client()).System.CpuTempSupport(context.Background())
	if err != nil {
		t.Fatalf("CpuTempSupport 失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("cputemp_support = %d, 期望 1", n)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_system")[0]),
		map[string]any{"TYPE": "cputemp_support"})
}

// TestSystemAppFlow 覆盖 TYPE=app_flow 与 minute 参数断言。
func TestSystemAppFlow(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_system", "show", map[string]any{
		"app_flow": []any{map[string]any{
			"Total": 516, "HTTP": 77, "Download": 103, "Transport": 0,
			"IM": 12, "Video": 81, "Common": 66, "Game": 0, "Others": 11,
			"Test": 0, "Unknown": 166,
		}},
	})
	rows, err := New(ts.Client()).System.AppFlow(context.Background(), 30, nil)
	if err != nil {
		t.Fatalf("AppFlow 失败: %v", err)
	}
	if len(rows) != 1 || rows[0].Total != 516 || rows[0].Unknown != 166 {
		t.Fatalf("app_flow 解码不符: %+v", rows)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_system")[0]),
		map[string]any{"TYPE": "app_flow", "minute": "30"})
}

// TestSystemAppFlowInvalidMinutes 非法分钟数返回 ErrInvalidParam。
func TestSystemAppFlowInvalidMinutes(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	if _, err := New(ts.Client()).System.AppFlow(context.Background(), 0, nil); err == nil {
		t.Fatal("minutes=0 应报错")
	}
}

// TestSystemError 覆盖固件错误路径。
func TestSystemError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("monitor_system", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95006, ErrMsg: "失败", FuncName: "monitor_system", Action: action}
	})
	_, err := New(ts.Client()).System.Curves(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95006 {
		t.Fatalf("期望 APIError(95006), got %v", err)
	}
}
