package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveUDPProxy 真机只读验证 UDP 代理规则与可选接口（udp_proxy.show
// TYPE=total,data / TYPE=all，真机探测实证）。只读，不做任何写动作。
func TestLiveUDPProxy(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).UDProxy
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("udp_proxy.show 失败: %v", err)
	}
	t.Logf("UDP 代理规则 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("规则: id=%d iface=%s port=%d renew=%d net=%s enabled=%s",
			r.ID, r.Interface, r.Port, r.RenewTime, r.NetEnabled, r.Enabled)
	}

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("udp_proxy.show TYPE=all 失败: %v", err)
	}
	t.Logf("可选接口: %d 项", len(info.Interface))
}
