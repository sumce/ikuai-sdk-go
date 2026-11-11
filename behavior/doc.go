// Package behavior 实现爱快 Web 控制台「行为管控」大类，共 17 个功能：
//
//	audit              行为审计设置
//	record_free_setting 审计豁免设置
//	audit_url_log      URL 审计日志
//	audit_im_log       IM 审计日志
//	audit_terminal_log 终端审计日志
//	audit_terminal_stat 终端审计统计
//	mac_comment        MAC 备注
//	acl_mac            MAC 访问控制
//	url_black          网页黑名单
//	domain_blacklist   域名黑名单
//	domain_group       域名分组
//	url_redirect       网页重定向
//	url_keywords       网页关键字
//	url_replace        网页替换
//	acl_l7             七层协议控制
//	acl_l2route        二层访问控制
//	acl_qq             QQ 准入控制
//
// 开发状态：已实现（行为管控 大类 17 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package behavior
