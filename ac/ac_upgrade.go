package ac

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// APUpgradeDevice 是一台待升级 AP（ac_upgrade.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 无 AP 设备，升级列表恒为空数组，
// 行字段无法实证。字段名按固件 i18n（ac 区块 ap-mac「AP MAC」/
// now_version「当前版本」/new_version「最新版本」/Model「型号」）推导。
type APUpgradeDevice struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// MAC AP 的 MAC 地址（i18n ac.ap-mac）。
	MAC string `json:"mac"`
	// Model AP 型号（i18n ac.Model）。
	Model string `json:"model"`
	// Version 当前固件版本（i18n ac.now_version）。
	Version string `json:"version"`
	// NewVersion 最新固件版本（i18n ac.new_version）。
	NewVersion string `json:"new_version"`
}

// UpgradeService 是 AP 系统升级（ac_upgrade）的子服务。
type UpgradeService struct {
	c *ikuai.Client
}

// List 获取 AP 可升级列表（分页）。
// 对应 /Action/call: ac_upgrade.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「AC管理 → AP系统升级」页面。
func (s *UpgradeService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[APUpgradeDevice], error) {
	resp, err := s.c.Call(ctx, "ac_upgrade", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[APUpgradeDevice](resp.Data)
}

// Upgrade 触发指定 AP 的固件升级（写动作）。
// 对应 /Action/call: ac_upgrade.set（action 名与参数形态待真机验证，
// 参数按 SDK 约定取 ikuai.IDParam：单个 → {"id":N}，多个 → {"ids":[...]}）。
// ⚠️ 写动作且不可逆（升级过程断电会导致 AP 变砖，升级完成后 AP 自动
// 重启）：仅离线测试覆盖，绝对禁止对真机调用。
func (s *UpgradeService) Upgrade(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ac_upgrade", "set", ikuai.IDParam(ids))
	return err
}
