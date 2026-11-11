package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// vrrp_config.show 的 TYPE 选择器常量（真机 3.7.24 验证）。
const (
	// VRRPTypeData 配置行区块选择器。注意本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)），List/Get 均使用 data。
	VRRPTypeData = "data"
)

// VRRPConfig 是主备热备的单行配置（vrrp_config.show data 数组元素，
// 观测恒为单行 id=1）。
type VRRPConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// Type 工作模式（观测值 1；枚举语义待真机验证，推测 1=主备热备）。
	Type int64 `json:"type"`
	// Enabled 是否启用（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// Prio VRRP 优先级（观测值 150）。
	Prio int64 `json:"prio"`
	// Method 抢占/通告方式（观测值 2；枚举语义待真机验证）。
	Method int64 `json:"method"`
	// Domain 检测域名（观测值 "www.baidu.com"）。
	Domain string `json:"domain"`
	// DNS 检测用 DNS 服务器（观测值 "114.114.114.114"）。
	DNS string `json:"dns"`
	// Gateway 检测网关（观测值与 Domain 相同，语义待真机验证）。
	Gateway string `json:"gateway"`
	// Ifnames 参与热备的接口名列表（逗号分隔，未启用时为空）。
	Ifnames string `json:"ifnames"`
	// Interval VRRP 通告间隔（秒，观测值 3）。
	Interval int64 `json:"interval"`
	// AutoSync 配置自动同步开关（0=关闭）。
	AutoSync int64 `json:"auto_sync"`
	// SingleLine 单线路模式开关（0=关闭）。
	SingleLine int64 `json:"single_line"`
	// IgnoreWanstatus 忽略外网状态开关（观测值 1）。
	IgnoreWanstatus int64 `json:"ignore_wanstatus"`
	// Interfaces 虚拟 IP 绑定的接口名列表（逗号分隔）。
	Interfaces string `json:"interfaces"`
	// VirtualIPs 虚拟 IP 列表（逗号分隔）。
	VirtualIPs string `json:"virtual_ips"`
	// HtIface 心跳接口（未启用时为空）。
	HtIface string `json:"ht_iface"`
	// RemoteAddr 对端（备机）管理地址（未启用时为空）。
	RemoteAddr string `json:"remote_addr"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (c VRRPConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// VRRPInfo 是主备热备页的聚合结果（vrrp_config.show TYPE=data）。
type VRRPInfo struct {
	// Data 配置行（观测恒为单行）。
	Data []VRRPConfig `json:"data"`
	// Version 上报该配置的固件版本（观测值 "3.7.24"）。
	Version string `json:"version"`
}

// Row 返回配置首行；data 为空数组时返回 nil。
func (v *VRRPInfo) Row() *VRRPConfig {
	if v == nil || len(v.Data) == 0 {
		return nil
	}
	return &v.Data[0]
}

// VRRPService 是主备热备（vrrp_config）的子服务。
type VRRPService struct {
	c *ikuai.Client
}

// Get 获取主备热备配置（含固件版本号）。
// 对应 /Action/call: vrrp_config.show（TYPE=data），
// 即 Web 控制台「系统设置 → 主备热备」页面。
func (s *VRRPService) Get(ctx context.Context) (*VRRPInfo, error) {
	var out VRRPInfo
	if err := s.c.CallData(ctx, "vrrp_config", "show",
		map[string]any{"TYPE": VRRPTypeData}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List 以分页形态获取主备热备配置行（固件恒返回单行）。
// 对应 /Action/call: vrrp_config.show（TYPE=data, limit="offset,rows"）。
// opts 为 nil 时仅发送 {"TYPE":"data"}；opts 中的分页参数会被固件接受，
// 但本 func 不支持 TYPE=total（超出的 TYPE 值会被归一为 data）。
func (s *VRRPService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[VRRPConfig], error) {
	param := typedShowParam(opts, VRRPTypeData, nil)
	resp, err := s.c.Call(ctx, "vrrp_config", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[VRRPConfig](resp.Data)
}

// Save 保存主备热备配置（写动作）。
// 对应 /Action/call: vrrp_config.save（action 名按同类单行配置 func 的
// alg.save 推断）。cfg 应为 Get 返回并修改后的完整行（固件按整行覆盖）。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *VRRPService) Save(ctx context.Context, cfg VRRPConfig) error {
	_, err := s.c.Call(ctx, "vrrp_config", "save", cfg)
	return err
}
