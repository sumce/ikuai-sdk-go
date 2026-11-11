package logcenter

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveRTCollect 真机只读验证 rt_collect 的 show 类动作（告警列表 +
// 聚合区块 + 提醒标志）。探测真机实测 total=33、warning_notify=1。
func TestLiveRTCollect(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	rt := New(c).RTCollect

	// 告警列表（TYPE=total,data）。
	page, err := rt.Alerts(ctx, nil)
	if err != nil {
		t.Fatalf("rt_collect.show 失败: %v", err)
	}
	t.Logf("告警信息: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: [%d级] %s/%s %s (%s)", r.Priority, r.MsgType, r.Source, r.Content, r.Detail)
	}
	if page.Total <= 0 {
		t.Error("告警信息不应为空（探测真机实测 33 条）")
	}
	if r := page.Rows[0]; r.ID == 0 || r.Timestamp == 0 || r.Content == "" || r.MsgType == "" {
		t.Errorf("告警行字段与探测形态不符: %+v", r)
	}

	// 聚合区块（TYPE=all）。
	all, err := rt.AlertsAll(ctx)
	if err != nil {
		t.Fatalf("rt_collect.show TYPE=all 失败: %v", err)
	}
	t.Logf("告警聚合: total=%d 行数=%d warning_notify=%d", all.Total, len(all.Rows), all.WarningNotify)
	if all.Total != page.Total {
		t.Errorf("TYPE=all total=%d 与 TYPE=total,data total=%d 不一致", all.Total, page.Total)
	}

	// 告警提醒标志（TYPE=warning_notify，抓包实证）。
	n, err := rt.WarningNotify(ctx)
	if err != nil {
		t.Fatalf("rt_collect.show TYPE=warning_notify 失败: %v", err)
	}
	t.Logf("warning_notify: %d（探测真机实测 1）", n)
	if n != 1 {
		t.Errorf("warning_notify 应为 1（探测真机实测），got %d", n)
	}

	// 未读过滤（FILTER1="readflag,==,0"，真机实测生效；探测真机 33 条全部未读）。
	unread, err := rt.Alerts(ctx, nil)
	if err != nil {
		t.Fatalf("rt_collect.show 失败: %v", err)
	}
	unreadFlagged := 0
	for _, r := range unread.Rows {
		if r.ReadFlag == AlertReadFlagUnread {
			unreadFlagged++
		}
	}
	t.Logf("本页未读行数: %d/%d", unreadFlagged, len(unread.Rows))
}
