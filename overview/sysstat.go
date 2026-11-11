package overview

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// sysstat.show 的 TYPE 选择器常量（抓包实证）。
const (
	// SysStatTypeVerinfo 固件版本信息（verinfo）。
	SysStatTypeVerinfo = "verinfo"
	// SysStatTypeCPU CPU 占用率采样（cpu，百分比字符串数组）。
	SysStatTypeCPU = "cpu"
	// SysStatTypeMemory 内存信息（memory）。
	SysStatTypeMemory = "memory"
	// SysStatTypeStream 连接数与流量（stream）。
	SysStatTypeStream = "stream"
	// SysStatTypeCpuTemp CPU 温度采样（cputemp，摄氏度数组）。
	SysStatTypeCpuTemp = "cputemp"
	// SysStatTypeFreq CPU 频率采样（freq，字符串数值数组）。
	SysStatTypeFreq = "freq"
)

// SysStatTypeRealtime 是控制台实时面板请求的 TYPE 组合
// （verinfo,cpu,memory,stream,cputemp，抓包实证），SysStatService.
// Realtime 的默认值。
const SysStatTypeRealtime = SysStatTypeVerinfo + "," + SysStatTypeCPU + "," +
	SysStatTypeMemory + "," + SysStatTypeStream + "," + SysStatTypeCpuTemp

// SysStatRealtime 是系统实时状态的聚合结果（sysstat.show）。各区块仅在
// 对应 TYPE 选择器被请求时出现（未请求的区块为零值 nil）。
type SysStatRealtime struct {
	// CPU CPU 占用率采样列表（TYPE 含 cpu；百分比字符串）。
	CPU []string `json:"cpu"`
	// CPUTemp CPU 温度采样列表（TYPE 含 cputemp；摄氏度）。
	CPUTemp []int64 `json:"cputemp"`
	// Freq CPU 频率采样列表（TYPE 含 freq；字符串数值）。
	Freq []string `json:"freq"`
	// Memory 内存信息（TYPE 含 memory）。
	Memory *MemoryInfo `json:"memory"`
	// Stream 连接数与流量（TYPE 含 stream）。
	Stream *StreamInfo `json:"stream"`
	// VerInfo 固件版本信息（TYPE 含 verinfo）。
	VerInfo *VerInfo `json:"verinfo"`
}

// CPUFreq 是 CPU 占用率与频率采样的聚合结果（sysstat.show
// TYPE=cpu,freq）。区块仅在对应 TYPE 被请求时出现（未请求时为零值 nil）。
type CPUFreq struct {
	// CPU CPU 占用率采样列表（百分比字符串，如 "10.00%"；观测 21 个）。
	CPU []string `json:"cpu"`
	// Freq CPU 频率采样列表（字符串数值；观测 20 个）。
	Freq []string `json:"freq"`
}

// SysStatService 是系统实时状态（sysstat）的子服务。
type SysStatService struct {
	c *ikuai.Client
}

// Realtime 获取系统实时状态（CPU/内存/连接/流量/版本信息）。
// 对应 /Action/call: sysstat.show（TYPE=verinfo,cpu,memory,stream,cputemp），
// 即 Web 控制台轮询的系统实时数据。
// opts 为 nil 时请求默认组合；单区块可用 opts.Type 指定（SysStatType*
// 常量逗号拼接），其余键可用 opts.Extra 透传。
func (s *SysStatService) Realtime(ctx context.Context, opts *ikuai.ListOptions) (*SysStatRealtime, error) {
	param := map[string]any{}
	if opts != nil {
		param = opts.BuildParam()
		if param == nil {
			param = map[string]any{}
		}
	}
	if _, ok := param["TYPE"]; !ok {
		param["TYPE"] = SysStatTypeRealtime
	}
	var out SysStatRealtime
	if err := s.c.CallData(ctx, "sysstat", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerInfo 获取固件版本信息。
// 对应 /Action/call: sysstat.show（TYPE=verinfo）。
func (s *SysStatService) VerInfo(ctx context.Context) (*VerInfo, error) {
	var out struct {
		VerInfo *VerInfo `json:"verinfo"`
	}
	if err := s.c.CallData(ctx, "sysstat", "show",
		map[string]any{"TYPE": SysStatTypeVerinfo}, &out); err != nil {
		return nil, err
	}
	return out.VerInfo, nil
}

// CPUFreq 获取 CPU 占用率与频率采样。
// 对应 /Action/call: sysstat.show（TYPE=cpu,freq）。
func (s *SysStatService) CPUFreq(ctx context.Context) (*CPUFreq, error) {
	var out CPUFreq
	if err := s.c.CallData(ctx, "sysstat", "show",
		map[string]any{"TYPE": SysStatTypeCPU + "," + SysStatTypeFreq}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
