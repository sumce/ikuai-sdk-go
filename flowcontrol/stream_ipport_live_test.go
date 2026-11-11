package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveIPPort 真机只读验证端口分流（stream_ipport.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveIPPort(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).IPPort

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("stream_ipport.show 列表失败: %v", err)
	}
	if page.Total == 0 {
		t.Log("端口分流规则为空（探测真机有 9 条，此设备可能已变更）")
	}
	t.Logf("端口分流规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		if r.ID == 0 || r.Interface == "" {
			t.Errorf("第 %d 行 id/interface 异常: %+v", i, r)
		}
		t.Logf("  行: id=%d enabled=%s interface=%s protocol=%s src_addr=%q comment=%q",
			r.ID, r.Enabled, r.Interface, r.Protocol, r.SrcAddr, r.Comment)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("stream_ipport.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	if len(sel.Protocol) == 0 {
		t.Error("协议选择器不应为空")
	}
	t.Logf("可选接口 %d 个, 协议: %v, ipgroup=%v", len(sel.Interface), sel.Protocol, sel.IPGroup)

	// 第二页（总数充足时）。
	if page.Total > 20 {
		p2, err := svc.List(ctx, &ikuai.ListOptions{Offset: 20, Limit: 20})
		if err != nil {
			t.Fatalf("stream_ipport.show 第二页失败: %v", err)
		}
		if len(p2.Rows) == 0 {
			t.Error("第二页应有数据")
		}
	}
}
