// Package accounting 实现爱快 Web 控制台「认证计费」大类，共 16 个功能：
//
//	ppp_online      在线用户
//	webauth         Web 认证服务
//	pppoe_server    PPPoE 服务端
//	pptp_server     PPTP 服务端
//	l2tp_server     L2TP 服务端
//	openvpn-server  OpenVPN 服务端
//	ike_server      IPSec IKEv2 服务端
//	ppp_package     套餐管理
//	pppuser         账号管理
//	ppp_passwd      自助密码
//	ppp_paylog      报表管理
//	coupon          优惠券管理
//	pppoe_proxy     代理账号/在线账号
//	notice_temp     通知模板
//	notice_cycle    周期通知
//	notice_remind   到期提醒
//	notice_expires  到期通知
//
// 开发状态：已实现（认证计费 大类 16 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package accounting
