package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是流控分流大类的服务入口，聚合 Web 控制台
// 「网络设置 → 流控分流」与「流控设置 → 限速」菜单下的全部 10 个功能。
type Service struct {
	c *ikuai.Client

	// MultiWan 多线负载（lb_pcc）：多线负载规则（相同运营商负载比重）。
	MultiWan *MultiWanService
	// Layer7 协议分流（stream_layer7）：按七层应用协议分流到指定接口。
	Layer7 *Layer7Service
	// IPPort 端口分流（stream_ipport）：按源地址/端口分流到指定接口。
	IPPort *IPPortService
	// Domain 域名分流（stream_domain）：按目标域名分流到指定接口。
	Domain *DomainService
	// UpDown 流量上下线分流（stream_updown）：按流量上下线条件分流。
	UpDown *UpDownService
	// SmartControl 智能流控（stream_control）：智能流控模式开关。
	SmartControl *SmartControlService
	// SimpleQoS IP 限速（simple_qos）：按 IP 的上下行限速规则。
	SimpleQoS *SimpleQoSService
	// MacQoS MAC 限速（mac_qos）：按 MAC 的上下行限速规则。
	MacQoS *MacQoSService
	// CustomProtos 自定义协议（dprotos）：自定义协议库（端口特征）。
	CustomProtos *CustomProtosService
	// AdvancedProtos 高级自定义协议（dprotos_l7）：高级自定义协议库。
	AdvancedProtos *AdvancedProtosService
}

// New 构造流控分流服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:              c,
		MultiWan:       &MultiWanService{c: c},
		Layer7:         &Layer7Service{c: c},
		IPPort:         &IPPortService{c: c},
		Domain:         &DomainService{c: c},
		UpDown:         &UpDownService{c: c},
		SmartControl:   &SmartControlService{c: c},
		SimpleQoS:      &SimpleQoSService{c: c},
		MacQoS:         &MacQoSService{c: c},
		CustomProtos:   &CustomProtosService{c: c},
		AdvancedProtos: &AdvancedProtosService{c: c},
	}
}

// 固件通用的启用/禁用字符串值（enabled 字段，与固件完全一致）。
const (
	EnabledYes = "yes"
	EnabledNo  = "no"
)

// SelectorsTypeAll 是流控分流各列表 func（lb_pcc/stream_layer7/stream_ipport/
// stream_domain/stream_updown/simple_qos/mac_qos/dprotos/dprotos_l7）show
// 动作的表单选择器 TYPE 值：响应在 total/data 之外附带表单可选值区块
// （接口列表、IP 分组、运营商列表等，真机 3.7.24 验证）。
const SelectorsTypeAll = "all"

// Raw 以任意 func_name/action/param 调用流控分流大类接口，作为未封装
// action/TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
// 注意：本包全部写动作（add/edit/del/seting 等）未对真机验证，只应通过
// Raw 在充分知晓后果的前提下调用未封装动作，且严禁对真机执行写测试
// （见 docs/RULES.md §9.2 与本包各写方法注释）。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// selectorShowParam 构造带固定默认键的 show 参数：opts == nil 时仅由
// defaults 生成参数；opts 非空时以 opts.BuildParam() 为底，defaults 只补
// 缺失键，用户在 opts.Type / opts.Extra 中显式给出的值优先。
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

// pageShowParam 构造标准分页类 func（本包 9 个列表型 func 均支持
// TYPE=total,data 的 show）的参数：opts == nil 时补
// ikuai.DefaultListOptions()（TYPE=total,data, limit="0,20"），defaults
// 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}
