// Package tools 实现爱快 Web 控制台「应用工具」大类，共 8 个功能：
//
//	Ping        在线 Ping
//	Traceroute  路由追踪
//	tcpdump     抓包工具
//	iperf       内网测速
//	subnet      子网计算
//	speedtest   外网测速
//	iksyscheck  系统诊断
//	watchdog    看门狗
//
// 开发状态：已实现（应用工具 大类 8 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package tools
