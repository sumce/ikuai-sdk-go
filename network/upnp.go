package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// UPnPConfig 是 UPnP 服务的全局配置（upnpd.show Data.data[0]，真机实证；
// 单行配置）。注意：upnpd.show 不接受 TYPE=total,data（真机回
// 30006 unknown TYPE (total)），必须空参数调用。
type UPnPConfig struct {
	// ID 行 ID（真机观测恒为 1）。
	ID int64 `json:"id"`
	// Enabled 是否启用 UPnP（"yes"/"no"，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// ExcludePort 排除端口段（真机观测 "1-1024"，即不对 1-1024 做映射）。
	ExcludePort string `json:"exclude_port"`
	// LANIP 允许发起 UPnP 映射的内网地址范围
	// （真机观测 "0.0.0.0-255.255.255.255"=不限制）。
	LANIP string `json:"lan_ip"`
	// Interface 指定生效线路（真机观测 "any"=任意）。
	Interface string `json:"interface"`
	// CheckLink 连接检查开关（0/1；语义待真机验证）。
	CheckLink int64 `json:"check_link"`
	// CheckInterval 连接检查间隔（分钟，真机观测 5）。
	CheckInterval int64 `json:"check_interval"`
	// RstSwitch 定时重置映射开关（0/1，真机观测 0）。
	RstSwitch int64 `json:"rst_switch"`
	// RstWeek 重置周期（"1234567"=每天，星期位串 0/1-6）。
	RstWeek string `json:"rst_week"`
	// RstTime 重置时刻（"HH:MM"，真机观测 "00:00"）。
	RstTime string `json:"rst_time"`
}

// IsEnabled 报告 UPnP 服务是否启用。
func (r UPnPConfig) IsEnabled() bool { return r.Enabled == EnabledYes }

// UPnPLease 是一条 UPnP 自动端口映射租约（upnpd_leases.show 行；
// 真机列表为空（UPnP 未启用），字段名由固件语义推导，待真机验证；
// 运行状态表，只读）。
type UPnPLease struct {
	// ID 行 ID（待真机验证）。
	ID int64 `json:"id"`
	// Interface 映射生效线路（待真机验证）。
	Interface string `json:"interface"`
	// Protocol 协议（tcp/udp，待真机验证）。
	Protocol string `json:"protocol"`
	// ExtPort 外部端口（待真机验证）。
	ExtPort int64 `json:"ext_port"`
	// LANIP 内网主机地址（待真机验证）。
	LANIP string `json:"lan_ip"`
	// LANPort 内部端口（待真机验证）。
	LANPort int64 `json:"lan_port"`
	// Description 映射描述（发起 UPnP 映射的应用名，待真机验证）。
	Description string `json:"description"`
}

// UPnPService 是 UPnP 族（upnpd/upnpd_leases）的子服务。
type UPnPService struct {
	c *ikuai.Client
}

// Get 获取 UPnP 服务全局配置（单行）。
// 对应 /Action/call: upnpd.show（{}，真机实证；TYPE=total,data 回
// 30006 unknown TYPE，故固定空参数），即 Web 控制台
// 「应用工具 → UPnP → UPnP 设置」页面。
// opts 允许通过 Extra 附加参数（真机不接受 TYPE/limit）。
func (s *UPnPService) Get(ctx context.Context) (*UPnPConfig, error) {
	var out struct {
		Data []UPnPConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "upnpd", "show", bareShowParam(nil), &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// Set 保存 UPnP 服务全局配置（写动作）。
// 对应 /Action/call: upnpd.set。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *UPnPService) Set(ctx context.Context, cfg UPnPConfig) error {
	_, err := s.c.Call(ctx, "upnpd", "set", cfg)
	return err
}

// LeaseList 获取 UPnP 自动端口映射租约列表（运行状态表，只读；
// show 路径真机实证返回空表）。
// 对应 /Action/call: upnpd_leases.show（TYPE=total,data），即 Web 控制台
// 「应用工具 → UPnP → UPnP 状态」页面。
func (s *UPnPService) LeaseList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[UPnPLease], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "upnpd_leases", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[UPnPLease](resp.Data)
}
