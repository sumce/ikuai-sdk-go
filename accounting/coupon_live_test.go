package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveCoupon 真机只读验证优惠券管理（coupon.show）。
func TestLiveCoupon(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.Coupon.List(ctx, nil)
	if err != nil {
		t.Fatalf("coupon.show 失败: %v", err)
	}
	t.Logf("上网码总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码上网码行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}
}
