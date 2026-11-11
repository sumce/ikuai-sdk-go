package overview

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sumce/ikuai-sdk-go"
)

// homepage.show 的 TYPE 选择器常量（抓包实证）。
const (
	// HomepageTypeSysStat 系统概况数据块：CPU/内存/在线用户/流量/版本（sysstat）。
	HomepageTypeSysStat = "sysstat"
	// HomepageTypeACStatus AP 统计数据块（ac_status）。
	HomepageTypeACStatus = "ac_status"
	// HomepageTypeDHCPAddrpoolNum DHCP 可分配地址数（dhcp_addrpool_num）。
	HomepageTypeDHCPAddrpoolNum = "dhcp_addrpool_num"
	// HomepageTypeEtherInfo 物理网口信息（ether_info）。
	HomepageTypeEtherInfo = "ether_info"
	// HomepageTypeSnapshoot 内外网接口快照（snapshoot，固件原始拼写），
	// 对应 Data 中的 snapshoot_lan / snapshoot_wan 两个数组。
	HomepageTypeSnapshoot = "snapshoot"
	// HomepageTypeAppFlow 应用流量分类快照（app_flow）。
	HomepageTypeAppFlow = "app_flow"
	// HomepageTypeAppHistorical 应用流量历史区间排行（app_historical）。
	HomepageTypeAppHistorical = "app_historical"
	// HomepageTypeMonitorSystem 首页带宽总量与每日统计（monitor_system）。
	HomepageTypeMonitorSystem = "monitor_system"
	// HomepageTypeAntipiracy 违禁监控状态（antipiracy）。
	HomepageTypeAntipiracy = "antipiracy"
)

// homepage.show 控制台实际使用的 TYPE 组合（抓包实证）。
const (
	// HomepageTypeSystemOverview 是「系统概况」面板的组合
	// （sysstat,ac_status），Service.SystemOverview 使用。
	HomepageTypeSystemOverview = HomepageTypeSysStat + "," + HomepageTypeACStatus
	// HomepageTypeEtherInfoSnapshoot 是「接口快照」面板的组合
	// （ether_info,snapshoot）。
	HomepageTypeEtherInfoSnapshoot = HomepageTypeEtherInfo + "," + HomepageTypeSnapshoot
	// HomepageTypeAppFlowHistorical 是「应用流量」面板的组合
	// （app_flow,app_historical）。
	HomepageTypeAppFlowHistorical = HomepageTypeAppFlow + "," + HomepageTypeAppHistorical
)

// SystemOverview 是系统概况仪表盘的聚合结果（homepage.show
// TYPE=sysstat,ac_status），即 Web 控制台首页顶部的总览面板。
// 两个区块仅在对应 TYPE 被请求时出现（未请求时为零值 nil）。
type SystemOverview struct {
	// SysStat 系统概况区块（TYPE 含 sysstat）。
	SysStat *SysStat `json:"sysstat"`
	// ACStatus AP 统计区块（TYPE 含 ac_status；未接入 AC 时计数值为 0）。
	ACStatus *ACStatus `json:"ac_status"`
}

// SysStat 是系统概况数据块（homepage.show TYPE=sysstat）。
type SysStat struct {
	// CPU CPU 占用率采样列表（百分比字符串，如 "6.98%"；观测 21 个采样）。
	CPU []string `json:"cpu"`
	// CPUTemp CPU 温度采样列表（摄氏度；观测单元素，如 [39]）。
	CPUTemp []int64 `json:"cputemp"`
	// Freq CPU 频率采样列表（字符串数值，观测 20 个、值域约 1000~1900，
	// 单位待真机验证）。
	Freq []string `json:"freq"`
	// GWID 网关设备唯一 ID（32 位十六进制，与 register.show 的 gwid 一致）。
	GWID string `json:"gwid"`
	// Hostname 主机名。
	Hostname string `json:"hostname"`
	// LinkStatus 链路状态标志（观测值 0；字段语义待真机验证）。
	LinkStatus int64 `json:"link_status"`
	// Memory 内存信息（含使用率字符串）。
	Memory MemoryInfo `json:"memory"`
	// OnlineUser 在线用户/终端统计。
	OnlineUser OnlineUserInfo `json:"online_user"`
	// Stream 连接数与实时/累计流量。
	Stream StreamInfo `json:"stream"`
	// Uptime 运行时长（秒；观测值 1629218 ≈ 18.9 天）。
	Uptime int64 `json:"uptime"`
	// VerInfo 固件版本信息。
	VerInfo VerInfo `json:"verinfo"`
}

// UptimeDuration 将 Uptime（秒）换算为 time.Duration；区块缺失时返回 0。
func (o *SystemOverview) UptimeDuration() time.Duration {
	if o == nil || o.SysStat == nil {
		return 0
	}
	return time.Duration(o.SysStat.Uptime) * time.Second
}

// CPUAverage 计算 CPU 占用率采样的平均值（百分比数值，如 9.49 表示
// 9.49%）。CPU 列表缺失/为空时返回 0；单个采样解析失败时报错。
func (o *SystemOverview) CPUAverage() (float64, error) {
	if o == nil || o.SysStat == nil || len(o.SysStat.CPU) == 0 {
		return 0, nil
	}
	total := 0.0
	for i, s := range o.SysStat.CPU {
		v, err := parsePercent(s)
		if err != nil {
			return 0, fmt.Errorf("overview: cpu[%d]: %w", i, err)
		}
		total += v
	}
	return total / float64(len(o.SysStat.CPU)), nil
}

// ACStatus 是 AP 统计数据块（homepage.show TYPE=ac_status）。
type ACStatus struct {
	// APCount AP 总数。
	APCount int64 `json:"ap_count"`
	// APOnline 在线 AP 数。
	APOnline int64 `json:"ap_online"`
}

// MemoryInfo 是内存信息子结构（sysstat.memory）。
// 各容量字段观测口径与 Linux /proc/meminfo 的 kB 一致（如
// total=47914500 ≈ 45.7 GiB），单位待真机验证。
type MemoryInfo struct {
	// Total 内存总量。
	Total int64 `json:"total"`
	// Available 可用内存（估算值）。
	Available int64 `json:"available"`
	// Free 空闲内存。
	Free int64 `json:"free"`
	// Cached 页缓存占用。
	Cached int64 `json:"cached"`
	// Buffers 块缓冲占用。
	Buffers int64 `json:"buffers"`
	// Used 内存使用率（百分比字符串，如 "6%"）。
	Used string `json:"used"`
}

// UsedPercent 解析使用率字符串（"6%" → 6.0）。空值返回 0。
func (m MemoryInfo) UsedPercent() (float64, error) {
	return parsePercent(m.Used)
}

// OnlineUserInfo 是在线用户统计子结构（sysstat.online_user）。
type OnlineUserInfo struct {
	// Count 在线用户总数。
	Count int64 `json:"count"`
	// Count2G 2.4G 无线在线用户数。
	Count2G int64 `json:"count_2g"`
	// Count5G 5G 无线在线用户数。
	Count5G int64 `json:"count_5g"`
	// CountWired 有线在线用户数。
	CountWired int64 `json:"count_wired"`
	// CountWireless 无线在线用户数（Count2G + Count5G）。
	CountWireless int64 `json:"count_wireless"`
}

// StreamInfo 是流量/连接子结构（sysstat.stream）。
// Upload/Download 观测值量级为每秒数十 MB，按实时速率（字节/秒）理解；
// TotalUp/TotalDown 为开机以来的累计流量（字节）。口径待真机验证。
type StreamInfo struct {
	// ConnectNum 当前连接数。
	ConnectNum int64 `json:"connect_num"`
	// Upload 实时上传速率（字节/秒，待真机验证）。
	Upload int64 `json:"upload"`
	// Download 实时下载速率（字节/秒，待真机验证）。
	Download int64 `json:"download"`
	// TotalUp 累计上传流量（字节）。
	TotalUp int64 `json:"total_up"`
	// TotalDown 累计下载流量（字节）。
	TotalDown int64 `json:"total_down"`
}

// VerInfo 是固件版本信息子结构（sysstat.verinfo）。
type VerInfo struct {
	// Modelname 设备型号名（观测值为空字符串）。
	Modelname string `json:"modelname"`
	// Verstring 版本全串（如 "3.7.24 x64 Build202608141210"）。
	Verstring string `json:"verstring"`
	// Version 版本号（如 "3.7.24"）。
	Version string `json:"version"`
	// BuildDate 固件构建时间（数值 YYYYMMDDHHMM，如 202608141210）。
	BuildDate int64 `json:"build_date"`
	// Arch CPU 架构（观测值 "x86"）。
	Arch string `json:"arch"`
	// Sysbit 系统位数（观测值 "x64"）。
	Sysbit string `json:"sysbit"`
	// Verflags 版本标志（观测值为空字符串；字段语义待真机验证）。
	Verflags string `json:"verflags"`
	// IsEnterprise 企业版标志（1=企业版，0=非企业版）。
	IsEnterprise int64 `json:"is_enterprise"`
	// SupportI18n 是否支持多语言（1=支持）。
	SupportI18n int64 `json:"support_i18n"`
	// SupportDingTalk 是否支持钉钉绑定（1=支持）。
	SupportDingTalk int64 `json:"support_dingtalk"`
	// SupportLCD 是否支持 LCD 屏（1=支持）。
	SupportLCD int64 `json:"support_lcd"`
	// BootGuide 启动介质（观测值 "hd"）。
	BootGuide string `json:"bootguide"`
}

// BuildTime 解析固件构建时间数值 BuildDate（YYYYMMDDHHMM）为本地时间；
// BuildDate 为 0 时返回零值 time.Time。
func (v VerInfo) BuildTime() (time.Time, error) {
	if v.BuildDate == 0 {
		return time.Time{}, nil
	}
	return time.ParseInLocation("200601021504",
		strconv.FormatInt(v.BuildDate, 10), time.Local)
}

// InterfaceSnapshot 是接口快照面板的聚合结果（homepage.show
// TYPE=ether_info,snapshoot）。三个区块仅在对应 TYPE 被请求时出现。
type InterfaceSnapshot struct {
	// EtherInfo 物理网口信息，键为网口名（观测：eth0~eth11/usb0/vnet）。
	EtherInfo map[string]EtherInfo `json:"ether_info"`
	// SnapshootLAN 内网接口快照行列表（TYPE 含 snapshoot）。
	SnapshootLAN []SnapshootLAN `json:"snapshoot_lan"`
	// SnapshootWAN 外网接口快照行列表（TYPE 含 snapshoot）。
	SnapshootWAN []SnapshootWAN `json:"snapshoot_wan"`
}

// EtherInfo 是一个物理网口的信息行（ether_info 的值对象）。
type EtherInfo struct {
	// Driver 网卡驱动名（观测：ice/igb/mlx4_core/r8125/kvm/usb）。
	Driver string `json:"driver"`
	// Duplex 双工模式标志（观测值 0/1；字段语义待真机验证）。
	Duplex int64 `json:"duplex"`
	// Interface 绑定的 WAN/LAN 接口名（未绑定为 "Null"）。
	Interface string `json:"interface"`
	// Link 物理连接状态（1=已连接，0=未连接）。
	Link int64 `json:"link"`
	// Lock 锁定标志（观测值 0；字段语义待真机验证）。
	Lock int64 `json:"lock"`
	// MAC 网卡 MAC 地址。
	MAC string `json:"mac"`
	// Model 网卡型号描述（PCI ID 数据库文本）。
	Model string `json:"model"`
	// Speed 协商速率（Mbps，未连接为 0）。
	Speed int64 `json:"speed"`
	// Type 端口类型（观测：FIBRE/TP/USB）。
	Type string `json:"type"`
}

// SnapshootLAN 是一个内网接口的快照行（snapshoot_lan 数组元素）。
type SnapshootLAN struct {
	// BandMode 工作模式标志（观测值 0；字段语义待真机验证）。
	BandMode int64 `json:"bandmode"`
	// Comment 备注。
	Comment string `json:"comment"`
	// ID 接口行 ID。
	ID int64 `json:"id"`
	// Interface 内网接口名（lan1、lan2...）。
	Interface string `json:"interface"`
	// IPAddr 接口 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// LinkMode 接口模式（观测值 4；字段语义待真机验证）。
	LinkMode int64 `json:"linkmode"`
	// MAC 接口 MAC 地址。
	MAC string `json:"mac"`
	// Member 桥接/绑定的物理网口名列表。
	Member []string `json:"member"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
}

// SnapshootWAN 是一个外网接口的快照行（snapshoot_wan 数组元素）。
// 拨号计数与检测类字段仅在对应接入方式下有非零值；其余字段按命名
// 推断，语义待真机验证。
type SnapshootWAN struct {
	// BandMode 工作模式标志（观测值 0；字段语义待真机验证）。
	BandMode int64 `json:"bandmode"`
	// CCID 4G 卡号（观测值为空）。
	CCID string `json:"ccid"`
	// CheckRes 外网连通检测结果（观测值 0；字段语义待真机验证）。
	CheckRes int64 `json:"check_res"`
	// Comment 备注。
	Comment string `json:"comment"`
	// CountCheckFail 外网检测失败次数。
	CountCheckFail int64 `json:"count_check_fail"`
	// CountDHCP DHCP 接入拨号计数。
	CountDHCP int64 `json:"count_dhcp"`
	// CountPPPoE PPPoE 接入拨号计数（观测值 13）。
	CountPPPoE int64 `json:"count_pppoe"`
	// CountStatic 静态 IP 接入计数。
	CountStatic int64 `json:"count_static"`
	// DefaultRoute 默认路由标志（字段语义待真机验证）。
	DefaultRoute int64 `json:"default_route"`
	// DNS1 首选 DNS。
	DNS1 string `json:"dns1"`
	// DNS2 备用 DNS。
	DNS2 string `json:"dns2"`
	// ErrMsg 拨号/检测错误信息。
	ErrMsg string `json:"errmsg"`
	// Gateway 网关地址。
	Gateway string `json:"gateway"`
	// ID 接口行 ID。
	ID int64 `json:"id"`
	// IMEI 4G 模块 IMEI（观测值为空）。
	IMEI string `json:"imei"`
	// Interface 外网接口名（wan1、wan2...）。
	Interface string `json:"interface"`
	// Internet 外网连通状态（字段语义待真机验证）。
	Internet int64 `json:"internet"`
	// IPAddr 接口 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// Isnr 字段语义待真机验证（观测值为空）。
	Isnr string `json:"isnr"`
	// ISP 运营商。
	ISP string `json:"isp"`
	// MAC 接口 MAC 地址。
	MAC string `json:"mac"`
	// Member 绑定的物理网口名列表。
	Member []string `json:"member"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
	// PCID 4G 卡 PCID（观测值为空）。
	PCID string `json:"pcid"`
	// Power 4G 信号强度（观测值为空）。
	Power string `json:"power"`
	// Qnw 4G 运营商（观测值为空）。
	Qnw string `json:"qnw"`
	// UpdateTime 快照更新时间戳（观测值 0；单位待真机验证）。
	UpdateTime int64 `json:"updatetime"`
}

// AppFlowSummary 是应用流量分类快照行（app_flow 数组元素），各分类
// 流量为字节计（键名与固件一致，首字母大写）。
type AppFlowSummary struct {
	// Total 全部应用流量合计。
	Total int64 `json:"Total"`
	// HTTP HTTP 应用流量。
	HTTP int64 `json:"HTTP"`
	// Download 下载类应用流量。
	Download int64 `json:"Download"`
	// Transport 传输类应用流量。
	Transport int64 `json:"Transport"`
	// IM 即时通讯类应用流量。
	IM int64 `json:"IM"`
	// Video 视频类应用流量。
	Video int64 `json:"Video"`
	// Common 常用应用流量。
	Common int64 `json:"Common"`
	// Game 游戏类应用流量。
	Game int64 `json:"Game"`
	// Others 其他应用流量。
	Others int64 `json:"Others"`
	// Test 测试类应用流量。
	Test int64 `json:"Test"`
	// Unknown 未识别应用流量。
	Unknown int64 `json:"Unknown"`
}

// AppFlowBlock 是应用流量分类快照区块（app_flow）。固件对它双重包裹：
// Data 中 app_flow 键的值仍为 {"app_flow": [行列表]}。
type AppFlowBlock struct {
	// Rows 应用流量分类快照行（观测单行，单位字节）。
	Rows []AppFlowSummary `json:"app_flow"`
}

// AppHistoricalBlock 是应用流量历史排行区块（app_historical）。固件对它
// 双重包裹：Data 中 app_historical 键的值仍为 {"app_historical": [[行]]}。
type AppHistoricalBlock struct {
	// Groups 二维数组：每个内层数组为一个时间区间的分类流量行
	// （观测恒为 11 行，分类固定为 Unknown/HTTP/Download/Transport/
	// IM/Video/Common/Game/Others/Test/Total；区间含义待真机验证）。
	Groups [][]AppHistoricalRow `json:"app_historical"`
}

// AppHistoricalRow 是一个时间区间内某应用分类的流量行。
// 固件以字符串表示流量值（如 "16540064"），原样保留。
type AppHistoricalRow struct {
	// Download 该区间下载流量（字符串数字，单位待真机验证）。
	Download string `json:"download"`
	// Name 应用分类名。
	Name string `json:"name"`
	// Upload 该区间上传流量（字符串数字，单位待真机验证）。
	Upload string `json:"upload"`
}

// AppFlowOverview 是应用流量面板的聚合结果（homepage.show
// TYPE=app_flow,app_historical）。两个区块仅在对应 TYPE 被请求时出现
// （未请求的区块为零值）。
type AppFlowOverview struct {
	// AppFlow 应用流量分类快照区块（TYPE 含 app_flow）。
	AppFlow AppFlowBlock `json:"app_flow"`
	// AppHistorical 应用流量历史排行区块（TYPE 含 app_historical）。
	AppHistorical AppHistoricalBlock `json:"app_historical"`
}

// HomepageMonitorSystem 是首页带宽/每日统计数据块（homepage.show
// TYPE=monitor_system 的 monitor_system 区块）。
type HomepageMonitorSystem struct {
	// BandwidthTotal 带宽总量统计（观测值全 0；语义待真机验证）。
	BandwidthTotal BandwidthTotal `json:"bandwidth_total"`
	// EveryDayStats 每日统计行列表。
	EveryDayStats []EveryDayStat `json:"every_day_stats"`
}

// BandwidthTotal 是带宽总量统计子结构（bandwidth_total）。
type BandwidthTotal struct {
	// Upload 上行带宽总量（观测值 0；单位/语义待真机验证）。
	Upload int64 `json:"upload"`
	// Download 下行带宽总量（观测值 0；单位/语义待真机验证）。
	Download int64 `json:"download"`
}

// EveryDayStat 是每日统计行（every_day_stats 数组元素）。
type EveryDayStat struct {
	// CleanTime 观测值 20261110，形如 YYYYMMDD 日期数值（语义待真机验证）。
	CleanTime int64 `json:"clean_time"`
	// ID 行 ID。
	ID int64 `json:"id"`
	// WiFi WiFi 相关统计（观测值 0；语义待真机验证）。
	WiFi int64 `json:"wifi"`
}

// HomepageService 是首页仪表盘（homepage）的子服务。
type HomepageService struct {
	c *ikuai.Client
}

// SystemOverview 获取系统概况仪表盘数据（CPU/内存/在线用户/流量/
// 版本/AP 统计）。对应 /Action/call: homepage.show（TYPE=sysstat,ac_status），
// 即 Web 控制台「系统概况」首页总览面板。
func (s *Service) SystemOverview(ctx context.Context) (*SystemOverview, error) {
	var out SystemOverview
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeSystemOverview}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DHCPAddrpoolNum 获取 DHCP 地址池剩余可分配地址数（available_num，
// 抓包观测 17、复测 0；语义按命名与控制台推断，待真机验证）。
// 对应 /Action/call: homepage.show（TYPE=dhcp_addrpool_num）。
func (s *HomepageService) DHCPAddrpoolNum(ctx context.Context) (int64, error) {
	var out struct {
		AvailableNum int64 `json:"available_num"`
	}
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeDHCPAddrpoolNum}, &out); err != nil {
		return 0, err
	}
	return out.AvailableNum, nil
}

// InterfaceSnapshot 获取接口快照面板数据（物理网口信息 + 内外网接口
// 快照）。对应 /Action/call: homepage.show（TYPE=ether_info,snapshoot），
// 即 Web 控制台「系统概况」页面的接口快照卡片。
func (s *HomepageService) InterfaceSnapshot(ctx context.Context) (*InterfaceSnapshot, error) {
	var out InterfaceSnapshot
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeEtherInfoSnapshoot}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AppHistorical 获取应用流量历史排行（每个时间区间一组分类流量行）。
// 对应 /Action/call: homepage.show（TYPE=app_historical）。
func (s *HomepageService) AppHistorical(ctx context.Context) (*AppHistoricalBlock, error) {
	var out struct {
		AppHistorical AppHistoricalBlock `json:"app_historical"`
	}
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeAppHistorical}, &out); err != nil {
		return nil, err
	}
	return &out.AppHistorical, nil
}

// AppFlow 获取应用流量面板数据（分类快照 + 历史排行）。
// 对应 /Action/call: homepage.show（TYPE=app_flow,app_historical），
// 即 Web 控制台「系统概况」页面的应用流量卡片。
func (s *HomepageService) AppFlow(ctx context.Context) (*AppFlowOverview, error) {
	var out AppFlowOverview
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeAppFlowHistorical}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MonitorSystem 获取首页带宽总量与每日统计数据。
// 对应 /Action/call: homepage.show（TYPE=monitor_system）。
// 注意与 monitoring 包的 monitor_system.show（负载曲线）为不同 func。
func (s *HomepageService) MonitorSystem(ctx context.Context) (*HomepageMonitorSystem, error) {
	var out struct {
		MonitorSystem HomepageMonitorSystem `json:"monitor_system"`
	}
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeMonitorSystem}, &out); err != nil {
		return nil, err
	}
	return &out.MonitorSystem, nil
}

// Antipiracy 获取违禁监控状态（观测值 -1；字段语义待真机验证）。
// 对应 /Action/call: homepage.show（TYPE=antipiracy）。
func (s *HomepageService) Antipiracy(ctx context.Context) (int64, error) {
	var out struct {
		Antipiracy int64 `json:"antipiracy"`
	}
	if err := s.c.CallData(ctx, "homepage", "show",
		map[string]any{"TYPE": HomepageTypeAntipiracy}, &out); err != nil {
		return 0, err
	}
	return out.Antipiracy, nil
}

// Show 以任意 TYPE 组合调用 homepage.show，返回原始响应封装。
// 对应 /Action/call: homepage.show（TYPE=<types 逗号拼接>）。
// types 传本包 HomepageType* 常量；Data 中按请求的 TYPE 分别含
// sysstat/ac_status/ether_info/snapshoot_lan/snapshoot_wan/app_flow/
// app_historical/monitor_system/dhcp_addrpool_num/antipiracy 区块，
// 可用本包类型（SystemOverview/InterfaceSnapshot/AppFlowOverview 等）
// 对 resp.Data 按需解码。
func (s *HomepageService) Show(ctx context.Context, types ...string) (*ikuai.Response, error) {
	if len(types) == 0 {
		return nil, fmt.Errorf("overview: homepage.show 至少需要一个 TYPE: %w", ikuai.ErrInvalidParam)
	}
	return s.c.Call(ctx, "homepage", "show",
		map[string]any{"TYPE": strings.Join(types, ",")})
}

// parsePercent 解析固件的百分比字符串（如 "6.98%"、"6%"）为数值。
// 空值返回 0；非数字内容时报错。
func parsePercent(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("overview: 百分比 %q 不是数字: %w", s, err)
	}
	return v, nil
}
