package accounting

import (
	"context"
	"encoding/base64"

	"github.com/sumce/ikuai-sdk-go"
)

// decodeNoticeContent 解码通知正文：固件以 base64 编码的 HTML 文档承载
// 通知内容（notice_remind/notice_expires.content 真机实证；notice_temp.
// content 空串同构）。空串返回空串。
func decodeNoticeContent(content string) (string, error) {
	if content == "" {
		return "", nil
	}
	b, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// NoticeTempConfig 是实时通知的单行配置（notice_temp.show 行，真机
// 3.7.24 实证 7 字段；语义按固件 i18n realtime_notice 页面标注）。
type NoticeTempConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Content 通知内容（base64 编码的 HTML 文档，真机观测为空）。
	Content string `json:"content"`
	// RedURL 认证页跳转地址（i18n JumpPage；真机观测 "http://www.baidu.com"）。
	RedURL string `json:"red_url"`
	// Timeout 显示倒计时（秒，0=不启用倒计时；真机观测 60）。
	Timeout int64 `json:"timeout"`
	// PPPUser 接收对象含拨号用户（0/1；真机观测 0）。
	PPPUser int64 `json:"pppuser"`
	// ListUser 接收对象含列表中的内网用户（0/1；真机观测 0）。
	ListUser int64 `json:"listuser"`
	// IPGroup 接收的 IPv4 分组（ListUser=1 时的 IP 列表分组；真机观测为空）。
	IPGroup string `json:"ipgroup"`
}

// DecodeContent 解码通知正文（base64 → HTML 文本）。
func (c NoticeTempConfig) DecodeContent() (string, error) {
	return decodeNoticeContent(c.Content)
}

// NoticeCycle 是一条周期通知（notice_cycle.show 行；菜单「周期通知」，
// 固件 i18n week_notice）。
//
// 真机（3.7.24，认证计费未启用）data 为空数组，行字段由固件 i18n
// week_notice 表单项推导，并与同族 notice_remind 实证行（time/red_url/
// content/timeout 同构）交叉印证，字段语义待真机验证。
type NoticeCycle struct {
	// ID 行 ID（add 时无需填写；edit/del 必填）。omitempty 保证 add 参数
	// 不携带 id。
	ID int64 `json:"id,omitempty"`
	// Content 通知内容（base64 编码的 HTML 文档，同 notice_remind）。
	Content string `json:"content"`
	// Week 推送周期（星期集合，如 "1234567"；形态待真机验证）。
	Week string `json:"week"`
	// Time 推送时间（如 "08:00"；形态参照 notice_remind.time）。
	Time string `json:"time"`
	// RedURL 认证页跳转地址（i18n JumpPage）。
	RedURL string `json:"red_url"`
	// Timeout 显示倒计时（秒，0=不启用倒计时）。
	Timeout int64 `json:"timeout"`
	// PPPUser 接收对象含拨号用户（0/1）。
	PPPUser int64 `json:"pppuser"`
	// ListUser 接收对象含列表中的内网用户（0/1）。
	ListUser int64 `json:"listuser"`
	// IPGroup 接收的 IPv4 分组（ListUser=1 时）。
	IPGroup string `json:"ipgroup"`
}

// DecodeContent 解码通知正文（base64 → HTML 文本）。
func (c NoticeCycle) DecodeContent() (string, error) {
	return decodeNoticeContent(c.Content)
}

// NoticeRemindConfig 是到期提醒的单行配置（notice_remind.show 行，真机
// 3.7.24 实证 6 字段；语义按固件 i18n expirenotice 页面标注）。
type NoticeRemindConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Enabled 到期提醒开关（"yes"/"no"；真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// Content 提前到期提醒的通知内容（base64 编码的 HTML 文档，真机实证）。
	Content string `json:"content"`
	// RedURL 认证页跳转地址（真机观测 "http://www.baidu.com"）。
	RedURL string `json:"red_url"`
	// Days 提前提醒天数（真机观测 5；i18n unit=天）。
	Days int64 `json:"days"`
	// Time 定时提醒时刻列表（真机观测 "08:00,16:00"）。
	Time string `json:"time"`
	// Timeout 显示倒计时（秒，0=不启用倒计时；真机观测 60）。
	Timeout int64 `json:"timeout"`
}

// IsEnabled 报告到期提醒是否开启。
func (c NoticeRemindConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// DecodeContent 解码通知正文（base64 → HTML 文本）。
func (c NoticeRemindConfig) DecodeContent() (string, error) {
	return decodeNoticeContent(c.Content)
}

// NoticeExpiresConfig 是 PPPoE 到期通知的单行配置（notice_expires.show
// 行，真机 3.7.24 实证 4 字段；语义按固件 i18n pppoe_notice 页面标注）。
type NoticeExpiresConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Content 到期通知内容（base64 编码的 HTML 文档，真机实证）。
	Content string `json:"content"`
	// WhiteIP 白名单公网 IP（多行，如 "192.168.6.1/24"；真机观测为空）。
	WhiteIP string `json:"whiteip"`
	// WhiteDomain 白名单域名（多行；真机观测为空）。
	WhiteDomain string `json:"whitedomain"`
}

// DecodeContent 解码通知正文（base64 → HTML 文本）。
func (c NoticeExpiresConfig) DecodeContent() (string, error) {
	return decodeNoticeContent(c.Content)
}

// NoticesService 是通知族（notice_temp/notice_cycle/notice_remind/
// notice_expires）的子服务。四个 func 对应 Web 控制台「认证计费 → 通知」
// 菜单：实时通知/周期通知/到期提醒/PPPoE 到期通知。
type NoticesService struct {
	c *ikuai.Client
}

// Temp 获取实时通知配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: notice_temp.show（TYPE=data，真机 3.7.24 实证）。
func (s *NoticesService) Temp(ctx context.Context) (*NoticeTempConfig, error) {
	var out struct {
		Data []NoticeTempConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "notice_temp", "show",
		map[string]any{"TYPE": cfgTypeData}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetTemp 保存实时通知配置（整行保存，随行携带 ID）。
// 对应 /Action/call: notice_temp.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NoticesService) SetTemp(ctx context.Context, cfg NoticeTempConfig) error {
	_, err := s.c.Call(ctx, "notice_temp", "save", cfg)
	return err
}

// CycleList 获取周期通知列表（分页）。
// 对应 /Action/call: notice_cycle.show（TYPE=total,data, limit="offset,rows"，
// 真机实证），即 Web 控制台「认证计费 → 通知 → 周期通知」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *NoticesService) CycleList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[NoticeCycle], error) {
	resp, err := s.c.Call(ctx, "notice_cycle", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[NoticeCycle](resp.Data)
}

// CycleAdd 新增周期通知，返回固件分配的行 ID。
// 对应 /Action/call: notice_cycle.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NoticesService) CycleAdd(ctx context.Context, row NoticeCycle) (int64, error) {
	resp, err := s.c.Call(ctx, "notice_cycle", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// CycleEdit 按 ID 修改周期通知。
// 对应 /Action/call: notice_cycle.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NoticesService) CycleEdit(ctx context.Context, row NoticeCycle) error {
	_, err := s.c.Call(ctx, "notice_cycle", "edit", row)
	return err
}

// CycleDelete 删除周期通知，支持单个或多个 ID。
// 对应 /Action/call: notice_cycle.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NoticesService) CycleDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "notice_cycle", "del", ikuai.IDParam(ids))
	return err
}

// Remind 获取到期提醒配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: notice_remind.show（TYPE=data，真机 3.7.24 实证）。
func (s *NoticesService) Remind(ctx context.Context) (*NoticeRemindConfig, error) {
	var out struct {
		Data []NoticeRemindConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "notice_remind", "show",
		map[string]any{"TYPE": cfgTypeData}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetRemind 保存到期提醒配置（整行保存，随行携带 ID）。
// 对应 /Action/call: notice_remind.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NoticesService) SetRemind(ctx context.Context, cfg NoticeRemindConfig) error {
	_, err := s.c.Call(ctx, "notice_remind", "save", cfg)
	return err
}

// Expires 获取 PPPoE 到期通知配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: notice_expires.show（TYPE=data，真机 3.7.24 实证）。
func (s *NoticesService) Expires(ctx context.Context) (*NoticeExpiresConfig, error) {
	var out struct {
		Data []NoticeExpiresConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "notice_expires", "show",
		map[string]any{"TYPE": cfgTypeData}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetExpires 保存 PPPoE 到期通知配置（整行保存，随行携带 ID）。
// 对应 /Action/call: notice_expires.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NoticesService) SetExpires(ctx context.Context, cfg NoticeExpiresConfig) error {
	_, err := s.c.Call(ctx, "notice_expires", "save", cfg)
	return err
}
