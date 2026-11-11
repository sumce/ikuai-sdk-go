package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestSmartControlGet 覆盖 stream_control.show TYPE=all 的成功路径
// （真机实测响应 {"stream_ctl_mode":1}）。
func TestSmartControlGet(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want int64
	}{
		{name: "开启", data: map[string]any{"stream_ctl_mode": 1}, want: 1},
		{name: "关闭", data: map[string]any{"stream_ctl_mode": 0}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("stream_control", "show", tt.data)
			cfg, err := New(ts.Client()).SmartControl.Get(context.Background())
			if err != nil {
				t.Fatalf("Get 失败: %v", err)
			}
			if cfg.StreamCtlMode != tt.want {
				t.Fatalf("StreamCtlMode = %d, 期望 %d", cfg.StreamCtlMode, tt.want)
			}
			if cfg.IsEnabled() != (tt.want == SmartControlModeOn) {
				t.Fatalf("IsEnabled = %v, 与模式 %d 不符", cfg.IsEnabled(), tt.want)
			}
			wantParam(t, paramOf(t, ts.CallsFor("stream_control")[0]),
				map[string]any{"TYPE": SmartControlTypeAll})
		})
	}
}

// TestSmartControlSet 覆盖 stream_control.seting 参数形态
// （写动作，开源生产集成实证 action 拼写，离线）。
func TestSmartControlSet(t *testing.T) {
	tests := []struct {
		name string
		mode int64
	}{
		{name: "开启", mode: SmartControlModeOn},
		{name: "关闭", mode: SmartControlModeOff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("stream_control", "seting", map[string]any{})
			if err := New(ts.Client()).SmartControl.Set(context.Background(),
				SmartControlConfig{StreamCtlMode: tt.mode}); err != nil {
				t.Fatalf("Set 失败: %v", err)
			}
			rec := ts.CallsFor("stream_control")[0]
			if rec.Action != "seting" {
				t.Fatalf("action = %q, 期望 seting（固件拼写）", rec.Action)
			}
			wantParam(t, paramOf(t, rec),
				map[string]any{"stream_ctl_mode": tt.mode})
		})
	}
}

// TestSmartControlError 覆盖固件错误路径。
func TestSmartControlError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("stream_control", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (data)", FuncName: "stream_control", Action: action}
	})
	svc := New(ts.Client()).SmartControl
	ctx := context.Background()
	if _, err := svc.Get(ctx); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := svc.Set(ctx, SmartControlConfig{StreamCtlMode: 1}); err == nil {
		t.Fatal("seting 错误应上抛")
	}
}

// TestSmartControlConstants 常量与真机/开源实证值一致。
func TestSmartControlConstants(t *testing.T) {
	if SmartControlTypeAll != "all" || SmartControlTypeMode != "stream_ctl_mode" {
		t.Fatalf("TYPE 常量不符: %q / %q", SmartControlTypeAll, SmartControlTypeMode)
	}
	if SmartControlModeOff != 0 || SmartControlModeOn != 1 {
		t.Fatalf("模式常量不符: %d / %d", SmartControlModeOff, SmartControlModeOn)
	}
}
