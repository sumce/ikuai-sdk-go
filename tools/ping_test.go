package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// pingSample 构造 Ping.show 的响应 Data（形态为真机 3.7.24 实测）。
func pingSample() map[string]any {
	return map[string]any{
		"host": "www.baidu.com", "interface": "adsl08", "proto": "ipv4",
		"count": 10, "l4proto": "icmp", "port": "", "dns": "",
		"data": []any{
			map[string]any{
				"status": 0,
				"response": "PING www.baidu.com (220.181.111.1) from adsl08: 56(84) bytes of data.\n" +
					"32 bytes from 220.181.111.1: icmp_req=1 ttl=54 time=18.09 ms\n" +
					"--- www.baidu.com ping statistics ---\n" +
					"2 packets transmitted, 2 received, 0% packet loss, time 2s\n",
			},
		},
	}
}

// TestPingResult 覆盖 Ping.show 空参数查询最近一次检测结果的成功路径
// 与参数断言。
func TestPingResult(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("Ping", "show", pingSample())
	state, err := New(ts.Client()).Ping.Result(context.Background())
	if err != nil {
		t.Fatalf("Result 失败: %v", err)
	}
	if state.Host != "www.baidu.com" || state.Interface != "adsl08" ||
		state.Proto != DiagStackIPv4 || state.Count != 10 ||
		state.L4Proto != DiagL4ICMP || state.Port != "" || state.DNS != "" {
		t.Fatalf("任务配置解码不符: %+v", state)
	}
	if len(state.Data) != 1 || state.Data[0].Status != 0 ||
		len(state.Data[0].Response) == 0 {
		t.Fatalf("结果行解码不符: %+v", state.Data)
	}
	rec := ts.CallsFor("Ping")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %s, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestPingExecute 覆盖 Ping.show 带参发起检测的参数断言（仅离线）。
func TestPingExecute(t *testing.T) {
	tests := []struct {
		name string
		req  PingRequest
		want map[string]any
	}{
		{
			name: "ICMP 探测",
			req: PingRequest{
				Host: "223.5.5.5", Interface: "auto", Proto: DiagStackIPv4,
				Count: 4, L4Proto: DiagL4ICMP, Port: "", DNS: "",
			},
			want: map[string]any{
				"host": "223.5.5.5", "interface": "auto", "proto": "ipv4",
				"count": float64(4), "l4proto": "icmp", "port": "", "dns": "",
			},
		},
		{
			name: "TCP 探测带端口",
			req: PingRequest{
				Host: "example.com", Interface: "lan1", Proto: DiagStackIPv4,
				Count: 3, L4Proto: DiagL4TCP, Port: "443", DNS: "",
			},
			want: map[string]any{
				"host": "example.com", "interface": "lan1", "proto": "ipv4",
				"count": float64(3), "l4proto": "tcp", "port": "443", "dns": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("Ping", "show", pingSample())
			state, err := New(ts.Client()).Ping.Execute(context.Background(), tt.req)
			if err != nil {
				t.Fatalf("Execute 失败: %v", err)
			}
			if state.Host != "www.baidu.com" || len(state.Data) != 1 {
				t.Fatalf("返回状态解码不符: %+v", state)
			}
			rec := ts.CallsFor("Ping")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %s, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), tt.want)
		})
	}
}

// TestPingStop 覆盖 Ping.stop_test 停止检测的参数断言（仅离线）。
func TestPingStop(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("Ping", "stop_test", map[string]any{})
	if err := New(ts.Client()).Ping.Stop(context.Background()); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	rec := ts.CallsFor("Ping")[0]
	if rec.Action != "stop_test" {
		t.Fatalf("action = %s, 期望 stop_test", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestPingError 覆盖固件错误路径（Result/Execute/Stop → *ikuai.APIError）。
func TestPingError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("Ping", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "Ping", Action: action}
	})
	svc := New(ts.Client()).Ping
	if _, err := svc.Result(context.Background()); err == nil {
		t.Fatal("Result 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.Execute(context.Background(), PingRequest{Host: "223.5.5.5"}); err == nil {
		t.Fatal("Execute 应失败")
	}
	if err := svc.Stop(context.Background()); err == nil {
		t.Fatal("Stop 应失败")
	}
}
