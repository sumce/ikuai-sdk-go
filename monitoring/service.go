package monitoring

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是状态监控大类的服务入口，聚合 Web 控制台
// 「状态监控」菜单下的全部 10 个功能。
type Service struct {
	c *ikuai.Client

	// Interface 线路监控（monitor_iface）：接口检测状态/实时流量/网口信息/快照。
	Interface *InterfaceService
	// LanIP 终端监控 IPv4（monitor_lanip）：内网终端列表，支持分页与关键字搜索。
	LanIP *LanIPService
	// LanIPv6 终端监控 IPv6（monitor_lanipv6）：内网 IPv6 终端与接口列表。
	LanIPv6 *LanIPv6Service
	// AppFlow 协议监控（monitor_app_flow）：应用流量历史与协议排行。
	AppFlow *AppFlowService
	// L7QoS 策略监控（monitor_l7qos）：流控策略状态。
	L7QoS *L7QoSService
	// System 负载监控（monitor_system）：CPU/内存/磁盘/终端数/连接数/温度曲线。
	System *SystemService
	// CFlow 分流监控（cflow）：连接统计汇总（今日/本周/昨日）。
	CFlow *CFlowService
	// Camera 摄像头监控（camera）：摄像头探测。
	Camera *CameraService
	// CloudSwitch 交换机监控（cloud_switch）：云管理交换机列表。
	CloudSwitch *CloudSwitchService
	// DevControl 周边设备监控（dev_control）：智能设备管控列表。
	DevControl *DevControlService
}

// New 构造状态监控服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:           c,
		Interface:   &InterfaceService{c: c},
		LanIP:       &LanIPService{c: c},
		LanIPv6:     &LanIPv6Service{c: c},
		AppFlow:     &AppFlowService{c: c},
		L7QoS:       &L7QoSService{c: c},
		System:      &SystemService{c: c},
		CFlow:       &CFlowService{c: c},
		Camera:      &CameraService{c: c},
		CloudSwitch: &CloudSwitchService{c: c},
		DevControl:  &DevControlService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用状态监控大类接口，作为未封装
// TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// selectorShowParam 构造 TYPE 选择器类 func（monitor_iface /
// monitor_system / monitor_app_flow / monitor_l7qos）的 show 参数。
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

// pageShowParam 构造分页类 func（monitor_lanip / camera / cloud_switch /
// dev_control 等）的 show 参数：opts == nil 时补 ikuai.DefaultListOptions()
// （TYPE=total,data, limit="0,20"），defaults 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}
