package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveNotices 真机只读验证通知族配置（notice_temp/notice_remind/
// notice_expires.show TYPE=data）。写动作（save）禁止对真机调用。
func TestLiveNotices(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Notices

	temp, err := svc.Temp(ctx)
	if err != nil {
		t.Fatalf("notice_temp.show 失败: %v", err)
	}
	if temp != nil {
		t.Logf("实时通知: red_url=%s timeout=%d pppuser=%d listuser=%d",
			temp.RedURL, temp.Timeout, temp.PPPUser, temp.ListUser)
	} else {
		t.Log("实时通知配置未落库")
	}

	remind, err := svc.Remind(ctx)
	if err != nil {
		t.Fatalf("notice_remind.show 失败: %v", err)
	}
	if remind != nil {
		t.Logf("到期提醒: enabled=%s days=%d time=%s timeout=%d red_url=%s",
			remind.Enabled, remind.Days, remind.Time, remind.Timeout, remind.RedURL)
	} else {
		t.Log("到期提醒配置未落库")
	}

	expires, err := svc.Expires(ctx)
	if err != nil {
		t.Fatalf("notice_expires.show 失败: %v", err)
	}
	if expires != nil {
		t.Logf("PPPoE 到期通知: whiteip=%q whitedomain=%q", expires.WhiteIP, expires.WhiteDomain)
	} else {
		t.Log("PPPoE 到期通知配置未落库")
	}
}

// TestLiveNoticeCycle 真机只读验证周期通知列表（notice_cycle.show）。
// 周期通知的增删改（写动作）禁止对真机调用。
func TestLiveNoticeCycle(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Notices

	page, err := svc.CycleList(ctx, nil)
	if err != nil {
		t.Fatalf("notice_cycle.show 失败: %v", err)
	}
	t.Logf("周期通知总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		t.Logf("周期通知: id=%d week=%s time=%s red_url=%s timeout=%d",
			row.ID, row.Week, row.Time, row.RedURL, row.Timeout)
	}
}
