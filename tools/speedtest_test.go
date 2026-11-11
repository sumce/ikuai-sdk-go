package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// speedtestSample 构造 speedtest.show 的响应 Data（形态为真机 3.7.24
// 实测：data 为按接口名为键的对象；速率字段为字符串）。
func speedtestSample() map[string]any {
	return map[string]any{
		"global_status":     0,
		"global_status_msg": "测速任务已停止",
		"task":              "",
		"data": map[string]any{
			"adsl01": map[string]any{
				"interface": "adsl01", "status": "0", "status_msg": "未测速",
				"download": "--", "upload": "--", "down_avg": "--", "up_avg": "--",
				"down_min": "--", "up_min": "--", "down_max": "--", "up_max": "--",
			},
			"adsl09": map[string]any{
				"interface": "adsl09", "status": "4", "status_msg": "正在测速下载",
				"download": "114442.36", "upload": "0", "down_avg": "0", "up_avg": "0",
				"down_min": "0", "up_min": "0", "down_max": "0", "up_max": "0",
			},
		},
	}
}

// TestSpeedTestStatus 覆盖 speedtest.show 状态查询的成功路径与参数断言。
func TestSpeedTestStatus(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("speedtest", "show", speedtestSample())
	state, err := New(ts.Client()).SpeedTest.Status(context.Background())
	if err != nil {
		t.Fatalf("Status 失败: %v", err)
	}
	if state.GlobalStatus != SpeedTestGlobalStatusStopped ||
		state.GlobalStatusMsg != "测速任务已停止" || state.Task != "" {
		t.Fatalf("全局状态解码不符: %+v", state)
	}
	if len(state.Lines) != 2 {
		t.Fatalf("线路数解码不符: %d", len(state.Lines))
	}
	idle, ok := state.Lines["adsl01"]
	if !ok || !idle.IsIdle() || idle.Interface != "adsl01" ||
		idle.StatusMsg != "未测速" || idle.Download != "--" {
		t.Fatalf("adsl01 空闲线路解码不符: %+v", idle)
	}
	busy := state.Lines["adsl09"]
	if busy.IsIdle() || busy.Status != SpeedTestLineStatusTestingDownload ||
		busy.StatusMsg != "正在测速下载" || busy.Download != "114442.36" {
		t.Fatalf("adsl09 测速中线路解码不符: %+v", busy)
	}
	rec := ts.CallsFor("speedtest")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %s, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestSpeedTestStart 覆盖 speedtest.start 发起测速的参数断言（仅离线）。
func TestSpeedTestStart(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("speedtest", "start", map[string]any{})
	if err := New(ts.Client()).SpeedTest.Start(context.Background(), "adsl01"); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	rec := ts.CallsFor("speedtest")[0]
	if rec.Action != "start" {
		t.Fatalf("action = %s, 期望 start", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"interface": "adsl01"})
}

// TestSpeedTestStop 覆盖 speedtest.stop 停止测速的参数断言（仅离线）。
func TestSpeedTestStop(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("speedtest", "stop", map[string]any{})
	if err := New(ts.Client()).SpeedTest.Stop(context.Background()); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	rec := ts.CallsFor("speedtest")[0]
	if rec.Action != "stop" {
		t.Fatalf("action = %s, 期望 stop", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestSpeedTestError 覆盖固件错误路径。
func TestSpeedTestError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("speedtest", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "speedtest", Action: action}
	})
	svc := New(ts.Client()).SpeedTest
	if _, err := svc.Status(context.Background()); err == nil {
		t.Fatal("Status 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := svc.Start(context.Background(), "adsl01"); err == nil {
		t.Fatal("Start 应失败")
	}
	if err := svc.Stop(context.Background()); err == nil {
		t.Fatal("Stop 应失败")
	}
}
