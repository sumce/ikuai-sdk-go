package logcenter

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveMessages 真机只读验证 ikmessages 的 show 类动作（消息列表 +
// 未读数）。
// ⚠️ 只测 show 类：ikmessages.start / ikmessages.set 属写动作，真机禁调
// （见 docs/RULES.md §9.2 与本包方法注释）。
func TestLiveMessages(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	msg := New(c).Messages

	// 消息列表（TYPE=total,data；探测真机消息表为空）。
	page, err := msg.List(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("ikmessages.show 失败: %v", err)
	}
	t.Logf("站内信消息: total=%d 本页行数=%d（行字段空表推导，待真机验证）",
		page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d title=%s read=%d", r.ID, r.Title, r.Read)
	}

	// 未读数（TYPE=total + FILTER1="read,==,0"，抓包实证形态）。
	n, err := msg.UnreadCount(ctx)
	if err != nil {
		t.Fatalf("ikmessages.show TYPE=total FILTER 失败: %v", err)
	}
	t.Logf("未读消息数: %d（探测真机实测 0）", n)
	if n != 0 {
		t.Errorf("未读消息数应为 0（探测真机实测），got %d", n)
	}
}
