package ac

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// ac_status.show 的 TYPE 选择器常量（真机 3.7.24 逐一 show 探测实证；
// 未知 TYPE 固件回 Result=30006 "unknown TYPE (x)"，可与列表 func 混用
// 逗号组合，控制台「无线概况」页使用 TYPE=all）。
const (
	// ACStatusTypeACStatus AC 智能控制开关状态区块（0=关闭 1=开启，语义待真机验证）。
	ACStatusTypeACStatus = "ac_status"
	// ACStatusTypeAll 概况页全量组合（控制台「AC管理 → 无线概况」实际使用的 TYPE）。
	ACStatusTypeAll = "all"
	// ACStatusTypeAPStatus AP 在线统计区块。
	ACStatusTypeAPStatus = "ap_status"
	// ACStatusTypeCltStatus 无线终端统计区块。
	ACStatusTypeCltStatus = "clt_status"
	// ACStatusTypeAPBadStat AP 异常统计区块（含全网累计上下行 total_upload/total_download）。
	ACStatusTypeAPBadStat = "ap_bad_stat"
	// ACStatusTypeAPChanInfo AP 信道信息区块（空表，行结构待真机验证）。
	ACStatusTypeAPChanInfo = "ap_chaninfo"
	// ACStatusTypeAPChanLoad AP 信道负载区块（仅 TYPE=all 内嵌，非独立合法 TYPE）。
	ACStatusTypeAPChanLoad = "ap_chanload"
	// ACStatusTypeAPCountInfo AP 数量分布区块（空表，行结构待真机验证）。
	ACStatusTypeAPCountInfo = "ap_countinfo"
	// ACStatusTypeAPEventData AP 事件列表区块（空表，行结构待真机验证）。
	ACStatusTypeAPEventData = "ap_event_data"
	// ACStatusTypeAPNetStatus AP 健康评分区块。
	ACStatusTypeAPNetStatus = "ap_net_status"
	// ACStatusTypeTotalCountChannel TOP5 信道/AP 占用区块。
	ACStatusTypeTotalCountChannel = "total_count_channel"
	// ACStatusTypeTotalCountFlow 流量统计区块（空表，行结构待真机验证）。
	ACStatusTypeTotalCountFlow = "total_count_flow"
	// ACStatusTypeTotalCountLowRSSI 弱信号终端区块（空表，行结构待真机验证）。
	ACStatusTypeTotalCountLowRSSI = "total_count_low_rssi"
	// ACStatusTypeTotalCountNetStatus 全网健康评分区块。
	ACStatusTypeTotalCountNetStatus = "total_count_net_status"
	// ACStatusTypeTotalCountNetQuality 网络质量曲线区块（空表，行结构待真机验证）。
	ACStatusTypeTotalCountNetQuality = "total_count_netquality"
	// ACStatusTypeTotalCountOnline 在线终端曲线区块（空表，行结构待真机验证）。
	ACStatusTypeTotalCountOnline = "total_count_online"
	// ACStatusTypeTotalOnlineActive 活跃终端分布区块（空表，行结构待真机验证）。
	ACStatusTypeTotalOnlineActive = "total_online_active"
	// ACStatusTypeTotalOnlineManufacturer 终端厂商分布区块（空表，行结构待真机验证）。
	ACStatusTypeTotalOnlineManufacturer = "total_online_manufacturer"
	// ACStatusTypeTotalOnlineSystemType 终端系统分布区块（空表，行结构待真机验证）。
	ACStatusTypeTotalOnlineSystemType = "total_online_system_type"
	// ACStatusTypeTotalOnlineWiFiProto 无线协议分布区块（空表，行结构待真机验证）。
	ACStatusTypeTotalOnlineWiFiProto = "total_online_wifi_proto"
	// ACStatusTypeWPHYUsage 信道利用率/底噪区块（-255 表示无数据）。
	ACStatusTypeWPHYUsage = "wphy_usage"
)

// ACAPStatus 是 AP 在线统计（ac_status.show TYPE=ap_status / all 内嵌）。
type ACAPStatus struct {
	// APCount 总 AP 数。
	APCount int64 `json:"ap_count"`
	// APOnline 正常在线 AP 数。
	APOnline int64 `json:"ap_online"`
	// APOffline 异常/非开启时段 AP 数（枚举语义待真机验证）。
	APOffline int64 `json:"ap_offline"`
	// APPerfer5G 开启「5G 优先」的 AP 数。
	APPerfer5G int64 `json:"ap_perfer_5g"`
	// APRoaming 开启「快速漫游」的 AP 数。
	APRoaming int64 `json:"ap_roaming"`
}

// ACCltStatus 是无线终端统计（ac_status.show TYPE=clt_status / all 内嵌）。
type ACCltStatus struct {
	// CltCount 在线无线终端总数。
	CltCount int64 `json:"clt_count"`
	// CltCount2G 2.4G 频段终端数。
	CltCount2G int64 `json:"clt_count_2g"`
	// CltCount5G 5G 频段终端数。
	CltCount5G int64 `json:"clt_count_5g"`
	// CltMaxOnline 历史最大在线数。
	CltMaxOnline int64 `json:"clt_max_online"`
	// CltActive 活跃终端数（活跃判定标准待真机验证）。
	CltActive int64 `json:"clt_active"`
	// CltInactive 非活跃终端数。
	CltInactive int64 `json:"clt_inactive"`
}

// ACAPLoad 是 AP 负载分布（仅 ac_status.show TYPE=all 内嵌返回；
// ap_load 不是合法的独立 TYPE，真机验证回 30006）。
type ACAPLoad struct {
	// APLoadLow 低负载 AP 数。
	APLoadLow int64 `json:"ap_load_low"`
	// APLoadMedium 中负载 AP 数。
	APLoadMedium int64 `json:"ap_load_medium"`
	// APLoadHigh 高负载 AP 数。
	APLoadHigh int64 `json:"ap_load_high"`
	// APNum 参与负载统计的 AP 数。
	APNum int64 `json:"ap_num"`
}

// ACBadStat 是 AP 异常统计（ac_status.show TYPE=ap_bad_stat / all 内嵌；
// 各计数单位与口径待真机验证）。
type ACBadStat struct {
	// DenyTotal 拒绝接入累计。
	DenyTotal int64 `json:"deny_total"`
	// DelayTotal 延时异常累计。
	DelayTotal int64 `json:"delay_total"`
	// DropptkTotal 丢包异常累计。
	DropptkTotal int64 `json:"dropptk_total"`
	// LowDownrateTotal 低下载速率累计。
	LowDownrateTotal int64 `json:"low_downrate_total"`
	// CoverageTotal 信号覆盖异常累计。
	CoverageTotal int64 `json:"coverage_total"`
	// ChannfTotal 底噪异常累计。
	ChannfTotal int64 `json:"channf_total"`
}

// ACBadStatResult 是 TYPE=ap_bad_stat 的完整返回（异常计数外含全网
// 累计上下行，字节；真机探测实证）。
type ACBadStatResult struct {
	// APBadStat 各类异常计数。
	APBadStat ACBadStat `json:"ap_bad_stat"`
	// TotalUpload 全网累计上行（字节）。
	TotalUpload int64 `json:"total_upload"`
	// TotalDownload 全网累计下行（字节）。
	TotalDownload int64 `json:"total_download"`
}

// ACNetStatus 是 AP 健康评分（ac_status.show TYPE=ap_net_status / all 内嵌；
// 评分区间与扣分口径待真机验证）。
type ACNetStatus struct {
	// ScoreChannel2G 2G 信道状态评分。
	ScoreChannel2G int64 `json:"score_channel_2g"`
	// ScoreChannel5G 5G 信道状态评分。
	ScoreChannel5G int64 `json:"score_channel_5g"`
	// ScoreDropptk2G 2G 丢包评分。
	ScoreDropptk2G int64 `json:"score_dropptk_2g"`
	// ScoreDropptk5G 5G 丢包评分。
	ScoreDropptk5G int64 `json:"score_dropptk_5g"`
	// ScoreConnLoad 接入终端负载评分。
	ScoreConnLoad int64 `json:"score_conn_load"`
	// ScoreDelay 延时评分。
	ScoreDelay int64 `json:"score_delay"`
	// Coverage 信号覆盖率（百分比口径待真机验证）。
	Coverage int64 `json:"coverage"`
	// Delay 延时（单位待真机验证）。
	Delay int64 `json:"delay"`
}

// ACTotalNetStatus 是全网健康评分（ac_status.show TYPE=total_count_net_status）。
type ACTotalNetStatus struct {
	// ScoreActiveUser 活跃用户评分。
	ScoreActiveUser int64 `json:"score_active_user"`
	// ScoreChutilLoad 信道利用率负载评分。
	ScoreChutilLoad int64 `json:"score_chutil_load"`
	// ScoreChannfLoad 底噪负载评分。
	ScoreChannfLoad int64 `json:"score_channf_load"`
	// ScoreConnSta 终端连接评分。
	ScoreConnSta int64 `json:"score_conn_sta"`
	// Delay 延时（单位待真机验证）。
	Delay int64 `json:"delay"`
	// Coverage 信号覆盖率（百分比口径待真机验证）。
	Coverage int64 `json:"coverage"`
	// Dropptk 丢包（单位待真机验证）。
	Dropptk int64 `json:"dropptk"`
}

// ACWPHYUsage 是全网信道利用率/底噪（ac_status.show TYPE=wphy_usage；
// 真机无 AP 时各值为 -255，即「无数据」哨兵值）。
type ACWPHYUsage struct {
	// Chutil2G 2.4G 信道利用率（百分比；-255=无数据）。
	Chutil2G int64 `json:"chutil_2g"`
	// Channf2G 2.4G 信道底噪（dBm；-255=无数据）。
	Channf2G int64 `json:"channf_2g"`
	// Chutil5G 5G 信道利用率（百分比；-255=无数据）。
	Chutil5G int64 `json:"chutil_5g"`
	// Channf5G 5G 信道底噪（dBm；-255=无数据）。
	Channf5G int64 `json:"channf_5g"`
}

// ACTopChannel 是 TOP5 信道/AP 占用（ac_status.show TYPE=total_count_channel；
// 两个子数组行结构待真机验证）。
type ACTopChannel struct {
	// APMacTop5 按 AP 维度的 TOP5（行结构待真机验证）。
	APMacTop5 []ACStatRow `json:"apmac_top5"`
	// ChannelTop5 按信道维度的 TOP5（行结构待真机验证）。
	ChannelTop5 []ACStatRow `json:"channel_top5"`
}

// ACStatusOverview 是无线概况页聚合（ac_status.show TYPE=all，真机探测
// 实证完整键集）。列表/图表区块行结构未知（真机无 AP 恒为空数组），
// 以 ACStatRow 保留原始键值。
type ACStatusOverview struct {
	// ACStatus AC 智能控制开关（0=关闭 1=开启，语义待真机验证）。
	ACStatus int64 `json:"ac_status"`
	// APBadStat AP 异常统计。
	APBadStat ACBadStat `json:"ap_bad_stat"`
	// TotalUpload 全网累计上行（字节）。
	TotalUpload int64 `json:"total_upload"`
	// TotalDownload 全网累计下行（字节）。
	TotalDownload int64 `json:"total_download"`
	// APChanInfo AP 信道信息（行结构待真机验证）。
	APChanInfo []ACStatRow `json:"ap_chaninfo"`
	// APChanLoad AP 信道负载（行结构待真机验证）。
	APChanLoad []ACStatRow `json:"ap_chanload"`
	// APClients AP 接入终端明细（行结构待真机验证）。
	APClients []ACStatRow `json:"ap_clients"`
	// APLoad AP 负载分布。
	APLoad ACAPLoad `json:"ap_load"`
	// APCountInfo AP 数量分布（行结构待真机验证）。
	APCountInfo []ACStatRow `json:"ap_countinfo"`
	// APEventData AP 事件列表（行结构待真机验证）。
	APEventData []ACStatRow `json:"ap_event_data"`
	// APEventTotal AP 事件总数。
	APEventTotal int64 `json:"ap_event_total"`
	// APNetStatus AP 健康评分。
	APNetStatus ACNetStatus `json:"ap_net_status"`
	// APPuniness 弱信号（puniness）AP 数（语义待真机验证）。
	APPuniness int64 `json:"ap_puniness"`
	// APStatus AP 在线统计。
	APStatus ACAPStatus `json:"ap_status"`
	// CltStatus 无线终端统计。
	CltStatus ACCltStatus `json:"clt_status"`
	// TotalCountChannel TOP5 信道/AP 占用。
	TotalCountChannel ACTopChannel `json:"total_count_channel"`
	// TotalCountFlow 流量统计（行结构待真机验证）。
	TotalCountFlow []ACStatRow `json:"total_count_flow"`
	// TotalCountLowRSSI 弱信号终端（行结构待真机验证）。
	TotalCountLowRSSI []ACStatRow `json:"total_count_low_rssi"`
	// TotalCountNetStatus 全网健康评分。
	TotalCountNetStatus ACTotalNetStatus `json:"total_count_net_status"`
	// TotalCountNetQuality 网络质量曲线（行结构待真机验证）。
	TotalCountNetQuality []ACStatRow `json:"total_count_netquality"`
	// TotalCountOnline 在线终端曲线（行结构待真机验证）。
	TotalCountOnline []ACStatRow `json:"total_count_online"`
	// TotalOnlineActive 活跃终端分布（行结构待真机验证）。
	TotalOnlineActive []ACStatRow `json:"total_online_active"`
	// TotalOnlineManufacturer 终端厂商分布（行结构待真机验证）。
	TotalOnlineManufacturer []ACStatRow `json:"total_online_manufacturer"`
	// TotalOnlineSystemType 终端系统分布（行结构待真机验证）。
	TotalOnlineSystemType []ACStatRow `json:"total_online_system_type"`
	// TotalOnlineWiFiProto 无线协议分布（行结构待真机验证）。
	TotalOnlineWiFiProto []ACStatRow `json:"total_online_wifi_proto"`
	// WPHYUsage 信道利用率/底噪。
	WPHYUsage ACWPHYUsage `json:"wphy_usage"`
}

// StatusService 是无线概况（ac_status）的子服务。
type StatusService struct {
	c *ikuai.Client
}

// Overview 获取无线概况页全量聚合（AC 开关/AP 与终端统计/异常计数/
// 健康评分/信道利用/各类图表数据）。
// 对应 /Action/call: ac_status.show（TYPE=all，真机探测实证），
// 即 Web 控制台「AC管理 → 无线概况」页面。
func (s *StatusService) Overview(ctx context.Context) (*ACStatusOverview, error) {
	var out ACStatusOverview
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAll}), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OverviewBy 以自定义 ListOptions（可携 TYPE/分页/过滤）获取无线概况，
// 作为 Overview 的参数化形态；opts 为 nil 时等价 Overview。
// 对应 /Action/call: ac_status.show。
func (s *StatusService) OverviewBy(ctx context.Context, opts *ikuai.ListOptions) (*ACStatusOverview, error) {
	var out ACStatusOverview
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(opts, map[string]any{"TYPE": ACStatusTypeAll}), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ControlStatus 获取 AC 智能控制开关状态（0=关闭 1=开启，语义待真机验证）。
// 对应 /Action/call: ac_status.show（TYPE=ac_status）。
func (s *StatusService) ControlStatus(ctx context.Context) (int64, error) {
	var out struct {
		ACStatus int64 `json:"ac_status"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeACStatus}), &out); err != nil {
		return 0, err
	}
	return out.ACStatus, nil
}

// APStatus 获取 AP 在线统计（总数/正常/异常/5G 优先/快速漫游）。
// 对应 /Action/call: ac_status.show（TYPE=ap_status）。
func (s *StatusService) APStatus(ctx context.Context) (*ACAPStatus, error) {
	var out struct {
		APStatus ACAPStatus `json:"ap_status"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAPStatus}), &out); err != nil {
		return nil, err
	}
	return &out.APStatus, nil
}

// ClientStatus 获取无线终端统计（总数/2.4G/5G/最大在线/活跃）。
// 对应 /Action/call: ac_status.show（TYPE=clt_status）。
func (s *StatusService) ClientStatus(ctx context.Context) (*ACCltStatus, error) {
	var out struct {
		CltStatus ACCltStatus `json:"clt_status"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeCltStatus}), &out); err != nil {
		return nil, err
	}
	return &out.CltStatus, nil
}

// APBadStat 获取 AP 异常统计与全网累计上下行。
// 对应 /Action/call: ac_status.show（TYPE=ap_bad_stat）。
func (s *StatusService) APBadStat(ctx context.Context) (*ACBadStatResult, error) {
	var out ACBadStatResult
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAPBadStat}), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// APNetStatus 获取 AP 健康评分（信道/丢包/负载/延时/覆盖）。
// 对应 /Action/call: ac_status.show（TYPE=ap_net_status）。
func (s *StatusService) APNetStatus(ctx context.Context) (*ACNetStatus, error) {
	var out struct {
		APNetStatus ACNetStatus `json:"ap_net_status"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAPNetStatus}), &out); err != nil {
		return nil, err
	}
	return &out.APNetStatus, nil
}

// WPHYUsage 获取全网信道利用率与底噪（无 AP 数据时各值为 -255）。
// 对应 /Action/call: ac_status.show（TYPE=wphy_usage）。
func (s *StatusService) WPHYUsage(ctx context.Context) (*ACWPHYUsage, error) {
	var out struct {
		WPHYUsage ACWPHYUsage `json:"wphy_usage"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeWPHYUsage}), &out); err != nil {
		return nil, err
	}
	return &out.WPHYUsage, nil
}

// APChanInfo 获取 AP 信道信息列表（行结构待真机验证）。
// 对应 /Action/call: ac_status.show（TYPE=ap_chaninfo）。
func (s *StatusService) APChanInfo(ctx context.Context) ([]ACStatRow, error) {
	var out struct {
		APChanInfo []ACStatRow `json:"ap_chaninfo"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAPChanInfo}), &out); err != nil {
		return nil, err
	}
	return out.APChanInfo, nil
}

// APCountInfo 获取 AP 数量分布列表（行结构待真机验证）。
// 对应 /Action/call: ac_status.show（TYPE=ap_countinfo）。
func (s *StatusService) APCountInfo(ctx context.Context) ([]ACStatRow, error) {
	var out struct {
		APCountInfo []ACStatRow `json:"ap_countinfo"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAPCountInfo}), &out); err != nil {
		return nil, err
	}
	return out.APCountInfo, nil
}

// APEvents 获取 AP 事件列表（行结构待真机验证）。
// 对应 /Action/call: ac_status.show（TYPE=ap_event_data）。
func (s *StatusService) APEvents(ctx context.Context) ([]ACStatRow, error) {
	var out struct {
		APEventData []ACStatRow `json:"ap_event_data"`
	}
	if err := s.c.CallData(ctx, "ac_status", "show",
		selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAPEventData}), &out); err != nil {
		return nil, err
	}
	return out.APEventData, nil
}
