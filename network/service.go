package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是网络设置大类的服务入口，聚合 Web 控制台
// 「网络设置」菜单下的全部 35 个功能（按功能族归并为子服务）。
type Service struct {
	c *ikuai.Client

	// LAN 内外网设置（lan）：内网接口列表/单接口/编辑。
	LAN *LANService
	// DHCP DHCP 服务族（dhcp_server/dhcp_static/dhcp_lease/dhcp_acl_mac）。
	DHCP *DHCPService
	// DNS DNS 设置族（dns/dns_replace）。
	DNS *DNSService
	// VLAN VLAN 设置（vlan）。
	VLAN *VLANService
	// Groups 分组族（ipgroup/ipv6group/macgroup/dtgroup）。
	Groups *GroupsService
	// Route 路由族（static_rt/static_rt_table）。
	Route *RouteService
	// VPNClient VPN 客户端族（pptp/l2tp/openvpn/ipsec/wireguard/ike/sdwan/wuying）。
	VPNClient *VPNClientService
	// UPnP UPnP 族（upnpd/upnpd_leases）。
	UPnP *UPnPService
	// NAT NAT 规则族（nat_rule/netmap）。
	NAT *NATService
	// DNAT 端口映射（dnat，抓包实证最完整，独立子服务）。
	DNAT *DNATService
	// IPv6 IPv6 族（ipv6/ipv6_stream/dhcp6_lease/ipv6_static/dhcp6_acl_mac/ipv6_neighbor）。
	IPv6 *IPv6Service
	// Multicast 组播族（igmp_proxy/iptv）。
	Multicast *MulticastService
}

// New 构造网络设置服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:         c,
		LAN:       &LANService{c: c},
		DHCP:      &DHCPService{c: c},
		DNS:       &DNSService{c: c},
		VLAN:      &VLANService{c: c},
		Groups:    &GroupsService{c: c},
		Route:     &RouteService{c: c},
		VPNClient: &VPNClientService{c: c},
		UPnP:      &UPnPService{c: c},
		NAT:       &NATService{c: c},
		DNAT:      &DNATService{c: c},
		IPv6:      &IPv6Service{c: c},
		Multicast: &MulticastService{c: c},
	}
}

// 固件通用的启用/禁用字符串值（enabled 字段，与固件完全一致）。
const (
	EnabledYes = "yes"
	EnabledNo  = "no"
)

// Raw 以任意 func_name/action/param 调用网络设置大类接口，作为未封装
// action/TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
// 注意：本包大量写动作（add/edit/del/set 等）虽有强类型封装，但除 dnat
// （抓包实证）外均未经过真机验证，严禁对真机执行未验证的写动作
// （见 docs/RULES.md §9.2）。
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

// pageShowParam 构造标准分页类 func（dhcp_*/ipgroup/vlan/static_rt/dnat/
// netmap/nat_rule/ipv6_neighbor 等 TYPE=total,data 型 func）的 show 参数：
// opts == nil 时补 ikuai.DefaultListOptions()（TYPE=total,data,
// limit="0,20"），defaults 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// forcedShowParam 构造 TYPE 固定的 show 参数（lan/support_wisp、dns/dns_config、
// dhcp6_lease/all 等）：opts 为 nil 或未显式指定 Type 时强制写入 defaultType
// （真机 3.7.24 对这类 func 传入其他 TYPE 回 30006 unknown TYPE）；opts.Type
// 显式给出的值被尊重。其余键仍尊重 opts（limit/ORDER_BY/ORDER 等）。
func forcedShowParam(opts *ikuai.ListOptions, defaultType string, defaults map[string]any) map[string]any {
	p := pageShowParam(opts, defaults)
	if opts == nil || opts.Type == "" {
		p["TYPE"] = defaultType
	}
	return p
}

// bareShowParam 构造空参数型 show 参数（upnpd/iptv/wuying/ik_web_sdwan 等
// 真机实证不接受 TYPE=total,data 的 func）：opts == nil 时返回空对象；
// opts 非空时返回 opts.BuildParam()（可为 nil，由核心包转空对象）。
func bareShowParam(opts *ikuai.ListOptions) map[string]any {
	if opts == nil {
		return map[string]any{}
	}
	return opts.BuildParam()
}
