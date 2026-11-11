package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveLanIPv6 真机只读验证终端监控 IPv6（monitor_lanipv6.show）。
func TestLiveLanIPv6(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.LanIPv6.List(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_lanipv6.show 列表失败: %v", err)
	}
	t.Logf("IPv6 终端总数=%d 本页行数=%d", page.Total, len(page.Rows))

	ifaces, err := svc.LanIPv6.Interfaces(ctx)
	if err != nil {
		t.Fatalf("monitor_lanipv6.show TYPE=interface 失败: %v", err)
	}
	if len(ifaces) == 0 {
		t.Error("接口列表不应为空")
	}
	t.Logf("可选接口: %v", ifaces)
}
