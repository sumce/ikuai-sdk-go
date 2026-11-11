package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveLanIP 真机只读验证终端监控 IPv4（monitor_lanip.show）。
func TestLiveLanIP(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.LanIP.List(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_lanip.show 列表失败: %v", err)
	}
	t.Logf("终端总数=%d 本页行数=%d", page.Total, len(page.Rows))
	if page.Total == 0 {
		t.Error("真机应有在线终端")
	}
	if len(page.Rows) > 0 && page.Rows[0].IPAddr == "" {
		t.Errorf("首行 ip_addr 为空: %+v", page.Rows[0])
	}

	// 关键字搜索（用第一行 IP 前缀保证有命中）。
	if len(page.Rows) > 0 {
		ip := page.Rows[0].IPAddr
		found, err := svc.LanIP.Search(ctx, ip, nil)
		if err != nil {
			t.Fatalf("monitor_lanip.show 搜索失败: %v", err)
		}
		if found.Total == 0 {
			t.Errorf("按已知 IP %q 搜索应有命中", ip)
		}
		t.Logf("搜索 %q 命中 %d 台", ip, found.Total)
	}

	// 接口列表。
	ifaces, err := svc.LanIP.Interfaces(ctx)
	if err != nil {
		t.Fatalf("monitor_lanip.show TYPE=interface 失败: %v", err)
	}
	if len(ifaces) == 0 {
		t.Error("接口列表不应为空")
	}
	t.Logf("可选接口: %v", ifaces)

	// 第二页。
	if page.Total > 20 {
		p2, err := svc.LanIP.List(ctx, &ikuai.ListOptions{Offset: 20, Limit: 20})
		if err != nil {
			t.Fatalf("monitor_lanip.show 第二页失败: %v", err)
		}
		if len(p2.Rows) == 0 {
			t.Error("第二页应有数据")
		}
	}
}
