package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDNS 真机只读验证 DNS 族（dns.show TYPE=dns_config 与
// dns_replace.show）。
func TestLiveDNS(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).DNS

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("dns.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("dns.show 应返回单行配置")
	}
	t.Logf("DNS 代理=%s dns1=%s dns2=%s cache_ttl=%d", cfg.Enabled, cfg.DNS1, cfg.DNS2, cfg.CacheTTL)
	if cfg.ID != 1 {
		t.Errorf("dns.id 应为 1, got %d", cfg.ID)
	}

	page, err := svc.ReplaceList(ctx, nil)
	if err != nil {
		t.Fatalf("dns_replace.show 失败: %v", err)
	}
	t.Logf("多线 DNS total=%d", page.Total)
}
