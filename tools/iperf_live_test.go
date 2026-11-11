package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveIPerfResult 真机只读验证内网测速状态查询（iperf.show 空参，
// 返回最近一次测速的任务配置与结果；不发起任何新测速）。
func TestLiveIPerfResult(t *testing.T) {
	c := livetest.Client(t)
	state, err := New(c).IPerf.Result(context.Background())
	if err != nil {
		t.Fatalf("iperf.show 失败: %v", err)
	}
	t.Logf("host=%q port=%d threads=%d dir=%q interface=%q l3proto=%q l4proto=%q time=%d interval=%d bandwidth=%q",
		state.Host, state.Port, state.Threads, state.Dir, state.Interface,
		state.L3Proto, state.L4Proto, state.Time, state.Interval, state.Bandwidth)
	if state.Port <= 0 {
		t.Errorf("port 应为正数: %d", state.Port)
	}
	if state.Dir != IPerfDirUpload && state.Dir != IPerfDirDownload {
		t.Errorf("dir 取值异常: %q", state.Dir)
	}
	if len(state.Data) == 0 {
		t.Fatal("结果行 data 不应为空数组")
	}
	for i, r := range state.Data {
		t.Logf("  行 %d: status=%d response=%.120q", i, r.Status, r.Response)
	}
}
