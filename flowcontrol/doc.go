// Package flowcontrol 实现爱快 Web 控制台「流控分流」大类，共 10 个功能：
//
//	lb_pcc          多线负载
//	stream_layer7   协议分流
//	stream_ipport   端口分流
//	stream_domain   域名分流
//	stream_updown   流量上下线分流
//	stream_control  智能流控
//	simple_qos      IP 限速
//	mac_qos         MAC 限速
//	dprotos         自定义协议
//	dprotos_l7      高级自定义协议
//
// 开发状态：已实现（流控分流 大类 10 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package flowcontrol
