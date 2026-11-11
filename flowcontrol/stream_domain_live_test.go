package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDomain 真机只读验证域名分流（stream_domain.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveDomain(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Domain

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("stream_domain.show 列表失败: %v", err)
	}
	t.Logf("域名分流规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d enabled=%s interface=%s domain=%q src_addr=%q",
			r.ID, r.Enabled, r.Interface, r.Domain, r.SrcAddr)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("stream_domain.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	t.Logf("可选接口 %d 个, ipgroup=%v macgroup=%d dtgroup=%d",
		len(sel.Interface), sel.IPGroup, len(sel.MACGroup), len(sel.DTGroup))
}
