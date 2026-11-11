package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// DomainRule 是一条域名分流规则行（stream_domain.show total,data 的
// data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空；字段名取自开源 3.x
// 实现（ztc1997/ikuai-bypass 生产使用的扁平字段形态），语义与真机一致
// 但行数据未经本机核对，标注待真机验证。
type DomainRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
	// Interface 出接口名。字段语义待真机验证。
	Interface string `json:"interface"`
	// Domain 目标域名串（逗号分隔，支持通配）。字段语义待真机验证。
	Domain string `json:"domain"`
	// SrcAddr 源地址串（IP/IP 分组名，逗号分隔）。字段语义待真机验证。
	SrcAddr string `json:"src_addr"`
	// Week 生效日（"1234567" = 周一至周日）。字段语义待真机验证。
	Week string `json:"week"`
	// Time 生效时段（如 "00:00-23:59"）。字段语义待真机验证。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r DomainRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// DomainSelectors 是域名分流表单的可选值区块（stream_domain.show TYPE=all
// 的区块，真机 3.7.24 实测）。
type DomainSelectors struct {
	// Interface 可选接口列表，每项 [接口名, 备注]。
	Interface [][]string `json:"interface"`
	// IPGroup 可选 IPv4 分组名列表。
	IPGroup []string `json:"ipgroup"`
	// MACGroup 可选 MAC 分组名列表。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表。
	DTGroup []string `json:"dtgroup"`
}

// DomainService 是域名分流（stream_domain）的子服务。
type DomainService struct {
	c *ikuai.Client
}

// List 获取域名分流规则列表（分页）。
// 对应 /Action/call: stream_domain.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「网络设置 → 流控分流 → 域名分流」页面。
func (s *DomainService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DomainRule], error) {
	resp, err := s.c.Call(ctx, "stream_domain", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DomainRule](resp.Data)
}

// Selectors 获取域名分流表单的可选值（接口/IP/MAC/时间分组等）。
// 对应 /Action/call: stream_domain.show（TYPE=all）。
func (s *DomainService) Selectors(ctx context.Context) (*DomainSelectors, error) {
	var out DomainSelectors
	if err := s.c.CallData(ctx, "stream_domain", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增域名分流规则，返回固件分配的行 ID。
// 对应 /Action/call: stream_domain.add（开源 3.x 实现为扁平字段形态）。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *DomainService) Add(ctx context.Context, rule DomainRule) (int64, error) {
	resp, err := s.c.Call(ctx, "stream_domain", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改域名分流规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: stream_domain.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *DomainService) Edit(ctx context.Context, rule DomainRule) error {
	_, err := s.c.Call(ctx, "stream_domain", "edit", rule)
	return err
}

// Delete 删除域名分流规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: stream_domain.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *DomainService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "stream_domain", "del", ikuai.IDParam(ids))
	return err
}
