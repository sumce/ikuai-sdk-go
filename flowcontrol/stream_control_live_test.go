package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSmartControl 真机只读验证智能流控（stream_control.show）。
// Set 为写动作（stream_control.seting），严禁对真机调用。
func TestLiveSmartControl(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	cfg, err := New(c).SmartControl.Get(ctx)
	if err != nil {
		t.Fatalf("stream_control.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("stream_control.show data 不应为空")
	}
	if cfg.StreamCtlMode != SmartControlModeOn && cfg.StreamCtlMode != SmartControlModeOff {
		t.Errorf("stream_ctl_mode 出现未知的模式值 %d（已知 0/1）", cfg.StreamCtlMode)
	}
	t.Logf("智能流控模式=%d (IsEnabled=%v)", cfg.StreamCtlMode, cfg.IsEnabled())
}
