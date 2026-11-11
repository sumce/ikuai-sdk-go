package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveVRRP 真机只读验证主备热备（vrrp_config.show TYPE=data）。
// Save 为写动作，严禁对真机调用。
func TestLiveVRRP(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).VRRP

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("vrrp_config.show 失败: %v", err)
	}
	if info.Version == "" {
		t.Error("version 不应为空")
	}
	row := info.Row()
	if row == nil {
		t.Fatal("vrrp_config.show data 不应为空（固件恒返回单行）")
	}
	if row.ID != 1 {
		t.Errorf("行 ID 应为 1: %+v", row)
	}
	t.Logf("热备启用=%v 模式=%d 优先级=%d 通告间隔=%ds version=%s",
		row.IsEnabled(), row.Type, row.Prio, row.Interval, info.Version)

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("vrrp_config.show List 失败: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Errorf("List 应返回单行: %d", len(page.Rows))
	}

	// 带分页参数的形态（真机接受 limit）。
	page2, err := svc.List(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("vrrp_config.show 分页形态失败: %v", err)
	}
	if len(page2.Rows) != 1 {
		t.Errorf("分页形态应返回单行: %d", len(page2.Rows))
	}
}
