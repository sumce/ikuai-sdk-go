package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestURLLogs 覆盖 audit_url_log.show 的无 TYPE 分页（真机实测 TYPE 被
// 30006 拒绝）与行解码。
func TestURLLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit_url_log", "show", map[string]any{
		"total": -1,
		"data": []any{
			map[string]any{"id": 1, "timestamp": 1790279676, "ip_addr": "172.22.1.2",
				"mac": "00:0c:29:3b:11:af", "host": "www.example.com"},
		},
	})
	page, err := New(ts.Client()).AuditLogs.URLLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("URLLogs 失败: %v", err)
	}
	if page.Total != -1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.Timestamp != 1790279676 || r.IPAddr != "172.22.1.2" || r.MAC != "00:0c:29:3b:11:af" ||
		r.Host != "www.example.com" {
		t.Fatalf("日志行解码不符: %+v", r)
	}
	rec := ts.CallsFor("audit_url_log")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	// nil opts → 默认分页但无 TYPE。
	wantParam(t, paramOf(t, rec), map[string]any{"limit": "0,20"})

	// 第二页同样不带 TYPE。
	ts.Reset()
	if _, err = New(ts.Client()).AuditLogs.URLLogs(context.Background(),
		&ikuai.ListOptions{Offset: 20, Limit: 20}); err != nil {
		t.Fatalf("URLLogs 第二页失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("audit_url_log")[0]),
		map[string]any{"limit": "20,20"})
}

// TestURLLogFiles 覆盖 audit_url_log.show TYPE=all（真机实测区块）。
func TestURLLogFiles(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit_url_log", "show", map[string]any{
		"total": -1, "data": []any{},
		"stream_filelist": []any{
			map[string]any{"name": "url_log_1.tar.gz", "size": 1024, "time": 1790279676},
		},
		"stream_filelist_total": 1,
	})
	out, err := New(ts.Client()).AuditLogs.URLLogFiles(context.Background())
	if err != nil {
		t.Fatalf("URLLogFiles 失败: %v", err)
	}
	if out.Total != -1 || len(out.Rows) != 0 {
		t.Fatalf("日志区块解码不符: %+v", out)
	}
	if len(out.Files) != 1 || out.Files[0].Name != "url_log_1.tar.gz" ||
		out.Files[0].Size != 1024 || out.Files[0].Time != 1790279676 {
		t.Fatalf("文件列表解码不符: %+v", out.Files)
	}
	if out.FilesTotal != 1 {
		t.Fatalf("FilesTotal = %d, 期望 1", out.FilesTotal)
	}
	wantParam(t, paramOf(t, ts.CallsFor("audit_url_log")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestIMLogs 覆盖 audit_im_log.show 分页与行解码（真机实测 TYPE=total,data）。
func TestIMLogs(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
		},
		{
			name: "第二页",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 20},
			want: map[string]any{"TYPE": "total,data", "limit": "20,20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("audit_im_log", "show", map[string]any{
				"total": 1,
				"data": []any{
					map[string]any{"id": 1, "im_type": IMTypeQQPC, "account": "123456789",
						"event": "上线", "ip": "172.22.1.2", "mac": "00:0c:29:3b:11:af",
						"date_time": "2026-09-25 03:54:36"},
				},
			})
			page, err := New(ts.Client()).AuditLogs.IMLogs(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("IMLogs 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.IMType != IMTypeQQPC || r.Account != "123456789" || r.Event != "上线" ||
				r.IP != "172.22.1.2" || r.DateTime != "2026-09-25 03:54:36" {
				t.Fatalf("IM 日志行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("audit_im_log")[0]), tt.want)
		})
	}
}

// TestIMTypes IM 类型常量与固件 i18n 值一致。
func TestIMTypes(t *testing.T) {
	want := map[string]string{
		IMTypeAlitalkPC:     "Alitalk_pc",
		IMTypeAlitalkMobile: "Alitalk_mobile",
		IMTypeYYPC:          "yy_pc",
		IMTypeYYMobile:      "yy_mobile",
		IMTypeQQPC:          "qq_pc",
		IMTypeQQMobile:      "qq_mobile",
	}
	for got, exp := range want {
		if got != exp {
			t.Fatalf("IM 类型常量 %q != %q", got, exp)
		}
	}
}

// terminalLogSample 构造真机探测形态的 audit_terminal_log.show 响应 Data
// （TYPE=total,data；真机 3.7.24 实测行结构）。
func terminalLogSample() map[string]any {
	return map[string]any{
		"total": 229127,
		"data": []any{
			map[string]any{
				"devtype": "Ruijie", "client_model": "", "id": 1,
				"timestamp": 1790279676, "logout_time": 1792101180,
				"online_time": 1821504, "ip_addr": "172.20.4.253",
				"date_time": "2026-09-25 03:54:36", "mac": "54:16:51:f6:31:03",
				"comment": "fff", "total_up": 366007936, "total_down": 87958856,
				"auth": 0, "systype": "Windows10", "client_typeid": 102999999,
			},
		},
	}
}

// TestTerminalLogs 覆盖 audit_terminal_log.show 分页（真机实测行结构）。
func TestTerminalLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit_terminal_log", "show", terminalLogSample())
	page, err := New(ts.Client()).AuditLogs.TerminalLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("TerminalLogs 失败: %v", err)
	}
	if page.Total != 229127 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Timestamp != 1790279676 || r.LogoutTime != 1792101180 ||
		r.OnlineTime != 1821504 || r.IPAddr != "172.20.4.253" ||
		r.MAC != "54:16:51:f6:31:03" || r.Comment != "fff" || r.Auth != 0 ||
		r.SysType != "Windows10" || r.DevType != "Ruijie" ||
		r.ClientTypeID != 102999999 || r.TotalUp != 366007936 || r.TotalDown != 87958856 {
		t.Fatalf("终端日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("audit_terminal_log")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestTerminalLogsAll 覆盖 audit_terminal_log.show TYPE=all 双栈区块
// （真机实测 data/data6/total/total6）。
func TestTerminalLogsAll(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit_terminal_log", "show", map[string]any{
		"total": 229127, "total6": 0,
		"data":  []any{map[string]any{"id": 1, "ip_addr": "172.20.4.253"}},
		"data6": []any{},
	})
	out, err := New(ts.Client()).AuditLogs.TerminalLogsAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("TerminalLogsAll 失败: %v", err)
	}
	if out.Total != 229127 || out.Total6 != 0 || len(out.Rows) != 1 || len(out.Rows6) != 0 {
		t.Fatalf("双栈区块解码不符: %+v", out)
	}
	wantParam(t, paramOf(t, ts.CallsFor("audit_terminal_log")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestTerminalStats 覆盖 audit_terminal_stat.show 分页（空表推导，待真机
// 验证；真机 30002 见 TestTerminalStatsError）。
func TestTerminalStats(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit_terminal_stat", "show", map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{"id": 1, "mac": "00:0c:29:3b:11:af", "account": "user01",
				"uptrend": 4389810181468, "downward": 3532845777258, "comment": ""},
		},
	})
	page, err := New(ts.Client()).AuditLogs.TerminalStats(context.Background(), nil)
	if err != nil {
		t.Fatalf("TerminalStats 失败: %v", err)
	}
	if page.Total != 2 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.MAC != "00:0c:29:3b:11:af" || r.Account != "user01" ||
		r.Uptrend != 4389810181468 || r.Downward != 3532845777258 {
		t.Fatalf("统计行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("audit_terminal_stat")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestAuditLogsError 覆盖固件错误路径（含真机 30002 形态）。
func TestAuditLogsError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("audit_url_log", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)",
			FuncName: "audit_url_log", Action: action}
	})
	ts.HandleFuncName("audit_terminal_stat", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname(audit_terminal_stat)",
			FuncName: "audit_terminal_stat", Action: action}
	})
	svc := New(ts.Client()).AuditLogs
	ctx := context.Background()

	if _, err := svc.URLLogs(ctx, nil); err == nil {
		t.Fatal("audit_url_log.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.URLLogFiles(ctx); err == nil {
		t.Fatal("audit_url_log.show TYPE=all 错误应上抛")
	}
	if _, err := svc.IMLogs(ctx, nil); err == nil {
		t.Fatal("audit_im_log.show 错误应上抛")
	}
	if _, err := svc.TerminalLogs(ctx, nil); err == nil {
		t.Fatal("audit_terminal_log.show 错误应上抛")
	}
	if _, err := svc.TerminalLogsAll(ctx, nil); err == nil {
		t.Fatal("audit_terminal_log.show TYPE=all 错误应上抛")
	}
	if _, err := svc.TerminalStats(ctx, nil); err == nil {
		t.Fatal("audit_terminal_stat.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
}
