package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// watchdogSample 构造 watchdog.show 的响应 Data（形态为真机 3.7.24
// 实测：data 单行 {watchdog_thresh, watchdog}）。
func watchdogSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{"watchdog_thresh": 30, "watchdog": 1}},
	}
}

// TestWatchdogGet 覆盖 watchdog.show 配置读取的成功路径与参数断言。
func TestWatchdogGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("watchdog", "show", watchdogSample())
	cfg, err := New(ts.Client()).Watchdog.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("配置不应为 nil")
	}
	if cfg.Watchdog != WatchdogOn || cfg.WatchdogThresh != 30 {
		t.Fatalf("配置解码不符: %+v", cfg)
	}
	if !cfg.IsEnabled() {
		t.Fatal("IsEnabled 应为 true")
	}
	rec := ts.CallsFor("watchdog")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %s, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestWatchdogGetEmpty 覆盖 data 为空数组时返回 nil。
func TestWatchdogGetEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("watchdog", "show", map[string]any{"data": []any{}})
	cfg, err := New(ts.Client()).Watchdog.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg != nil {
		t.Fatalf("空 data 应返回 nil, got %+v", cfg)
	}
}

// TestWatchdogSet 覆盖 watchdog.edit 保存配置的参数断言（仅离线，
// 真机写动作零调用）。
func TestWatchdogSet(t *testing.T) {
	tests := []struct {
		name string
		cfg  WatchdogConfig
		want map[string]any
	}{
		{
			name: "开启并设置阈值",
			cfg:  WatchdogConfig{Watchdog: WatchdogOn, WatchdogThresh: 60},
			want: map[string]any{"watchdog": float64(1), "watchdog_thresh": float64(60)},
		},
		{
			name: "关闭",
			cfg:  WatchdogConfig{Watchdog: WatchdogOff, WatchdogThresh: 30},
			want: map[string]any{"watchdog": float64(0), "watchdog_thresh": float64(30)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("watchdog", "edit", map[string]any{})
			if err := New(ts.Client()).Watchdog.Set(context.Background(), tt.cfg); err != nil {
				t.Fatalf("Set 失败: %v", err)
			}
			rec := ts.CallsFor("watchdog")[0]
			if rec.Action != "edit" {
				t.Fatalf("action = %s, 期望 edit", rec.Action)
			}
			wantParam(t, paramOf(t, rec), tt.want)
		})
	}
}

// TestWatchdogError 覆盖固件错误路径。
func TestWatchdogError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("watchdog", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "watchdog", Action: action}
	})
	svc := New(ts.Client()).Watchdog
	if _, err := svc.Get(context.Background()); err == nil {
		t.Fatal("Get 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := svc.Set(context.Background(), WatchdogConfig{Watchdog: WatchdogOn, WatchdogThresh: 30}); err == nil {
		t.Fatal("Set 应失败")
	}
}
