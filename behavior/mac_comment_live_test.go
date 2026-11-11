package behavior

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveMACComment 真机只读验证 MAC 备注（mac_comment.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveMACComment(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).MACComment

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("mac_comment.show 失败: %v", err)
	}
	t.Logf("MAC 备注总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 5 {
			break
		}
		t.Logf("  行: id=%d mac=%s comment=%s", r.ID, r.MAC, r.Comment)
	}
	if len(page.Rows) == 0 {
		t.Log("本机无 MAC 备注记录")
		return
	}
	if r := page.Rows[0]; r.MAC == "" {
		t.Errorf("MAC 备注行 mac 字段为空, 行结构与探测形态不符: %+v", r)
	}
}
