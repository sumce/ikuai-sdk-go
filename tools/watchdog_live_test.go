package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWatchdogGet 真机只读验证看门狗配置读取（watchdog.show 空参，
// 真机实测 data 单行 {watchdog_thresh:30, watchdog:1}）。
func TestLiveWatchdogGet(t *testing.T) {
	c := livetest.Client(t)
	cfg, err := New(c).Watchdog.Get(context.Background())
	if err != nil {
		t.Fatalf("watchdog.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("配置不应为 nil（真机实测恒返回一行）")
	}
	t.Logf("watchdog=%d watchdog_thresh=%d (秒)", cfg.Watchdog, cfg.WatchdogThresh)
	if cfg.Watchdog != WatchdogOn && cfg.Watchdog != WatchdogOff {
		t.Errorf("watchdog 开关取值异常: %d", cfg.Watchdog)
	}
	if cfg.WatchdogThresh <= 0 {
		t.Errorf("watchdog_thresh 应为正数: %d", cfg.WatchdogThresh)
	}
}
