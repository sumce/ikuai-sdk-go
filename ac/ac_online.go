package ac

import (
	"context"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// ac_online_clt.show 的参数常量。
const (
	// OnlineClientsOrderByIPAddr 控制台按 IPv4 整型值排序的默认字段
	// （真机 ORDER_BY 探测实证为合法列）。
	OnlineClientsOrderByIPAddr = "ip_addr_int"
	// OnlineClientsFindsDefault 关键字搜索的默认字段集（真机 FINDS 探测
	// 实证全部为合法列；依据 i18n searchTip「MAC/SSID/备注」与
	// search_holder「选择或输入AP MAC」）。
	OnlineClientsFindsDefault = "mac,ip_addr,ssid,comment,apmac"
)

// OnlineClient 是一台在线无线终端（ac_online_clt.show data 数组元素）。
//
// 行字段经真机 ORDER_BY/FINDS 列名探测实证（非法列名固件回
// "no such column"），为真机可验证的全部 21 个列；字段中文语义按固件
// i18n online_user 区块对齐。timestamp/last_time/uprate/downrate/
// frequencies/manufacturer 的精确语义待真机验证（有数据后比对）。
type OnlineClient struct {
	// ID 终端行 ID。
	ID int64 `json:"id"`
	// MAC 终端 MAC 地址（i18n online_user.mac「终端 MAC」）。
	MAC string `json:"mac"`
	// IPAddr 终端 IPv4 地址（i18n online_user.ip_addr「终端 IP」）。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt IPv4 地址的整型表示（排序键）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// APMAC 所属 AP 的 MAC（i18n online_user.apmac「AP信息」）。
	APMAC string `json:"apmac"`
	// APName 所属 AP 名称。
	APName string `json:"apname"`
	// BSSID 所属 AP 的 BSSID。
	BSSID string `json:"bssid"`
	// SSID 接入的无线 SSID（i18n online_user.ssid）。
	SSID string `json:"ssid"`
	// Signal 终端信号强度（i18n online_user.signal「终端信号强度」，dBm）。
	Signal string `json:"signal"`
	// Frequencies 无线频段（2.4G/5G，格式待真机验证）。
	Frequencies string `json:"frequencies"`
	// Upload 上行速率（字节/秒，i18n online_user.upload「上行速率」）。
	Upload int64 `json:"upload"`
	// Download 下行速率（字节/秒，i18n online_user.download「下行速率」）。
	Download int64 `json:"download"`
	// Uprate 上行连接速率（i18n Upstream_rate「上行连接速率」，如链路速率文本，待真机验证）。
	Uprate string `json:"uprate"`
	// Downrate 下行连接速率（i18n Downlink_rate「下行连接速率」，待真机验证）。
	Downrate string `json:"downrate"`
	// TotalUp 累计上行流量（字节，i18n online_user.total_up「累计上行」）。
	TotalUp int64 `json:"total_up"`
	// TotalDown 累计下行流量（字节，i18n online_user.total_down「累计下行」）。
	TotalDown int64 `json:"total_down"`
	// ConnectTime 连接时长（i18n online_user.connect_time；单位待真机验证）。
	ConnectTime int64 `json:"connect_time"`
	// Timestamp 状态采集时间（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
	// LastTime 最近上线时间（Unix 时间戳，秒；语义待真机验证）。
	LastTime int64 `json:"last_time"`
	// Manufacturer 终端厂商/品牌（i18n terminal_brand「终端品牌」）。
	Manufacturer string `json:"manufacturer"`
	// Comment 终端备注（i18n online_user.comment）。
	Comment string `json:"comment"`
}

// OnlineClientsService 是无线终端列表（ac_online_clt）的子服务。
type OnlineClientsService struct {
	c *ikuai.Client
}

// List 获取在线无线终端列表（分页，支持 opts 的分页/排序/过滤）。
// 对应 /Action/call: ac_online_clt.show（TYPE=total,data, limit="offset,rows"，
// ORDER_BY 真机探测实证合法），
// 即 Web 控制台「AC管理 → 无线终端列表」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions() 并按 ip_addr_int 排序。
func (s *OnlineClientsService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[OnlineClient], error) {
	param := pageShowParam(opts, map[string]any{"ORDER_BY": OnlineClientsOrderByIPAddr})
	resp, err := s.c.Call(ctx, "ac_online_clt", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[OnlineClient](resp.Data)
}

// Search 按关键字搜索在线无线终端（KEYWORDS 匹配 FINDS 字段集，覆盖
// 控制台「选择或输入AP MAC」的按 AP 过滤场景）。
// 对应 /Action/call: ac_online_clt.show（FINDS=mac,ip_addr,ssid,comment,apmac）。
// keywords 为空返回 ikuai.ErrInvalidParam；默认字段集可用
// opts.Extra["FINDS"] 覆盖（字段集经真机 FINDS 探测实证合法）。
func (s *OnlineClientsService) Search(ctx context.Context, keywords string, opts *ikuai.ListOptions) (*ikuai.Page[OnlineClient], error) {
	if keywords == "" {
		return nil, fmt.Errorf("ac: keywords 不能为空: %w", ikuai.ErrInvalidParam)
	}
	param := pageShowParam(opts, map[string]any{
		"ORDER_BY": OnlineClientsOrderByIPAddr,
		"FINDS":    OnlineClientsFindsDefault,
	})
	param["KEYWORDS"] = keywords
	resp, err := s.c.Call(ctx, "ac_online_clt", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[OnlineClient](resp.Data)
}
