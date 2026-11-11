package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSpeedTestStatus 真机只读验证外网测速状态查询（speedtest.show
// 空参；不发起任何新测速）。
func TestLiveSpeedTestStatus(t *testing.T) {
	c := livetest.Client(t)
	state, err := New(c).SpeedTest.Status(context.Background())
	if err != nil {
		t.Fatalf("speedtest.show 失败: %v", err)
	}
	t.Logf("global_status=%d global_status_msg=%q task=%q 线路数=%d",
		state.GlobalStatus, state.GlobalStatusMsg, state.Task, len(state.Lines))
	if len(state.Lines) == 0 {
		t.Fatal("线路 data 不应为空对象")
	}
	for key, line := range state.Lines {
		if line.Interface != key {
			t.Errorf("data 键 %q 与行 interface %q 不一致", key, line.Interface)
		}
		if line.StatusMsg == "" {
			t.Errorf("线路 %q status_msg 不应为空", key)
		}
		t.Logf("  %s: status=%q msg=%q download=%q upload=%q down_avg=%q up_avg=%q",
			key, line.Status, line.StatusMsg, line.Download, line.Upload, line.DownAvg, line.UpAvg)
	}
}
