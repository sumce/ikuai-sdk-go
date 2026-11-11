package security

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// advanced.show 的 TYPE 选择器常量（真机 3.7.24 验证）。
const (
	// AdvancedTypeData 配置行区块选择器。本 func 不接受 TYPE=total
	// （真机对 TYPE=total,data 回 30006 unknown TYPE (total)）。
	AdvancedTypeData = "data"
)

// AdvancedConfig 是安全高级设置的单行配置（advanced.show data 数组元素，
// 真机 3.7.24 实测恒为单行 id=1；开关字段为 0/1 数值，语义对照固件
// i18n firewall.advanced 区块）。
type AdvancedConfig struct {
	// ID 行 ID（真机实测恒为 1）。
	ID int64 `json:"id"`
	// NoPingLAN 禁止内网 PING 路由（i18n advanced.intranet_ping；
	// 实测 0=不禁止）。
	NoPingLAN int64 `json:"noping_lan"`
	// NoPingWAN 禁止外网 PING 路由（i18n advanced.outside_ping；
	// 实测 1=禁止）。
	NoPingWAN int64 `json:"noping_wan"`
	// NoTracert 禁止客户端 tracert 路由追踪（i18n advanced.client_tracert；
	// 实测 0=不禁止）。
	NoTracert int64 `json:"notracert"`
	// HijackPing 劫持所有 PING 值，游戏的 Ping 值显示为 0
	// （i18n advanced.hijack_ping；实测 0=关闭；使用时请勿同时禁 Ping）。
	HijackPing int64 `json:"hijack_ping"`
	// Invalid 丢弃无效的连接访问路由（i18n advanced.Invalid_connection；
	// 实测 0=关闭）。
	Invalid int64 `json:"invalid"`
	// DOSDefenseLAN 启用内网 DOS 攻击防御（i18n advanced.dns_defense；
	// 实测 1=启用）。
	DOSDefenseLAN int64 `json:"dos_lan"`
	// DOSDefenseNum 每个 IP 允许的最大访问并发数（i18n advanced.limit_num；
	// 实测 300；若用路由做 DNS 代理服务可填较大数值）。
	DOSDefenseNum int64 `json:"dos_lan_num"`
	// LimitTCP2P TCP 2P 防御开关（真机实测 0；固件 i18n 无对应标签，
	// 枚举语义待真机验证）。
	LimitTCP2P int64 `json:"limit_tcp2p"`
	// LimitTCP2PNum TCP 2P 防御阈值（真机实测 50；语义待真机验证）。
	LimitTCP2PNum int64 `json:"limit_tcp2p_num"`
	// LimitUDP2P UDP 2P 防御开关（真机实测 0；语义待真机验证）。
	LimitUDP2P int64 `json:"limit_udp2p"`
	// LimitUDP2PNum UDP 2P 防御阈值（真机实测 300；语义待真机验证）。
	LimitUDP2PNum int64 `json:"limit_udp2p_num"`
	// TCPMSS 启用 TCP 最大报文长度（i18n advanced.tcp_length；
	// 实测 1=启用）。
	TCPMSS int64 `json:"tcp_mss"`
	// TCPMSSNum TCP-MSS 值（i18n advanced.tcp_mss；实测 1400；取值
	// 1000-1500 且为 4 的整数倍，单位：字节）。
	TCPMSSNum int64 `json:"tcp_mss_num"`
}

// AdvancedService 是安全高级设置（advanced）的子服务。
type AdvancedService struct {
	c *ikuai.Client
}

// Get 获取安全高级设置（禁 Ping/tracert、DOS 防御、TCP-MSS 等开关）。
// 对应 /Action/call: advanced.show（TYPE=data），
// 即 Web 控制台「安全设置 → 高级设置」页面。
func (s *AdvancedService) Get(ctx context.Context) (*AdvancedConfig, error) {
	page, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(page.Rows) == 0 {
		return nil, nil
	}
	return &page.Rows[0], nil
}

// List 以分页形态获取安全高级设置配置行（固件恒返回单行）。
// 对应 /Action/call: advanced.show（TYPE=data, limit="offset,rows"）。
// 本 func 不接受 TYPE=total（真机回 30006 unknown TYPE (total)）。
func (s *AdvancedService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[AdvancedConfig], error) {
	param := typedShowParam(opts, AdvancedTypeData, nil)
	resp, err := s.c.Call(ctx, "advanced", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[AdvancedConfig](resp.Data)
}

// Save 保存安全高级设置（写动作）。
// 对应 /Action/call: advanced.save（action 名按同类单行配置 func 的
// alg.save/remote_control.save 推断）。cfg 应为 Get 返回并修改后的完整行。
// ⚠️ 写动作，action/字段待真机验证：仅离线测试覆盖，禁止对真机调用
// （修改本配置会直接影响路由器防火墙行为）。
func (s *AdvancedService) Save(ctx context.Context, cfg AdvancedConfig) error {
	_, err := s.c.Call(ctx, "advanced", "save", cfg)
	return err
}
