package logcenter

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// rt_collect（实时采集告警信息，Web 控制台「日志中心 → 系统日志 → 告警信息」
// 页面）的 show 形态说明（真机 3.7.24 实测）：
//   - show TYPE=total,data：标准分页 {"total":N,"data":[...]}（实测 total=33，
//     FILTER1="readflag,==,0" 过滤未读生效）。
//   - show TYPE=all：聚合区块 total/data/warning_notify（实测 33 行 +
//     warning_notify=1）。
//   - show TYPE=warning_notify：仅 {"warning_notify":1}（抓包实证，
//     analysis/calls/rt_collect__show.json；Web 控制台用其做菜单未读角标）。
//   - 实时推送走 WebSocket 通道 /Action/rtcollect/websocket（抓包实证
//     101 Switching Protocols），SDK 不封装 WebSocket。
//   - 固件 i18n perceivableLog 列名与行字段对应：msgtype=消息类型、
//     priority=消息级别（1 信息/2 警告/3 故障）、source=来源、
//     content=内容描述、readflag=已读/未读。

// 告警消息级别值（rt_collect priority 字段，固件 i18n perceivableLog.level
// 枚举，真机实测 1/2 两值）。
const (
	AlertPriorityInfo    = 1 // 信息（真机实测）
	AlertPriorityWarning = 2 // 警告（真机实测）
	AlertPriorityFault   = 3 // 故障（i18n 枚举，真机样本未出现）
)

// 告警已读标记值（rt_collect readflag 字段；未读=0 真机实测，已读取值待真机验证）。
const (
	AlertReadFlagUnread = 0 // 未读（真机实测）
	AlertReadFlagRead   = 1 // 已读（推测，待真机验证）
)

// AlertRow 是一条实时采集告警信息行（rt_collect.show 的 data 元素；
// 真机 3.7.24 实测行结构）。
type AlertRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// MsgType 消息类型（真机实测值如 "网络设置"/"状态监控"/"驱动插件"）。
	MsgType string `json:"msgtype"`
	// Source 来源（真机实测值如 "本地设备(vwan1)"/"本地设备"/"本机设备"）。
	Source string `json:"source"`
	// Priority 消息级别（1=信息 2=警告 3=故障，见 AlertPriority* 常量）。
	Priority int64 `json:"priority"`
	// Content 内容描述（真机实测值如 "网段冲突"/"路由端口状态变更"）。
	Content string `json:"content"`
	// Detail 详情（真机实测值如 "vwan1与lan3网段冲突,网段详情：192.168.1.23/24"）。
	Detail string `json:"detail"`
	// Solution 处理建议（真机实测无建议时为 "--"）。
	Solution string `json:"solution"`
	// ReadFlag 已读标记（0=未读，见 AlertReadFlag* 常量）。
	ReadFlag int64 `json:"readflag"`
}

// AlertsAll 是告警信息页的聚合结果（rt_collect.show TYPE=all，真机
// 3.7.24 实测区块 total/data/warning_notify）。
type AlertsAll struct {
	// Total 告警总数。
	Total int64 `json:"total"`
	// Rows 告警行。
	Rows []AlertRow `json:"data"`
	// WarningNotify 告警提醒标志（真机实测 1；语义推测为菜单未读角标
	// 开关，待真机验证）。
	WarningNotify int64 `json:"warning_notify"`
}

// RTCollectService 是实时采集告警信息（rt_collect，只读）的子服务。
type RTCollectService struct {
	c *ikuai.Client
}

// Alerts 获取告警信息列表（只读分页）。
// 对应 /Action/call: rt_collect.show（TYPE=total,data，真机实测），
// 即 Web 控制台「日志中心 → 系统日志 → 告警信息」页面。
// 按已读状态过滤可用 opts.Filters 添加 {"readflag,==,0"}（真机实测生效）。
func (s *RTCollectService) Alerts(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[AlertRow], error) {
	resp, err := s.c.Call(ctx, "rt_collect", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[AlertRow](resp.Data)
}

// AlertsAll 获取告警信息聚合区块（总数 + 行 + 告警提醒标志，只读）。
// 对应 /Action/call: rt_collect.show（TYPE=all，真机实测区块
// total/data/warning_notify）。
func (s *RTCollectService) AlertsAll(ctx context.Context) (*AlertsAll, error) {
	var out AlertsAll
	if err := s.c.CallData(ctx, "rt_collect", "show",
		map[string]any{"TYPE": RTCollectTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WarningNotify 获取告警提醒标志（只读）。
// 对应 /Action/call: rt_collect.show（TYPE=warning_notify，抓包实证 →
// {"warning_notify":1}），Web 控制台用其做「告警信息」菜单角标。
func (s *RTCollectService) WarningNotify(ctx context.Context) (int64, error) {
	var p struct {
		WarningNotify int64 `json:"warning_notify"`
	}
	if err := s.c.CallData(ctx, "rt_collect", "show",
		map[string]any{"TYPE": RTCollectTypeWarningNotify}, &p); err != nil {
		return 0, err
	}
	return p.WarningNotify, nil
}
