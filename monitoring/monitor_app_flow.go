package monitoring

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// monitor_app_flow.show 的 TYPE 选择器常量（抓包实证）。
const (
	// AppFlowTypeHistory 按时点的协议分类流量历史（app_history）。
	AppFlowTypeHistory = "app_history"
	// AppFlowTypeInterface 可选接口嵌套列表（interface）。
	AppFlowTypeInterface = "interface"
	// AppFlowTypeCleanTime 协议流量统计清零时间戳（app_history_cleantime）。
	AppFlowTypeCleanTime = "app_history_cleantime"
	// AppFlowTypeFlow 协议流量排行（flow）。
	AppFlowTypeFlow = "flow"
)

// monitor_app_flow.show 的 datetype / time_range 参数常量。
const (
	// AppFlowDatetypeDay 按天聚合（抓包实证）。
	AppFlowDatetypeDay = "day"
	// AppFlowDatetypeHour 按小时聚合（真机验证可用）。
	AppFlowDatetypeHour = "hour"
	// AppFlowTimeRangeAll 全部时间范围（抓包实证）。
	AppFlowTimeRangeAll = "all"
	// AppFlowTimeRangeDay 最近一天（真机验证可用）。
	AppFlowTimeRangeDay = "day"
)

// AppFlowInterfaceAll 接口选择器的「全部接口」值。
const AppFlowInterfaceAll = "all"

// AppFlowHistory 是协议分类流量历史的一个时点行（app_history），
// 各分类字段值为该时点的流量（字节）。
type AppFlowHistory struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
	// Total 全部协议流量合计。
	Total int64 `json:"Total"`
	// TotalUp 全部协议上传合计。
	TotalUp int64 `json:"Total_up"`
	// TotalDown 全部协议下载合计。
	TotalDown int64 `json:"Total_down"`
	// HTTPUp/HTTPDown HTTP 应用上传/下载流量。
	HTTPUp   int64 `json:"HTTP_up"`
	HTTPDown int64 `json:"HTTP_down"`
	// DownloadUp/DownloadDown 下载类应用上传/下载流量。
	DownloadUp   int64 `json:"Download_up"`
	DownloadDown int64 `json:"Download_down"`
	// TransportUp/TransportDown 传输类应用上传/下载流量。
	TransportUp   int64 `json:"Transport_up"`
	TransportDown int64 `json:"Transport_down"`
	// IMUp/IMDown 即时通讯类应用上传/下载流量。
	IMUp   int64 `json:"IM_up"`
	IMDown int64 `json:"IM_down"`
	// VideoUp/VideoDown 视频类应用上传/下载流量。
	VideoUp   int64 `json:"Video_up"`
	VideoDown int64 `json:"Video_down"`
	// CommonUp/CommonDown 常用应用上传/下载流量。
	CommonUp   int64 `json:"Common_up"`
	CommonDown int64 `json:"Common_down"`
	// GameUp/GameDown 游戏类应用上传/下载流量。
	GameUp   int64 `json:"Game_up"`
	GameDown int64 `json:"Game_down"`
	// OthersUp/OthersDown 其他应用上传/下载流量。
	OthersUp   int64 `json:"Others_up"`
	OthersDown int64 `json:"Others_down"`
	// TestUp/TestDown 测试类应用上传/下载流量。
	TestUp   int64 `json:"Test_up"`
	TestDown int64 `json:"Test_down"`
	// UnknownUp/UnknownDown 未识别应用上传/下载流量。
	UnknownUp   int64 `json:"Unknown_up"`
	UnknownDown int64 `json:"Unknown_down"`
}

// AppFlowProto 是协议流量排行的一行（flow）。
type AppFlowProto struct {
	// ID 协议排行 ID。
	ID int64 `json:"id"`
	// AppName 应用/协议名（如 Unknown）。
	AppName string `json:"app_name"`
	// ConnectNum 该协议当前连接数。
	ConnectNum int64 `json:"connect_num"`
	// Upload 当前上传速率（字节/秒）。
	Upload int64 `json:"upload"`
	// Download 当前下载速率（字节/秒）。
	Download int64 `json:"download"`
	// Total 当前上传+下载速率合计（字节/秒）。
	Total int64 `json:"total"`
	// TotalUp 累计上传流量（字节）。
	TotalUp int64 `json:"total_up"`
	// TotalDown 累计下载流量（字节）。
	TotalDown int64 `json:"total_down"`
}

// AppFlowInterfaces 是 monitor_app_flow.show
// TYPE=interface,app_history_cleantime 的结果。
type AppFlowInterfaces struct {
	// Interface 可选接口嵌套列表（[接口名, 备注]，首项为 ["all", 接口总览?]；
	// 固件无总览项时首项为 ["all"]）。
	Interface [][]string `json:"interface"`
	// CleanTime 协议流量历史清零时间戳（0=未清零；字段语义待真机验证）。
	CleanTime int64 `json:"app_history_cleantime"`
}

// AppFlowService 是协议监控（monitor_app_flow）的子服务。
type AppFlowService struct {
	c *ikuai.Client
}

// History 获取协议分类流量历史（按接口/聚合粒度/时间范围）。
// 对应 /Action/call: monitor_app_flow.show（TYPE=app_history, interface=all,
// datetype=day, time_range=all），
// 即 Web 控制台「状态监控 → 协议监控」页面的历史曲线数据。
// opts 为 nil 时使用抓包默认值；interface/datetype/time_range 可通过
// opts.Extra 覆盖（值常量见 AppFlowDatetype* / AppFlowTimeRange* /
// AppFlowInterfaceAll）。
func (s *AppFlowService) History(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[AppFlowHistory], error) {
	param := selectorShowParam(opts, map[string]any{
		"TYPE":       AppFlowTypeHistory,
		"interface":  AppFlowInterfaceAll,
		"datetype":   AppFlowDatetypeDay,
		"time_range": AppFlowTimeRangeAll,
	})
	resp, err := s.c.Call(ctx, "monitor_app_flow", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[AppFlowHistory](resp.Data)
}

// Flow 获取协议流量排行（Top 协议的上传/下载/连接数）。
// 对应 /Action/call: monitor_app_flow.show（TYPE=flow）。
func (s *AppFlowService) Flow(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[AppFlowProto], error) {
	param := selectorShowParam(opts, map[string]any{"TYPE": AppFlowTypeFlow})
	resp, err := s.c.Call(ctx, "monitor_app_flow", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[AppFlowProto](resp.Data)
}

// Interfaces 获取协议监控的可选接口列表与统计清零时间戳。
// 对应 /Action/call: monitor_app_flow.show（TYPE=interface,app_history_cleantime）。
func (s *AppFlowService) Interfaces(ctx context.Context) (*AppFlowInterfaces, error) {
	var out AppFlowInterfaces
	if err := s.c.CallData(ctx, "monitor_app_flow", "show",
		map[string]any{
			"TYPE": AppFlowTypeInterface + "," + AppFlowTypeCleanTime,
		}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
