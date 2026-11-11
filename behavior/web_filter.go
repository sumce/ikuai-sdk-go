package behavior

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// url_black / url_redirect / url_keywords / url_replace 的 show 形态说明
// （真机 3.7.24 实测）：TYPE=total,data 标准分页可用；TYPE=all 额外返回
// 表单选择器区块（url_black 含 url_white_refer，其余含 pattern 匹配模式
// 候选值，ipgroup/macgroup/dtgroup 为分组名字符串数组）。

// url_black 控制模式字段（mode）的值（固件 i18n black_list.event 实证）。
const (
	// URLBlackModeBlack 黑名单模式（默认所有域名可访问，名单内不可访问）。
	URLBlackModeBlack int64 = 0
	// URLBlackModeWhite 白名单模式（默认所有域名不可访问，名单内可访问）。
	URLBlackModeWhite int64 = 1
)

// URL 匹配模式字段（mode）的值（固件 i18n url_jump/keywords_replace/
// param-replace 的 event 实证，真机 TYPE=all 的 pattern 区块同样回
// ["exact","vague"]）。
const (
	// URLMatchExact 精确匹配。
	URLMatchExact = "exact"
	// URLMatchVague 模糊匹配。
	URLMatchVague = "vague"
)

// URLBlackRule 是一条网页黑白名单规则行（url_black.show total,data 的
// data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；mode 的 0/1 语义按固件 i18n black_list.event 实证，ip_addr 字段
// 按开源 PyIKuaiClient 生产实证，domain 按 i18n 控制域名列建模，除 id 外
// 全部标注待真机验证。
type URLBlackRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Mode 控制模式：URLBlackModeBlack=黑名单 / URLBlackModeWhite=白名单
	// （固件 i18n 实证）。字段语义待真机验证。
	Mode int64 `json:"mode"`
	// Domain 控制域名（如 www.baidu.com，多值逗号分隔；仅黑名单模式）。
	// 字段语义待真机验证。
	Domain string `json:"domain"`
	// IPAddr 内网 IP/IP 分组名（逗号分隔）。字段语义待真机验证。
	IPAddr string `json:"ip_addr"`
	// Time 生效时段（如 "00:00-23:59"）。字段语义待真机验证。
	Time string `json:"time"`
	// Week 生效日（"1234567"）。字段语义待真机验证。
	Week string `json:"week"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r URLBlackRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// WebFilterSelectors 是网页过滤表单的可选值区块（url_black.show TYPE=all
// 的区块，真机 3.7.24 实测）。
type WebFilterSelectors struct {
	// IPGroup 可选 IPv4 分组名列表（真机实测为分组名字符串数组）。
	IPGroup []string `json:"ipgroup"`
	// MACGroup 可选 MAC 分组名列表（真机实测为空数组）。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表（真机实测为空数组）。
	DTGroup []string `json:"dtgroup"`
	// URLWhiteRefer 网页白名单 refer 开关（真机观测 0；语义待真机验证）。
	URLWhiteRefer int64 `json:"url_white_refer"`
}

// URLRedirectRule 是一条网页重定向规则行（url_redirect.show total,data 的
// data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按固件 i18n url_jump 列名（comment 标题/mode 匹配模式/
// src_url/dst_url/excluded 排除值/hit_rate 跳转率/prio 匹配优先级/week/
// time）建模，除 id 外全部标注待真机验证。
type URLRedirectRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 标题（i18n 列名 comment=标题）。字段语义待真机验证。
	Comment string `json:"comment"`
	// Mode 匹配模式：URLMatchExact=精确 / URLMatchVague=模糊（固件
	// i18n 实证）。字段语义待真机验证。
	Mode string `json:"mode"`
	// SrcURL 源地址（被匹配的 URL）。字段语义待真机验证。
	SrcURL string `json:"src_url"`
	// DstURL 目的地址（跳转目标，不能与源地址相同）。字段语义待真机验证。
	DstURL string `json:"dst_url"`
	// Excluded 排除值（URL 中包含此字段时不跳转）。字段语义待真机验证。
	Excluded string `json:"excluded"`
	// HitRate 跳转率（按百分比跳转至指定页面）。字段语义待真机验证。
	HitRate int64 `json:"hit_rate"`
	// Prio 匹配优先级（1~63，数值越小优先级越高）。字段语义待真机验证。
	Prio int64 `json:"prio"`
	// Time 生效时段（如 "00:00-08:00,09:00-11:00"）。字段语义待真机验证。
	Time string `json:"time"`
	// Week 生效日（"1234567"）。字段语义待真机验证。
	Week string `json:"week"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r URLRedirectRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// URLKeywordRule 是一条网页关键字替换规则行（url_keywords.show total,data
// 的 data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按固件 i18n keywords_replace 列名（ori_keyword 替换关键字/
// rep_keyword 替换值）建模，除 id 外全部标注待真机验证。
type URLKeywordRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 标题。字段语义待真机验证。
	Comment string `json:"comment"`
	// Mode 匹配模式：URLMatchExact=精确 / URLMatchVague=模糊。字段语义
	// 待真机验证。
	Mode string `json:"mode"`
	// SrcURL 源地址（被匹配的 URL）。字段语义待真机验证。
	SrcURL string `json:"src_url"`
	// OriKeyword 替换关键字（页面中被替换的文本）。字段语义待真机验证。
	OriKeyword string `json:"ori_keyword"`
	// RepKeyword 替换值（替换后的文本，中文仅支持 UTF-8 编码网站）。
	// 字段语义待真机验证。
	RepKeyword string `json:"rep_keyword"`
	// Excluded 排除值（URL 中包含此字段时不替换）。字段语义待真机验证。
	Excluded string `json:"excluded"`
	// HitRate 替换率（按百分比生效）。字段语义待真机验证。
	HitRate int64 `json:"hit_rate"`
	// Prio 匹配优先级（数值越小优先级越高）。字段语义待真机验证。
	Prio int64 `json:"prio"`
	// Time 生效时段。字段语义待真机验证。
	Time string `json:"time"`
	// Week 生效日（"1234567"）。字段语义待真机验证。
	Week string `json:"week"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r URLKeywordRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// URLReplaceRule 是一条网页参数替换规则行（url_replace.show total,data 的
// data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按固件 i18n param-replace 列名（param_keyword 主键/
// rep_keyword 替换值）建模，除 id 外全部标注待真机验证。
type URLReplaceRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 标题。字段语义待真机验证。
	Comment string `json:"comment"`
	// Mode 匹配模式：URLMatchExact=精确 / URLMatchVague=模糊。字段语义
	// 待真机验证。
	Mode string `json:"mode"`
	// SrcURL 源地址（被匹配的 URL）。字段语义待真机验证。
	SrcURL string `json:"src_url"`
	// ParamKeyword 主键（URL 参数名）。字段语义待真机验证。
	ParamKeyword string `json:"param_keyword"`
	// RepKeyword 替换值。字段语义待真机验证。
	RepKeyword string `json:"rep_keyword"`
	// Excluded 排除值（URL 中包含此字段时不替换）。字段语义待真机验证。
	Excluded string `json:"excluded"`
	// HitRate 替换率（按百分比生效）。字段语义待真机验证。
	HitRate int64 `json:"hit_rate"`
	// Prio 匹配优先级（数值越小优先级越高）。字段语义待真机验证。
	Prio int64 `json:"prio"`
	// Time 生效时段。字段语义待真机验证。
	Time string `json:"time"`
	// Week 生效日（"1234567"）。字段语义待真机验证。
	Week string `json:"week"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r URLReplaceRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// WebFilterService 是网页过滤（url_black/url_redirect/url_keywords/
// url_replace）的子服务。
type WebFilterService struct {
	c *ikuai.Client
}

// BlackList 获取网页黑白名单规则列表（分页）。
// 对应 /Action/call: url_black.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → 网页过滤 → 黑名单」页面。
func (s *WebFilterService) BlackList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[URLBlackRule], error) {
	resp, err := s.c.Call(ctx, "url_black", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[URLBlackRule](resp.Data)
}

// BlackSelectors 获取网页黑白名单表单的可选值（IP/MAC/时间分组名列表与
// 白名单 refer 开关）。
// 对应 /Action/call: url_black.show（TYPE=all）。
func (s *WebFilterService) BlackSelectors(ctx context.Context) (*WebFilterSelectors, error) {
	var out WebFilterSelectors
	if err := s.c.CallData(ctx, "url_black", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BlackAdd 新增网页黑白名单规则，返回固件分配的行 ID。
// 对应 /Action/call: url_black.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) BlackAdd(ctx context.Context, rule URLBlackRule) (int64, error) {
	resp, err := s.c.Call(ctx, "url_black", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// BlackEdit 按 ID 修改网页黑白名单规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: url_black.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) BlackEdit(ctx context.Context, rule URLBlackRule) error {
	_, err := s.c.Call(ctx, "url_black", "edit", rule)
	return err
}

// BlackDelete 删除网页黑白名单规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: url_black.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *WebFilterService) BlackDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "url_black", "del", ikuai.IDParam(ids))
	return err
}

// RedirectList 获取网页重定向规则列表（分页）。
// 对应 /Action/call: url_redirect.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → 网页过滤 → 网页重定向」页面。
func (s *WebFilterService) RedirectList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[URLRedirectRule], error) {
	resp, err := s.c.Call(ctx, "url_redirect", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[URLRedirectRule](resp.Data)
}

// RedirectSelectors 获取网页重定向表单的可选值（IP/MAC/时间分组名列表与
// 匹配模式候选值）。
// 对应 /Action/call: url_redirect.show（TYPE=all，真机实测 pattern 区块）。
func (s *WebFilterService) RedirectSelectors(ctx context.Context) (*ACLSelectors, error) {
	return s.groupSelectors(ctx, "url_redirect")
}

// RedirectAdd 新增网页重定向规则，返回固件分配的行 ID。
// 对应 /Action/call: url_redirect.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) RedirectAdd(ctx context.Context, rule URLRedirectRule) (int64, error) {
	resp, err := s.c.Call(ctx, "url_redirect", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// RedirectEdit 按 ID 修改网页重定向规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: url_redirect.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) RedirectEdit(ctx context.Context, rule URLRedirectRule) error {
	_, err := s.c.Call(ctx, "url_redirect", "edit", rule)
	return err
}

// RedirectDelete 删除网页重定向规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: url_redirect.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *WebFilterService) RedirectDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "url_redirect", "del", ikuai.IDParam(ids))
	return err
}

// KeywordsList 获取网页关键字替换规则列表（分页）。
// 对应 /Action/call: url_keywords.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → 网页过滤 → 关键字替换」页面。
func (s *WebFilterService) KeywordsList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[URLKeywordRule], error) {
	resp, err := s.c.Call(ctx, "url_keywords", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[URLKeywordRule](resp.Data)
}

// KeywordsSelectors 获取关键字替换表单的可选值（IP/MAC/时间分组名列表与
// 匹配模式候选值）。
// 对应 /Action/call: url_keywords.show（TYPE=all，真机实测 pattern 区块）。
func (s *WebFilterService) KeywordsSelectors(ctx context.Context) (*ACLSelectors, error) {
	return s.groupSelectors(ctx, "url_keywords")
}

// KeywordsAdd 新增关键字替换规则，返回固件分配的行 ID。
// 对应 /Action/call: url_keywords.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) KeywordsAdd(ctx context.Context, rule URLKeywordRule) (int64, error) {
	resp, err := s.c.Call(ctx, "url_keywords", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// KeywordsEdit 按 ID 修改关键字替换规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: url_keywords.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) KeywordsEdit(ctx context.Context, rule URLKeywordRule) error {
	_, err := s.c.Call(ctx, "url_keywords", "edit", rule)
	return err
}

// KeywordsDelete 删除关键字替换规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: url_keywords.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *WebFilterService) KeywordsDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "url_keywords", "del", ikuai.IDParam(ids))
	return err
}

// ReplaceList 获取网页参数替换规则列表（分页）。
// 对应 /Action/call: url_replace.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → 网页过滤 → 参数替换」页面。
func (s *WebFilterService) ReplaceList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[URLReplaceRule], error) {
	resp, err := s.c.Call(ctx, "url_replace", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[URLReplaceRule](resp.Data)
}

// ReplaceSelectors 获取参数替换表单的可选值（IP/MAC/时间分组名列表与
// 匹配模式候选值）。
// 对应 /Action/call: url_replace.show（TYPE=all，真机实测 pattern 区块）。
func (s *WebFilterService) ReplaceSelectors(ctx context.Context) (*ACLSelectors, error) {
	return s.groupSelectors(ctx, "url_replace")
}

// ReplaceAdd 新增参数替换规则，返回固件分配的行 ID。
// 对应 /Action/call: url_replace.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) ReplaceAdd(ctx context.Context, rule URLReplaceRule) (int64, error) {
	resp, err := s.c.Call(ctx, "url_replace", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// ReplaceEdit 按 ID 修改参数替换规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: url_replace.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *WebFilterService) ReplaceEdit(ctx context.Context, rule URLReplaceRule) error {
	_, err := s.c.Call(ctx, "url_replace", "edit", rule)
	return err
}

// ReplaceDelete 删除参数替换规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: url_replace.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *WebFilterService) ReplaceDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "url_replace", "del", ikuai.IDParam(ids))
	return err
}

// groupSelectors 解码 url_redirect/url_keywords/url_replace 共用的分组名
// 选择器区块（真机 TYPE=all 实测：ipgroup/macgroup/dtgroup + pattern）。
func (s *WebFilterService) groupSelectors(ctx context.Context, funcName string) (*ACLSelectors, error) {
	var out ACLSelectors
	if err := s.c.CallData(ctx, funcName, "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
