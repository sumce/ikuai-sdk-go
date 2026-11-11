package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// TracerouteRequest 是一次路由追踪的请求参数（Traceroute.show 响应的
// 任务配置字段即本结构；字段集为真机 3.7.24 实测）。
//
// 空闲时固件返回 max=0、wait=0；控制台常用默认值（如最大跳数 30）
// 未实证，由调用方显式指定。
type TracerouteRequest struct {
	// Host 目标 IP 地址或域名（i18n routetrace.IPorDomain）。
	Host string `json:"host"`
	// Interface 指定接口（i18n routetrace.Interface）。
	Interface string `json:"interface"`
	// Max 最大路由跳数（i18n MaxCount）。
	Max int64 `json:"max"`
	// Wait 超时等待时间，单位秒（i18n WaitTime；单位按语义推导，
	// 待真机验证）。
	Wait int64 `json:"wait"`
	// DNS 指定解析 DNS（真机观测默认空串，语义待真机验证）。
	DNS string `json:"dns"`
}

// TracerouteState 是路由追踪任务的当前状态（Traceroute.show 的 Data，
// 真机 3.7.24 实测：show {} 返回任务配置字段 + data 结果行，空闲时
// response 为空串）。
type TracerouteState struct {
	// Host 目标 IP 地址或域名。
	Host string `json:"host"`
	// Interface 指定接口。
	Interface string `json:"interface"`
	// Max 最大路由跳数。
	Max int64 `json:"max"`
	// Wait 超时等待时间。
	Wait int64 `json:"wait"`
	// DNS 指定解析 DNS。
	DNS string `json:"dns"`
	// Data 检测结果行（response 为 traceroute 命令的原始文本输出）。
	Data []DiagResult `json:"data"`
}

// TracerouteService 是路由追踪（Traceroute）的子服务。func_name 首字母
// 大写，以固件 menu.json 为准（tools.route-trace「路由追踪」页面）。
type TracerouteService struct {
	c *ikuai.Client
}

// Result 查询最近一次路由追踪的任务配置与结果（只读）。
// 对应 /Action/call: Traceroute.show（空参数 {}，真机 3.7.24 实测），
// 即 Web 控制台「应用工具 → 路由追踪」页面。
//
// ⚠️ 注意：本方法只发空参数 show（固件以 show 携带完整参数发起检测，
// 见 Execute），不会触发新的检测。
func (s *TracerouteService) Result(ctx context.Context) (*TracerouteState, error) {
	var out TracerouteState
	if err := s.c.CallData(ctx, "Traceroute", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Execute 发起一次路由追踪（诊断动作）。
// 对应 /Action/call: Traceroute.show（param 为完整 TracerouteRequest——
// 固件以 show 携带参数发起检测，与 Ping 的异步任务模式一致，真机
// show {} 实测返回同一形态；带参执行未对真机调用，action 名待真机
// 验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用：会向目标地址发出真实探测流量，
// 仅离线测试覆盖。
func (s *TracerouteService) Execute(ctx context.Context, req TracerouteRequest) (*TracerouteState, error) {
	var out TracerouteState
	if err := s.c.CallData(ctx, "Traceroute", "show", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stop 停止进行中的路由追踪（诊断动作）。
// 对应 /Action/call: Traceroute.stop_test（action 名按 i18n
// routetrace.StopDetection「停止检测」按钮推断，未对真机调用，
// 待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用，仅离线测试覆盖。
func (s *TracerouteService) Stop(ctx context.Context) error {
	_, err := s.c.Call(ctx, "Traceroute", "stop_test", map[string]any{})
	return err
}
