package monitoring

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sumce/ikuai-sdk-go"
)

// monitor_lanip.show 的 TYPE 选择器与抓包参数常量。
const (
	// LanIPTypeDataTotal 是控制台列表页抓包使用的 TYPE（data,total）；
	// SDK 默认分页发送 total,data，对固件等价（已真机验证）。
	LanIPTypeDataTotal = "data,total"
	// LanIPTypeInterface 内网接口嵌套列表选择器（抓包实证）。
	LanIPTypeInterface = "interface"

	// LanIPOrderByIPAddr 控制台默认按 IPv4 整型值排序的字段。
	LanIPOrderByIPAddr = "ip_addr_int"
	// LanIPOrderTypeIP 控制台抓包始终携带的 orderType 附加参数值。
	LanIPOrderTypeIP = "IP"
	// LanIPFindsDefault 关键字搜索的默认字段集（抓包实证）。
	LanIPFindsDefault = "ip_addr,mac,comment,username,hostname"
)

// LanIPClient 是一台内网在线终端的监控行（monitor_lanip.show TYPE=data,total）。
type LanIPClient struct {
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
	// ClientDevice 终端厂商（如 vivo、VMware）。
	ClientDevice string `json:"client_device"`
	// ClientModel 终端型号，固件原样返回 URL 编码文本（如 vivo%20T4%205G）。
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
	// IPAddr 终端 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt IPv4 地址的整型表示（排序键）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// LinkAddr 链路层地址（抓包中为空串；字段语义待真机验证）。
	LinkAddr string `json:"link_addr"`
	// MAC 终端 MAC 地址。
	MAC string `json:"mac"`
	// PPType PPPoE 接入类型（非 PPPoE 终端为空；字段语义待真机验证）。
	PPType string `json:"ppptype"`
	// Reject 被拒绝计数（字段语义待真机验证）。
	Reject int64 `json:"reject"`
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

// ClientModelDecoded 返回 URL 解码后的终端型号（固件对型号做了百分号
// 编码，如 "vivo%20T4%205G" → "vivo T4 5G"）。
func (r LanIPClient) ClientModelDecoded() string {
	s, err := url.QueryUnescape(r.ClientModel)
	if err != nil {
		return r.ClientModel
	}
	return s
}

// LanIPService 是终端监控 IPv4（monitor_lanip）的子服务。
type LanIPService struct {
	c *ikuai.Client
}

// List 获取内网在线终端列表（分页）。
// 对应 /Action/call: monitor_lanip.show（TYPE=data,total, ORDER_BY=ip_addr_int,
// orderType=IP, limit="offset,rows"；控制台另用 FINDS/KEYWORDS 搜索），
// 即 Web 控制台「状态监控 → 终端监控」IPv4 页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()；关键字搜索请用 Search。
func (s *LanIPService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[LanIPClient], error) {
	param := pageShowParam(opts, map[string]any{
		"ORDER_BY":  LanIPOrderByIPAddr,
		"orderType": LanIPOrderTypeIP,
	})
	resp, err := s.c.Call(ctx, "monitor_lanip", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[LanIPClient](resp.Data)
}

// Search 按关键字搜索内网在线终端（KEYWORDS 匹配 FINDS 字段集）。
// 对应 /Action/call: monitor_lanip.show（FINDS=ip_addr,mac,comment,username,hostname,
// KEYWORDS=keywords）。keywords 为空返回 ikuai.ErrInvalidParam。
// 默认字段集可用 opts.Extra["FINDS"] 覆盖。
func (s *LanIPService) Search(ctx context.Context, keywords string, opts *ikuai.ListOptions) (*ikuai.Page[LanIPClient], error) {
	if keywords == "" {
		return nil, fmt.Errorf("monitoring: keywords 不能为空: %w", ikuai.ErrInvalidParam)
	}
	param := pageShowParam(opts, map[string]any{
		"ORDER_BY":  LanIPOrderByIPAddr,
		"orderType": LanIPOrderTypeIP,
		"FINDS":     LanIPFindsDefault,
	})
	param["KEYWORDS"] = keywords
	resp, err := s.c.Call(ctx, "monitor_lanip", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[LanIPClient](resp.Data)
}

// Interfaces 获取终端监控的可选接口列表（嵌套数组：[接口名] 或
// [接口名, 备注]，首项固定为 ["all"]）。
// 对应 /Action/call: monitor_lanip.show（TYPE=interface）。
func (s *LanIPService) Interfaces(ctx context.Context) ([][]string, error) {
	var out struct {
		Interface [][]string `json:"interface"`
	}
	if err := s.c.CallData(ctx, "monitor_lanip", "show",
		map[string]any{"TYPE": LanIPTypeInterface}, &out); err != nil {
		return nil, err
	}
	return out.Interface, nil
}
