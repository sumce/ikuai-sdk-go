package logcenter

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是日志中心大类的服务入口，聚合 Web 控制台「日志中心」菜单下的
// 全部 11 个功能（菜单树 func 与页面的对应关系见 docs/API_CATALOG.md §12）。
type Service struct {
	c *ikuai.Client

	// SystemLog syslog 系统日志表（syslog-pppauth/syslog-arp/syslog-apaction/
	// syslog-dhcpd/syslog-ddns/syslog-wanpppoe/syslog-notice/syslog-sysevent/
	// syslog-webadmin 共 9 张只读日志表，统一 List + 关键字检索模式）。
	SystemLog *SystemLogService
	// Messages 站内信系统消息（ikmessages：消息通知页，含消息推送 start）。
	Messages *MessagesService
	// RTCollect 实时采集告警信息（rt_collect：告警信息页）。
	RTCollect *RTCollectService
}

// New 构造日志中心服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:         c,
		SystemLog: &SystemLogService{c: c},
		Messages:  &MessagesService{c: c},
		RTCollect: &RTCollectService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用日志中心大类接口，作为未封装
// action/TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
// 注意：日志中心绝大多数 func 为只读 show；本包未实测的写动作（如批量
// 标记已读/清空日志）未经真机验证，只应在充分知晓后果的前提下通过 Raw
// 调用，且严禁对真机执行写测试（见 docs/RULES.md §9.2 与本包各写方法注释）。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// SyslogFunc 是 syslog 日志表的 func_name（带连字符，与固件完全一致）。
type SyslogFunc string

// 9 张 syslog 日志表的 func_name（固件 menu.json 实证）。
const (
	// SyslogPPPAuth 认证日志（syslog-pppauth，Web 控制台「日志中心 → 用户日志 → 认证日志」）。
	SyslogPPPAuth SyslogFunc = "syslog-pppauth"
	// SyslogARP ARP 日志（syslog-arp，「用户日志 → ARP日志」）。
	SyslogARP SyslogFunc = "syslog-arp"
	// SyslogAPAction 无线终端日志（syslog-apaction，「用户日志 → 无线终端日志」）。
	SyslogAPAction SyslogFunc = "syslog-apaction"
	// SyslogDHCPd DHCP 日志（syslog-dhcpd，「功能日志 → DHCP日志」）。
	SyslogDHCPd SyslogFunc = "syslog-dhcpd"
	// SyslogDDNS 动态域名日志（syslog-ddns，「功能日志 → 动态域名日志」）。
	SyslogDDNS SyslogFunc = "syslog-ddns"
	// SyslogWANPPPoE 外网拨号日志（syslog-wanpppoe，「功能日志 → 外网拨号日志」）。
	SyslogWANPPPoE SyslogFunc = "syslog-wanpppoe"
	// SyslogNotice 推送通知日志（syslog-notice，「功能日志 → 推送通知日志」）。
	SyslogNotice SyslogFunc = "syslog-notice"
	// SyslogSysEvent 系统日志（syslog-sysevent，「系统日志 → 系统日志」）。
	SyslogSysEvent SyslogFunc = "syslog-sysevent"
	// SyslogWebAdmin 操作日志（syslog-webadmin，「系统日志 → 操作日志」）。
	SyslogWebAdmin SyslogFunc = "syslog-webadmin"
)

// 本包 show 动作的 TYPE 选择器常量（真机 3.7.24 实测）。
const (
	// SyslogTypeAll 是 syslog-*.show 的表单选择器 TYPE 值：响应在 total/data
	// 之外附带 interface 区块（接口/线路下拉可选值，9 张表实测均返回 13 项）。
	SyslogTypeAll = "all"
	// MessagesTypeTotal 只取消息总数（ikmessages.show TYPE=total，
	// 配合 FILTER1="read,==,0" 即未读消息数，抓包实证）。
	MessagesTypeTotal = "total"
	// RTCollectTypeAll 是 rt_collect.show 的聚合 TYPE 值：响应包含
	// total/data/warning_notify 三个区块（真机实测）。
	RTCollectTypeAll = "all"
	// RTCollectTypeWarningNotify 只取告警提醒标志（rt_collect.show
	// TYPE=warning_notify → {"warning_notify":1}，抓包实证）。
	RTCollectTypeWarningNotify = "warning_notify"
)

// selectorShowParam 构造带固定默认键的 show 参数：opts == nil 时仅由
// defaults 生成参数；opts 非空时以 opts.BuildParam() 为底，defaults 只补
// 缺失键，用户在 opts.Type / opts.Extra 中显式给出的值优先。
func selectorShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	var p map[string]any
	if opts != nil {
		p = opts.BuildParam()
	}
	if p == nil {
		p = map[string]any{}
	}
	for k, v := range defaults {
		if _, ok := p[k]; !ok {
			p[k] = v
		}
	}
	return p
}

// pageShowParam 构造标准分页类 func（本包接受 TYPE=total,data 的列表型
// func）的 show 参数：opts == nil 时补 ikuai.DefaultListOptions()
// （TYPE=total,data, limit="0,20"），defaults 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// listSyslog 以标准分页参数调用 syslog-*.show 并解码为强类型行。
// 9 张 syslog 表真机实测均接受 TYPE=total,data 标准分页与
// FILTER/FINDS/KEYWORDS（如 FILTER1="interface,==,lan4"）。
func listSyslog[T any](ctx context.Context, c *ikuai.Client, fn SyslogFunc, opts *ikuai.ListOptions) (*ikuai.Page[T], error) {
	resp, err := c.Call(ctx, string(fn), "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[T](resp.Data)
}
