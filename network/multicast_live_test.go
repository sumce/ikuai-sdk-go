package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveMulticast 真机只读验证组播族（igmp_proxy.show 与 iptv.show）。
func TestLiveMulticast(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Multicast

	igmp, err := svc.IGMPGet(ctx)
	if err != nil {
		t.Fatalf("igmp_proxy.show 失败: %v", err)
	}
	if igmp == nil {
		t.Fatal("igmp_proxy.show 应返回单行配置")
	}
	t.Logf("IGMP enabled=%s upstream=%s downstream=%s version=%s",
		igmp.Enabled, igmp.Upstream, igmp.Downstream, igmp.Version)
	if igmp.ID != 1 {
		t.Errorf("igmp_proxy.id 应为 1, got %d", igmp.ID)
	}

	iptv, err := svc.IPTVGet(ctx)
	if err != nil {
		t.Fatalf("iptv.show 失败: %v", err)
	}
	if iptv == nil {
		t.Fatal("iptv.show 应返回单行配置")
	}
	t.Logf("IPTV enabled=%s mode=%d wan=%s(vlan %d) lan=%s(vlan %d)",
		iptv.Enabled, iptv.Mode, iptv.WANIface, iptv.WANVLANID, iptv.LANIface, iptv.LANVLANID)
	if iptv.ID != 1 {
		t.Errorf("iptv.id 应为 1, got %d", iptv.ID)
	}
}
