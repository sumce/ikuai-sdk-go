package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveAppFlow 真机只读验证协议监控（monitor_app_flow.show）。
func TestLiveAppFlow(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	hist, err := svc.AppFlow.History(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_app_flow.show TYPE=app_history 失败: %v", err)
	}
	t.Logf("app_history 时点行数=%d", len(hist.Rows))
	if len(hist.Rows) > 0 && hist.Rows[0].Timestamp == 0 {
		t.Errorf("首行 timestamp 为 0: %+v", hist.Rows[0])
	}

	flow, err := svc.AppFlow.Flow(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_app_flow.show TYPE=flow 失败: %v", err)
	}
	t.Logf("flow 协议数=%d", len(flow.Rows))

	ifaces, err := svc.AppFlow.Interfaces(ctx)
	if err != nil {
		t.Fatalf("monitor_app_flow.show TYPE=interface,app_history_cleantime 失败: %v", err)
	}
	t.Logf("可选接口=%v cleantime=%d", ifaces.Interface, ifaces.CleanTime)
	if len(ifaces.Interface) == 0 {
		t.Error("接口列表不应为空")
	}
}
