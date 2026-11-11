package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// iperf 测速方向枚举值（dir 字段；upload 为真机实测默认值，
// download 按 i18n throughput_test.dirUp/dirDown 推导）。
const (
	// IPerfDirUpload 上行（真机实测默认值）。
	IPerfDirUpload = "upload"
	// IPerfDirDownload 下行（按 i18n 选项推导）。
	IPerfDirDownload = "download"
)

// IPerfRequest 是一次内网测速（iperf 客户端）的请求参数
// （iperf.show 响应的任务配置字段即本结构；字段集为真机 3.7.24 实测）。
//
// 控制台默认值参考（真机实测 show {}）：port=5201、threads=1、
// dir=upload、interface=auto、l3proto=ipv4、l4proto=udp、time=10、
// interval=1、bandwidth="100"。
type IPerfRequest struct {
	// Host 服务端 IP 地址或域名（i18n throughput_test.ip；作为服务端
	// 使用时为空）。
	Host string `json:"host"`
	// Port 服务端口（i18n servePort，iperf 默认 5201）。
	Port int64 `json:"port"`
	// Threads 线程数量（i18n throughput_test.number）。
	Threads int64 `json:"threads"`
	// Dir 测速方向（"upload"/"download"，枚举 IPerfDir*）。
	Dir string `json:"dir"`
	// Interface 本端接口（i18n throughput_test.interface，"auto"=自动）。
	Interface string `json:"interface"`
	// L3Proto 协议栈（"ipv4"/"ipv6"，枚举 DiagStack*）。
	L3Proto string `json:"l3proto"`
	// L4Proto 协议类型（"tcp"/"udp"，枚举 DiagL4*）。
	L4Proto string `json:"l4proto"`
	// Time 测试周期，单位秒（i18n throughput_test.period，真机观测 10）。
	Time int64 `json:"time"`
	// Interval 间隔时间（打印周期），单位秒（i18n throughput_test.intervals，
	// 真机观测 1）。
	Interval int64 `json:"interval"`
	// Bandwidth 发送带宽（i18n throughput_test.bandwidth；固件以字符串
	// 返回，实测 "100"；单位待真机验证）。
	Bandwidth string `json:"bandwidth"`
}

// IPerfState 是内网测速任务的当前状态（iperf.show 的 Data，真机
// 3.7.24 实测：show {} 返回任务配置字段 + data 结果行，空闲时
// response 为空串）。
type IPerfState struct {
	// Host 服务端 IP 地址或域名。
	Host string `json:"host"`
	// Port 服务端口。
	Port int64 `json:"port"`
	// Threads 线程数量。
	Threads int64 `json:"threads"`
	// Dir 测速方向。
	Dir string `json:"dir"`
	// Interface 本端接口。
	Interface string `json:"interface"`
	// L3Proto 协议栈。
	L3Proto string `json:"l3proto"`
	// L4Proto 协议类型。
	L4Proto string `json:"l4proto"`
	// Time 测试周期，单位秒。
	Time int64 `json:"time"`
	// Interval 间隔时间，单位秒。
	Interval int64 `json:"interval"`
	// Bandwidth 发送带宽（字符串形态）。
	Bandwidth string `json:"bandwidth"`
	// Data 检测结果行（response 为 iperf 的原始文本输出）。
	Data []DiagResult `json:"data"`
}

// IPerfService 是内网测速（iperf）的子服务
// （tools.throughput-test「吞吐测试」页面）。
//
// 页面另有 iperf 服务端开关（i18n throughput_test.func_switch/
// server，说明文案「服务端是否启用，不影响客户端使用」），其开关
// 状态未在 show 响应中出现，未提供封装（待真机验证；需要时经
// Service.Raw 显式调用）。
type IPerfService struct {
	c *ikuai.Client
}

// Result 查询最近一次内网测速的任务配置与结果（只读）。
// 对应 /Action/call: iperf.show（空参数 {}，真机 3.7.24 实测），
// 即 Web 控制台「应用工具 → 吞吐测试」页面。
//
// ⚠️ 注意：本方法只发空参数 show（固件以 show 携带完整参数发起检测，
// 见 Execute），不会触发新的检测。
func (s *IPerfService) Result(ctx context.Context) (*IPerfState, error) {
	var out IPerfState
	if err := s.c.CallData(ctx, "iperf", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Execute 发起一次内网测速（诊断动作）。
// 对应 /Action/call: iperf.show（param 为完整 IPerfRequest——固件以
// show 携带参数发起检测，与 Ping 的异步任务模式一致，真机 show {}
// 实测返回同一形态；带参执行未对真机调用，action 名待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用：会向服务端发出真实测速流量，
// 仅离线测试覆盖。
func (s *IPerfService) Execute(ctx context.Context, req IPerfRequest) (*IPerfState, error) {
	var out IPerfState
	if err := s.c.CallData(ctx, "iperf", "show", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Stop 停止进行中的内网测速（诊断动作）。
// 对应 /Action/call: iperf.stop（action 名按 i18n
// throughput_test.stop「停止测试」按钮推断，未对真机调用，待真机
// 验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用，仅离线测试覆盖。
func (s *IPerfService) Stop(ctx context.Context) error {
	_, err := s.c.Call(ctx, "iperf", "stop", map[string]any{})
	return err
}
