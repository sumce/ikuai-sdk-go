package security

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// firewall.show 的 TYPE 选择器常量（真机 3.7.24 穷举验证：仅 all/status
// 有效，其余 TYPE 回 30006 unknown TYPE）。
const (
	// FirewallTypeAll 云防火墙状态总览区块（含 CPU 核心数）。
	FirewallTypeAll = "all"
	// FirewallTypeStatus 仅插件状态区块。
	FirewallTypeStatus = "status"
)

// FirewallStatus 是云防火墙插件页的状态（firewall.show TYPE=all 的 Data，
// 真机 3.7.24 实测）。云防火墙通过插件方式集成防火墙功能模块，需绑定
// 爱快云平台后安装（i18n helpTip.firewallPluginTip / Plugin_management）。
type FirewallStatus struct {
	// CPUCores CPU 核心数（实测 20；用于云防火墙容量评估，语义待真机验证）。
	CPUCores int64 `json:"cpu_cores"`
	// Status 插件状态（实测 0；推测 0=未安装/未开通，非 0 的枚举语义
	// 待真机验证）。
	Status int64 `json:"status"`
}

// FirewallService 是云防火墙（firewall）的子服务。
//
// 该 func 为云平台插件页：真机仅实证 show TYPE=all/status 两个只读
// TYPE；安装/开通等动作由云平台流程完成，未封装（可用 Service.Raw 在
// 充分知晓后果的前提下调用）。
type FirewallService struct {
	c *ikuai.Client
}

// Get 获取云防火墙插件状态总览（插件状态 + CPU 核心数）。
// 对应 /Action/call: firewall.show（TYPE=all），
// 即 Web 控制台「安全设置 → 云防火墙」页面。
func (s *FirewallService) Get(ctx context.Context) (*FirewallStatus, error) {
	var out FirewallStatus
	if err := s.c.CallData(ctx, "firewall", "show",
		map[string]any{"TYPE": FirewallTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Status 获取云防火墙插件状态（TYPE=status 区块）。
// 对应 /Action/call: firewall.show（TYPE=status）。
func (s *FirewallService) Status(ctx context.Context) (int64, error) {
	var out struct {
		Status int64 `json:"status"`
	}
	if err := s.c.CallData(ctx, "firewall", "show",
		map[string]any{"TYPE": FirewallTypeStatus}, &out); err != nil {
		return 0, err
	}
	return out.Status, nil
}

// Save 保存云防火墙状态配置（写动作）。
// 对应 /Action/call: firewall.save（action 名按同类配置 func 的 save
// 惯例推断，本 func 无任何写动作实证）。
// ⚠️ 写动作，action/字段均为推断，待真机验证：仅离线测试覆盖，
// 禁止对真机调用；云防火墙的开通/安装请走云平台流程。
func (s *FirewallService) Save(ctx context.Context, st FirewallStatus) error {
	_, err := s.c.Call(ctx, "firewall", "save", st)
	return err
}
