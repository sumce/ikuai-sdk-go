package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// watchdog 开关枚举值（watchdog 字段；1 为真机实测开启值，0 按开关
// 语义推导）。
const (
	// WatchdogOff 关闭主动健康检测。
	WatchdogOff = 0
	// WatchdogOn 开启主动健康检测（真机实测值）。
	WatchdogOn = 1
)

// WatchdogConfig 是看门狗（健康检测）的单行配置（watchdog.show 的
// data 数组元素，真机 3.7.24 实测单行）。
type WatchdogConfig struct {
	// Watchdog 主动健康检测开关（1=开启/0=关闭，枚举 WatchdogOn/Off）。
	Watchdog int64 `json:"watchdog"`
	// WatchdogThresh 等待时间，单位秒（i18n watchdog.WaitTime；i18n
	// 单位键 minute 实标「秒」，真机观测 30；当系统持续非健康状态
	// 超过该时长时自动重启）。
	WatchdogThresh int64 `json:"watchdog_thresh"`
}

// IsEnabled 主动健康检测是否开启（watchdog == 1）。
func (c WatchdogConfig) IsEnabled() bool { return c.Watchdog == WatchdogOn }

// watchdogPayload 是 watchdog.show 的 Data 包裹形态。
type watchdogPayload struct {
	Data []WatchdogConfig `json:"data"`
}

// WatchdogService 是看门狗（watchdog）的子服务
// （service.watchdog「健康检测」页面，挂在应用工具菜单下）。
type WatchdogService struct {
	c *ikuai.Client
}

// Get 获取看门狗（主动健康检测）配置。
// 对应 /Action/call: watchdog.show（空参数 {}，真机 3.7.24 实测返回
// data 单行 {watchdog_thresh:30, watchdog:1}）。
//
// 未配置时返回 nil。
func (s *WatchdogService) Get(ctx context.Context) (*WatchdogConfig, error) {
	var out watchdogPayload
	if err := s.c.CallData(ctx, "watchdog", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	return &out.Data[0], nil
}

// Set 保存看门狗（主动健康检测）配置（写动作）。
// 对应 /Action/call: watchdog.edit（action 名按固件单行配置页惯例
// 推断，param 为完整 WatchdogConfig——未对真机调用，action 名与
// 字段取值待真机验证）。
//
// ⚠️ 写动作，禁止对真机调用：该配置可在系统持续非健康时触发自动
// 重启（i18n watchdog.help），仅离线测试覆盖。
func (s *WatchdogService) Set(ctx context.Context, cfg WatchdogConfig) error {
	_, err := s.c.Call(ctx, "watchdog", "edit", cfg)
	return err
}
