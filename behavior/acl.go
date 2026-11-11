package behavior

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// acl_mac.show 的 TYPE 形态说明（真机 3.7.24 实测）：TYPE=total,data 标准
// 分页可用；TYPE=all 额外返回表单选择器区块（实测为标量 0，无区块内容，
// 故 MAC 规则不提供 Selectors 方法）。
//
// acl_l7 / acl_qq / acl_l2route 的 TYPE=total,data 同样可用，TYPE=all 额外
// 返回 dtgroup/ipgroup/macgroup 分组名选择器区块（真机实测 ipgroup 为分组
// 名字符串数组）。

// ACLSelectors 是访问控制族表单的可选值区块（acl_l7/acl_qq/acl_l2route
// 等 show TYPE=all 的区块，真机 3.7.24 实测）。
type ACLSelectors struct {
	// IPGroup 可选 IPv4 分组名列表（真机实测为分组名字符串数组）。
	IPGroup []string `json:"ipgroup"`
	// MACGroup 可选 MAC 分组名列表（真机实测为空数组）。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表（真机实测为空数组）。
	DTGroup []string `json:"dtgroup"`
}

// acl_l7 动作字段（action）的值（固件 i18n pro_control.event 实证，
// 与开源 PyIKuaiClient 生产实证一致）。
const (
	// L7ActionAccept 允许（优先级默认 31，固件 i18n 实证）。
	L7ActionAccept = "accept"
	// L7ActionDrop 阻断（优先级默认 32，固件 i18n 实证）。
	L7ActionDrop = "drop"
)

// MACRule 是一条 MAC 访问控制规则行（acl_mac.show total,data 的 data 元素；
// 行字段为真机 3.7.24 实测）。
type MACRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// MAC 受控终端 MAC 地址（如 "6c:4b:90:b0:20:ca"）。
	MAC string `json:"mac"`
	// Comment 备注（真机观测 "滑子豪"）。
	Comment string `json:"comment"`
	// Comment1 第二备注列（真机行实测存在键 "comment:1"，观测为空串；
	// 语义待真机验证，疑似分组/展示列）。
	Comment1 string `json:"comment:1"`
	// Enabled 是否启用（"yes"/"no"，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// Time 生效时段（真机观测 "00:00-23:59"）。
	Time string `json:"time"`
	// Week 生效日（真机观测 "1234567" = 周一至周日）。
	Week string `json:"week"`
	// Expires 到期时间（真机观测 0；0=永不过期 推断，语义待真机验证）。
	Expires int64 `json:"expires"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r MACRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// L7Rule 是一条七层协议控制规则行（acl_l7.show total,data 的 data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按开源 PyIKuaiClient 生产实证（add/edit 成功载荷）与固件
// i18n pro_control 列名（terminal/app_proto/action/src_addr/dst_addr/week/
// time/优先级 0-63）建模，除 id 外全部标注待真机验证。
type L7Rule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Action 动作：L7ActionAccept=允许 / L7ActionDrop=阻断。字段语义待真机验证。
	Action string `json:"action"`
	// AppProto 七层协议名（多个逗号分隔）。字段语义待真机验证。
	AppProto string `json:"app_proto"`
	// SrcAddr 源地址（内网 IP/IP 分组名，逗号分隔）。字段语义待真机验证。
	SrcAddr string `json:"src_addr"`
	// DstAddr 目的地址（可空）。字段语义待真机验证。
	DstAddr string `json:"dst_addr"`
	// Prio 优先级：0(最高)-63(最低)，允许默认 31、阻断默认 32（固件
	// i18n 实证）；数值相同时先匹配先配置的策略。字段语义待真机验证。
	Prio int64 `json:"prio"`
	// Time 生效时段（如 "00:00-23:59"）。字段语义待真机验证。
	Time string `json:"time"`
	// Week 生效日（"1234567"）。字段语义待真机验证。
	Week string `json:"week"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r L7Rule) IsEnabled() bool { return r.Enabled == EnabledYes }

// L2RouteConfig 是二层访问控制（禁止二级路由）的单行配置
// （acl_l2route.show 的 data 元素；行字段为真机 3.7.24 实测，观测恒为
// 单行 id=1）。
type L2RouteConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// Nol2rt 禁止二级路由开关（观测 0；1=开启 推断，语义待真机验证）。
	Nol2rt int64 `json:"nol2rt"`
	// Nol2rtIP 允许 IP 列表/例外地址范围（固件 i18n secondaryroute 的
	// ipList/nol2rt_ip 列名实证；逗号分隔，格式如 192.168.100.1/24）。
	Nol2rtIP string `json:"nol2rt_ip"`
	// Times 禁止时间（观测 "00:00-23:59"；i18n ForbidTime）。
	Times string `json:"times"`
	// TTLNum 自定义 TTL 值（观测 1；i18n customTtL，语义待真机验证）。
	TTLNum int64 `json:"ttl_num"`
}

// IsEnabled 禁止二级路由是否开启（非 0 视为开启，语义待真机验证）。
func (c L2RouteConfig) IsEnabled() bool { return c.Nol2rt != 0 }

// acl_qq 控制模式字段（mode）的值（固件 i18n qq_control.event 实证）。
const (
	// QQModeBlack 黑名单模式（默认所有 QQ 可登录，名单内不允许）。
	QQModeBlack int64 = 0
	// QQModeWhite 白名单模式（针对设置的 IP，仅名单内 QQ 可登录）。
	QQModeWhite int64 = 1
)

// QQRule 是一条 QQ 准入控制规则行（acl_qq.show total,data 的 data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段按固件 i18n qq_control 列名（mode/qq/ip_addr/week/time）与
// 访问控制族通用字段（enabled/comment）建模，除 id 外全部标注待真机验证。
type QQRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Mode 控制模式：QQModeBlack=黑名单 / QQModeWhite=白名单（固件
	// i18n 实证）。字段语义待真机验证。
	Mode int64 `json:"mode"`
	// QQ QQ 号码（多个逗号分隔；i18n 格式 "123456789 备注" 每行一条）。
	// 字段语义待真机验证。
	QQ string `json:"qq"`
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
func (r QQRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// ACLService 是访问控制族（acl_mac/acl_l7/acl_l2route/acl_qq）的子服务。
type ACLService struct {
	c *ikuai.Client
}

// MACList 获取 MAC 访问控制规则列表（分页）。
// 对应 /Action/call: acl_mac.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → MAC 访问控制」页面。
func (s *ACLService) MACList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[MACRule], error) {
	resp, err := s.c.Call(ctx, "acl_mac", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[MACRule](resp.Data)
}

// MACAdd 新增 MAC 访问控制规则，返回固件分配的行 ID。
// 对应 /Action/call: acl_mac.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) MACAdd(ctx context.Context, rule MACRule) (int64, error) {
	resp, err := s.c.Call(ctx, "acl_mac", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// MACEdit 按 ID 修改 MAC 访问控制规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: acl_mac.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) MACEdit(ctx context.Context, rule MACRule) error {
	_, err := s.c.Call(ctx, "acl_mac", "edit", rule)
	return err
}

// MACDelete 删除 MAC 访问控制规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: acl_mac.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) MACDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "acl_mac", "del", ikuai.IDParam(ids))
	return err
}

// L7List 获取七层协议控制规则列表（分页）。
// 对应 /Action/call: acl_l7.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → 应用协议控制（专业管控）」页面。
func (s *ACLService) L7List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[L7Rule], error) {
	resp, err := s.c.Call(ctx, "acl_l7", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[L7Rule](resp.Data)
}

// L7Selectors 获取七层协议控制表单的可选值（IP/MAC/时间分组名列表）。
// 对应 /Action/call: acl_l7.show（TYPE=all）。
func (s *ACLService) L7Selectors(ctx context.Context) (*ACLSelectors, error) {
	var out ACLSelectors
	if err := s.c.CallData(ctx, "acl_l7", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// L7Add 新增七层协议控制规则，返回固件分配的行 ID。
// 对应 /Action/call: acl_l7.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *ACLService) L7Add(ctx context.Context, rule L7Rule) (int64, error) {
	resp, err := s.c.Call(ctx, "acl_l7", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// L7Edit 按 ID 修改七层协议控制规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: acl_l7.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *ACLService) L7Edit(ctx context.Context, rule L7Rule) error {
	_, err := s.c.Call(ctx, "acl_l7", "edit", rule)
	return err
}

// L7Delete 删除七层协议控制规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: acl_l7.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) L7Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "acl_l7", "del", ikuai.IDParam(ids))
	return err
}

// L2RouteGet 获取二层访问控制（禁止二级路由）配置（观测恒为单行）。
// 对应 /Action/call: acl_l2route.show（无 TYPE，param={}，真机实测；
// TYPE=total,data 回 30006 unknown TYPE (total)），
// 即 Web 控制台「行为管控 → 二级路由」页面。
func (s *ACLService) L2RouteGet(ctx context.Context) (*L2RouteConfig, error) {
	var out struct {
		Data []L2RouteConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "acl_l2route", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// L2RouteSelectors 获取二层访问控制表单的可选值（IP/MAC/时间分组名列表）。
// 对应 /Action/call: acl_l2route.show（TYPE=all）。
func (s *ACLService) L2RouteSelectors(ctx context.Context) (*ACLSelectors, error) {
	var out ACLSelectors
	if err := s.c.CallData(ctx, "acl_l2route", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// L2RouteSet 保存二层访问控制配置（写动作）。
// 对应 /Action/call: acl_l2route.edit（行携带 id，随 JSON 自动提交；
// action 名按单行配置 func 惯例推断，无抓包实证）。
// cfg 应为 L2RouteGet 返回并修改后的完整行（固件按整行覆盖）。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) L2RouteSet(ctx context.Context, cfg L2RouteConfig) error {
	_, err := s.c.Call(ctx, "acl_l2route", "edit", cfg)
	return err
}

// QQList 获取 QQ 准入控制规则列表（分页）。
// 对应 /Action/call: acl_qq.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「行为管控 → QQ 控制」页面。
func (s *ACLService) QQList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[QQRule], error) {
	resp, err := s.c.Call(ctx, "acl_qq", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[QQRule](resp.Data)
}

// QQSelectors 获取 QQ 准入控制表单的可选值（IP/MAC/时间分组名列表）。
// 对应 /Action/call: acl_qq.show（TYPE=all）。
func (s *ACLService) QQSelectors(ctx context.Context) (*ACLSelectors, error) {
	var out ACLSelectors
	if err := s.c.CallData(ctx, "acl_qq", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QQAdd 新增 QQ 准入控制规则，返回固件分配的行 ID。
// 对应 /Action/call: acl_qq.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *ACLService) QQAdd(ctx context.Context, rule QQRule) (int64, error) {
	resp, err := s.c.Call(ctx, "acl_qq", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// QQEdit 按 ID 修改 QQ 准入控制规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: acl_qq.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *ACLService) QQEdit(ctx context.Context, rule QQRule) error {
	_, err := s.c.Call(ctx, "acl_qq", "edit", rule)
	return err
}

// QQDelete 删除 QQ 准入控制规则，支持单个或多个 ID（内部自动选
// {"id"} / {"ids"}）。
// 对应 /Action/call: acl_qq.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) QQDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "acl_qq", "del", ikuai.IDParam(ids))
	return err
}
