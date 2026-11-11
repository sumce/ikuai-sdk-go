package logcenter

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSyslogLists 真机只读验证 9 张 syslog 日志表的 show（TYPE=total,data）。
// 全部为只读日志查询，无写动作。探测真机（3.7.24）实测：
// syslog-arp(26)/syslog-dhcpd(5808)/syslog-ddns(2066)/syslog-sysevent(4168)/
// syslog-wanpppoe(5000)/syslog-webadmin(610) 有数据；
// syslog-pppauth/syslog-apaction/syslog-notice 为空表（total=0）。
func TestLiveSyslogLists(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	sys := New(c).SystemLog

	// ARP 日志（真机实测 total=26）。
	page, err := sys.ARPLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-arp.show 失败: %v", err)
	}
	t.Logf("ARP 日志: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d content=%s", r.ID, r.Content)
	}
	if page.Total <= 0 {
		t.Error("ARP 日志不应为空（探测真机实测 26 条）")
	}
	if r := page.Rows[0]; r.ID == 0 || r.Timestamp == 0 || r.Content == "" {
		t.Errorf("ARP 日志行字段与探测形态不符: %+v", r)
	}

	// DHCP 日志（真机实测 total=5808，含 DHCPv6 行）。
	dhcp, err := sys.DHCPLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-dhcpd.show 失败: %v", err)
	}
	t.Logf("DHCP 日志: total=%d 本页行数=%d", dhcp.Total, len(dhcp.Rows))
	for i, r := range dhcp.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: msgtype=%s iface=%s mac=%s ip=%s event=%s", r.MsgType, r.Interface, r.MAC, r.IPAddr, r.Event)
	}
	if dhcp.Total <= 0 {
		t.Error("DHCP 日志不应为空（探测真机实测 5808 条）")
	}
	if r := dhcp.Rows[0]; r.MsgType == "" || r.Interface == "" {
		t.Errorf("DHCP 日志行字段与探测形态不符: %+v", r)
	}

	// DDNS 日志（真机实测 total=2066）。
	ddns, err := sys.DDNSLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-ddns.show 失败: %v", err)
	}
	t.Logf("DDNS 日志: total=%d 本页行数=%d", ddns.Total, len(ddns.Rows))
	if r := ddns.Rows[0]; r.Interface == "" || r.Result == "" {
		t.Errorf("DDNS 日志行字段与探测形态不符: %+v", r)
	}

	// 外网拨号日志（真机实测 total=5000）。
	pppoe, err := sys.WANPPPoELogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-wanpppoe.show 失败: %v", err)
	}
	t.Logf("外网拨号日志: total=%d 本页行数=%d", pppoe.Total, len(pppoe.Rows))
	if r := pppoe.Rows[0]; r.Interface == "" || r.Content == "" {
		t.Errorf("外网拨号日志行字段与探测形态不符: %+v", r)
	}

	// 系统日志（真机实测 total=4168）。
	se, err := sys.SysEventLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-sysevent.show 失败: %v", err)
	}
	t.Logf("系统日志: total=%d 本页行数=%d", se.Total, len(se.Rows))
	if r := se.Rows[0]; r.Timestamp == 0 || r.Content == "" {
		t.Errorf("系统日志行字段与探测形态不符: %+v", r)
	}

	// 操作日志（真机实测 total=610）。
	wa, err := sys.WebAdminLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-webadmin.show 失败: %v", err)
	}
	t.Logf("操作日志: total=%d 本页行数=%d", wa.Total, len(wa.Rows))
	if r := wa.Rows[0]; r.Username == "" || r.Event == "" {
		t.Errorf("操作日志行字段与探测形态不符: %+v", r)
	}

	// 认证日志（真机空表 total=0；仅验证调用成功与字段解码）。
	pa, err := sys.PPPAuthLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-pppauth.show 失败: %v", err)
	}
	t.Logf("认证日志: total=%d 本页行数=%d（行字段空表推导，待真机验证）", pa.Total, len(pa.Rows))

	// 无线终端日志（真机空表）。
	ap, err := sys.APActionLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-apaction.show 失败: %v", err)
	}
	t.Logf("无线终端日志: total=%d 本页行数=%d（行字段空表推导，待真机验证）", ap.Total, len(ap.Rows))

	// 推送通知日志（真机空表）。
	no, err := sys.NoticeLogs(ctx, &ikuai.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("syslog-notice.show 失败: %v", err)
	}
	t.Logf("推送通知日志: total=%d 本页行数=%d（行字段空表推导，待真机验证）", no.Total, len(no.Rows))
}

// TestLiveSyslogInterfaces 真机只读验证 TYPE=all 的接口选择器区块
// （探测真机 9 张表均返回 13 项线路/接口）。
func TestLiveSyslogInterfaces(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	sys := New(c).SystemLog
	for _, fn := range []SyslogFunc{SyslogPPPAuth, SyslogARP, SyslogAPAction,
		SyslogDHCPd, SyslogDDNS, SyslogWANPPPoE, SyslogNotice, SyslogSysEvent, SyslogWebAdmin} {
		ifaces, err := sys.Interfaces(ctx, fn)
		if err != nil {
			t.Fatalf("%s.show TYPE=all 失败: %v", fn, err)
		}
		t.Logf("%s 接口选择器: %v", fn, ifaces)
		if len(ifaces) == 0 {
			t.Errorf("%s 接口选择器不应为空（探测真机实测 13 项）", fn)
		}
	}
}

// TestLiveSyslogKeywordSearch 真机只读验证关键字检索形态
// （探测真机实测 KEYWORDS="DHCPACK" + FINDS 命中 1088 条）。
func TestLiveSyslogKeywordSearch(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	page, err := New(c).SystemLog.DHCPLogs(ctx, &ikuai.ListOptions{Limit: 3,
		Keywords: "DHCPACK", Finds: "msgtype,event,mac,ip_addr"})
	if err != nil {
		t.Fatalf("syslog-dhcpd.show KEYWORDS 检索失败: %v", err)
	}
	t.Logf("DHCP 日志 DHCPACK 检索: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: msgtype=%s mac=%s ip=%s", r.MsgType, r.MAC, r.IPAddr)
	}
	if page.Total <= 0 {
		t.Error("DHCPACK 检索不应为空（探测真机实测 1088 条）")
	}
}

// TestLiveSyslogInterfaceFilter 真机只读验证按接口过滤形态
// （探测真机实测 FILTER1="interface,==,lan4" 生效）。
func TestLiveSyslogInterfaceFilter(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	page, err := New(c).SystemLog.DHCPLogs(ctx, &ikuai.ListOptions{Limit: 3,
		Filters: []ikuai.Filter{{Field: "interface", Op: "==", Value: "lan4"}}})
	if err != nil {
		t.Fatalf("syslog-dhcpd.show FILTER 失败: %v", err)
	}
	t.Logf("DHCP 日志 lan4 过滤: total=%d 本页行数=%d", page.Total, len(page.Rows))
	if page.Total <= 0 {
		t.Error("lan4 接口过滤不应为空（探测真机实测 3266 条）")
	}
}
