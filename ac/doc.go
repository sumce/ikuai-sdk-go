// Package ac 实现爱快 Web 控制台「AC管理」（无线控制器）大类，
// 共 8 个功能：
//
//	ac_status        AP 统计
//	ac_server        AP 配置（AC 服务）
//	ac_group         AP 分组
//	ac_upgrade       AP 升级
//	wls_black        无线黑名单
//	wls_mvlan        无线终端 VLAN
//	ac_online_clt    在线用户
//	ac_net_optimize  无线优化
//
// 开发状态：已实现（AC管理 大类 8 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package ac
