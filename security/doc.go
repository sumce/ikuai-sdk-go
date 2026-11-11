// Package security 实现爱快 Web 控制台「安全设置」大类，共 5 个功能：
//
//	acl          访问控制（IP 规则）
//	arp          ARP 表
//	conn_limit   连接数限制
//	advanced     安全高级设置
//	firewall     防火墙
//
// 开发状态：已实现（安全设置 大类 5 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package security
