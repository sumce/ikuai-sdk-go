package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// IGMPProxy 是 IGMP 代理配置（igmp_proxy.show Data.data[0]，真机实证；
// 单行配置，show 接受 TYPE=total,data 与空参数两种形态）。
type IGMPProxy struct {
	// ID 行 ID（真机观测恒为 1）。
	ID int64 `json:"id"`
	// Enabled 是否启用 IGMP 代理（"yes"/"no"，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// Upstream 上游接口（连运营商侧的 WAN 线路，未配置为空串）。
	Upstream string `json:"upstream"`
	// Downstream 下游接口（面向组播接收者的 LAN 接口，逗号分隔，未配置为空串）。
	Downstream string `json:"downstream"`
	// Version IGMP 协议版本（真机观测 "3"，固件以字符串返回）。
	Version string `json:"version"`
}

// IsEnabled 报告 IGMP 代理是否启用。
func (r IGMPProxy) IsEnabled() bool { return r.Enabled == EnabledYes }

// Downstreams 按逗号拆分下游接口列表。
func (r IGMPProxy) Downstreams() []string { return splitList(r.Downstream) }

// IPTVConfig 是 IPTV 透传配置（iptv.show Data.data[0]，真机实证；
// 单行配置。iptv.show 不接受 TYPE=total/data（真机回 30006
// unknown TYPE），必须空参数调用）。
type IPTVConfig struct {
	// ID 行 ID（真机观测恒为 1）。
	ID int64 `json:"id"`
	// Enabled 是否启用 IPTV 透传（"yes"/"no"，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// Mode 模式（真机观测 0；枚举语义待真机验证）。
	Mode int64 `json:"mode"`
	// WANIface 上联口（ IPTV 信号接入的 WAN 口，未配置为空串）。
	WANIface string `json:"wan_iface"`
	// WANVLANID 上联口 VLAN ID（0=不打 VLAN 标签）。
	WANVLANID int64 `json:"wan_vlanid"`
	// LANIface 输出口（IPTV 机顶盒接入的 LAN 口，启用后将被解绑用于转发）。
	LANIface string `json:"lan_iface"`
	// LANVLANID 输出口 VLAN ID（0=不打 VLAN 标签）。
	LANVLANID int64 `json:"lan_vlanid"`
}

// IsEnabled 报告 IPTV 透传是否启用。
func (r IPTVConfig) IsEnabled() bool { return r.Enabled == EnabledYes }

// MulticastService 是组播族（igmp_proxy/iptv）的子服务。
type MulticastService struct {
	c *ikuai.Client
}

// IGMPGet 获取 IGMP 代理配置（单行）。
// 对应 /Action/call: igmp_proxy.show（TYPE=total,data，真机实证），
// 即 Web 控制台「网络设置 → IGMP 代理」页面。
func (s *MulticastService) IGMPGet(ctx context.Context) (*IGMPProxy, error) {
	var out struct {
		Data []IGMPProxy `json:"data"`
	}
	if err := s.c.CallData(ctx, "igmp_proxy", "show",
		map[string]any{"TYPE": "total,data"}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// IGMPSet 保存 IGMP 代理配置（写动作）。
// 对应 /Action/call: igmp_proxy.set。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *MulticastService) IGMPSet(ctx context.Context, cfg IGMPProxy) error {
	_, err := s.c.Call(ctx, "igmp_proxy", "set", cfg)
	return err
}

// IPTVGet 获取 IPTV 透传配置（单行）。
// 对应 /Action/call: iptv.show（{}，真机实证；TYPE=total,data 回
// 30006 unknown TYPE，故固定空参数），即 Web 控制台「网络设置 → IPTV」页面。
func (s *MulticastService) IPTVGet(ctx context.Context) (*IPTVConfig, error) {
	var out struct {
		Data []IPTVConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "iptv", "show", bareShowParam(nil), &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// IPTVSet 保存 IPTV 透传配置（写动作）。
// 对应 /Action/call: iptv.set。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *MulticastService) IPTVSet(ctx context.Context, cfg IPTVConfig) error {
	_, err := s.c.Call(ctx, "iptv", "set", cfg)
	return err
}
