package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// tcpdump 抓包协议类型枚举值（i18n capturePacket.help.tip1「支持tcp、
// udp、icmp、arp等协议类型」推导；真机 show 观测为空串=全部协议，
// 具体取值待真机验证）。
const (
	// TCPDumpProtoTCP TCP 协议。
	TCPDumpProtoTCP = "tcp"
	// TCPDumpProtoUDP UDP 协议。
	TCPDumpProtoUDP = "udp"
	// TCPDumpProtoICMP ICMP 协议。
	TCPDumpProtoICMP = "icmp"
	// TCPDumpProtoARP ARP 协议。
	TCPDumpProtoARP = "arp"
)

// TCPDumpRequest 是一次抓包任务的过滤条件（tcpdump.show 状态行的
// 配置字段即本结构；字段集为真机 3.7.24 实测）。
//
// 过滤字段留空匹配全部（i18n capturePacket.help.tip2/tip3/tip4）。
type TCPDumpRequest struct {
	// Interface 抓包接口（i18n capturePacket.captureInterface 选择抓包接口）。
	Interface string `json:"interface"`
	// IPAddr IP 过滤（i18n capturePacket.IP，留空匹配全部地址）。
	IPAddr string `json:"ip_addr"`
	// MAC MAC 过滤（i18n capturePacket.MAC，留空匹配全部）。
	MAC string `json:"mac"`
	// Port 端口过滤（i18n capturePacket.Port，留空匹配全部端口）。
	Port string `json:"port"`
	// Proto 协议类型（"tcp"/"udp"/"icmp"/"arp"，枚举 TCPDumpProto*；
	// 留空抓取全部数据包）。
	Proto string `json:"proto"`
	// HitNum 抓包数量（i18n capturePacket.packetNumber，i18n 提示范围
	// 1-80000，真机观测默认 100；show 状态行内为字符串形态）。
	HitNum int64 `json:"hit_num"`
	// Filename 抓包文件保存路径与名称（i18n capturePacket.fileName
	// 文件绝对路径，如 "/Document/001"）。
	Filename string `json:"filename"`
}

// TCPDumpState 是抓包任务的当前状态行（tcpdump.show 的 data 数组
// 元素，真机 3.7.24 实测单行；空闲时 filename 为空、size 为 "0"）。
type TCPDumpState struct {
	// Interface 抓包接口。
	Interface string `json:"interface"`
	// MAC MAC 过滤。
	MAC string `json:"mac"`
	// HitNum 抓包数量（固件以字符串返回，实测 "100"）。
	HitNum string `json:"hit_num"`
	// Proto 协议类型。
	Proto string `json:"proto"`
	// IPAddr IP 过滤。
	IPAddr string `json:"ip_addr"`
	// Port 端口过滤。
	Port string `json:"port"`
	// Filename 抓包文件绝对路径（进行中/完成后非空）。
	Filename string `json:"filename"`
	// Size 抓包文件大小（固件以字符串返回，实测 "0"；单位待真机验证）。
	Size string `json:"size"`
	// LastTime 文件时间（i18n capturePacket.fileTime；格式待真机验证）。
	LastTime string `json:"last_time"`
}

// tcpdumpState 是 tcpdump.show 的 Data 包裹形态。
type tcpdumpState struct {
	Data []TCPDumpState `json:"data"`
}

// TCPDumpService 是抓包工具（tcpdump）的子服务
// （tools.capture-packet「抓包工具」页面）。
type TCPDumpService struct {
	c *ikuai.Client
}

// Status 查询抓包任务的当前状态（只读）。
// 对应 /Action/call: tcpdump.show（空参数 {}，真机 3.7.24 实测返回
// 单行状态：过滤条件 + 抓包文件信息；空闲时 filename 为空、size
// 为 "0"），
// 即 Web 控制台「应用工具 → 抓包工具」页面。
//
// 无抓包任务时返回 nil（固件仍返回一行为空的占位行）。
func (s *TCPDumpService) Status(ctx context.Context) (*TCPDumpState, error) {
	var out tcpdumpState
	if err := s.c.CallData(ctx, "tcpdump", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	return &out.Data[0], nil
}

// Start 发起一次抓包（诊断动作）。
// 对应 /Action/call: tcpdump.show（param 为完整 TCPDumpRequest——固件
// 与 Ping/Traceroute 一致以 show 携带参数发起异步任务，真机 show {}
// 实测返回同一状态行形态；带参执行未对真机调用，action 名待真机
// 验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用：会在路由器上启动真实抓包并占用
// 内存/磁盘（存储位置为「系统内存」时第二次抓包覆盖旧文件，i18n
// tip5），仅离线测试覆盖。
func (s *TCPDumpService) Start(ctx context.Context, req TCPDumpRequest) (*TCPDumpState, error) {
	var out tcpdumpState
	if err := s.c.CallData(ctx, "tcpdump", "show", req, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	return &out.Data[0], nil
}

// Stop 停止进行中的抓包（诊断动作）。
// 对应 /Action/call: tcpdump.stop（action 名按 i18n
// capturePacket.StopPacket「停止抓包」按钮推断，未对真机调用，
// 待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用，仅离线测试覆盖。
func (s *TCPDumpService) Stop(ctx context.Context) error {
	_, err := s.c.Call(ctx, "tcpdump", "stop", map[string]any{})
	return err
}
