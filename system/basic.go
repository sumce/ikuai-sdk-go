package system

import (
	"context"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// basic.show 的 TYPE 选择器常量（burp 抓包实证）。
const (
	// BasicTypeTimestamp 系统时间戳区块（unix 秒）。
	BasicTypeTimestamp = "timestamp"
	// BasicTypeData 基础设置单行数据区块。
	BasicTypeData = "data"
	// BasicTypeInterface 接口名列表区块（listenport/backport 的候选值）。
	BasicTypeInterface = "interface"
	// BasicTypeSupportFast 快速转发支持标志区块。
	BasicTypeSupportFast = "support_fast"
)

// BasicTypeAll 是控制台「基础设置」页实际使用的组合
// （timestamp,data,interface,support_fast），Basic.Get 使用。
const BasicTypeAll = BasicTypeTimestamp + "," + BasicTypeData + "," + BasicTypeInterface + "," + BasicTypeSupportFast

// basic.set_language 观测到的语言代码（真机 support_i18n=0 仅观测到 1）。
const (
	// BasicLanguageZhCN 简体中文（抓包实证值）。
	BasicLanguageZhCN = int64(1)
)

// BasicSettings 是基础设置的单行数据（basic.show data 数组元素）。
type BasicSettings struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Hostname 主机名。
	Hostname string `json:"hostname"`
	// Language 语言代码（观测值 1=简体中文）。
	Language int64 `json:"language"`
	// TimeZone 时区小时偏移（观测值 8=UTC+8）。
	TimeZone int64 `json:"time_zone"`
	// TimeZoneFull 时区完整表示（观测值 "0800"）。
	TimeZoneFull string `json:"time_zone_full"`
	// SwitchNat NAT 模式开关（观测值 2；枚举语义待真机验证）。
	SwitchNat int64 `json:"switch_nat"`
	// SwitchDpi DPI 协议库开关（1=开启）。
	SwitchDpi int64 `json:"switch_dpi"`
	// SwitchNtp NTP 校时开关（1=开启）。
	SwitchNtp int64 `json:"switch_ntp"`
	// SwitchNtpd NTP 服务端开关（0=关闭；语义待真机验证）。
	SwitchNtpd int64 `json:"switch_ntpd"`
	// SwitchNtpserver 自定义 NTP 服务器开关（0=使用默认列表）。
	SwitchNtpserver int64 `json:"switch_ntpserver"`
	// NtpserverList 自定义 NTP 服务器列表（逗号分隔，未自定义时为空）。
	NtpserverList string `json:"ntpserver_list"`
	// NtpSyncCycle NTP 同步周期（分钟，观测值 60）。
	NtpSyncCycle int64 `json:"ntp_sync_cycle"`
	// LinkMode 内外网模式标志（观测值 0；枚举语义待真机验证）。
	LinkMode int64 `json:"link_mode"`
	// FastNat 快速转发开关（0=关闭）。
	FastNat int64 `json:"fast_nat"`
	// LanNat 内网 NAT 开关（1=开启）。
	LanNat int64 `json:"lan_nat"`
	// Listenport 监听（内网）接口名。
	Listenport string `json:"listenport"`
	// Backport 回源（外网）接口名。
	Backport string `json:"backport"`
}

// BasicInfo 是基础设置页的聚合结果（basic.show
// TYPE=timestamp,data,interface,support_fast）。
type BasicInfo struct {
	// Timestamp 服务端系统时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Data 基础设置数据（观测恒为单行）。
	Data []BasicSettings `json:"data"`
	// Interface 接口名列表（lan1.../adsl01...，供 listenport/backport 选择）。
	Interface []string `json:"interface"`
	// SupportFast 快速转发支持标志（观测值 0；语义待真机验证）。
	SupportFast int64 `json:"support_fast"`
}

// Row 返回基础设置首行；data 为空数组时返回 nil。
func (b *BasicInfo) Row() *BasicSettings {
	if b == nil || len(b.Data) == 0 {
		return nil
	}
	return &b.Data[0]
}

// BasicService 是基础设置（basic）的子服务。
type BasicService struct {
	c *ikuai.Client
}

// Get 获取基础设置（主机名/NAT/NTP/语言/时区/接口列表）。
// 对应 /Action/call: basic.show（TYPE=timestamp,data,interface,support_fast），
// 即 Web 控制台「系统设置 → 基础设置」页面。
func (s *BasicService) Get(ctx context.Context) (*BasicInfo, error) {
	var out BasicInfo
	if err := s.c.CallData(ctx, "basic", "show",
		map[string]any{"TYPE": BasicTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetLanguage 设置控制台语言（写动作）。
// 对应 /Action/call: basic.set_language（param={"language":N}，抓包实证
// N=1 简体中文）。lang 必须 > 0，否则返回 ikuai.ErrInvalidParam。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *BasicService) SetLanguage(ctx context.Context, lang int64) error {
	if lang <= 0 {
		return fmt.Errorf("system: lang 必须为正整数（观测值 %d）: %w", BasicLanguageZhCN, ikuai.ErrInvalidParam)
	}
	_, err := s.c.Call(ctx, "basic", "set_language", map[string]any{"language": lang})
	return err
}

// SyncTime 立即触发一次 NTP 校时（写动作）。
// 对应 /Action/call: basic.sync_time（抓包实证 param 为空）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *BasicService) SyncTime(ctx context.Context) error {
	_, err := s.c.Call(ctx, "basic", "sync_time", map[string]any{})
	return err
}
