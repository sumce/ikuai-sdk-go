package accounting

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// OnlineUser 是一条认证在线用户记录（ppp_online.show 行）。
//
// 真机（3.7.24，认证计费未启用）data 为空数组，行字段无法实证；控制台
// 列（固件 i18n auth_expenses.account_online）为：账号 username / 姓名
// name / 认证类型 auth_type / 上线时间 online_time / 到期时间 end_time /
// 联系方式 phone / IP 地址 / MAC / 上行 upload / 下行 download / 接口
// interface / 代拨账号 agent_account / 备注 comment。
// 故行对象以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type OnlineUser struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (u *OnlineUser) UnmarshalJSON(b []byte) error {
	u.Raw = append(u.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (u OnlineUser) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(u.Raw, &m); err != nil {
		return nil, fmt.Errorf("accounting: 解码 ppp_online 行失败: %w", err)
	}
	return m, nil
}

// OnlineUserService 是认证在线用户（ppp_online）的子服务。
type OnlineUserService struct {
	c *ikuai.Client
}

// List 获取认证在线用户列表（分页，支持关键字/过滤）。
// 对应 /Action/call: ppp_online.show（TYPE=total,data, limit="offset,rows"，
// 真机实证；TYPE=data/total/all 亦可用），即 Web 控制台
// 「认证计费 → 认证用户 → 在线用户」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *OnlineUserService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[OnlineUser], error) {
	resp, err := s.c.Call(ctx, "ppp_online", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[OnlineUser](resp.Data)
}
