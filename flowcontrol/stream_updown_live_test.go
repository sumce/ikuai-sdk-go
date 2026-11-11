package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveUpDown 真机只读验证流量上下线分流（stream_updown.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveUpDown(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).UpDown

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("stream_updown.show 列表失败: %v", err)
	}
	t.Logf("流量上下线分流规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d enabled=%s interface=%s protocol=%s",
			r.ID, r.Enabled, r.Interface, r.Protocol)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("stream_updown.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	if len(sel.Protocol) == 0 {
		t.Error("协议选择器不应为空")
	}
	t.Logf("可选接口 %d 个, 协议: %v, ipgroup=%v", len(sel.Interface), sel.Protocol, sel.IPGroup)
}
