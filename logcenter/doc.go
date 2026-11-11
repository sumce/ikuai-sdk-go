// Package logcenter 实现爱快 Web 控制台「日志中心」大类，共 11 个功能：
//
//	syslog-pppauth   PPP 拨号日志
//	syslog-arp       ARP 日志
//	syslog-apaction  AP 操作日志
//	syslog-dhcpd     DHCP 日志
//	syslog-ddns      DDNS 日志
//	syslog-wanpppoe  WAN PPPoE 日志
//	syslog-notice    通知日志
//	syslog-sysevent  系统事件
//	syslog-webadmin  管理日志
//	ikmessages       系统消息（站内信，含 start）
//	rt_collect       实时采集配置
//
// 开发状态：已实现（日志中心大类 11 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package logcenter
