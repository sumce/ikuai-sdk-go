package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveMultiWan 真机只读验证多线负载（lb_pcc.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveMultiWan(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).MultiWan

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("lb_pcc.show 列表失败: %v", err)
	}
	t.Logf("多线负载规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d enabled=%s interface=%s isp=%s weight=%d",
			r.ID, r.Enabled, r.Interface, r.ISP, r.Weight)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("lb_pcc.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	if len(sel.ISPList) == 0 || sel.ISPList[0] != "all" {
		t.Errorf("运营商选择器不符: %v", sel.ISPList)
	}
	t.Logf("可选接口 %d 个, 运营商: %v", len(sel.Interface), sel.ISPList)

	// 第二页（总数充足时）。
	if page.Total > 20 {
		p2, err := svc.List(ctx, &ikuai.ListOptions{Offset: 20, Limit: 20})
		if err != nil {
			t.Fatalf("lb_pcc.show 第二页失败: %v", err)
		}
		t.Logf("第二页行数=%d", len(p2.Rows))
	}
}
