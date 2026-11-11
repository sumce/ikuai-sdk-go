package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSystem 真机只读验证负载监控（monitor_system.show）。
func TestLiveSystem(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	m, err := svc.System.Curves(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_system.show 全曲线失败: %v", err)
	}
	t.Logf("cpu=%d memory=%d disk=%d terminal=%d conn=%d temp1=%d temp2=%d",
		len(m.CPU), len(m.Memory), len(m.DiskSpaceUsed), len(m.OnTerminal),
		len(m.ConnNum), len(m.CpuTemp1), len(m.CpuTemp2))
	if len(m.CPU) == 0 || len(m.Memory) == 0 || len(m.ConnNum) == 0 {
		t.Error("真机应有曲线数据")
	}

	rs, err := svc.System.RateStat(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_system.show TYPE=rate_stat 失败: %v", err)
	}
	t.Logf("rate_stat 时点行数=%d", len(rs.RateStat))
	if len(rs.RateStat) == 0 {
		t.Error("真机应有峰值速率数据")
	}

	support, err := svc.System.CpuTempSupport(ctx)
	if err != nil {
		t.Fatalf("monitor_system.show TYPE=cputemp_support 失败: %v", err)
	}
	t.Logf("cputemp_support=%d", support)

	af, err := svc.System.AppFlow(ctx, 30, nil)
	if err != nil {
		t.Fatalf("monitor_system.show TYPE=app_flow minute=30 失败: %v", err)
	}
	t.Logf("app_flow 行数=%d", len(af))
	if len(af) == 0 {
		t.Error("真机应有 app_flow 快照")
	}
}
