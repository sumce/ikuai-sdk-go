package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// stream_control.show 的 TYPE 选择器常量（真机 3.7.24 实测：本 func 只接受
// all 与 stream_ctl_mode 两个 TYPE 值，其余回 30006 unknown TYPE）。
const (
	// SmartControlTypeAll 全量区块选择器，响应 {"stream_ctl_mode":N}
	// （真机实测）。
	SmartControlTypeAll = "all"
	// SmartControlTypeMode 模式区块选择器，响应与 TYPE=all 相同
	// （真机实测；开源生产集成常用）。
	SmartControlTypeMode = "stream_ctl_mode"
)

// stream_ctl_mode 字段的模式值（真机实测 1，开源生产集成实证 0/1 两态）。
const (
	// SmartControlModeOff 智能流控关闭。
	SmartControlModeOff int64 = 0
	// SmartControlModeOn 智能流控开启（自动模式）。
	SmartControlModeOn int64 = 1
)

// SmartControlConfig 是智能流控的单行配置（stream_control.show 的 Data）。
type SmartControlConfig struct {
	// StreamCtlMode 智能流控模式（真机实测 1=开启；开源生产集成实证
	// 0=关闭、1=开启）。字段语义待真机验证（是否还有更多模式值）。
	StreamCtlMode int64 `json:"stream_ctl_mode"`
}

// IsEnabled 智能流控是否开启。
func (c SmartControlConfig) IsEnabled() bool { return c.StreamCtlMode == SmartControlModeOn }

// SmartControlService 是智能流控（stream_control）的子服务。
type SmartControlService struct {
	c *ikuai.Client
}

// Get 获取智能流控配置（模式开关）。
// 对应 /Action/call: stream_control.show（TYPE=all），
// 即 Web 控制台「流控设置 → 智能流控」页面。
func (s *SmartControlService) Get(ctx context.Context) (*SmartControlConfig, error) {
	var out SmartControlConfig
	if err := s.c.CallData(ctx, "stream_control", "show",
		map[string]any{"TYPE": SmartControlTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Set 保存智能流控配置（写动作）。
// 对应 /Action/call: stream_control.seting（固件 action 即为该拼写，
// 开源生产集成 dscao/ikuai 实证；param 仅含 stream_ctl_mode）。
// cfg 应为 Get 返回并修改后的配置。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *SmartControlService) Set(ctx context.Context, cfg SmartControlConfig) error {
	_, err := s.c.Call(ctx, "stream_control", "seting", cfg)
	return err
}
