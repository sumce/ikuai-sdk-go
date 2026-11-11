package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLivePackages 真机只读验证套餐管理（ppp_package.show）。
func TestLivePackages(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.Packages.List(ctx, nil)
	if err != nil {
		t.Fatalf("ppp_package.show 失败: %v", err)
	}
	t.Logf("套餐总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, pkg := range page.Rows {
		t.Logf("套餐: id=%d name=%s packtime=%d unit=%s price=%.2f up=%d down=%d",
			pkg.ID, pkg.PackName, pkg.PackTime, pkg.PeriodUnit, pkg.Price,
			pkg.UpSpeed, pkg.DownSpeed)
	}
}
