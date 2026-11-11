package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Layer7Rule 是一条协议分流规则行（stream_layer7.show total,data 的
// data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段名按同族分流规则行（stream_ipport 实测）与行为管控
// acl_l7 的 app_proto 字段（开源 3.x 实现实证）保守推断，除 id 外
// 全部标注待真机验证。
type Layer7Rule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
	// Interface 出接口名。字段语义待真机验证。
	Interface string `json:"interface"`
	// AppProto 七层应用协议名。字段语义待真机验证。
	AppProto string `json:"app_proto"`
	// SrcAddr 源地址串（IP/IP 分组名，逗号分隔）。字段语义待真机验证。
	SrcAddr string `json:"src_addr"`
	// DstAddr 目的地址串。字段语义待真机验证。
	DstAddr string `json:"dst_addr"`
	// Week 生效日（"1234567" = 周一至周日）。字段语义待真机验证。
	Week string `json:"week"`
	// Time 生效时段（如 "00:00-23:59"）。字段语义待真机验证。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r Layer7Rule) IsEnabled() bool { return r.Enabled == EnabledYes }

// Layer7Selectors 是协议分流表单的可选值区块（stream_layer7.show TYPE=all
// 的区块，真机 3.7.24 实测；syn_proxy 为页面级开关状态值）。
type Layer7Selectors struct {
	// Interface 可选接口列表，每项 [接口名, 备注]。
	Interface [][]string `json:"interface"`
	// IPGroup 可选 IPv4 分组名列表。
	IPGroup []string `json:"ipgroup"`
	// MACGroup 可选 MAC 分组名列表。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表。
	DTGroup []string `json:"dtgroup"`
	// SynProxy 页面级防代理（syn_proxy）状态值（实测 0；语义待真机验证）。
	SynProxy int64 `json:"syn_proxy"`
}

// Layer7Service 是协议分流（stream_layer7）的子服务。
type Layer7Service struct {
	c *ikuai.Client
}

// List 获取协议分流规则列表（分页）。
// 对应 /Action/call: stream_layer7.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「网络设置 → 流控分流 → 协议分流」页面。
func (s *Layer7Service) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[Layer7Rule], error) {
	resp, err := s.c.Call(ctx, "stream_layer7", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[Layer7Rule](resp.Data)
}

// Selectors 获取协议分流表单的可选值（接口/IP/MAC/时间分组等）。
// 对应 /Action/call: stream_layer7.show（TYPE=all）。
func (s *Layer7Service) Selectors(ctx context.Context) (*Layer7Selectors, error) {
	var out Layer7Selectors
	if err := s.c.CallData(ctx, "stream_layer7", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增协议分流规则，返回固件分配的行 ID。
// 对应 /Action/call: stream_layer7.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *Layer7Service) Add(ctx context.Context, rule Layer7Rule) (int64, error) {
	resp, err := s.c.Call(ctx, "stream_layer7", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改协议分流规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: stream_layer7.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *Layer7Service) Edit(ctx context.Context, rule Layer7Rule) error {
	_, err := s.c.Call(ctx, "stream_layer7", "edit", rule)
	return err
}

// Delete 删除协议分流规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: stream_layer7.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *Layer7Service) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "stream_layer7", "del", ikuai.IDParam(ids))
	return err
}
