package ac

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是 AC 管理（无线控制器）大类的服务入口，聚合 Web 控制台
// 「AC管理」菜单下的全部 8 个功能。
type Service struct {
	c *ikuai.Client

	// Status 无线概况（ac_status）：AC 开关状态、AP/终端统计、健康评分、信道与负载图表。
	Status *StatusService
	// Server AP 列表（ac_server）：接入 AC 的 AP 设备列表。
	Server *ServerService
	// Group AP 分组（ac_group）：AP 分组管理与组内 AP 分配。
	Group *GroupService
	// Upgrade AP 系统升级（ac_upgrade）：AP 固件可升级列表与升级动作。
	Upgrade *UpgradeService
	// Blacklist 无线黑白名单（wls_black）：按 MAC/SSID 控制无线终端接入。
	Blacklist *BlacklistService
	// Mvlan 无线终端 VLAN（wls_mvlan）：按 MAC/SSID 指定无线终端所属 VLAN。
	Mvlan *MvlanService
	// OnlineClients 无线终端列表（ac_online_clt）：在线无线终端与信号/速率信息。
	OnlineClients *OnlineClientsService
	// NetOptimize 无线网优（ac_net_optimize）：无线优化配置与体检结果。
	NetOptimize *NetOptimizeService
}

// New 构造 AC 管理服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:             c,
		Status:        &StatusService{c: c},
		Server:        &ServerService{c: c},
		Group:         &GroupService{c: c},
		Upgrade:       &UpgradeService{c: c},
		Blacklist:     &BlacklistService{c: c},
		Mvlan:         &MvlanService{c: c},
		OnlineClients: &OnlineClientsService{c: c},
		NetOptimize:   &NetOptimizeService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用 AC 管理大类接口，作为未封装
// TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// ACStatRow 是 ac_status 概况区块与 ac_net_optimize 结果中行结构未知
// 的通用行（真机 3.7.24 无 AP 设备，相关数组恒为空，字段无法实证），
// 保留原始键值供上层按需解析；字段语义待真机验证。
type ACStatRow map[string]any

// selectorShowParam 构造 TYPE 选择器类 func（ac_status）的 show 参数。
//
//	opts == nil 时仅由 defaults 生成参数（与控制台抓包一致，不带分页
//	limit）；opts 非空时以 opts.BuildParam() 为底，defaults 只补缺失键，
//	用户在 opts.Type / opts.Extra 中显式给出的值优先。
func selectorShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	var p map[string]any
	if opts != nil {
		p = opts.BuildParam()
	}
	if p == nil {
		p = map[string]any{}
	}
	for k, v := range defaults {
		if _, ok := p[k]; !ok {
			p[k] = v
		}
	}
	return p
}

// pageShowParam 构造分页类 func（ac_server/ac_group/ac_upgrade/wls_black/
// wls_mvlan/ac_online_clt/ac_net_optimize 等）的 show 参数：opts == nil 时
// 补 ikuai.DefaultListOptions()（TYPE=total,data, limit="0,20"），
// defaults 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}
