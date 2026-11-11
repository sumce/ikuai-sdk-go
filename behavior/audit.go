package behavior

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// audit.show 的 TYPE 形态说明（真机 3.7.24 实测）：
//   - 无 TYPE（param={}）→ {"data":[配置行]}
//   - TYPE=total,data    → 30006 unknown TYPE (total)（不接受分页 TYPE）
//   - TYPE=all           → {"clean_type":[...],"data":[配置行]}
//
// 因此本 func 不走 typedShowParam/defaultType，Get 直接发空 param。

// audit.show TYPE=all 附带的 clean_type 可清理内容类型值（真机实测
// ["all","im","url","terminal"]，与固件 i18n recordset 的清理内容选项一致）。
const (
	AuditCleanTypeAll      = "all"      // 全部
	AuditCleanTypeIM       = "im"       // IM 上线记录
	AuditCleanTypeURL      = "url"      // 网址浏览记录
	AuditCleanTypeTerminal = "terminal" // 终端上下线记录
)

// AuditConfig 是行为记录配置的单行数据（audit.show 的 data 元素，真机
// 3.7.24 实测行字段；观测恒为单行，行内无 id 字段）。
type AuditConfig struct {
	// TotalSize 审计存储总容量（字节，观测 60579692）。
	TotalSize int64 `json:"total_size"`
	// UseSize 已使用容量（字节，观测 82700）。
	UseSize int64 `json:"use_size"`
	// OpenURLRecord 网址浏览记录开关（观测 0；1=开启 推断，语义待真机验证）。
	OpenURLRecord int64 `json:"open_url_record"`
	// OpenIMRecord IM 上线记录开关（观测 0；1=开启 推断，语义待真机验证）。
	OpenIMRecord int64 `json:"open_im_record"`
	// OpenTerminalRecord 终端上下线记录开关（观测 0；1=开启 推断，语义待真机验证）。
	OpenTerminalRecord int64 `json:"open_terminal_record"`
	// OpenAppidRecord 应用详情统计开关（观测 0；1=开启 推断，语义待真机验证）。
	OpenAppidRecord int64 `json:"open_appid_record"`
}

// IsURLRecordEnabled 网址浏览记录是否开启（非 0 视为开启）。
func (c AuditConfig) IsURLRecordEnabled() bool { return c.OpenURLRecord != 0 }

// IsIMRecordEnabled IM 上线记录是否开启（非 0 视为开启）。
func (c AuditConfig) IsIMRecordEnabled() bool { return c.OpenIMRecord != 0 }

// IsTerminalRecordEnabled 终端上下线记录是否开启（非 0 视为开启）。
func (c AuditConfig) IsTerminalRecordEnabled() bool { return c.OpenTerminalRecord != 0 }

// IsAppidRecordEnabled 应用详情统计是否开启（非 0 视为开启）。
func (c AuditConfig) IsAppidRecordEnabled() bool { return c.OpenAppidRecord != 0 }

// AuditInfo 是行为记录配置页的聚合结果（audit.show TYPE=all，真机实测
// 区块 clean_type + data）。
type AuditInfo struct {
	// CleanType 可清理内容类型列表（真机实测 ["all","im","url","terminal"]，
	// 值域见 AuditCleanType* 常量）。
	CleanType []string `json:"clean_type"`
	// Rows 配置行（观测恒为单行）。
	Rows []AuditConfig `json:"data"`
}

// Row 返回配置首行；data 为空数组时返回 nil。
func (a *AuditInfo) Row() *AuditConfig {
	if a == nil || len(a.Rows) == 0 {
		return nil
	}
	return &a.Rows[0]
}

// AuditService 是审计设置（audit + record_free_setting）的子服务。
type AuditService struct {
	c *ikuai.Client
}

// Get 获取行为记录配置（存储容量与 URL/IM/终端/应用详情记录开关）。
// 对应 /Action/call: audit.show（无 TYPE，param={}，真机实测），
// 即 Web 控制台「行为管控 → 行为记录管理 → 行为记录配置」页面。
func (s *AuditService) Get(ctx context.Context) (*AuditConfig, error) {
	var out struct {
		Data []AuditConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "audit", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// Info 获取行为记录配置页聚合信息（配置行 + 可清理内容类型列表）。
// 对应 /Action/call: audit.show（TYPE=all，真机实测）。
func (s *AuditService) Info(ctx context.Context) (*AuditInfo, error) {
	var out AuditInfo
	if err := s.c.CallData(ctx, "audit", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Save 保存行为记录配置（写动作）。
// 对应 /Action/call: audit.save（action 名按同类表单配置 func 的
// alg.save/ik_sysctl.save 推断，固件 burp 无本 func 抓包）。
// cfg 应为 Get 返回并修改后的完整行（固件按整行覆盖）。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *AuditService) Save(ctx context.Context, cfg AuditConfig) error {
	_, err := s.c.Call(ctx, "audit", "save", cfg)
	return err
}

// RecordFreeConfig 是审计豁免配置（record_free_setting.show）。
//
// ⚠️ 探测真机（3.7.24）对该 func 全部 show 形态回 30002 Not found
// funcname(record_free_setting)——本固件未注册该 func（菜单仍挂「审计豁免」
// 页，i18n 节 ip_white_list/mac_white_list）；空表推导，待真机验证：结构按
// 豁免页两张白名单表（审计白名单 IP / 审计白名单 MAC，列：值、说明
// explain、组 group）推导，IP/MAC 值字段名（ip/mac）为保守猜测。
type RecordFreeConfig struct {
	// IPWhiteList 审计白名单 IP 列表（配置生效后表内 IP 无需行为记录）。
	IPWhiteList []RecordFreeIP `json:"ip_white_list"`
	// MACWhiteList 审计白名单 MAC 列表（配置生效后表内 MAC 无需行为记录）。
	MACWhiteList []RecordFreeMAC `json:"mac_white_list"`
}

// RecordFreeIP 是审计豁免 IP 白名单的一行（空表推导，待真机验证）。
type RecordFreeIP struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// IP 豁免 IP/网段。字段名待真机验证。
	IP string `json:"ip"`
	// Explain 说明（固件 i18n 列名实证）。
	Explain string `json:"explain"`
	// Group 所属分组（固件 i18n 列名实证）。
	Group string `json:"group"`
}

// RecordFreeMAC 是审计豁免 MAC 白名单的一行（空表推导，待真机验证）。
type RecordFreeMAC struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// MAC 豁免 MAC 地址。字段名待真机验证。
	MAC string `json:"mac"`
	// Explain 说明（固件 i18n 列名实证）。
	Explain string `json:"explain"`
	// Group 所属分组（固件 i18n 列名实证）。
	Group string `json:"group"`
}

// RecordFreeGet 获取审计豁免配置（IP/MAC 白名单）。
// 对应 /Action/call: record_free_setting.show（param={}）。
// 注意：探测真机（3.7.24）回 30002 Not found funcname，字段为空表推导，
// 待真机验证。
func (s *AuditService) RecordFreeGet(ctx context.Context) (*RecordFreeConfig, error) {
	var out RecordFreeConfig
	if err := s.c.CallData(ctx, "record_free_setting", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RecordFreeSet 保存审计豁免配置（写动作）。
// 对应 /Action/call: record_free_setting.save（action 名按同类表单配置
// func 推断，无抓包实证）。
// cfg 应为 RecordFreeGet 返回并修改后的完整配置。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *AuditService) RecordFreeSet(ctx context.Context, cfg RecordFreeConfig) error {
	_, err := s.c.Call(ctx, "record_free_setting", "save", cfg)
	return err
}
