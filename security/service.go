package security

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是安全设置大类的服务入口，聚合 Web 控制台「安全设置」菜单下
// 的全部 5 个功能。
type Service struct {
	c *ikuai.Client

	// ACL 访问控制（acl）：五元组 ACL 规则（进/出接口、协议、动作）。
	ACL *ACLService
	// ARP ARP 表（arp）：ARP 绑定列表（运行时表，只读查询）。
	ARP *ARPService
	// ConnLimit 连接数限制（conn_limit）：按内网 IP 的并发连接数限制规则。
	ConnLimit *ConnLimitService
	// Advanced 安全高级设置（advanced）：禁 Ping/防 DOS/TCP-MSS 等单行配置。
	Advanced *AdvancedService
	// Firewall 云防火墙（firewall）：防火墙插件状态查询。
	Firewall *FirewallService
}

// New 构造安全设置服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:         c,
		ACL:       &ACLService{c: c},
		ARP:       &ARPService{c: c},
		ConnLimit: &ConnLimitService{c: c},
		Advanced:  &AdvancedService{c: c},
		Firewall:  &FirewallService{c: c},
	}
}

// 固件通用的启用/禁用字符串值（enabled 字段，与固件完全一致）。
const (
	EnabledYes = "yes"
	EnabledNo  = "no"
)

// SelectorsTypeAll 是安全设置各列表 func（acl/arp/conn_limit）show 动作
// 的表单选择器 TYPE 值：响应在 total/data 之外附带表单可选值区块（接口
// 列表、IP/MAC/时间分组等，真机 3.7.24 验证）。
const SelectorsTypeAll = "all"

// Raw 以任意 func_name/action/param 调用安全设置大类接口，作为未封装
// action/TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
// 注意：本包全部写动作（acl.add/edit/del、conn_limit.add/edit/del、
// advanced.save、firewall.save 等）未对真机验证，只应通过 Raw 在充分
// 知晓后果的前提下调用未封装动作，且严禁对真机执行写测试
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

// pageShowParam 构造标准分页类 func（acl/arp/conn_limit 均支持
// TYPE=total,data 的 show，真机 3.7.24 验证）的参数：opts == nil 时补
// ikuai.DefaultListOptions()（TYPE=total,data, limit="0,20"），defaults
// 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// typedShowParam 构造 TYPE 选择器类 func（advanced 以 TYPE=data 区块为主）
// 的 show 参数：opts 为 nil 或未显式指定 Type 时强制写入 defaultType
// （固件对这类 func 的 TYPE=total 回 30006 unknown TYPE (total)，真机
// 3.7.24 验证）；opts.Type 显式给出的值被尊重。其余键仍尊重 opts
// （limit/ORDER_BY/ORDER 等）。
func typedShowParam(opts *ikuai.ListOptions, defaultType string, defaults map[string]any) map[string]any {
	p := selectorShowParam(opts, defaults)
	if opts == nil || opts.Type == "" {
		p["TYPE"] = defaultType
	}
	return p
}
