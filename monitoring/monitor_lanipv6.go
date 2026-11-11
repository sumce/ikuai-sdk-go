package monitoring

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// monitor_lanipv6.show 的 TYPE 选择器常量（抓包实证）。
const (
	// LanIPv6TypeDataTotal 是控制台列表页抓包使用的 TYPE（data,total）；
	// SDK 默认分页发送 total,data，对固件等价（已真机验证）。
	LanIPv6TypeDataTotal = "data,total"
	// LanIPv6TypeInterface 接口嵌套列表选择器（抓包实证）。
	LanIPv6TypeInterface = "interface"
)

// LanIPv6Client 是一台内网 IPv6 终端的监控行
// （monitor_lanipv6.show TYPE=data,total）。字段集合与 monitor_lanip
// 一致（无 reject 字段），ip_addr 为 IPv6 地址。
type LanIPv6Client struct {
	// ACGid 认证计费用户组 ID（0=未分组）。
	ACGid int64 `json:"ac_gid"`
	// APMAC 所属 AP 的 MAC（有线终端为空）。
	APMAC string `json:"apmac"`
	// APName 所属 AP 名称（有线终端为空）。
	APName string `json:"apname"`
	// AuthType 认证类型（0=无需认证；编码含义待真机验证）。
	AuthType int64 `json:"auth_type"`
	// BSSID 所属 AP 的 BSSID（有线终端为空）。
	BSSID string `json:"bssid"`
	// ClientDevice 终端厂商。
	ClientDevice string `json:"client_device"`
	// ClientModel 终端型号，固件原样返回 URL 编码文本。
	ClientModel string `json:"client_model"`
	// ClientType 终端系统类型（如 Android/Windows/Unknown）。
	ClientType string `json:"client_type"`
	// ClientTypeID 终端类型内部编码。
	ClientTypeID int64 `json:"client_typeid"`
	// Comment 终端备注。
	Comment string `json:"comment"`
	// ConnectNum 当前连接数。
	ConnectNum int64 `json:"connect_num"`
	// Download 当前下载速率（字节/秒）。
	Download int64 `json:"download"`
	// Downrate 下载速率的格式化文本（抓包中为空串；字段语义待真机验证）。
	Downrate string `json:"downrate"`
	// DtalkName 钉钉关联名称（字段语义待真机验证）。
	DtalkName string `json:"dtalk_name"`
	// Frequencies 无线频段（抓包中为空串；字段语义待真机验证）。
	Frequencies string `json:"frequencies"`
	// Hostname 终端主机名。
	Hostname string `json:"hostname"`
	// ID 终端行 ID。
	ID int64 `json:"id"`
	// IPAddr 终端 IPv6 地址。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt 地址排序键（IPv6 场景下的整型排序值；字段语义待真机验证）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// LinkAddr 链路层地址（抓包中为空串；字段语义待真机验证）。
	LinkAddr string `json:"link_addr"`
	// MAC 终端 MAC 地址。
	MAC string `json:"mac"`
	// PPType PPPoE 接入类型（非 PPPoE 终端为空；字段语义待真机验证）。
	PPType string `json:"ppptype"`
	// Signal 无线信号强度（有线终端为空；字段语义待真机验证）。
	Signal string `json:"signal"`
	// SSID 接入的无线 SSID（有线终端为空）。
	SSID string `json:"ssid"`
	// Timestamp 状态采集时间（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
	// TotalUp 累计上传流量（字节）。
	TotalUp int64 `json:"total_up"`
	// TotalDown 累计下载流量（字节）。
	TotalDown int64 `json:"total_down"`
	// Upload 当前上传速率（字节/秒）。
	Upload int64 `json:"upload"`
	// Uprate 上传速率的格式化文本（抓包中为空串；字段语义待真机验证）。
	Uprate string `json:"uprate"`
	// Uptime 上线时间文本（如 2026-11-06 10:51:00）。
	Uptime string `json:"uptime"`
	// Username 认计费账号（未认证为空）。
	Username string `json:"username"`
	// WebID Web 认证会话 ID（0=无；字段语义待真机验证）。
	WebID int64 `json:"webid"`
}

// LanIPv6Service 是终端监控 IPv6（monitor_lanipv6）的子服务。
type LanIPv6Service struct {
	c *ikuai.Client
}

// List 获取内网 IPv6 终端列表（分页）。
// 对应 /Action/call: monitor_lanipv6.show（TYPE=data,total,
// ORDER_BY=ip_addr_int, orderType=IP, limit="offset,rows"），
// 即 Web 控制台「状态监控 → 终端监控」IPv6 页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *LanIPv6Service) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[LanIPv6Client], error) {
	param := pageShowParam(opts, map[string]any{
		"ORDER_BY":  LanIPOrderByIPAddr,
		"orderType": LanIPOrderTypeIP,
	})
	resp, err := s.c.Call(ctx, "monitor_lanipv6", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[LanIPv6Client](resp.Data)
}

// Interfaces 获取 IPv6 终端监控的可选接口列表（嵌套数组：[接口名] 或
// [接口名, 备注]，首项固定为 ["all"]）。
// 对应 /Action/call: monitor_lanipv6.show（TYPE=interface）。
func (s *LanIPv6Service) Interfaces(ctx context.Context) ([][]string, error) {
	var out struct {
		Interface [][]string `json:"interface"`
	}
	if err := s.c.CallData(ctx, "monitor_lanipv6", "show",
		map[string]any{"TYPE": LanIPv6TypeInterface}, &out); err != nil {
		return nil, err
	}
	return out.Interface, nil
}
