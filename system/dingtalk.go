package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// dingtalk.show 的 TYPE 选择器常量（抓包 + 真机 3.7.24 验证）。
const (
	// DingTalkTypeBindStatus 绑定状态/二维码区块。
	DingTalkTypeBindStatus = "bind_status"
	// DingTalkTypeBindID 绑定 ID 区块。
	DingTalkTypeBindID = "bindid"
	// DingTalkTypeAppBindSupport 应用绑定支持能力区块。
	DingTalkTypeAppBindSupport = "appbind_support"
)

// dingtalk.show bind_status.status 的观测枚举值（其余取值待真机验证）。
const (
	// DingTalkStatusQRReady 二维码就绪，等待扫码（观测 errmsg=
	// "使用钉钉扫描二维码,绑定到钉钉"，url 为二维码链接）。
	DingTalkStatusQRReady = int64(1)
	// DingTalkStatusGettingQR 正在获取二维码（观测 url 为空）。
	DingTalkStatusGettingQR = int64(2)
)

// DingTalkBindStatus 是钉钉绑定状态（dingtalk.show TYPE=bind_status）。
type DingTalkBindStatus struct {
	// Status 绑定流程状态（观测值 1=二维码就绪待扫，2=正在获取二维码；
	// 其余取值待真机验证）。
	Status int64 `json:"status"`
	// URL 扫码绑定链接（status=1 时为钉钉页面地址，否则为空）。
	URL string `json:"url"`
	// ErrMsg 状态描述文本（如 "正在获取二维码"）。
	ErrMsg string `json:"errmsg"`
}

// DingTalkAppBindSupport 是应用绑定能力支持情况
// （dingtalk.show TYPE=appbind_support）。
type DingTalkAppBindSupport struct {
	// SupportDtalk 是否支持钉钉绑定（1=支持）。
	SupportDtalk int64 `json:"support_dtalk"`
	// SupportWecom 是否支持企业微信绑定（1=支持）。
	SupportWecom int64 `json:"support_wecom"`
}

// DingTalkService 是应用绑定（dingtalk）的子服务。
type DingTalkService struct {
	c *ikuai.Client
}

// BindStatus 获取钉钉绑定状态（含扫码链接）。
// 对应 /Action/call: dingtalk.show（TYPE=bind_status），
// 即 Web 控制台「系统设置 → 应用绑定」页面的钉钉绑定卡片。
func (s *DingTalkService) BindStatus(ctx context.Context) (*DingTalkBindStatus, error) {
	var out DingTalkBindStatus
	if err := s.c.CallData(ctx, "dingtalk", "show",
		map[string]any{"TYPE": DingTalkTypeBindStatus}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BindID 获取钉钉绑定 ID（未绑定为 0）。
// 对应 /Action/call: dingtalk.show（TYPE=bindid）。
func (s *DingTalkService) BindID(ctx context.Context) (int64, error) {
	var out struct {
		BindID int64 `json:"bindid"`
	}
	if err := s.c.CallData(ctx, "dingtalk", "show",
		map[string]any{"TYPE": DingTalkTypeBindID}, &out); err != nil {
		return 0, err
	}
	return out.BindID, nil
}

// AppBindSupport 获取应用绑定能力支持情况（钉钉/企业微信）。
// 对应 /Action/call: dingtalk.show（TYPE=appbind_support）。
func (s *DingTalkService) AppBindSupport(ctx context.Context) (*DingTalkAppBindSupport, error) {
	var out DingTalkAppBindSupport
	if err := s.c.CallData(ctx, "dingtalk", "show",
		map[string]any{"TYPE": DingTalkTypeAppBindSupport}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
