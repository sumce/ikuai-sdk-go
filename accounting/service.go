package accounting

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// 固件 enabled 字段的布尔语义（"yes"/"no"，真机实证：webauth/pppoe_server/
// pptp_server/l2tp_server/openvpn-server/pppuser/ppp_passwd 均如此）。
const (
	// EnabledYes 已启用。
	EnabledYes = "yes"
	// EnabledNo 已停用。
	EnabledNo = "no"
)

// Service 是认证计费大类的服务入口，聚合 Web 控制台
// 「认证计费」（auth）菜单下的全部 16 个功能。
type Service struct {
	c *ikuai.Client

	// OnlineUsers 在线用户（ppp_online）：认证在线终端列表。
	OnlineUsers *OnlineUserService
	// WebAuth Web 认证服务（webauth）：认证方式开关/Radius/LDAP/白名单等全局配置。
	WebAuth *WebAuthService
	// VPNServers VPN 服务端族（pppoe_server/pptp_server/l2tp_server/
	// openvpn-server/ike_server）：PPPoE/PPTP/L2TP/OpenVPN/IKEv2 服务端配置。
	VPNServers *VPNServersService
	// Packages 套餐管理（ppp_package）：计费套餐 CRUD。
	Packages *PackagesService
	// Users 账号管理（pppuser）：计费账号 CRUD。
	Users *UsersService
	// SelfPasswd 自助密码（ppp_passwd）：允许哪些接入方式自助改密的开关配置。
	SelfPasswd *SelfPasswdService
	// PayLog 报表管理（ppp_paylog）：开户/缴费/退订等收费记录（只读）。
	PayLog *PayLogService
	// Coupon 优惠券管理（coupon）：上网码列表与批量生成。
	Coupon *CouponService
	// Proxy 代理账号/在线账号（pppoe_proxy）：代拨账号族。
	Proxy *ProxyService
	// Notices 通知族（notice_temp/notice_cycle/notice_remind/notice_expires）：
	// 实时通知/周期通知/到期提醒/PPPoE 到期通知。
	Notices *NoticesService
}

// New 构造认证计费服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:           c,
		OnlineUsers: &OnlineUserService{c: c},
		WebAuth:     &WebAuthService{c: c},
		VPNServers:  &VPNServersService{c: c},
		Packages:    &PackagesService{c: c},
		Users:       &UsersService{c: c},
		SelfPasswd:  &SelfPasswdService{c: c},
		PayLog:      &PayLogService{c: c},
		Coupon:      &CouponService{c: c},
		Proxy:       &ProxyService{c: c},
		Notices:     &NoticesService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用认证计费大类接口，作为未封装
// TYPE 组合与写动作（如上网码批量生成、账号套餐续费）的逃生口
// （每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// selectorShowParam 构造带 TYPE 选择器的 show 参数：
//
//	opts == nil 时仅由 defaults 生成参数；opts 非空时以 opts.BuildParam()
//	为底，defaults 只补缺失键，用户在 opts.Type / opts.Extra 中显式给出
//	的值优先。
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

// pageShowParam 构造分页类 func（ppp_online/ppp_package/pppuser/ppp_paylog/
// coupon/notice_cycle/pppoe_proxy 等）的 show 参数：opts == nil 时补
// ikuai.DefaultListOptions()（TYPE=total,data, limit="0,20"）。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// cfgTypeData 是单行配置类 func（webauth/pppoe_server/pptp_server/
// l2tp_server/openvpn-server/ppp_passwd/notice_temp/notice_remind/
// notice_expires）show 的 TYPE 选择器。真机 3.7.24 实证：TYPE=data 返回
// 单行配置（与抓包 alg.show TYPE=data 同构）；TYPE=total,data 会回
// 30006 unknown TYPE。
const cfgTypeData = "data"
