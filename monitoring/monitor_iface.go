package monitoring

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// monitor_iface.show 的 TYPE 选择器常量（抓包实证，可任意组合）。
const (
	// IfaceTypeCheck 线路检测状态区块（iface_check）。
	IfaceTypeCheck = "iface_check"
	// IfaceTypeStream 接口实时流量区块（iface_stream）。
	IfaceTypeStream = "iface_stream"
	// IfaceTypeEtherInfo 网口物理信息区块（ether_info）。
	IfaceTypeEtherInfo = "ether_info"
	// IfaceTypeSnapshoot 内外网口配置快照区块（snapshoot）。
	IfaceTypeSnapshoot = "snapshoot"
)

// IfaceTypeAll 是 Web 控制台「线路监控」页面默认请求的全量 TYPE 组合
// （iface_check,iface_stream,ether_info,snapshoot，抓包实证）。
const IfaceTypeAll = IfaceTypeCheck + "," + IfaceTypeStream + "," +
	IfaceTypeEtherInfo + "," + IfaceTypeSnapshoot

// IfaceMonitor 是 monitor_iface.show 的聚合结果，各区块仅在对应 TYPE
// 选择器被请求时出现（未请求的区块为零值 nil）。
type IfaceMonitor struct {
	// IfaceCheck 外网线路检测状态行（TYPE 含 iface_check 时返回）。
	IfaceCheck []IfaceCheck `json:"iface_check"`
	// IfaceStream 全部接口的实时流量行（TYPE 含 iface_stream 时返回）。
	IfaceStream []IfaceStream `json:"iface_stream"`
	// EtherInfo 网口名 → 物理信息（TYPE 含 ether_info 时返回）。
	EtherInfo map[string]EtherInfo `json:"ether_info"`
	// SnapshootLAN 内网口配置快照行（TYPE 含 snapshoot 时返回）。
	SnapshootLAN []IfaceSnapshootLAN `json:"snapshoot_lan"`
	// SnapshootWAN 外网口配置快照行（TYPE 含 snapshoot 时返回）。
	SnapshootWAN []IfaceSnapshootWAN `json:"snapshoot_wan"`
}

// IfaceCheck 是一条外网线路的检测状态行（iface_check）。
type IfaceCheck struct {
	// ID 线路 ID。
	ID int64 `json:"id"`
	// Interface 接口名（如 adsl01）。
	Interface string `json:"interface"`
	// ParentInterface 所属物理口（如 wan1）。
	ParentInterface string `json:"parent_interface"`
	// IPAddr 线路当前 IP。
	IPAddr string `json:"ip_addr"`
	// Gateway 线路网关。
	Gateway string `json:"gateway"`
	// Internet 接入方式（如 PPPOE / STATIC / DHCP）。
	Internet string `json:"internet"`
	// Updatetime 检测更新时间，固件以字符串形式返回的 Unix 时间戳（秒）。
	Updatetime string `json:"updatetime"`
	// AutoSwitch 切换模式描述（如「自动」）。
	AutoSwitch string `json:"auto_switch"`
	// Result 检测结果（如 success）。
	Result string `json:"result"`
	// ErrMsg 检测结果描述（如「线路检测成功」）。
	ErrMsg string `json:"errmsg"`
	// Comment 线路备注。
	Comment string `json:"comment"`
}

// IfaceStream 是一个接口的实时流量行（iface_stream）。
type IfaceStream struct {
	// Interface 接口名。
	Interface string `json:"interface"`
	// Comment 接口备注。
	Comment string `json:"comment"`
	// IPAddr 接口 IP。
	IPAddr string `json:"ip_addr"`
	// ConnectNum 连接数；固件对无连接接口返回 "--"（字符串占位）。
	ConnectNum string `json:"connect_num"`
	// Upload 当前上传速率（字节/秒）。
	Upload int64 `json:"upload"`
	// Download 当前下载速率（字节/秒）。
	Download int64 `json:"download"`
	// TotalUp 累计上传流量（字节）。
	TotalUp int64 `json:"total_up"`
	// TotalDown 累计下载流量（字节）。
	TotalDown int64 `json:"total_down"`
	// Updropped 累计发送丢包数。
	Updropped int64 `json:"updropped"`
	// Downdropped 累计接收丢包数。
	Downdropped int64 `json:"downdropped"`
	// Uppacked 累计发送包数。
	Uppacked int64 `json:"uppacked"`
	// Downpacked 累计接收包数。
	Downpacked int64 `json:"downpacked"`
}

// EtherInfo 是一个物理网口的链路信息（ether_info，键为网口名 eth0/vnet/usb0 等）。
type EtherInfo struct {
	// Driver 网卡驱动名（如 ice/e1000e）。
	Driver string `json:"driver"`
	// Type 网口类型（如 FIBRE/1G）。
	Type string `json:"type"`
	// MAC 网口 MAC 地址。
	MAC string `json:"mac"`
	// Link 链路状态（1=UP）。
	Link int64 `json:"link"`
	// Speed 协商速率（Mbps）。
	Speed int64 `json:"speed"`
	// Duplex 双工模式（1=全双工）。
	Duplex int64 `json:"duplex"`
	// Model 网卡型号描述。
	Model string `json:"model"`
	// Interface 绑定的接口名（可为空）。
	Interface string `json:"interface"`
	// Lock 端口锁定状态（0=未锁定；字段语义待真机验证）。
	Lock int64 `json:"lock"`
}

// IfaceSnapshootLAN 是一个内网口的配置快照行（snapshoot_lan）。
type IfaceSnapshootLAN struct {
	// ID 内网口 ID。
	ID int64 `json:"id"`
	// Comment 备注。
	Comment string `json:"comment"`
	// Interface 接口名（如 lan1）。
	Interface string `json:"interface"`
	// Bandmode 带宽模式（0=不限制；字段语义待真机验证）。
	Bandmode int64 `json:"bandmode"`
	// Linkmode 链路模式（编码含义待真机验证）。
	Linkmode int64 `json:"linkmode"`
	// MAC 接口 MAC 地址。
	MAC string `json:"mac"`
	// Member 绑定的物理网口名列表。
	Member []string `json:"member"`
	// IPAddr 接口 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
}

// IfaceSnapshootWAN 是一个外网口的配置快照行（snapshoot_wan）。
type IfaceSnapshootWAN struct {
	// ID 外网口 ID。
	ID int64 `json:"id"`
	// Comment 备注。
	Comment string `json:"comment"`
	// Interface 接口名（如 wan1）。
	Interface string `json:"interface"`
	// MAC 接口 MAC 地址。
	MAC string `json:"mac"`
	// Member 绑定的物理网口名列表。
	Member []string `json:"member"`
	// Bandmode 带宽模式（0=不限制；字段语义待真机验证）。
	Bandmode int64 `json:"bandmode"`
	// DefaultRoute 默认路由标志（0=否；字段语义待真机验证）。
	DefaultRoute int64 `json:"default_route"`
	// Internet 接入方式编码（4=PPPoE；字段语义待真机验证）。
	Internet int64 `json:"internet"`
	// IPAddr 接口 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
	// Gateway 网关。
	Gateway string `json:"gateway"`
	// DNS1 主 DNS。
	DNS1 string `json:"dns1"`
	// DNS2 备 DNS。
	DNS2 string `json:"dns2"`
	// CountStatic 静态接入切换计数。
	CountStatic int64 `json:"count_static"`
	// CountDHCP DHCP 接入切换计数。
	CountDHCP int64 `json:"count_dhcp"`
	// CountPPPoE PPPoE 接入切换计数。
	CountPPPoE int64 `json:"count_pppoe"`
	// CountCheckFail 线路检测失败计数。
	CountCheckFail int64 `json:"count_check_fail"`
	// Updatetime 检测更新时间（Unix 时间戳，秒）。
	Updatetime int64 `json:"updatetime"`
	// CheckRes 检测结果编码（字段语义待真机验证）。
	CheckRes int64 `json:"check_res"`
	// ErrMsg 检测结果描述。
	ErrMsg string `json:"errmsg"`
	// Power 4G 模块信号强度（无 4G 模块时为空；字段语义待真机验证）。
	Power string `json:"power"`
	// ISP 运营商（无 4G 模块时为空；字段语义待真机验证）。
	ISP string `json:"isp"`
	// IMEI 4G 模块 IMEI（字段语义待真机验证）。
	IMEI string `json:"imei"`
	// QNW 4G 网络制式（字段语义待真机验证）。
	QNW string `json:"qnw"`
	// CCID 4G SIM 卡 CCID（字段语义待真机验证）。
	CCID string `json:"ccid"`
	// Isnr 4G 模块 SNR（字段语义待真机验证）。
	Isnr string `json:"isnr"`
	// PCID 4G 模块 PCI（字段语义待真机验证）。
	PCID string `json:"pcid"`
}

// InterfaceService 是线路监控（monitor_iface）的子服务。
type InterfaceService struct {
	c *ikuai.Client
}

// Monitor 获取线路监控数据（线路检测状态、接口实时流量、网口物理信息、
// 内外网口快照，按 TYPE 选择器返回对应区块）。
// 对应 /Action/call: monitor_iface.show（TYPE=iface_check,iface_stream[,ether_info][,snapshoot]），
// 即 Web 控制台「状态监控 → 线路监控」页面。
// opts 为 nil 时请求全量 TYPE；仅需要部分区块时通过
// opts.Type = IfaceTypeCheck+","+IfaceTypeStream 等组合传入。
func (s *InterfaceService) Monitor(ctx context.Context, opts *ikuai.ListOptions) (*IfaceMonitor, error) {
	param := selectorShowParam(opts, map[string]any{"TYPE": IfaceTypeAll})
	var out IfaceMonitor
	if err := s.c.CallData(ctx, "monitor_iface", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
