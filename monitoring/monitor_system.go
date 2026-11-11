package monitoring

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sumce/ikuai-sdk-go"
)

// monitor_system.show 的 TYPE 选择器常量（抓包实证）。
const (
	// SystemTypeRateStat 峰值速率统计曲线（rate_stat）。
	SystemTypeRateStat = "rate_stat"
	// SystemTypeCPU CPU 占用率曲线（cpu）。
	SystemTypeCPU = "cpu"
	// SystemTypeMemory 内存曲线（memory）。
	SystemTypeMemory = "memory"
	// SystemTypeDiskSpaceUsed 磁盘空间使用率曲线（disk_space_used）。
	SystemTypeDiskSpaceUsed = "disk_space_used"
	// SystemTypeOnTerminal 在线终端数曲线（on_terminal）。
	SystemTypeOnTerminal = "on_terminal"
	// SystemTypeConnNum 连接数曲线（conn_num）。
	SystemTypeConnNum = "conn_num"
	// SystemTypeCpuTemp1 CPU 温度曲线 1（cputemp1）。
	SystemTypeCpuTemp1 = "cputemp1"
	// SystemTypeCpuTemp2 CPU 温度曲线 2（cputemp2）。
	SystemTypeCpuTemp2 = "cputemp2"
	// SystemTypeCpuTempSupport CPU 温度支持标志（cputemp_support）。
	SystemTypeCpuTempSupport = "cputemp_support"
	// SystemTypeAppFlow 首页应用流量快照（app_flow，按 minute 取最近区间）。
	SystemTypeAppFlow = "app_flow"
)

// SystemTypeAllCurves 是控制台「系统监控」页面的全部曲线 TYPE 组合
// （cpu,memory,disk_space_used,on_terminal,conn_num,cputemp1,cputemp2，
// 抓包实证）。
const SystemTypeAllCurves = SystemTypeCPU + "," + SystemTypeMemory + "," +
	SystemTypeDiskSpaceUsed + "," + SystemTypeOnTerminal + "," +
	SystemTypeConnNum + "," + SystemTypeCpuTemp1 + "," + SystemTypeCpuTemp2

// monitor_system.show 的 datetype / math 参数常量。
const (
	// SystemDatetypeHour 按小时聚合（抓包实证）。
	SystemDatetypeHour = "hour"
	// SystemDatetypeDay 按天聚合（真机验证可用）。
	SystemDatetypeDay = "day"
	// SystemMathAvg 取均值（抓包实证）。
	SystemMathAvg = "avg"
	// SystemMathMax 取峰值（真机验证可用）。
	SystemMathMax = "max"
)

// SystemRateStatPoint 是峰值速率统计的一个时点行（rate_stat）。
type SystemRateStatPoint struct {
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
	// MaxUpload 区间上传峰值（字节/秒）。
	MaxUpload int64 `json:"max_upload"`
	// MaxDownload 区间下载峰值（字节/秒）。
	MaxDownload int64 `json:"max_download"`
	// RxPackages 区间接收包数。
	RxPackages int64 `json:"rx_packages"`
	// TxPackages 区间发送包数。
	TxPackages int64 `json:"tx_packages"`
}

// SystemCPUPoint 是 CPU 占用率曲线的一个时点行（cpu）。
// CPU 值为百分比；固件在 math=max 等聚合下返回小数（真机实证），
// 故采用 float64。
type SystemCPUPoint struct {
	// CPU CPU 占用率（百分比，可能带小数）。
	CPU float64 `json:"cpu"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
}

// SystemMemoryPoint 是内存曲线的一个时点行（memory）。
type SystemMemoryPoint struct {
	// Memory 已用内存（字节）。
	Memory int64 `json:"memory"`
	// MemoryUse 内存使用率（百分比整数）。
	MemoryUse int64 `json:"memory_use"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
}

// SystemDiskPoint 是磁盘空间使用率曲线的一个时点行（disk_space_used）。
type SystemDiskPoint struct {
	// DiskSpaceUsed 磁盘空间使用率（百分比整数）。
	DiskSpaceUsed int64 `json:"disk_space_used"`
	// DiskSpaceUse 磁盘空间使用增量（抓包中为 0；字段语义待真机验证）。
	DiskSpaceUse int64 `json:"disk_space_use"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
}

// SystemTerminalPoint 是在线终端数曲线的一个时点行（on_terminal）。
type SystemTerminalPoint struct {
	// OnTerminal 在线终端数。
	OnTerminal int64 `json:"on_terminal"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
}

// SystemConnPoint 是连接数曲线的一个时点行（conn_num）。
type SystemConnPoint struct {
	// ConnNum 当前连接数。
	ConnNum int64 `json:"conn_num"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
}

// SystemCPUTempPoint 是 CPU 温度曲线的一个时点行（cputemp1/cputemp2），
// 温度单位为摄氏度。
type SystemCPUTempPoint struct {
	// CpuTemp1 CPU 1 温度（cputemp1；单路 CPU 为 0）。
	CpuTemp1 int64 `json:"cputemp1"`
	// CpuTemp2 CPU 2 温度（cputemp2）。
	CpuTemp2 int64 `json:"cputemp2"`
	// Timestamp 时点（Unix 时间戳，秒）。
	Timestamp int64 `json:"timestamp"`
}

// SystemAppFlow 是应用流量快照行（app_flow，单行），值为最近 minute
// 分钟区间的各分类流量（字节）。
type SystemAppFlow struct {
	// Total 全部应用流量合计。
	Total int64 `json:"Total"`
	// HTTP HTTP 应用流量。
	HTTP int64 `json:"HTTP"`
	// Download 下载类应用流量。
	Download int64 `json:"Download"`
	// Transport 传输类应用流量。
	Transport int64 `json:"Transport"`
	// IM 即时通讯类应用流量。
	IM int64 `json:"IM"`
	// Video 视频类应用流量。
	Video int64 `json:"Video"`
	// Common 常用应用流量。
	Common int64 `json:"Common"`
	// Game 游戏类应用流量。
	Game int64 `json:"Game"`
	// Others 其他应用流量。
	Others int64 `json:"Others"`
	// Test 测试类应用流量。
	Test int64 `json:"Test"`
	// Unknown 未识别应用流量。
	Unknown int64 `json:"Unknown"`
}

// SystemMonitor 是系统监控曲线的聚合结果，各区块仅在对应 TYPE 选择器
// 被请求时出现（未请求的区块为零值 nil）。
type SystemMonitor struct {
	// CPU CPU 占用率曲线（TYPE 含 cpu）。
	CPU []SystemCPUPoint `json:"cpu"`
	// Memory 内存曲线（TYPE 含 memory）。
	Memory []SystemMemoryPoint `json:"memory"`
	// DiskSpaceUsed 磁盘使用率曲线（TYPE 含 disk_space_used）。
	DiskSpaceUsed []SystemDiskPoint `json:"disk_space_used"`
	// OnTerminal 在线终端数曲线（TYPE 含 on_terminal）。
	OnTerminal []SystemTerminalPoint `json:"on_terminal"`
	// ConnNum 连接数曲线（TYPE 含 conn_num）。
	ConnNum []SystemConnPoint `json:"conn_num"`
	// CpuTemp1 CPU 温度曲线 1（TYPE 含 cputemp1）。
	CpuTemp1 []SystemCPUTempPoint `json:"cputemp1"`
	// CpuTemp2 CPU 温度曲线 2（TYPE 含 cputemp2）。
	CpuTemp2 []SystemCPUTempPoint `json:"cputemp2"`
}

// SystemRateStat 是峰值速率统计曲线结果（rate_stat）。
type SystemRateStat struct {
	// RateStat 峰值速率时点行列表。
	RateStat []SystemRateStatPoint `json:"rate_stat"`
}

// SystemService 是负载监控（monitor_system）的子服务。
type SystemService struct {
	c *ikuai.Client
}

// Curves 获取系统负载监控曲线（CPU/内存/磁盘/在线终端/连接数/CPU 温度）。
// 对应 /Action/call: monitor_system.show（TYPE=cpu,memory,disk_space_used,
// on_terminal,conn_num,cputemp1,cputemp2, datetype=hour, math=avg），
// 即 Web 控制台「状态监控 → 系统监控」页面。
// opts 为 nil 时请求全部曲线；datetype/math 可用 opts.Extra 覆盖
// （值常量见 SystemDatetype* / SystemMath*），单条曲线可用 opts.Type 指定。
func (s *SystemService) Curves(ctx context.Context, opts *ikuai.ListOptions) (*SystemMonitor, error) {
	param := selectorShowParam(opts, map[string]any{
		"TYPE":     SystemTypeAllCurves,
		"datetype": SystemDatetypeHour,
		"math":     SystemMathAvg,
	})
	var out SystemMonitor
	if err := s.c.CallData(ctx, "monitor_system", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RateStat 获取峰值速率统计曲线（上传/下载峰值与收发包数）。
// 对应 /Action/call: monitor_system.show（TYPE=rate_stat, datetype=hour, math=avg）。
// opts 为 nil 时使用抓包默认聚合参数。
func (s *SystemService) RateStat(ctx context.Context, opts *ikuai.ListOptions) (*SystemRateStat, error) {
	param := selectorShowParam(opts, map[string]any{
		"TYPE":     SystemTypeRateStat,
		"datetype": SystemDatetypeHour,
		"math":     SystemMathAvg,
	})
	var out SystemRateStat
	if err := s.c.CallData(ctx, "monitor_system", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CpuTempSupport 查询设备是否支持 CPU 温度监控（cputemp_support，
// 1=支持，0=不支持）。
// 对应 /Action/call: monitor_system.show（TYPE=cputemp_support）。
func (s *SystemService) CpuTempSupport(ctx context.Context) (int64, error) {
	var out struct {
		CpuTempSupport int64 `json:"cputemp_support"`
	}
	if err := s.c.CallData(ctx, "monitor_system", "show",
		map[string]any{"TYPE": SystemTypeCpuTempSupport}, &out); err != nil {
		return 0, err
	}
	return out.CpuTempSupport, nil
}

// AppFlow 获取最近 minutes 分钟区间的应用流量分类快照（app_flow）。
// 对应 /Action/call: monitor_system.show（TYPE=app_flow, minute="30"）；
// 真机抓包出现的分钟值：30 / 60 / 1440。
// opts 仅用于透传分页外的其他键（一般传 nil）。
func (s *SystemService) AppFlow(ctx context.Context, minutes int64, opts *ikuai.ListOptions) ([]SystemAppFlow, error) {
	if minutes <= 0 {
		return nil, fmt.Errorf("monitoring: minutes 必须为正数: %w", ikuai.ErrInvalidParam)
	}
	param := selectorShowParam(opts, map[string]any{
		"TYPE":   SystemTypeAppFlow,
		"minute": strconv.FormatInt(minutes, 10), // 抓包以字符串形式传 minute。
	})
	var out struct {
		AppFlow []SystemAppFlow `json:"app_flow"`
	}
	if err := s.c.CallData(ctx, "monitor_system", "show", param, &out); err != nil {
		return nil, err
	}
	return out.AppFlow, nil
}
