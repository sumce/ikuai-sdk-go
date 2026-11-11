package flowcontrol

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveLayer7 真机只读验证协议分流（stream_layer7.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveLayer7(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Layer7

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("stream_layer7.show 列表失败: %v", err)
	}
	t.Logf("协议分流规则总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d enabled=%s app_proto=%s interface=%s src_addr=%s",
			r.ID, r.Enabled, r.AppProto, r.Interface, r.SrcAddr)
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("stream_layer7.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	t.Logf("可选接口 %d 个, ipgroup=%v macgroup=%d dtgroup=%d syn_proxy=%d",
		len(sel.Interface), sel.IPGroup, len(sel.MACGroup), len(sel.DTGroup), sel.SynProxy)
}
