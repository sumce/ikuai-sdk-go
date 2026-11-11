package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSimpleQoS 真机只读验证 IP 限速（simple_qos.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveSimpleQoS(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).SimpleQoS

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("simple_qos.show 列表失败: %v", err)
	}
	t.Logf("IP 限速规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		if r.ID == 0 {
			t.Errorf("第 %d 行 id 异常: %+v", i, r)
		}
		t.Logf("  行: id=%d enabled=%s ip_addr=%s up=%d down=%d iface=%q",
			r.ID, r.Enabled, r.IPAddr, r.Upload, r.Download, r.Interface)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("simple_qos.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	t.Logf("可选接口 %d 个, ipgroup=%v", len(sel.Interface), sel.IPGroup)

	// 第二页（总数充足时）。
	if page.Total > 20 {
		p2, err := svc.List(ctx, &ikuai.ListOptions{Offset: 20, Limit: 20})
		if err != nil {
			t.Fatalf("simple_qos.show 第二页失败: %v", err)
		}
		t.Logf("第二页行数=%d", len(p2.Rows))
	}
}

// TestLiveMacQoS 真机只读验证 MAC 限速（mac_qos.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveMacQoS(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).MacQoS

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("mac_qos.show 列表失败: %v", err)
	}
	t.Logf("MAC 限速规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		if r.MACAddr == "" {
			t.Errorf("第 %d 行 mac_addr 为空: %+v", i, r)
		}
		t.Logf("  行: id=%d enabled=%s mac=%s ip_type=%s up=%d down=%d attr=%d type=%d",
			r.ID, r.Enabled, r.MACAddr, r.IPType, r.Upload, r.Download, r.Attr, r.Type)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("mac_qos.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	t.Logf("可选接口 %d 个, macgroup=%d dtgroup=%d", len(sel.Interface), len(sel.MACGroup), len(sel.DTGroup))
}
