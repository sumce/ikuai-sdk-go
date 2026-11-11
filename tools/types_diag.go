package tools

// 诊断类 func（Ping/Traceroute/iperf）共用类型与枚举值。三个功能的
// show 响应均为「任务配置字段 + data:[{status,response}]」的异步任务
// 形态（真机 3.7.24 实测），行结构一致，归并为 DiagResult。

// 诊断类 func 共用的协议栈（proto/l3proto 字段）与传输层协议
// （l4proto 字段）枚举值：协议栈 ipv4/ipv6 与 icmp/tcp/udp 为真机
// show 实测值（iperf 默认 l4proto=udp、Ping 默认 l4proto=icmp）；
// udp/tcp 亦用于 Ping 的 l4proto（i18n pingtest.protocol_type
// 协议类型 + port 字段，组合语义待真机验证）。
const (
	// DiagStackIPv4 IPv4 协议栈（真机实测默认值）。
	DiagStackIPv4 = "ipv4"
	// DiagStackIPv6 IPv6 协议栈（按 i18n 协议栈选项推导）。
	DiagStackIPv6 = "ipv6"

	// DiagL4ICMP ICMP 协议（真机实测 Ping 默认值）。
	DiagL4ICMP = "icmp"
	// DiagL4TCP TCP 协议。
	DiagL4TCP = "tcp"
	// DiagL4UDP UDP 协议（真机实测 iperf 默认值）。
	DiagL4UDP = "udp"
)

// DiagResult 是诊断类任务结果行（Ping/Traceroute/iperf show 响应的
// data 数组元素，真机 3.7.24 实测形态 {"status":0,"response":"..."}）。
type DiagResult struct {
	// Status 任务状态（真机观测 0；其余取值语义待真机验证）。
	Status int64 `json:"status"`
	// Response 检测结果的原始文本输出（如 ping 命令的逐行输出）。
	Response string `json:"response"`
}
