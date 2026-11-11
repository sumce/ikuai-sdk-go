package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// wecom.show 的 TYPE 选择器常量（burp 抓包实证）。
const (
	// WeComTypeDataBindID 控制台「应用绑定」页企业微信卡片使用的组合
	// （data,bindid）：绑定配置与绑定 ID 一并返回。
	WeComTypeDataBindID = "data,bindid"
)

// WeComStatus 是企业微信绑定状态（wecom.show TYPE=data,bindid）。
// 企业微信是应用绑定（dingtalk）的姊妹功能，与钉钉共用
// Web 控制台「系统设置 → 应用绑定」页面。
type WeComStatus struct {
	// Enable 企业微信绑定功能开关（"yes"/"no"）。
	Enable string `json:"enable"`
	// SN 设备序列号（未绑定时为空）。
	SN string `json:"sn"`
	// URL 扫码绑定链接（未绑定时为空；绑定时为企业微信授权页地址，
	// 可类比钉钉 bind_status.url）。
	URL string `json:"url"`
	// Status 绑定流程状态（未绑定为 0；枚举语义待真机验证，
	// 可类比钉钉 bind_status 的 1=二维码就绪/2=获取中）。
	Status int64 `json:"status"`
	// BindID 绑定 ID（未绑定为 0，语义同 dingtalk.show TYPE=bindid）。
	BindID int64 `json:"bindid"`
}

// IsEnabled 固件 enable 字段为 "yes"/"no"。
func (w WeComStatus) IsEnabled() bool { return w.Enable == EnabledYes }

// WeComService 是企业微信绑定（wecom）的子服务。
type WeComService struct {
	c *ikuai.Client
}

// Status 获取企业微信绑定状态（功能开关/序列号/扫码链接/绑定状态/绑定 ID）。
// 对应 /Action/call: wecom.show（TYPE=data,bindid，burp 抓包实证），
// 即 Web 控制台「系统设置 → 应用绑定」页面的企业微信绑定卡片
// （钉钉 dingtalk 的姊妹功能）。
func (s *WeComService) Status(ctx context.Context) (*WeComStatus, error) {
	var out WeComStatus
	if err := s.c.CallData(ctx, "wecom", "show",
		map[string]any{"TYPE": WeComTypeDataBindID}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
