package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// irqbalance.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
// 注意本 func 对空参数与其余 TYPE 值均回 30006 unknown TYPE (...)。
const (
	// IRQBalanceTypeHardIRQ 硬中断分布区块（分页 limit 被固件接受）。
	IRQBalanceTypeHardIRQ = "hardirq"
	// IRQBalanceTypeSoftIRQ 软中断统计区块（分页 limit 被固件接受）。
	IRQBalanceTypeSoftIRQ = "softirq"
	// IRQBalanceTypeCPUList CPU 列表与调频策略区块（cpulist,freqlist）。
	IRQBalanceTypeCPUList = "cpulist,freqlist"
)

// HardIRQEntry 是一条硬中断分布行（irqbalance.show TYPE=hardirq 的
// hardirq_data 数组元素）。
type HardIRQEntry struct {
	// Cpuid 中断绑定的 CPU（如 "cpu0"，"auto" 表示自动分配）。
	Cpuid string `json:"cpuid"`
	// IrqNum 硬中断号（固件以字符串表示，原样保留）。
	IrqNum string `json:"irq_num"`
	// Ifname 关联的网卡队列名（如 "ice-eth0-TxRx-0"）。
	Ifname string `json:"ifname"`
	// SetCpuid 用户指定的绑定 CPU（"auto" 或 "cpuN"；语义待真机验证）。
	SetCpuid string `json:"set_cpuid"`
}

// SoftIRQEntry 是一条软中断/负载统计行（irqbalance.show TYPE=softirq 的
// softirq_data 数组元素）。
type SoftIRQEntry struct {
	// Cpuid CPU 编号（固件以字符串表示，如 "0"）。
	Cpuid string `json:"cpuid"`
	// Freq 当前 CPU 频率（固件以字符串表示，单位 MHz，待真机验证）。
	Freq string `json:"freq"`
	// Phyid 物理核 ID（固件以字符串表示）。
	Phyid string `json:"phyid"`
	// Coreid 超线程逻辑核 ID（固件以字符串表示）。
	Coreid string `json:"coreid"`
	// Used CPU 占用率百分比文本（如 "10.00%"）。
	Used string `json:"used"`
	// Softirq 该核允许软中断（1=是）。
	Softirq int64 `json:"softirq"`
	// Hardirq 该核允许硬中断（1=是）。
	Hardirq int64 `json:"hardirq"`
}

// CPUListInfo 是 CPU 列表与调频策略块（irqbalance.show
// TYPE=cpulist,freqlist）。
type CPUListInfo struct {
	// Cpulist 可选的中断绑定 CPU 列表（首项固定 "auto"，其后为 cpu0...）。
	Cpulist []string `json:"cpulist"`
	// CpufreqSupport 是否支持 CPU 调频设置（1=支持）。
	CpufreqSupport int64 `json:"cpufreq_support"`
	// CurrentCpufreq 当前调频策略（观测值 "ondemand"）。
	CurrentCpufreq string `json:"current_cpufreq"`
	// CurrentTurbo 当前睿频（Turbo）开关（0=关闭）。
	CurrentTurbo int64 `json:"current_turbo"`
	// CpufreqList 支持的调频策略列表（观测值 conservative/ondemand/
	// powersave/performance/schedutil）。
	CpufreqList []string `json:"cpufreq_list"`
	// TurboSupport 是否支持睿频开关（1=支持）。
	TurboSupport int64 `json:"turbo_support"`
}

// IRQBalanceService 是负载监控（irqbalance）的子服务。
type IRQBalanceService struct {
	c *ikuai.Client
}

// HardIRQ 获取硬中断分布（网卡队列中断与 CPU 绑定关系）。
// 对应 /Action/call: irqbalance.show（TYPE=hardirq, limit="offset,rows",
// ORDER_BY, ORDER；burp 抓包实证），
// 即 Web 控制台「高级设置 → 负载监控」页面的硬中断表。
// opts 为 nil 时仅发送 {"TYPE":"hardirq"}，分页/排序参数经 opts 指定
// （固件接受 limit 但实测不生效，数据全量返回；本 func 不支持
// TYPE=total，空参数与其余 TYPE 值回 30006 unknown TYPE）。
func (s *IRQBalanceService) HardIRQ(ctx context.Context, opts *ikuai.ListOptions) ([]HardIRQEntry, error) {
	param := typedShowParam(opts, IRQBalanceTypeHardIRQ, nil)
	var out struct {
		HardIRQData []HardIRQEntry `json:"hardirq_data"`
	}
	if err := s.c.CallData(ctx, "irqbalance", "show", param, &out); err != nil {
		return nil, err
	}
	return out.HardIRQData, nil
}

// SoftIRQ 获取各 CPU 的软中断与负载统计。
// 对应 /Action/call: irqbalance.show（TYPE=softirq, limit="offset,rows",
// ORDER_BY, ORDER；burp 抓包实证）。
// opts 的处理方式与 HardIRQ 相同。
func (s *IRQBalanceService) SoftIRQ(ctx context.Context, opts *ikuai.ListOptions) ([]SoftIRQEntry, error) {
	param := typedShowParam(opts, IRQBalanceTypeSoftIRQ, nil)
	var out struct {
		SoftIRQData []SoftIRQEntry `json:"softirq_data"`
	}
	if err := s.c.CallData(ctx, "irqbalance", "show", param, &out); err != nil {
		return nil, err
	}
	return out.SoftIRQData, nil
}

// CPUList 获取可选 CPU 列表与调频策略（调频策略/睿频的候选值与当前值）。
// 对应 /Action/call: irqbalance.show（TYPE=cpulist,freqlist；burp 抓包实证）。
func (s *IRQBalanceService) CPUList(ctx context.Context) (*CPUListInfo, error) {
	var out CPUListInfo
	if err := s.c.CallData(ctx, "irqbalance", "show",
		map[string]any{"TYPE": IRQBalanceTypeCPUList}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
