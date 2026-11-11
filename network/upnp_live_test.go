package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveUPnP 真机只读验证 UPnP 族（upnpd.show 空参数与
// upnpd_leases.show）。
func TestLiveUPnP(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).UPnP

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("upnpd.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("upnpd.show 应返回单行配置")
	}
	t.Logf("UPnP enabled=%s exclude_port=%s lan_ip=%s if=%s",
		cfg.Enabled, cfg.ExcludePort, cfg.LANIP, cfg.Interface)
	if cfg.ID != 1 {
		t.Errorf("upnpd.id 应为 1, got %d", cfg.ID)
	}

	leases, err := svc.LeaseList(ctx, nil)
	if err != nil {
		t.Fatalf("upnpd_leases.show 失败: %v", err)
	}
	t.Logf("UPnP 租约 total=%d", leases.Total)
}
