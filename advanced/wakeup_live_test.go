package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWakeup 真机只读验证网络唤醒定时任务列表（wakeup.show，真机
// 探测实证）。只读；立即唤醒 Wake 绝对禁止对真机调用。
func TestLiveWakeup(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Wakeup.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("wakeup.show 失败: %v", err)
	}
	t.Logf("定时唤醒任务 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("任务: id=%d mac=%s week=%s time=%s comment=%s",
			r.ID, r.MAC, r.Week, r.Time, r.Comment)
		if r.MAC == "" {
			t.Error("唤醒任务 mac 为空（行字段空表推导，待真机验证）")
		}
	}
}
