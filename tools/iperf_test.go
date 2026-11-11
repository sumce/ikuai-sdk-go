package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// iperfSample 构造 iperf.show 的响应 Data（形态为真机 3.7.24 实测）。
func iperfSample() map[string]any {
	return map[string]any{
		"host": "", "port": 5201, "threads": 1, "dir": "upload",
		"interface": "auto", "l3proto": "ipv4", "l4proto": "udp",
		"time": 10, "interval": 1, "bandwidth": "100",
		"data": []any{map[string]any{"status": 0, "response": ""}},
	}
}

// TestIPerfResult 覆盖 iperf.show 空参数查询最近一次测速结果。
func TestIPerfResult(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("iperf", "show", iperfSample())
	state, err := New(ts.Client()).IPerf.Result(context.Background())
	if err != nil {
		t.Fatalf("Result 失败: %v", err)
	}
	if state.Host != "" || state.Port != 5201 || state.Threads != 1 ||
		state.Dir != IPerfDirUpload || state.Interface != "auto" ||
		state.L3Proto != DiagStackIPv4 || state.L4Proto != DiagL4UDP ||
		state.Time != 10 || state.Interval != 1 || state.Bandwidth != "100" {
		t.Fatalf("任务配置解码不符: %+v", state)
	}
	if len(state.Data) != 1 || state.Data[0].Status != 0 {
		t.Fatalf("结果行解码不符: %+v", state.Data)
	}
	rec := ts.CallsFor("iperf")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %s, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestIPerfExecute 覆盖 iperf.show 带参发起测速的参数断言（仅离线）。
func TestIPerfExecute(t *testing.T) {
	tests := []struct {
		name string
		req  IPerfRequest
		want map[string]any
	}{
		{
			name: "UDP 上行默认参数",
			req: IPerfRequest{
				Host: "192.168.1.2", Port: 5201, Threads: 1, Dir: IPerfDirUpload,
				Interface: "auto", L3Proto: DiagStackIPv4, L4Proto: DiagL4UDP,
				Time: 10, Interval: 1, Bandwidth: "100",
			},
			want: map[string]any{
				"host": "192.168.1.2", "port": float64(5201), "threads": float64(1),
				"dir": "upload", "interface": "auto", "l3proto": "ipv4",
				"l4proto": "udp", "time": float64(10), "interval": float64(1),
				"bandwidth": "100",
			},
		},
		{
			name: "TCP 下行多线程",
			req: IPerfRequest{
				Host: "192.168.1.2", Port: 5201, Threads: 4, Dir: IPerfDirDownload,
				Interface: "lan1", L3Proto: DiagStackIPv4, L4Proto: DiagL4TCP,
				Time: 30, Interval: 5, Bandwidth: "0",
			},
			want: map[string]any{
				"host": "192.168.1.2", "port": float64(5201), "threads": float64(4),
				"dir": "download", "interface": "lan1", "l3proto": "ipv4",
				"l4proto": "tcp", "time": float64(30), "interval": float64(5),
				"bandwidth": "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("iperf", "show", iperfSample())
			state, err := New(ts.Client()).IPerf.Execute(context.Background(), tt.req)
			if err != nil {
				t.Fatalf("Execute 失败: %v", err)
			}
			if state.Port != 5201 || len(state.Data) != 1 {
				t.Fatalf("返回状态解码不符: %+v", state)
			}
			rec := ts.CallsFor("iperf")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %s, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), tt.want)
		})
	}
}

// TestIPerfStop 覆盖 iperf.stop 停止测速的参数断言（仅离线）。
func TestIPerfStop(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("iperf", "stop", map[string]any{})
	if err := New(ts.Client()).IPerf.Stop(context.Background()); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	rec := ts.CallsFor("iperf")[0]
	if rec.Action != "stop" {
		t.Fatalf("action = %s, 期望 stop", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestIPerfError 覆盖固件错误路径。
func TestIPerfError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("iperf", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "iperf", Action: action}
	})
	svc := New(ts.Client()).IPerf
	if _, err := svc.Result(context.Background()); err == nil {
		t.Fatal("Result 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.Execute(context.Background(), IPerfRequest{Host: "192.168.1.2"}); err == nil {
		t.Fatal("Execute 应失败")
	}
	if err := svc.Stop(context.Background()); err == nil {
		t.Fatal("Stop 应失败")
	}
}
