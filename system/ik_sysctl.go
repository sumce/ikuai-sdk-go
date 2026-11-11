package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// ik_sysctl.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
const (
	// SysctlTypeData 配置行区块选择器。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）。
	SysctlTypeData = "data"
)

// SysctlConfig 是内核参数的单行配置（ik_sysctl.show data 数组元素，
// 观测恒为单行 id=1）。各 timeout 字段单位为秒。
type SysctlConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// BBR TCP BBR 拥塞控制开关（0=关闭）。
	BBR int64 `json:"bbr"`
	// SynRecvTimeout SYN_RECV 状态超时（秒，观测值 5）。
	SynRecvTimeout int64 `json:"syn_recv_timeout"`
	// SynSendTimeout SYN_SEND 状态超时（秒，观测值 5）。
	SynSendTimeout int64 `json:"syn_send_timeout"`
	// EstablishedTimeout ESTABLISHED 状态超时（秒，观测值 1800）。
	EstablishedTimeout int64 `json:"established_timeout"`
	// FinWaitTimeout FIN_WAIT 状态超时（秒，观测值 10）。
	FinWaitTimeout int64 `json:"fin_wait_timeout"`
	// LastAckTimeout LAST_ACK 状态超时（秒，观测值 10）。
	LastAckTimeout int64 `json:"last_ack_timeout"`
	// CloseWaitTimeout CLOSE_WAIT 状态超时（秒，观测值 10）。
	CloseWaitTimeout int64 `json:"close_wait_timeout"`
	// TimeWaitTimeout TIME_WAIT 状态超时（秒，观测值 10）。
	TimeWaitTimeout int64 `json:"time_wait_timeout"`
	// CloseTimeout CLOSE 状态超时（秒，观测值 5）。
	CloseTimeout int64 `json:"close_timeout"`
	// UDPTimeout UDP 会话超时（秒，观测值 10）。
	UDPTimeout int64 `json:"udp_timeout"`
	// UDPStreamTimeout UDP 流会话超时（秒，观测值 60）。
	UDPStreamTimeout int64 `json:"udp_stream_timeout"`
	// ICMPTimeout ICMP 会话超时（秒，观测值 10）。
	ICMPTimeout int64 `json:"icmp_timeout"`
}

// SysctlService 是内核参数（ik_sysctl）的子服务。
type SysctlService struct {
	c *ikuai.Client
}

// Get 获取内核参数配置（TCP/UDP/ICMP 会话超时与 BBR）。
// 对应 /Action/call: ik_sysctl.show（TYPE=data），
// 即 Web 控制台「高级设置 → 内核参数」页面（挂在系统设置大类下）。
func (s *SysctlService) Get(ctx context.Context) (*SysctlConfig, error) {
	page, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(page.Rows) == 0 {
		return nil, nil
	}
	return &page.Rows[0], nil
}

// List 以分页形态获取内核参数配置行（固件恒返回单行）。
// 对应 /Action/call: ik_sysctl.show（TYPE=data, limit="offset,rows"）。
// 本 func 不支持 TYPE=total（超出的 TYPE 值会被归一为 data）。
func (s *SysctlService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SysctlConfig], error) {
	param := typedShowParam(opts, SysctlTypeData, nil)
	resp, err := s.c.Call(ctx, "ik_sysctl", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[SysctlConfig](resp.Data)
}

// Save 保存内核参数配置（写动作）。
// 对应 /Action/call: ik_sysctl.save（action 名按同类单行配置 func 的
// alg.save 推断）。cfg 应为 Get 返回并修改后的完整行。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *SysctlService) Save(ctx context.Context, cfg SysctlConfig) error {
	_, err := s.c.Call(ctx, "ik_sysctl", "save", cfg)
	return err
}
