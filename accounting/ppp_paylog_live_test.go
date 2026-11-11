package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLivePayLog 真机只读验证报表管理（ppp_paylog.show）。
func TestLivePayLog(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.PayLog.List(ctx, nil)
	if err != nil {
		t.Fatalf("ppp_paylog.show 失败: %v", err)
	}
	t.Logf("收费记录总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, e := range page.Rows {
		t.Logf("记录: id=%d username=%s admin=%s action=%s packname=%s feemoney=%.2f time=%s",
			e.ID, e.Username, e.Admin, e.Action, e.Packname, e.FeeMoney, e.Time())
	}
}
