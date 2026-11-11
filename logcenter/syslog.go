package logcenter

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// 9 张 syslog 日志表的 show 形态说明（真机 3.7.24 实测）：
//   - 全部接受 TYPE=total,data 标准分页（{"total":N,"data":[...]}），支持
//     limit/ORDER_BY/ORDER/FILTER1..n/FINDS/KEYWORDS（实测
//     FILTER1="interface,==,lan4"、KEYWORDS="DHCPACK"+FINDS 均生效）。
//   - TYPE=all 在 total/data 之外附带 interface 区块：本机全部接口/线路名
//     下拉可选值（9 张表实测均为 13 项，如 adsl01..adsl14/vlan1211 等）。
//   - 无参数 {} 时不带 total、默认返回 100 行；SDK 统一走 TYPE=total,data。
//   - 有数据的 6 张表行结构以真机数据为准逐字段核对；空表（pppauth/
//     apaction/notice，total=0）行字段按固件 i18n 列名推导，待真机验证。

// PPPAuthLogRow 是一条认证日志行（syslog-pppauth.show 的 data 元素，
// Web 控制台「日志中心 → 用户日志 → 认证日志」页面）。
//
// ⚠️ 探测真机（3.7.24）上该表为空（total=0），行字段无法从真机数据核对；
// 字段按固件 i18n webauthLog 列名（auth_type/mac/ip_addr/event）推导，
// id/timestamp 为各 syslog 表共有的行结构。空表推导，待真机验证。
type PPPAuthLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// AuthType 认证类型（值域见 PPPAuthType* 常量）。空表推导，待真机验证。
	AuthType string `json:"auth_type"`
	// MAC 认证 IP/MAC。空表推导，待真机验证。
	MAC string `json:"mac"`
	// IPAddr 认证 IP。空表推导，待真机验证。
	IPAddr string `json:"ip_addr"`
	// Event 事件描述（认证成功/认证失败/退出登录等）。空表推导，待真机验证。
	Event string `json:"event"`
}

// 认证类型值（固件 i18n webauthLog ppptype_map 键值，与固件一致；行字段
// 与值的落点待真机验证）。
const (
	PPPAuthTypeUser      = "1"    // WEB-用户认证
	PPPAuthTypeCoupons   = "2"    // WEB-上网码认证
	PPPAuthTypeFixedCode = "3"    // WEB-固定密码
	PPPAuthTypePhone     = "4"    // WEB-手机认证
	PPPAuthTypeQQ        = "5"    // WEB-QQ认证
	PPPAuthTypeMicroblog = "6"    // WEB-微博认证
	PPPAuthTypeKey       = "7"    // WEB-一键认证/倒计时认证
	PPPAuthTypeWeChat    = "8"    // WEB-微信认证
	PPPAuthTypeTry       = "11"   // WEB-试用认证
	PPPAuthTypeWikiWiFi  = "12"   // WEB-微信连WiFi
	PPPAuthTypeRadius    = "13"   // WEB-Radius对接
	PPPAuthTypeInternet  = "16"   // WEB-红包上网
	PPPAuthTypeThird     = "99"   // WEB-第三方对接
	PPPAuthTypePPTP      = "pptp" // PPTP 拨号
	PPPAuthTypeL2TP      = "l2tp" // L2TP 拨号
	PPPAuthTypeOpenVPN   = "openvpn"
	PPPAuthTypePPPoE     = "pppoe" // WAN 口 PPPoE 拨号
)

// ARPLogRow 是一条 ARP 日志行（syslog-arp.show 的 data 元素，Web 控制台
// 「日志中心 → 用户日志 → ARP日志」页面；真机 3.7.24 实测行结构，
// total=26）。与系统日志行同构（id/timestamp/content）。
type ARPLogRow = SyslogRow

// SyslogRow 是 syslog-arp / syslog-sysevent 共用的日志行结构
// （真机 3.7.24 实测，两表行字段完全一致）。
type SyslogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Content 日志内容（如 ARP 欺骗告警 "检测到一个ARP地址欺骗在接口(_vlan3): ..."、
	// 系统事件 "系统启动"）。
	Content string `json:"content"`
}

// SysEventLogRow 是一条系统日志行（syslog-sysevent.show 的 data 元素，
// Web 控制台「日志中心 → 系统日志 → 系统日志」页面；真机 3.7.24 实测，
// total=4168）。与 ARP 日志行同构。
type SysEventLogRow = SyslogRow

// APActionLogRow 是一条 AP 操作日志行（syslog-apaction.show 的 data 元素，
// Web 控制台「日志中心 → 用户日志 → 无线终端日志」页面）。
//
// ⚠️ 探测真机（3.7.24）上该表为空（total=0），行字段无法从真机数据核对；
// 字段按固件 i18n terminalLog 列名（client_mac/ap_mac/signal/action）推导，
// id/timestamp 为各 syslog 表共有的行结构。空表推导，待真机验证。
type APActionLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// ClientMAC 设备（无线终端）MAC。空表推导，待真机验证。
	ClientMAC string `json:"client_mac"`
	// APMAC AP 的 MAC。空表推导，待真机验证。
	APMAC string `json:"ap_mac"`
	// Signal 信号强度（固件同类字段 monitor_lanip.signal 为字符串）。
	// 空表推导，类型与语义待真机验证。
	Signal string `json:"signal"`
	// Action 操作行为（登录/退出）。空表推导，待真机验证。
	Action string `json:"action"`
}

// DHCPLogRow 是一条 DHCP 日志行（syslog-dhcpd.show 的 data 元素，
// Web 控制台「日志中心 → 功能日志 → DHCP日志」页面；真机 3.7.24 实测
// 行结构，total=5808，同时覆盖 DHCPv4/DHCPv6 行）。
type DHCPLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// MsgType 消息类型（真机实测值：DHCPDISCOVER/DHCPOFFER/DHCPREQUEST/
	// DHCPACK 及 DHCPV6SOLICIT/DHCPV6REQUEST/DHCPV6RENEW/DHCPV6REBIND 等）。
	MsgType string `json:"msgtype"`
	// Interface 接口（如 lan4/vlan1211）。
	Interface string `json:"interface"`
	// Event 事件描述（真机实测值如 "IA_NA"/"--"）。
	Event string `json:"event"`
	// MAC MAC 地址/本地链接地址（IPv6 行为客户端链路本地地址）。
	MAC string `json:"mac"`
	// IPAddr IP 地址（IPv6 行含前缀长度与尾随空格，如
	// "2408:8220:32a:b3d1::597/128 "，固件原样返回，SDK 不做修剪）。
	IPAddr string `json:"ip_addr"`
}

// DDNSLogRow 是一条动态域名日志行（syslog-ddns.show 的 data 元素，
// Web 控制台「日志中心 → 功能日志 → 动态域名日志」页面；真机 3.7.24
// 实测行结构，total=2066）。
type DDNSLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Interface 解析网卡/线路（如 adsl07）。
	Interface string `json:"interface"`
	// Domain 域名（如 "ddns.example.com"）。
	Domain string `json:"domain"`
	// Event 事件描述（真机实测值如 "错误: (adsl07)网络接口异常"）。
	Event string `json:"event"`
	// Result 更新结果（真机实测值如 "失败"/"成功"）。
	Result string `json:"result"`
	// IPAddr 更新的 IP 地址（失败时为 "--"）。
	IPAddr string `json:"ip_addr"`
}

// WANPPPoELogRow 是一条外网拨号日志行（syslog-wanpppoe.show 的 data 元素，
// Web 控制台「日志中心 → 功能日志 → 外网拨号日志」页面；真机 3.7.24
// 实测行结构，total=5000，固件按存储上限轮转，total 小于最大行 ID）。
type WANPPPoELogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Interface 线路（如 adsl12；页面下拉「全部线路」即 TYPE=all 的
	// interface 区块）。
	Interface string `json:"interface"`
	// Content 日志详情（真机实测字段名为 content，如
	// "Timeout waiting for PADO packets"；固件 i18n 列名「日志详情」）。
	Content string `json:"content"`
}

// NoticeLogRow 是一条推送通知日志行（syslog-notice.show 的 data 元素，
// Web 控制台「日志中心 → 功能日志 → 推送通知日志」页面）。
//
// ⚠️ 探测真机（3.7.24）上该表为空（total=0），行字段无法从真机数据核对；
// 字段按固件 i18n noticeLog 列名（type/event）推导，id/timestamp 为各
// syslog 表共有的行结构。空表推导，待真机验证。
type NoticeLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Type 通知类型（值域见 NoticeType* 常量）。空表推导，待真机验证。
	Type string `json:"type"`
	// Event 事件描述。空表推导，待真机验证。
	Event string `json:"event"`
}

// 通知类型值（固件 i18n noticeLog 键值，与固件一致；行字段与值的落点
// 待真机验证）。
const (
	NoticeTypeReal   = "real_inform"   // 实时通知
	NoticeTypeCycle  = "cycle_inform"  // 定期通知
	NoticeTypeExpire = "expire_inform" // 到期提醒
	NoticeTypePast   = "past_inform"   // 拨号用户过期通知
)

// WebAdminLogRow 是一条管理员操作日志行（syslog-webadmin.show 的 data
// 元素，Web 控制台「日志中心 → 系统日志 → 操作日志」页面；真机 3.7.24
// 实测行结构，total=610）。
type WebAdminLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Username 管理员账号（真机实测值如 "admin"）。
	Username string `json:"username"`
	// IPAddr 操作来源 IP（真机实测值如公网/内网 IP）。
	IPAddr string `json:"ip_addr"`
	// Function 功能（真机实测值如 "--"）。
	Function string `json:"function"`
	// Event 事件（真机实测值如 "登录"）。
	Event string `json:"event"`
}

// SystemLogService 是 9 张 syslog 日志表（全部只读）的子服务。
type SystemLogService struct {
	c *ikuai.Client
}

// PPPAuthLogs 获取认证日志列表（只读分页）。
// 对应 /Action/call: syslog-pppauth.show（TYPE=total,data），
// 即 Web 控制台「日志中心 → 用户日志 → 认证日志」页面。
// 行字段为空表推导，待真机验证。
func (s *SystemLogService) PPPAuthLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[PPPAuthLogRow], error) {
	return listSyslog[PPPAuthLogRow](ctx, s.c, SyslogPPPAuth, opts)
}

// ARPLogs 获取 ARP 日志列表（只读分页）。
// 对应 /Action/call: syslog-arp.show（TYPE=total,data，真机实测 total=26），
// 即 Web 控制台「日志中心 → 用户日志 → ARP日志」页面。
func (s *SystemLogService) ARPLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ARPLogRow], error) {
	return listSyslog[ARPLogRow](ctx, s.c, SyslogARP, opts)
}

// APActionLogs 获取 AP 操作日志列表（只读分页）。
// 对应 /Action/call: syslog-apaction.show（TYPE=total,data），
// 即 Web 控制台「日志中心 → 用户日志 → 无线终端日志」页面。
// 行字段为空表推导，待真机验证。
func (s *SystemLogService) APActionLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[APActionLogRow], error) {
	return listSyslog[APActionLogRow](ctx, s.c, SyslogAPAction, opts)
}

// DHCPLogs 获取 DHCP 日志列表（只读分页，含 DHCPv4/DHCPv6 行）。
// 对应 /Action/call: syslog-dhcpd.show（TYPE=total,data，真机实测 total=5808），
// 即 Web 控制台「日志中心 → 功能日志 → DHCP日志」页面。
func (s *SystemLogService) DHCPLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DHCPLogRow], error) {
	return listSyslog[DHCPLogRow](ctx, s.c, SyslogDHCPd, opts)
}

// DDNSLogs 获取动态域名日志列表（只读分页）。
// 对应 /Action/call: syslog-ddns.show（TYPE=total,data，真机实测 total=2066），
// 即 Web 控制台「日志中心 → 功能日志 → 动态域名日志」页面。
func (s *SystemLogService) DDNSLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DDNSLogRow], error) {
	return listSyslog[DDNSLogRow](ctx, s.c, SyslogDDNS, opts)
}

// WANPPPoELogs 获取外网拨号日志列表（只读分页）。
// 对应 /Action/call: syslog-wanpppoe.show（TYPE=total,data，真机实测
// total=5000，含 WAN 口与 VPN 服务拨号），
// 即 Web 控制台「日志中心 → 功能日志 → 外网拨号日志」页面。
func (s *SystemLogService) WANPPPoELogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[WANPPPoELogRow], error) {
	return listSyslog[WANPPPoELogRow](ctx, s.c, SyslogWANPPPoE, opts)
}

// NoticeLogs 获取推送通知日志列表（只读分页）。
// 对应 /Action/call: syslog-notice.show（TYPE=total,data），
// 即 Web 控制台「日志中心 → 功能日志 → 推送通知日志」页面。
// 行字段为空表推导，待真机验证。
func (s *SystemLogService) NoticeLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[NoticeLogRow], error) {
	return listSyslog[NoticeLogRow](ctx, s.c, SyslogNotice, opts)
}

// SysEventLogs 获取系统日志列表（只读分页）。
// 对应 /Action/call: syslog-sysevent.show（TYPE=total,data，真机实测
// total=4168），即 Web 控制台「日志中心 → 系统日志 → 系统日志」页面。
func (s *SystemLogService) SysEventLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SysEventLogRow], error) {
	return listSyslog[SysEventLogRow](ctx, s.c, SyslogSysEvent, opts)
}

// WebAdminLogs 获取管理员操作日志列表（只读分页）。
// 对应 /Action/call: syslog-webadmin.show（TYPE=total,data，真机实测
// total=610），即 Web 控制台「日志中心 → 系统日志 → 操作日志」页面。
func (s *SystemLogService) WebAdminLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[WebAdminLogRow], error) {
	return listSyslog[WebAdminLogRow](ctx, s.c, SyslogWebAdmin, opts)
}

// Interfaces 获取指定 syslog 日志表的接口/线路选择器可选值（只读）。
// 对应 /Action/call: <func>.show（TYPE=all，真机实测响应附带 interface
// 区块，如 adsl01..adsl14/vlan1211 等 13 项）。
// fn 取 SyslogFunc 常量之一；关键字检索与按接口过滤通过
// List 的 opts.Keywords/opts.Filters 完成（真机实测 FILTER1=
// "interface,==,<线路名>" 生效）。
func (s *SystemLogService) Interfaces(ctx context.Context, fn SyslogFunc) ([]string, error) {
	var out struct {
		Interface []string `json:"interface"`
	}
	if err := s.c.CallData(ctx, string(fn), "show",
		map[string]any{"TYPE": SyslogTypeAll}, &out); err != nil {
		return nil, err
	}
	return out.Interface, nil
}
