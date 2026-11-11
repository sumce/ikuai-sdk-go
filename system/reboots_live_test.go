package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveReboots 真机只读验证定时重启任务列表（reboots.show）。
// ⚠️ 重启/定时任务写动作严禁对真机调用。
func TestLiveReboots(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Reboots

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("reboots.show 失败: %v", err)
	}
	if page.Total != int64(len(page.Rows)) {
		t.Errorf("total=%d 与行数=%d 不一致", page.Total, len(page.Rows))
	}
	t.Logf("定时重启任务数=%d", page.Total)
	for _, r := range page.Rows {
		m, err := r.Map()
		if err != nil {
			t.Errorf("行 map 解码失败: %v", err)
		}
		t.Logf("任务行: %v", m)
	}

	// 第二页形态。
	page2, err := svc.List(ctx, &ikuai.ListOptions{Offset: 0, Limit: 5})
	if err != nil {
		t.Fatalf("reboots.show 分页形态失败: %v", err)
	}
	if len(page2.Rows) > 5 {
		t.Errorf("分页 limit=5 不应超过 5 行: %d", len(page2.Rows))
	}
}
