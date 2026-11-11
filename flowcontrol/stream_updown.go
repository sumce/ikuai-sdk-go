package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// UpDownRule 是一条流量上下线分流规则行（stream_updown.show total,data
// 的 data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；stream_updown.show TYPE=all 实测携带 protocol 选择器区块，说明
// 规则行含协议维度，字段名按同族分流规则行（stream_ipport 实测）保守
// 推断，除 id 外全部标注待真机验证。
type UpDownRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
	// Interface 出接口名。字段语义待真机验证。
	Interface string `json:"interface"`
	// Protocol 协议（值域参考 TYPE=all 的 protocol 区块：
	// any/tcp/udp/tcp+udp/icmp）。字段语义待真机验证。
	Protocol string `json:"protocol"`
	// SrcAddr 源地址串（IP/IP 分组名，逗号分隔）。字段语义待真机验证。
	SrcAddr string `json:"src_addr"`
	// DstAddr 目的地址串。字段语义待真机验证。
	DstAddr string `json:"dst_addr"`
	// SrcPort 源端口串。字段语义待真机验证。
	SrcPort string `json:"src_port"`
	// DstPort 目的端口串。字段语义待真机验证。
	DstPort string `json:"dst_port"`
	// Week 生效日（"1234567" = 周一至周日）。字段语义待真机验证。
	Week string `json:"week"`
	// Time 生效时段（如 "00:00-23:59"）。字段语义待真机验证。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r UpDownRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// UpDownSelectors 是流量上下线分流表单的可选值区块（stream_updown.show
// TYPE=all 的区块，真机 3.7.24 实测）。
type UpDownSelectors struct {
	// Interface 可选接口列表，每项 [接口名, 备注]。
	Interface [][]string `json:"interface"`
	// IPGroup 可选 IPv4 分组名列表。
	IPGroup []string `json:"ipgroup"`
	// MACGroup 可选 MAC 分组名列表。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表。
	DTGroup []string `json:"dtgroup"`
	// Protocol 可选协议列表（实测 ["tcp","udp","tcp+udp","icmp","any"]）。
	Protocol []string `json:"protocol"`
}

// UpDownService 是流量上下线分流（stream_updown）的子服务。
type UpDownService struct {
	c *ikuai.Client
}

// List 获取流量上下线分流规则列表（分页）。
// 对应 /Action/call: stream_updown.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「网络设置 → 流控分流 → 流量上下线分流」页面。
func (s *UpDownService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[UpDownRule], error) {
	resp, err := s.c.Call(ctx, "stream_updown", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[UpDownRule](resp.Data)
}

// Selectors 获取流量上下线分流表单的可选值（接口/分组/协议等）。
// 对应 /Action/call: stream_updown.show（TYPE=all）。
func (s *UpDownService) Selectors(ctx context.Context) (*UpDownSelectors, error) {
	var out UpDownSelectors
	if err := s.c.CallData(ctx, "stream_updown", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增流量上下线分流规则，返回固件分配的行 ID。
// 对应 /Action/call: stream_updown.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *UpDownService) Add(ctx context.Context, rule UpDownRule) (int64, error) {
	resp, err := s.c.Call(ctx, "stream_updown", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改流量上下线分流规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: stream_updown.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *UpDownService) Edit(ctx context.Context, rule UpDownRule) error {
	_, err := s.c.Call(ctx, "stream_updown", "edit", rule)
	return err
}

// Delete 删除流量上下线分流规则，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: stream_updown.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *UpDownService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "stream_updown", "del", ikuai.IDParam(ids))
	return err
}
