package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDHCP 真机只读验证 DHCP 族 4 个 func 的 show。
func TestLiveDHCP(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).DHCP

	// dhcp_server.show。
	servers, err := svc.ServerList(ctx, nil)
	if err != nil {
		t.Fatalf("dhcp_server.show 失败: %v", err)
	}
	t.Logf("DHCP 服务端 total=%d", servers.Total)
	if len(servers.Rows) > 0 {
		r := servers.Rows[0]
		t.Logf("服务端首行: interface=%s pool=%s status=%d available=%d",
			r.Interface, r.AddrPool, r.Status, r.Available)
		if r.Interface == "" {
			t.Error("服务端首行 interface 为空")
		}
	}

	// dhcp_static.show。
	statics, err := svc.StaticList(ctx, nil)
	if err != nil {
		t.Fatalf("dhcp_static.show 失败: %v", err)
	}
	t.Logf("DHCP 静态分配 total=%d", statics.Total)

	// dhcp_lease.show（运行状态表）。
	leases, err := svc.LeaseList(ctx, nil)
	if err != nil {
		t.Fatalf("dhcp_lease.show 失败: %v", err)
	}
	t.Logf("DHCP 租约 total=%d", leases.Total)
	if len(leases.Rows) > 0 && leases.Rows[0].IPAddr == "" {
		t.Error("租约首行 ip_addr 为空")
	}

	// dhcp_acl_mac.show。
	acl, err := svc.ACLList(ctx, nil)
	if err != nil {
		t.Fatalf("dhcp_acl_mac.show 失败: %v", err)
	}
	t.Logf("DHCP 黑白名单 total=%d", acl.Total)
}
