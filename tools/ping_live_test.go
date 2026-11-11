package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLivePingResult 真机只读验证 Ping 检测状态查询（Ping.show 空参，
// 返回最近一次检测的任务配置与结果；不发起任何新检测）。
func TestLivePingResult(t *testing.T) {
	c := livetest.Client(t)
	state, err := New(c).Ping.Result(context.Background())
	if err != nil {
		t.Fatalf("Ping.show 失败: %v", err)
	}
	t.Logf("host=%q interface=%q proto=%q count=%d l4proto=%q port=%q dns=%q",
		state.Host, state.Interface, state.Proto, state.Count, state.L4Proto, state.Port, state.DNS)
	if state.Count <= 0 {
		t.Errorf("count 应为正数: %d", state.Count)
	}
	if len(state.Data) == 0 {
		t.Fatal("结果行 data 不应为空数组")
	}
	for i, r := range state.Data {
		t.Logf("  行 %d: status=%d response=%.120q", i, r.Status, r.Response)
	}
}

// TestLiveTracerouteResult 真机只读验证路由追踪状态查询
// （Traceroute.show 空参；不发起任何新检测）。
func TestLiveTracerouteResult(t *testing.T) {
	c := livetest.Client(t)
	state, err := New(c).Traceroute.Result(context.Background())
	if err != nil {
		t.Fatalf("Traceroute.show 失败: %v", err)
	}
	t.Logf("host=%q interface=%q max=%d wait=%d dns=%q",
		state.Host, state.Interface, state.Max, state.Wait, state.DNS)
	if len(state.Data) == 0 {
		t.Fatal("结果行 data 不应为空数组")
	}
	for i, r := range state.Data {
		t.Logf("  行 %d: status=%d response=%.120q", i, r.Status, r.Response)
	}
}
