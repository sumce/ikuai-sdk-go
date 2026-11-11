package behavior

import (
	"context"
	"encoding/json"

	"github.com/sumce/ikuai-sdk-go"
)

// domain_blacklist / domain_group 的 show 形态说明（真机 3.7.24 实测）：
// TYPE=total,data 标准分页可用；TYPE=all 额外返回表单选择器区块
// （domain_blacklist 含 dtgroup/ipgroup/macgroup；domain_group 含
// url_type 系统网址类型树与 sys_url）。

// DomainBlacklistRule 是一条域名黑名单规则行（domain_blacklist.show
// total,data 的 data 元素；固件页面语义为「禁止娱乐网站/网站类型控制」）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按开源 PyIKuaiClient 生产实证（add/edit 成功载荷）建模。
// 注意两个与其他 func 不同的拼写：本 func 的内网 IP 字段为 ipaddr（非
// ip_addr）、生效日字段为 weekdays（非 week）——固件 i18n 列名显示
// ip_addr/week，存在固件版本差异可能，待真机验证。
type DomainBlacklistRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// DomainGroup 网站类型/域名分组名（逗号分隔；值可参考
	// DomainGroupSelectors.URLType 的类型 ID）。字段语义待真机验证。
	DomainGroup string `json:"domain_group"`
	// IPAddr 内网 IP/IP 分组名（逗号分隔；注意固件字段拼写为 ipaddr）。
	// 字段语义待真机验证。
	IPAddr string `json:"ipaddr"`
	// Time 生效时段（如 "00:00-23:59"）。字段语义待真机验证。
	Time string `json:"time"`
	// Weekdays 生效日（"1234567"；注意固件字段拼写为 weekdays）。
	// 字段语义待真机验证。
	Weekdays string `json:"weekdays"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r DomainBlacklistRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// DomainCategory 是系统网址类型树的一个节点（domain_group.show TYPE=all
// 的 url_type 区块元素，真机 3.7.24 实测：9 个大类，每类挂 children 子类）。
type DomainCategory struct {
	// Name 类型名（如 "交通旅游"）。
	Name string `json:"name"`
	// ID 类型 ID（大类与大类名相同；子类为 "大类-子类" 形态，如
	// "交通旅游-旅游网站"，供 domain_blacklist 的 domain_group 字段引用）。
	ID string `json:"id"`
	// Children 子类型列表（叶子节点为空数组）。
	Children []DomainCategory `json:"children"`
}

// DomainGroupSelectors 是域名分组表单的可选值区块（domain_group.show
// TYPE=all，真机 3.7.24 实测区块 url_type/sys_url）。
type DomainGroupSelectors struct {
	// URLType 系统网址类型树（真机实测 9 大类：交通旅游/休闲娱乐/
	// 体育健身/医疗健康/新闻媒体/生活服务/论坛门户/购物网站/金融理财）。
	URLType []DomainCategory `json:"url_type"`
	// SysURL 系统网址区块（真机实测为空数组；元素结构无法核对，保留为
	// RawMessage，待真机验证）。
	SysURL []json.RawMessage `json:"sys_url"`
}

// DomainGroupRow 是一条用户自定义域名分组行（domain_group.show total,data
// 的 data 元素；固件页面语义为「网址库」——自定义网站类型分组，供
// domain_blacklist 的 domain_group 字段引用）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按同族分组表（macgroup 的 group_name/addr_pool/comment 形态，
// 开源 PyIKuaiClient 生产实证）与分组语义推导：group_name 分组名、
// domain 域名串（逗号分隔），空表推导，待真机验证。
type DomainGroupRow struct {
	// ID 分组行 ID。
	ID int64 `json:"id,omitempty"`
	// Name 分组名。空表推导，待真机验证。
	Name string `json:"group_name"`
	// Domain 域名串（逗号分隔，每行一条的表单录入后固件存储形态）。
	// 空表推导，待真机验证。
	Domain string `json:"domain"`
	// Comment 备注。空表推导，待真机验证。
	Comment string `json:"comment"`
}

// DomainFilterService 是域名过滤（domain_blacklist/domain_group）的子服务。
type DomainFilterService struct {
	c *ikuai.Client
}

// BlacklistList 获取域名黑名单规则列表（分页）。
// 对应 /Action/call: domain_blacklist.show（TYPE=total,data,
// limit="offset,rows"），
// 即 Web 控制台「行为管控 → 网页过滤 → 禁止指定网站」页面。
func (s *DomainFilterService) BlacklistList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DomainBlacklistRule], error) {
	resp, err := s.c.Call(ctx, "domain_blacklist", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DomainBlacklistRule](resp.Data)
}

// BlacklistSelectors 获取域名黑名单表单的可选值（IP/MAC/时间分组名列表）。
// 对应 /Action/call: domain_blacklist.show（TYPE=all）。
func (s *DomainFilterService) BlacklistSelectors(ctx context.Context) (*ACLSelectors, error) {
	var out ACLSelectors
	if err := s.c.CallData(ctx, "domain_blacklist", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BlacklistAdd 新增域名黑名单规则，返回固件分配的行 ID。
// 对应 /Action/call: domain_blacklist.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *DomainFilterService) BlacklistAdd(ctx context.Context, rule DomainBlacklistRule) (int64, error) {
	resp, err := s.c.Call(ctx, "domain_blacklist", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// BlacklistEdit 按 ID 修改域名黑名单规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: domain_blacklist.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *DomainFilterService) BlacklistEdit(ctx context.Context, rule DomainBlacklistRule) error {
	_, err := s.c.Call(ctx, "domain_blacklist", "edit", rule)
	return err
}

// BlacklistDelete 删除域名黑名单规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: domain_blacklist.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *DomainFilterService) BlacklistDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "domain_blacklist", "del", ikuai.IDParam(ids))
	return err
}

// GroupList 获取自定义域名分组列表（分页）。
// 对应 /Action/call: domain_group.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → 网页过滤 → 网址库」页面。
func (s *DomainFilterService) GroupList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DomainGroupRow], error) {
	resp, err := s.c.Call(ctx, "domain_group", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DomainGroupRow](resp.Data)
}

// GroupSelectors 获取网址库页的系统网址类型树与系统网址区块。
// 对应 /Action/call: domain_group.show（TYPE=all，真机实测 url_type 为
// 9 大类树形结构）。
func (s *DomainFilterService) GroupSelectors(ctx context.Context) (*DomainGroupSelectors, error) {
	var out DomainGroupSelectors
	if err := s.c.CallData(ctx, "domain_group", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GroupAdd 新增自定义域名分组，返回固件分配的行 ID。
// 对应 /Action/call: domain_group.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *DomainFilterService) GroupAdd(ctx context.Context, group DomainGroupRow) (int64, error) {
	resp, err := s.c.Call(ctx, "domain_group", "add", group)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// GroupEdit 按 ID 修改自定义域名分组（group.ID 需携带目标行 ID）。
// 对应 /Action/call: domain_group.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *DomainFilterService) GroupEdit(ctx context.Context, group DomainGroupRow) error {
	_, err := s.c.Call(ctx, "domain_group", "edit", group)
	return err
}

// GroupDelete 删除自定义域名分组，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: domain_group.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *DomainFilterService) GroupDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "domain_group", "del", ikuai.IDParam(ids))
	return err
}
