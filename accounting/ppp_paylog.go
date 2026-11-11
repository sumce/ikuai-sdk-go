package accounting

import (
	"context"
	"time"

	"github.com/sumce/ikuai-sdk-go"
)

// PayLogAction 常量：收费记录的操作描述（ppp_paylog.action，真机实证
// "开户"；其余取值按固件 i18n accountManagement 业务动作推导，待真机验证）。
const (
	// PayLogActionOpen 开户（真机实证取值）。
	PayLogActionOpen = "开户"
	// PayLogActionRenew 套餐续费。
	PayLogActionRenew = "套餐续费"
	// PayLogActionChange 变更当前套餐。
	PayLogActionChange = "变更当前套餐"
	// PayLogActionUnsubscribe 退订套餐。
	PayLogActionUnsubscribe = "退订套餐"
)

// PayLogEntry 是一条收费记录（ppp_paylog.show 行，真机 3.7.24 实证
// 9 字段；语义按固件 i18n reportManagement 页面标注）。
type PayLogEntry struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Username 账号（真机观测 "admin"）。
	Username string `json:"username"`
	// Name 用户姓名。
	Name string `json:"name"`
	// Admin 收费人员（真机观测 "admin"）。
	Admin string `json:"admin"`
	// Action 描述/操作类型（真机观测 "开户"，见 PayLogAction* 常量）。
	Action string `json:"action"`
	// Packname 套餐名称（真机观测 "自定义"）。
	Packname string `json:"packname"`
	// FeeMoney 收费金额（元）。
	FeeMoney float64 `json:"feemoney"`
	// Timestamp 收费时间（Unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// Comment 备注（真机观测 "自定义,账号到期时间:2029-12-21,"）。
	Comment string `json:"comment"`
}

// Time 收费时间。
func (e PayLogEntry) Time() time.Time { return time.Unix(e.Timestamp, 0) }

// PayLogService 是报表管理（ppp_paylog）的子服务。
type PayLogService struct {
	c *ikuai.Client
}

// List 获取收费记录列表（分页，支持关键字/过滤）。
// 对应 /Action/call: ppp_paylog.show（TYPE=total,data, limit="offset,rows"，
// 真机实证），即 Web 控制台「认证计费 → 认证账号 → 报表管理」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *PayLogService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[PayLogEntry], error) {
	resp, err := s.c.Call(ctx, "ppp_paylog", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[PayLogEntry](resp.Data)
}
