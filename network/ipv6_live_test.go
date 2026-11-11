package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveIPv6 真机只读验证 IPv6 族 6 个 func 的 show。
func TestLiveIPv6(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).IPv6

	// ipv6.show。
	cfgs, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("ipv6.show 失败: %v", err)
	}
	t.Logf("IPv6 线路 total=%d", cfgs.Total)
	for _, r := range cfgs.Rows {
		t.Logf("ipv6: id=%d if=%s internet=%s dhcp6=%s", r.ID, r.Interface, r.Internet, r.DHCP6IPAddr)
	}

	// ipv6_stream.show（运行状态）。
	stream, err := svc.StreamList(ctx, nil)
	if err != nil {
		t.Fatalf("ipv6_stream.show 失败: %v", err)
	}
	t.Logf("IPv6 分流 total=%d", stream.Total)
	if len(stream.Rows) > 0 {
		r := stream.Rows[0]
		t.Logf("分流首行: if=%s conn=%d total_down=%s", r.Interface, r.Conn, r.TotalDownload)
		if r.Interface == "" || r.TotalDownload == "" {
			t.Error("分流首行字段为空")
		}
	}

	// dhcp6_lease.show TYPE=all（真机实证形态）。
	leases, err := svc.LeaseList(ctx, nil)
	if err != nil {
		t.Fatalf("dhcp6_lease.show 失败: %v", err)
	}
	t.Logf("DHCPv6 租约 client_total=%d", leases.ClientTotal)
	if len(leases.Rows) > 0 && leases.Rows[0].IPv6AddrTrimmed() == "" {
		t.Error("租约首行 ipv6_addr 为空")
	}

	// ipv6_static.show。
	statics, err := svc.StaticList(ctx, nil)
	if err != nil {
		t.Fatalf("ipv6_static.show 失败: %v", err)
	}
	t.Logf("DHCPv6 静态分配 total=%d", statics.Total)

	// dhcp6_acl_mac.show。
	acl, err := svc.ACLList(ctx, nil)
	if err != nil {
		t.Fatalf("dhcp6_acl_mac.show 失败: %v", err)
	}
	t.Logf("DHCPv6 黑白名单 total=%d", acl.Total)

	// ipv6_neighbor.show（运行状态表）。
	neighbors, err := svc.NeighborList(ctx, nil)
	if err != nil {
		t.Fatalf("ipv6_neighbor.show 失败: %v", err)
	}
	t.Logf("IPv6 邻居 total=%d", neighbors.Total)
	if len(neighbors.Rows) > 0 && neighbors.Rows[0].IPv6Addr == "" {
		t.Error("邻居首行 ipv6_addr 为空")
	}
}
