// Package monitoring 实现爱快 Web 控制台「状态监控」大类
// （system-setting → 系统状态），共 10 个功能：
//
//	monitor_iface     接口监控
//	monitor_lanip     内网监控 IPv4
//	monitor_lanipv6   内网监控 IPv6
//	monitor_app_flow  协议监控（应用流量历史）
//	monitor_l7qos     策略监控（L7 流控策略状态）
//	monitor_system    系统监控（CPU/内存/磁盘/终端/连接数/温度曲线）
//	cflow             流量监控（连接统计汇总）
//	camera            摄像头监控
//	cloud_switch      交换机监控
//	dev_control       智能设备管控
//
// 开发状态：已实现（状态监控 大类 10 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package monitoring
