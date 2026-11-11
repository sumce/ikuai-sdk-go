package security

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// 连接数限制的协议枚举值（i18n firewall.conn 区块 protocol/any）。
const (
	// ConnLimitProtocolAny 任意协议（i18n conn.any）。
	ConnLimitProtocolAny = "any"
	// ConnLimitProtocolTCP TCP。
	ConnLimitProtocolTCP = "tcp"
	// ConnLimitProtocolUDP UDP。
	ConnLimitProtocolUDP = "udp"
)

// ConnLimitRule 是一条连接数限制规则行（conn_limit.show total,data 的
// data 元素）。
//
// 空表推导，待真机验证：真机 3.7.24 的连接数限制列表为空（total=0），
// 行字段无法实证。字段名按固件 i18n（firewall.conn 区块：ip_addr「内网IP」/
// wan_port「外网端口」/limits「连接数」/protocol「协议」）与表单语义推导。
type ConnLimitRule struct {
	// ID 规则行 ID（add 时无需填写；edit/del 必填）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"；推导值，待真机验证）。
	Enabled string `json:"enabled"`
	// Comment 备注（推导字段，待真机验证）。
	Comment string `json:"comment"`
	// IPAddr 内网IP（i18n conn.ip_addr；可为单 IP/网段/IP 分组名，分组
	// 可选项见 conn_limit.show TYPE=all 的 ipgroup 区块）。
	IPAddr string `json:"ip_addr"`
	// Protocol 协议（i18n conn.protocol；枚举 ConnLimitProtocol*，空/any=任意）。
	Protocol string `json:"protocol"`
	// WANPort 外网端口（i18n conn.wan_port；"any"/空串=任意，端口段形态
	// 推导同 ACL 端口串，待真机验证）。
	WANPort string `json:"wan_port"`
	// Limits 连接数上限（i18n conn.limits「连接数」；单位：个）。
	Limits int64 `json:"limits"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"（推导，待真机验证）。
func (r ConnLimitRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// ConnLimitSelectors 是连接数限制表单的可选值区块（conn_limit.show
// TYPE=all 的区块，真机 3.7.24 实测）。
type ConnLimitSelectors struct {
	// IPGroup 可选 IPv4 分组名列表（真机实测有值）。
	IPGroup []string `json:"ipgroup"`
}

// ConnLimitService 是连接数限制（conn_limit）的子服务。
type ConnLimitService struct {
	c *ikuai.Client
}

// List 获取连接数限制规则列表（分页）。
// 对应 /Action/call: conn_limit.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「安全设置 → 连接数限制」页面。
func (s *ConnLimitService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ConnLimitRule], error) {
	resp, err := s.c.Call(ctx, "conn_limit", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ConnLimitRule](resp.Data)
}

// Selectors 获取连接数限制表单的可选值（IP 分组）。
// 对应 /Action/call: conn_limit.show（TYPE=all）。
func (s *ConnLimitService) Selectors(ctx context.Context) (*ConnLimitSelectors, error) {
	var out ConnLimitSelectors
	if err := s.c.CallData(ctx, "conn_limit", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增连接数限制规则，返回固件分配的行 ID。
// 对应 /Action/call: conn_limit.add。
// ⚠️ 写动作，字段集待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *ConnLimitService) Add(ctx context.Context, rule ConnLimitRule) (int64, error) {
	resp, err := s.c.Call(ctx, "conn_limit", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改连接数限制规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: conn_limit.edit。
// ⚠️ 写动作，字段集待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *ConnLimitService) Edit(ctx context.Context, rule ConnLimitRule) error {
	_, err := s.c.Call(ctx, "conn_limit", "edit", rule)
	return err
}

// Delete 删除连接数限制规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: conn_limit.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *ConnLimitService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "conn_limit", "del", ikuai.IDParam(ids))
	return err
}
