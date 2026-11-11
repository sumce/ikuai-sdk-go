package ac

import (
	"context"
	"fmt"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// wls_black.control_mode 的控制模式枚举（语义依据 i18n wireless_list
// black_mode_tip「禁止规则中mac连接指定SSID」/ white_mode_tip「只允许
// 规则中mac连接指定SSID」；具体取值字符串待真机验证）。
const (
	// WlsModeBlack 黑名单模式（禁止规则中 MAC 连接指定 SSID）。
	WlsModeBlack = "black"
	// WlsModeWhite 白名单模式（只允许规则中 MAC 连接指定 SSID）。
	WlsModeWhite = "white"
)

// WlsBlackRule 是一条无线黑白名单规则（wls_black.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 无 AP 设备，规则列表恒为空数组，
// 行字段无法实证。字段名按固件 i18n（wireless_list 区块 control_mode/
// mac/object/week/time/comment）推导。
type WlsBlackRule struct {
	// ID 规则 ID（add 时无需填写；edit/del 必填）。
	ID int64 `json:"id,omitempty"`
	// ControlMode 控制模式（black/white，见 WlsModeBlack/WlsModeWhite；
	// 取值字符串待真机验证）。
	ControlMode string `json:"control_mode"`
	// MAC 受控终端 MAC 地址。
	MAC string `json:"mac"`
	// Object 作用对象（SSID 名称，i18n wireless_list.object；"all" 是否
	// 表示全部 SSID 待真机验证）。
	Object string `json:"object"`
	// Week 周期（i18n wireless_list.week；格式如 "1111100" 待真机验证）。
	Week string `json:"week"`
	// Time 生效时间（i18n wireless_list.time；格式如 "00:00-23:59" 待真机验证）。
	Time string `json:"time"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// BlacklistService 是无线黑白名单（wls_black）的子服务。
type BlacklistService struct {
	c *ikuai.Client
}

// List 获取无线黑白名单规则列表（分页）。
// 对应 /Action/call: wls_black.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「AC管理 → 无线黑白名单」页面。
func (s *BlacklistService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[WlsBlackRule], error) {
	resp, err := s.c.Call(ctx, "wls_black", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[WlsBlackRule](resp.Data)
}

// Add 新增无线黑白名单规则，返回固件分配的行 ID（写动作）。
// 对应 /Action/call: wls_black.add。mac 为必填，为空（或全空白）返回
// ikuai.ErrInvalidParam。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *BlacklistService) Add(ctx context.Context, rule WlsBlackRule) (int64, error) {
	if strings.TrimSpace(rule.MAC) == "" {
		return 0, fmt.Errorf("ac: 终端 MAC 不能为空: %w", ikuai.ErrInvalidParam)
	}
	resp, err := s.c.Call(ctx, "wls_black", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改无线黑白名单规则（写动作）。
// 对应 /Action/call: wls_black.edit（rule.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *BlacklistService) Edit(ctx context.Context, rule WlsBlackRule) error {
	_, err := s.c.Call(ctx, "wls_black", "edit", rule)
	return err
}

// Delete 删除无线黑白名单规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）（写动作）。
// 对应 /Action/call: wls_black.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *BlacklistService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "wls_black", "del", ikuai.IDParam(ids))
	return err
}

// WlsMvlanRule 是一条无线终端 VLAN 绑定（wls_mvlan.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 无 AP 设备，绑定列表恒为空数组，
// 行字段无法实证。字段名按固件语义推导：i18n helpTip.
// wirelessTerminalVlanTip「支持基于mac地址和SSID灵活设置无线终端所属vlan」，
// VLAN 字段名参照 network.vlan 的 vlan_id。
type WlsMvlanRule struct {
	// ID 规则 ID（add 时无需填写；edit/del 必填）。
	ID int64 `json:"id,omitempty"`
	// MAC 终端 MAC 地址（基于 MAC 指定 VLAN）。
	MAC string `json:"mac"`
	// SSID SSID 名称（基于 SSID 指定 VLAN）。
	SSID string `json:"ssid"`
	// VlanID 绑定的 VLAN ID（取值范围 1~4090）。
	VlanID int64 `json:"vlan_id"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// MvlanService 是无线终端 VLAN（wls_mvlan）的子服务。
type MvlanService struct {
	c *ikuai.Client
}

// List 获取无线终端 VLAN 绑定列表（分页）。
// 对应 /Action/call: wls_mvlan.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「AC管理 → 无线终端VLAN」页面。
func (s *MvlanService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[WlsMvlanRule], error) {
	resp, err := s.c.Call(ctx, "wls_mvlan", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[WlsMvlanRule](resp.Data)
}

// Add 新增无线终端 VLAN 绑定，返回固件分配的行 ID（写动作）。
// 对应 /Action/call: wls_mvlan.add。mac 与 ssid 至少填一项（固件语义
// 「基于 mac 地址和 SSID 设置 VLAN」），均为空返回 ikuai.ErrInvalidParam；
// vlanid 必须在 1~4090。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *MvlanService) Add(ctx context.Context, rule WlsMvlanRule) (int64, error) {
	if strings.TrimSpace(rule.MAC) == "" && strings.TrimSpace(rule.SSID) == "" {
		return 0, fmt.Errorf("ac: mac 与 ssid 至少填一项: %w", ikuai.ErrInvalidParam)
	}
	if rule.VlanID < 1 || rule.VlanID > 4090 {
		return 0, fmt.Errorf("ac: vlan_id 必须在 1~4090: %w", ikuai.ErrInvalidParam)
	}
	resp, err := s.c.Call(ctx, "wls_mvlan", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改无线终端 VLAN 绑定（写动作）。
// 对应 /Action/call: wls_mvlan.edit（rule.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *MvlanService) Edit(ctx context.Context, rule WlsMvlanRule) error {
	_, err := s.c.Call(ctx, "wls_mvlan", "edit", rule)
	return err
}

// Delete 删除无线终端 VLAN 绑定，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）（写动作）。
// 对应 /Action/call: wls_mvlan.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *MvlanService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "wls_mvlan", "del", ikuai.IDParam(ids))
	return err
}
