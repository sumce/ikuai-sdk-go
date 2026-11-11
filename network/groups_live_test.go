package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveGroups 真机只读验证分组族 4 个 func 的 show。
func TestLiveGroups(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Groups

	ip, err := svc.IPList(ctx, nil)
	if err != nil {
		t.Fatalf("ipgroup.show 失败: %v", err)
	}
	t.Logf("IPv4 分组 total=%d", ip.Total)
	if len(ip.Rows) > 0 && ip.Rows[0].GroupName == "" {
		t.Error("IPv4 分组首行 group_name 为空")
	}

	v6, err := svc.IPv6List(ctx, nil)
	if err != nil {
		t.Fatalf("ipv6group.show 失败: %v", err)
	}
	t.Logf("IPv6 分组 total=%d", v6.Total)

	mac, err := svc.MACList(ctx, nil)
	if err != nil {
		t.Fatalf("macgroup.show 失败: %v", err)
	}
	t.Logf("MAC 分组 total=%d", mac.Total)

	dt, err := svc.DTList(ctx, nil)
	if err != nil {
		t.Fatalf("dtgroup.show 失败: %v", err)
	}
	t.Logf("时间分组 total=%d", dt.Total)
}
