package logcenter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestSyslogContentLogs 覆盖 syslog-arp / syslog-sysevent（真机实测两表
// 行结构同构 {id,timestamp,content}）的 List 分页与行解码。
func TestSyslogContentLogs(t *testing.T) {
	tests := []struct {
		name   string
		fn     SyslogFunc
		total  int64
		sample map[string]any
		call   func(*SystemLogService, context.Context, *ikuai.ListOptions) (*ikuai.Page[SyslogRow], error)
	}{
		{
			name:  "ARPLogs",
			fn:    SyslogARP,
			total: 26,
			sample: map[string]any{
				"id": 1, "timestamp": 1791724070,
				"content": "检测到一个ARP地址欺骗在接口(_vlan3): 172.22.3.2 00:0c:29:79:da:95->00:0c:29:79:da:8b",
			},
			call: func(s *SystemLogService, ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SyslogRow], error) {
				return s.ARPLogs(ctx, opts)
			},
		},
		{
			name:  "SysEventLogs",
			fn:    SyslogSysEvent,
			total: 4168,
			sample: map[string]any{
				"id": 1, "timestamp": 1790279706, "content": "系统启动",
			},
			call: func(s *SystemLogService, ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SyslogRow], error) {
				return s.SysEventLogs(ctx, opts)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse(string(tt.fn), "show", map[string]any{
				"total": tt.total, "data": []any{tt.sample},
			})
			page, err := tt.call(New(ts.Client()).SystemLog, context.Background(), nil)
			if err != nil {
				t.Fatalf("%s 失败: %v", tt.name, err)
			}
			if page.Total != tt.total || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			wantID := int64(tt.sample["id"].(int))
			wantTS := int64(tt.sample["timestamp"].(int))
			if r.ID != wantID || r.Timestamp != wantTS || r.Content != tt.sample["content"] {
				t.Fatalf("日志行解码不符: %+v", r)
			}
			rec := ts.CallsFor(string(tt.fn))[0]
			if rec.Action != "show" {
				t.Fatalf("action = %q, 期望 show", rec.Action)
			}
			// nil opts → 默认分页 TYPE=total,data。
			wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "total,data", "limit": "0,20"})
		})
	}
}

// TestPPPAuthLogs 覆盖 syslog-pppauth.show（真机空表 total=0，行字段按
// i18n webauthLog 推导，待真机验证）。
func TestPPPAuthLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-pppauth", "show", map[string]any{
		"total": 0,
		"data": []any{
			map[string]any{"id": 1, "timestamp": 1790279675, "auth_type": PPPAuthTypePPPoE,
				"mac": "00:0c:29:3b:11:af", "ip_addr": "172.22.1.2", "event": "认证成功"},
		},
	})
	page, err := New(ts.Client()).SystemLog.PPPAuthLogs(context.Background(),
		&ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("PPPAuthLogs 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Timestamp != 1790279675 || r.AuthType != PPPAuthTypePPPoE ||
		r.MAC != "00:0c:29:3b:11:af" || r.IPAddr != "172.22.1.2" || r.Event != "认证成功" {
		t.Fatalf("认证日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-pppauth")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,5"})
}

// TestAPActionLogs 覆盖 syslog-apaction.show（真机空表 total=0，行字段按
// i18n terminalLog 推导，待真机验证）。
func TestAPActionLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-apaction", "show", map[string]any{
		"total": 0,
		"data": []any{
			map[string]any{"id": 3, "timestamp": 1790279675, "client_mac": "54:16:51:f6:31:03",
				"ap_mac": "d0:72:dc:12:34:56", "signal": "-58", "action": "login"},
		},
	})
	page, err := New(ts.Client()).SystemLog.APActionLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("APActionLogs 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 3 || r.ClientMAC != "54:16:51:f6:31:03" || r.APMAC != "d0:72:dc:12:34:56" ||
		r.Signal != "-58" || r.Action != "login" {
		t.Fatalf("AP 操作日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-apaction")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestDHCPLogs 覆盖 syslog-dhcpd.show（真机实测 total=5808，DHCPv4/v6 双栈
// 行结构 {id,timestamp,msgtype,interface,event,mac,ip_addr}）。
func TestDHCPLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-dhcpd", "show", map[string]any{
		"total": 5808,
		"data": []any{
			map[string]any{"timestamp": 1790280848, "msgtype": "DHCPV6SOLICIT",
				"interface": "vlan1211", "event": "IA_NA", "mac": "fe80::20c:29ff:fe0e:105f",
				"ip_addr": "2408:8220:32a:b3d1::597/128 ", "id": 1},
			map[string]any{"timestamp": 1790281100, "msgtype": "DHCPREQUEST",
				"interface": "lan4", "event": "--", "mac": "00:db:d4:e1:9e:84",
				"ip_addr": "192.168.2.4", "id": 2},
		},
	})
	page, err := New(ts.Client()).SystemLog.DHCPLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("DHCPLogs 失败: %v", err)
	}
	if page.Total != 5808 || len(page.Rows) != 2 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	v6 := page.Rows[0]
	if v6.ID != 1 || v6.MsgType != "DHCPV6SOLICIT" || v6.Interface != "vlan1211" ||
		v6.Event != "IA_NA" || v6.MAC != "fe80::20c:29ff:fe0e:105f" ||
		v6.IPAddr != "2408:8220:32a:b3d1::597/128 " {
		t.Fatalf("DHCPv6 行解码不符: %+v", v6)
	}
	v4 := page.Rows[1]
	if v4.ID != 2 || v4.MsgType != "DHCPREQUEST" || v4.Interface != "lan4" ||
		v4.Event != "--" || v4.MAC != "00:db:d4:e1:9e:84" || v4.IPAddr != "192.168.2.4" {
		t.Fatalf("DHCPv4 行解码不符: %+v", v4)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-dhcpd")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestDDNSLogs 覆盖 syslog-ddns.show（真机实测 total=2066）。
func TestDDNSLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-ddns", "show", map[string]any{
		"total": 2066,
		"data": []any{
			map[string]any{"interface": "adsl07", "result": "失败", "domain": "ddns.example.com",
				"event": "错误: (adsl07)网络接口异常", "ip_addr": "--", "id": 1, "timestamp": 1790279675},
		},
	})
	page, err := New(ts.Client()).SystemLog.DDNSLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("DDNSLogs 失败: %v", err)
	}
	if page.Total != 2066 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Timestamp != 1790279675 || r.Interface != "adsl07" ||
		r.Domain != "ddns.example.com" || r.Event != "错误: (adsl07)网络接口异常" ||
		r.Result != "失败" || r.IPAddr != "--" {
		t.Fatalf("DDNS 日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-ddns")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestWANPPPoELogs 覆盖 syslog-wanpppoe.show（真机实测 total=5000）。
func TestWANPPPoELogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-wanpppoe", "show", map[string]any{
		"total": 5000,
		"data": []any{
			map[string]any{"timestamp": 1793576081, "interface": "adsl12",
				"content": "Timeout waiting for PADO packets", "id": 26933},
		},
	})
	page, err := New(ts.Client()).SystemLog.WANPPPoELogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("WANPPPoELogs 失败: %v", err)
	}
	if page.Total != 5000 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 26933 || r.Timestamp != 1793576081 || r.Interface != "adsl12" ||
		r.Content != "Timeout waiting for PADO packets" {
		t.Fatalf("外网拨号日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-wanpppoe")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestNoticeLogs 覆盖 syslog-notice.show（真机空表 total=0，行字段按
// i18n noticeLog 推导，待真机验证）。
func TestNoticeLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-notice", "show", map[string]any{
		"total": 0,
		"data": []any{
			map[string]any{"id": 7, "timestamp": 1790279675,
				"type": NoticeTypeExpire, "event": "拨号账号即将到期"},
		},
	})
	page, err := New(ts.Client()).SystemLog.NoticeLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("NoticeLogs 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 7 || r.Timestamp != 1790279675 || r.Type != NoticeTypeExpire ||
		r.Event != "拨号账号即将到期" {
		t.Fatalf("推送通知日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-notice")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestWebAdminLogs 覆盖 syslog-webadmin.show（真机实测 total=610）。
func TestWebAdminLogs(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-webadmin", "show", map[string]any{
		"total": 610,
		"data": []any{
			map[string]any{"username": "admin", "ip_addr": "112.224.223.90",
				"function": "--", "event": "登录", "id": 1, "timestamp": 1790279784},
		},
	})
	page, err := New(ts.Client()).SystemLog.WebAdminLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("WebAdminLogs 失败: %v", err)
	}
	if page.Total != 610 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Timestamp != 1790279784 || r.Username != "admin" ||
		r.IPAddr != "112.224.223.90" || r.Function != "--" || r.Event != "登录" {
		t.Fatalf("操作日志行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("syslog-webadmin")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestSyslogListOptions 覆盖分页/过滤/关键字检索参数的透传（真机实测
// FILTER1="interface,==,lan4" 与 KEYWORDS+FINDS 均生效）。
func TestSyslogListOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "第二页",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 20},
			want: map[string]any{"TYPE": "total,data", "limit": "20,20"},
		},
		{
			name: "按接口过滤（真机实测形态）",
			opts: &ikuai.ListOptions{Limit: 3,
				Filters: []ikuai.Filter{{Field: "interface", Op: "==", Value: "lan4"}}},
			want: map[string]any{"TYPE": "total,data", "limit": "0,3", "FILTER1": "interface,==,lan4"},
		},
		{
			name: "关键字检索（真机实测形态）",
			opts: &ikuai.ListOptions{Limit: 3, Keywords: "DHCPACK", Finds: "msgtype,event,mac,ip_addr"},
			want: map[string]any{"TYPE": "total,data", "limit": "0,3",
				"KEYWORDS": "DHCPACK", "FINDS": "msgtype,event,mac,ip_addr"},
		},
		{
			name: "排序",
			opts: &ikuai.ListOptions{Limit: 3, OrderBy: "timestamp", Order: "desc"},
			want: map[string]any{"TYPE": "total,data", "limit": "0,3",
				"ORDER_BY": "timestamp", "ORDER": "desc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("syslog-dhcpd", "show", map[string]any{"total": 0, "data": []any{}})
			if _, err := New(ts.Client()).SystemLog.DHCPLogs(context.Background(), tt.opts); err != nil {
				t.Fatalf("DHCPLogs 失败: %v", err)
			}
			wantParam(t, paramOf(t, ts.CallsFor("syslog-dhcpd")[0]), tt.want)
		})
	}
}

// TestSyslogInterfaces 覆盖 syslog-*.show TYPE=all 的接口选择器区块
// （真机实测 9 张表均返回 interface 列表）。
func TestSyslogInterfaces(t *testing.T) {
	for _, fn := range []SyslogFunc{SyslogPPPAuth, SyslogARP, SyslogAPAction,
		SyslogDHCPd, SyslogDDNS, SyslogWANPPPoE, SyslogNotice, SyslogSysEvent, SyslogWebAdmin} {
		t.Run(string(fn), func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse(string(fn), "show", map[string]any{
				"total": 0, "data": []any{},
				"interface": []any{"adsl01", "adsl02", "lan4", "vlan1211"},
			})
			ifaces, err := New(ts.Client()).SystemLog.Interfaces(context.Background(), fn)
			if err != nil {
				t.Fatalf("Interfaces 失败: %v", err)
			}
			if len(ifaces) != 4 || ifaces[0] != "adsl01" || ifaces[3] != "vlan1211" {
				t.Fatalf("接口选择器解码不符: %v", ifaces)
			}
			wantParam(t, paramOf(t, ts.CallsFor(string(fn))[0]), map[string]any{"TYPE": "all"})
		})
	}
}

// TestSyslogError 覆盖 9 张表的固件错误路径（Result!=30000 → *ikuai.APIError）。
func TestSyslogError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []SyslogFunc{SyslogPPPAuth, SyslogARP, SyslogAPAction,
		SyslogDHCPd, SyslogDDNS, SyslogWANPPPoE, SyslogNotice, SyslogSysEvent, SyslogWebAdmin} {
		ts.HandleFuncName(string(fn), func(action string, param json.RawMessage) (any, error) {
			return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)",
				FuncName: string(fn), Action: action}
		})
	}
	svc := New(ts.Client()).SystemLog
	ctx := context.Background()
	calls := []struct {
		name string
		call func() error
	}{
		{"PPPAuthLogs", func() error { _, err := svc.PPPAuthLogs(ctx, nil); return err }},
		{"ARPLogs", func() error { _, err := svc.ARPLogs(ctx, nil); return err }},
		{"APActionLogs", func() error { _, err := svc.APActionLogs(ctx, nil); return err }},
		{"DHCPLogs", func() error { _, err := svc.DHCPLogs(ctx, nil); return err }},
		{"DDNSLogs", func() error { _, err := svc.DDNSLogs(ctx, nil); return err }},
		{"WANPPPoELogs", func() error { _, err := svc.WANPPPoELogs(ctx, nil); return err }},
		{"NoticeLogs", func() error { _, err := svc.NoticeLogs(ctx, nil); return err }},
		{"SysEventLogs", func() error { _, err := svc.SysEventLogs(ctx, nil); return err }},
		{"WebAdminLogs", func() error { _, err := svc.WebAdminLogs(ctx, nil); return err }},
		{"Interfaces", func() error { _, err := svc.Interfaces(ctx, SyslogDHCPd); return err }},
	}
	for _, c := range calls {
		if err := c.call(); err == nil {
			t.Fatalf("%s 错误应上抛", c.name)
		} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
			t.Fatalf("%s 期望 APIError(30006), got %v", c.name, err)
		}
	}
}
