package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveNAT 真机只读验证 NAT 规则族（nat_rule.show 两种 TYPE 与
// netmap.show）。
func TestLiveNAT(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).NAT

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("nat_rule.show 失败: %v", err)
	}
	t.Logf("NAT 规则 total=%d", page.Total)

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("nat_rule.show TYPE=interface,action 失败: %v", err)
	}
	if len(sel.Action) < 3 {
		t.Fatalf("action 候选应含 filter/snat/dnat, got %v", sel.Action)
	}
	t.Logf("接口候选 %d 个, 动作候选 %v", len(sel.Interface), sel.Action)

	dmz, err := svc.NetmapList(ctx, nil)
	if err != nil {
		t.Fatalf("netmap.show 失败: %v", err)
	}
	t.Logf("DMZ total=%d", dmz.Total)
}
