package behavior

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveAuditLogs 真机只读验证审计日志四 func 的 show。
// 全部为只读日志类，无写动作。
func TestLiveAuditLogs(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).AuditLogs

	// URL 审计日志（无 TYPE 分页）。
	page, err := svc.URLLogs(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("audit_url_log.show 失败: %v", err)
	}
	t.Logf("URL 审计日志: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: ip=%s mac=%s host=%s", r.IPAddr, r.MAC, r.Host)
	}

	files, err := svc.URLLogFiles(ctx)
	if err != nil {
		t.Fatalf("audit_url_log.show TYPE=all 失败: %v", err)
	}
	t.Logf("URL 日志磁盘文件: %d 个", len(files.Files))

	// IM 审计日志。
	im, err := svc.IMLogs(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("audit_im_log.show 失败: %v", err)
	}
	t.Logf("IM 审计日志: total=%d 本页行数=%d", im.Total, len(im.Rows))
	for i, r := range im.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: type=%s account=%s ip=%s", r.IMType, r.Account, r.IP)
	}

	// 终端上下线日志（真机实测 total=229127）。
	tl, err := svc.TerminalLogs(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("audit_terminal_log.show 失败: %v", err)
	}
	t.Logf("终端上下线日志: total=%d 本页行数=%d", tl.Total, len(tl.Rows))
	for i, r := range tl.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: ip=%s mac=%s sys=%s dev=%s 在线=%ds 上行=%d 下行=%d",
			r.IPAddr, r.MAC, r.SysType, r.DevType, r.OnlineTime, r.TotalUp, r.TotalDown)
	}
	if len(tl.Rows) == 0 {
		t.Error("终端上下线日志不应为空（探测真机实测 229127 条）")
	}
	if r := tl.Rows[0]; r.ID == 0 || r.Timestamp == 0 || r.SysType == "" {
		t.Errorf("终端日志行字段与探测形态不符: %+v", r)
	}

	// 双栈区块。
	tla, err := svc.TerminalLogsAll(ctx, nil)
	if err != nil {
		t.Fatalf("audit_terminal_log.show TYPE=all 失败: %v", err)
	}
	t.Logf("终端日志双栈: v4=%d v6=%d 行数 v4=%d v6=%d",
		tla.Total, tla.Total6, len(tla.Rows), len(tla.Rows6))

	// 终端流量统计：探测真机未注册该 func（30002），按预期记录失败形态。
	if _, err := svc.TerminalStats(ctx, nil); err != nil {
		t.Logf("audit_terminal_stat.show 按预期失败（本固件未注册）: %v", err)
	} else {
		t.Log("audit_terminal_stat.show 成功（本固件已注册该 func）")
	}
}
