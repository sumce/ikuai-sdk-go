package accounting

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// SelfPasswdConfig 是自助修改密码服务的单行配置（ppp_passwd.show 行，
// 真机 3.7.24 实证 7 字段；语义按固件 i18n self_password 页面标注）。
type SelfPasswdConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Enabled 自助修改密码总开关（"yes"/"no"；真机观测 "no"）。
	Enabled string `json:"enabled"`
	// AllowWeb 允许 WEB 认证用户修改密码（0/1；真机观测 1）。
	AllowWeb int64 `json:"allow_web"`
	// AllowPPPoE 允许 PPPoE 用户修改密码（0/1）。
	AllowPPPoE int64 `json:"allow_pppoe"`
	// AllowPPTP 允许 PPTP 用户修改密码（0/1）。
	AllowPPTP int64 `json:"allow_pptp"`
	// AllowL2TP 允许 L2TP 用户修改密码（0/1）。
	AllowL2TP int64 `json:"allow_l2tp"`
	// AllowOVPN 允许 OpenVPN 用户修改密码（0/1）。
	AllowOVPN int64 `json:"allow_ovpn"`
	// AllowQuit 允许 WEB 认证用户退出登录（0/1）。
	AllowQuit int64 `json:"allow_quit"`
}

// IsEnabled 报告自助修改密码服务是否开启。
func (c SelfPasswdConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// SelfPasswdService 是自助密码（ppp_passwd）的子服务。
type SelfPasswdService struct {
	c *ikuai.Client
}

// Get 获取自助修改密码配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: ppp_passwd.show（TYPE=data，真机 3.7.24 实证），
// 即 Web 控制台「认证计费 → 认证账号 → 自助密码」页面。
func (s *SelfPasswdService) Get(ctx context.Context) (*SelfPasswdConfig, error) {
	var out struct {
		Data []SelfPasswdConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "ppp_passwd", "show",
		map[string]any{"TYPE": cfgTypeData}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// Set 保存自助修改密码配置（整行保存，随行携带 ID）。
// 对应 /Action/call: ppp_passwd.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *SelfPasswdService) Set(ctx context.Context, cfg SelfPasswdConfig) error {
	_, err := s.c.Call(ctx, "ppp_passwd", "save", cfg)
	return err
}
