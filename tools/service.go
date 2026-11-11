package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是应用工具大类的服务入口，聚合 Web 控制台「应用工具」菜单
// 下的全部 8 个功能。
type Service struct {
	c *ikuai.Client

	// Ping 在线 Ping（Ping）：PING测试（诊断执行为写形态动作）。
	Ping *PingService
	// Traceroute 路由追踪（Traceroute）：路由追踪（诊断执行为写形态动作）。
	Traceroute *TracerouteService
	// TCPDump 抓包工具（tcpdump）：抓包任务状态查询与启停。
	TCPDump *TCPDumpService
	// IPerf 内网测速（iperf）：吞吐测试（诊断执行为写形态动作）。
	IPerf *IPerfService
	// Subnet 子网计算（subnet）：固件无服务端接口，本地纯计算。
	Subnet *SubnetService
	// SpeedTest 外网测速（speedtest）：线路测速状态与启停。
	SpeedTest *SpeedTestService
	// SysCheck 系统诊断（iksyscheck）：路由体检结果与启停。
	SysCheck *SysCheckService
	// Watchdog 看门狗（watchdog）：健康检测配置读写。
	Watchdog *WatchdogService
}

// New 构造应用工具服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:          c,
		Ping:       &PingService{c: c},
		Traceroute: &TracerouteService{c: c},
		TCPDump:    &TCPDumpService{c: c},
		IPerf:      &IPerfService{c: c},
		Subnet:     &SubnetService{c: c},
		SpeedTest:  &SpeedTestService{c: c},
		SysCheck:   &SysCheckService{c: c},
		Watchdog:   &WatchdogService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用应用工具大类接口，作为未封装
// action/param 组合的逃生口（每个大类 Service 仅此一个，见
// docs/RULES.md §6）。
//
// ⚠️ 本包多数 func 的诊断执行动作（Ping/Traceroute/tcpdump/iperf 的
// 带参 show、speedtest/iksyscheck 的启停）会产生真实副作用，经 Raw
// 调用未封装动作前务必确认后果；真机写动作一律禁止
// （见 docs/RULES.md §9.2 与本包各诊断方法注释）。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}
