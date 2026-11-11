package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveLAN 真机只读验证内外网设置（lan.show）。
func TestLiveLAN(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).LAN

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("lan.show 失败: %v", err)
	}
	t.Logf("接口总数=%d support_wisp=%d", page.Total, page.SupportWisp)
	if page.Total == 0 || len(page.Rows) == 0 {
		t.Fatal("真机应有内网接口")
	}
	r := page.Rows[0]
	t.Logf("首行: id=%d name=%s ip_mask=%s bandeth=%s", r.ID, r.Name, r.IPMask, r.BandETH)
	if r.Name == "" {
		t.Error("首行 name 为空")
	}

	iface, err := svc.Interface(ctx, r.ID)
	if err != nil {
		t.Fatalf("lan.show 单接口失败: %v", err)
	}
	if iface == nil || iface.ID != r.ID || iface.Name != r.Name {
		t.Fatalf("单接口回显不符: %+v", iface)
	}
}
