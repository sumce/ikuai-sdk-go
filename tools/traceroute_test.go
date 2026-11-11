package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// tracerouteSample 构造 Traceroute.show 的响应 Data（空闲形态为真机
// 3.7.24 实测；response 内容按 traceroute 输出语义构造）。
func tracerouteSample(withResult bool) map[string]any {
	resp := ""
	if withResult {
		resp = "traceroute to 223.5.5.5 (223.5.5.5), 30 hops max, 60 byte packets\n" +
			" 1  192.168.1.1  1.02 ms\n"
	}
	return map[string]any{
		"host": "223.5.5.5", "interface": "", "max": 30, "wait": 5, "dns": "",
		"data": []any{map[string]any{"status": 0, "response": resp}},
	}
}

// TestTracerouteResult 覆盖 Traceroute.show 空参数查询（含空闲态）。
func TestTracerouteResult(t *testing.T) {
	tests := []struct {
		name       string
		withResult bool
	}{
		{name: "空闲态", withResult: false},
		{name: "已有结果", withResult: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("Traceroute", "show", tracerouteSample(tt.withResult))
			state, err := New(ts.Client()).Traceroute.Result(context.Background())
			if err != nil {
				t.Fatalf("Result 失败: %v", err)
			}
			if state.Host != "223.5.5.5" || state.Max != 30 || state.Wait != 5 ||
				state.Interface != "" || state.DNS != "" {
				t.Fatalf("任务配置解码不符: %+v", state)
			}
			if len(state.Data) != 1 || state.Data[0].Status != 0 {
				t.Fatalf("结果行解码不符: %+v", state.Data)
			}
			if tt.withResult && state.Data[0].Response == "" {
				t.Fatal("结果态 response 不应为空")
			}
			if !tt.withResult && state.Data[0].Response != "" {
				t.Fatalf("空闲态 response 应为空: %q", state.Data[0].Response)
			}
			rec := ts.CallsFor("Traceroute")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %s, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), map[string]any{})
		})
	}
}

// TestTracerouteExecute 覆盖 Traceroute.show 带参发起检测的参数断言
// （仅离线）。
func TestTracerouteExecute(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("Traceroute", "show", tracerouteSample(true))
	req := TracerouteRequest{Host: "223.5.5.5", Interface: "auto", Max: 30, Wait: 5, DNS: ""}
	state, err := New(ts.Client()).Traceroute.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}
	if state.Host != "223.5.5.5" || state.Max != 30 {
		t.Fatalf("返回状态解码不符: %+v", state)
	}
	rec := ts.CallsFor("Traceroute")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %s, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{
		"host": "223.5.5.5", "interface": "auto", "max": float64(30),
		"wait": float64(5), "dns": "",
	})
}

// TestTracerouteStop 覆盖 Traceroute.stop_test 停止检测的参数断言
// （仅离线）。
func TestTracerouteStop(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("Traceroute", "stop_test", map[string]any{})
	if err := New(ts.Client()).Traceroute.Stop(context.Background()); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	rec := ts.CallsFor("Traceroute")[0]
	if rec.Action != "stop_test" {
		t.Fatalf("action = %s, 期望 stop_test", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestTracerouteError 覆盖固件错误路径。
func TestTracerouteError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("Traceroute", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "Traceroute", Action: action}
	})
	svc := New(ts.Client()).Traceroute
	if _, err := svc.Result(context.Background()); err == nil {
		t.Fatal("Result 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.Execute(context.Background(), TracerouteRequest{Host: "223.5.5.5"}); err == nil {
		t.Fatal("Execute 应失败")
	}
	if err := svc.Stop(context.Background()); err == nil {
		t.Fatal("Stop 应失败")
	}
}
