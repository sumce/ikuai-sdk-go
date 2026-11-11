package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveCustomProtos 真机只读验证自定义协议（dprotos.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveCustomProtos(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).CustomProtos

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("dprotos.show 列表失败: %v", err)
	}
	t.Logf("自定义协议总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d name=%q protocol=%q ports=%q",
			r.ID, r.Name, r.Protocol, r.Ports)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("dprotos.show TYPE=all 失败: %v", err)
	}
	t.Logf("ipgroup=%v", sel.IPGroup)
}

// TestLiveAdvancedProtos 真机只读验证高级自定义协议（dprotos_l7.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveAdvancedProtos(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	page, err := New(c).AdvancedProtos.List(ctx, nil)
	if err != nil {
		t.Fatalf("dprotos_l7.show 列表失败: %v", err)
	}
	if page.Total > 0 && len(page.Rows) == 0 {
		t.Error("total>0 但未解出数据行")
	}
	t.Logf("高级自定义协议总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d name=%q comment=%q", r.ID, r.Name, r.Comment)
	}
}
