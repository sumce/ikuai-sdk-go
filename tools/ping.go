package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// PingRequest 是一次 Ping 检测的请求参数（Ping.show 响应的任务配置
// 字段即本结构；真机 3.7.24 实测字段集）。
//
// 控制台默认值参考（i18n pingtest）：源接口 auto（自动）、协议栈
// ipv4、协议类型 icmp；interface 取值 "auto" 的语义待真机验证。
type PingRequest struct {
	// Host 目标 IP 地址或域名（i18n pingtest.IPorDomain）。
	Host string `json:"host"`
	// Interface 指定源接口（i18n AssignSourceInterface；"auto"=自动）。
	Interface string `json:"interface"`
	// Proto 协议栈（"ipv4"/"ipv6"，枚举 DiagStack*）。
	Proto string `json:"proto"`
	// Count PING 包次数（i18n PingNum，真机观测默认 10）。
	Count int64 `json:"count"`
	// L4Proto 协议类型（"icmp"/"tcp"/"udp"，枚举 DiagL4*）。
	L4Proto string `json:"l4proto"`
	// Port 端口（l4proto 为 tcp/udp 时使用，1-65535；icmp 时为空串）。
	Port string `json:"port"`
	// DNS 指定解析 DNS（真机观测默认空串，语义待真机验证）。
	DNS string `json:"dns"`
}

// PingState 是 Ping 检测任务的当前状态（Ping.show 的 Data，真机
// 3.7.24 实测：show {} 返回最近一次检测的任务配置与结果）。
type PingState struct {
	// Host 目标 IP 地址或域名。
	Host string `json:"host"`
	// Interface 指定源接口。
	Interface string `json:"interface"`
	// Proto 协议栈。
	Proto string `json:"proto"`
	// Count PING 包次数。
	Count int64 `json:"count"`
	// L4Proto 协议类型。
	L4Proto string `json:"l4proto"`
	// Port 端口。
	Port string `json:"port"`
	// DNS 指定解析 DNS。
	DNS string `json:"dns"`
	// Data 检测结果行（response 为 ping 命令的原始文本输出）。
	Data []DiagResult `json:"data"`
}

// PingService 是在线 Ping（Ping）的子服务。func_name 首字母大写，
// 以固件 menu.json 为准（tools.ping-test「PING测试」页面）。
type PingService struct {
	c *ikuai.Client
}

// Result 查询最近一次 Ping 检测的任务配置与结果（只读）。
// 对应 /Action/call: Ping.show（空参数 {}，真机 3.7.24 实测返回
// 任务配置字段 + data 结果行），
// 即 Web 控制台「应用工具 → PING测试」页面。
//
// ⚠️ 注意：本方法只发空参数 show（固件以 show 携带完整参数发起检测，
// 见 Execute），不会触发新的检测。
func (s *PingService) Result(ctx context.Context) (*PingState, error) {
	var out PingState
	if err := s.c.CallData(ctx, "Ping", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Execute 发起一次 Ping 检测（诊断动作）。
// 对应 /Action/call: Ping.show（param 为完整 PingRequest——固件以
// show 携带参数发起检测并返回任务状态，与 Result 查询共用同一
// action，真机 show {} 实测返回同一形态；带参执行未对真机调用，
// action 名待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用：会向目标地址发出真实探测流量，
// 仅离线测试覆盖。
func (s *PingService) Execute(ctx context.Context, req PingRequest) (*PingState, error) {
	var out PingState
	if err := s.c.CallData(ctx, "Ping", "show", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stop 停止进行中的 Ping 检测（诊断动作）。
// 对应 /Action/call: Ping.stop_test（action 名按 i18n pingtest.stop_test
// 「停止检测」按钮推断，未对真机调用，待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用，仅离线测试覆盖。
func (s *PingService) Stop(ctx context.Context) error {
	_, err := s.c.Call(ctx, "Ping", "stop_test", map[string]any{})
	return err
}
