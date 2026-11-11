package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// IPPortRule 是一条端口分流规则行（stream_ipport.show total,data 的
// data 元素，真机 3.7.24 实测 9 行）。
type IPPortRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，实测值）。
	Enabled string `json:"enabled"`
	// Comment 备注（实测样例 "出租上网"）。
	Comment string `json:"comment"`
	// Nexthop 下一跳网关（实测为空串；语义待真机验证）。
	Nexthop string `json:"nexthop"`
	// Interface 出接口名（实测支持逗号分隔多接口，如
	// "adsl03,adsl04,adsl05"）。
	Interface string `json:"interface"`
	// Mode 分流模式（实测值 0；编码语义待真机验证）。
	Mode int64 `json:"mode"`
	// Type 规则类型（实测值 0；编码语义待真机验证）。
	Type int64 `json:"type"`
	// SrcAddr 源地址串（实测混用 IP 与 IP 分组名，逗号分隔，
	// 如 "监控,运营,192.168.2.1/24"）。
	SrcAddr string `json:"src_addr"`
	// DstAddr 目的地址串（实测为空表示不限制）。
	DstAddr string `json:"dst_addr"`
	// Protocol 协议（实测值域 any/tcp/udp/tcp+udp/icmp，见
	// IPPortSelectors.Protocol）。
	Protocol string `json:"protocol"`
	// SrcPort 源端口串（如 "1000-2000,3000"；空表示不限制）。
	SrcPort string `json:"src_port"`
	// DstPort 目的端口串（空表示不限制）。
	DstPort string `json:"dst_port"`
	// DstType 目的地址类型（实测值 0；编码语义待真机验证）。
	DstType int64 `json:"dst_type"`
	// AreaCode 地区编码（实测为空串；字段语义待真机验证）。
	AreaCode string `json:"area_code"`
	// IfaceBand 接口带宽（实测值 1；单位与语义待真机验证）。
	IfaceBand int64 `json:"iface_band"`
	// Week 生效日（"1234567" = 周一至周日，实测值）。
	Week string `json:"week"`
	// Time 生效时段（实测 "00:00-23:59"）。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r IPPortRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// IPPortSelectors 是端口分流表单的可选值区块（stream_ipport.show TYPE=all
// 的区块，真机 3.7.24 实测）。
type IPPortSelectors struct {
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
	// IPLibStatus 地址库状态（实测 0；字段语义待真机验证）。
	IPLibStatus int64 `json:"iplib_status"`
}

// IPPortService 是端口分流（stream_ipport）的子服务。
type IPPortService struct {
	c *ikuai.Client
}

// List 获取端口分流规则列表（分页）。
// 对应 /Action/call: stream_ipport.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「网络设置 → 流控分流 → 端口分流」页面。
func (s *IPPortService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPPortRule], error) {
	resp, err := s.c.Call(ctx, "stream_ipport", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPPortRule](resp.Data)
}

// Selectors 获取端口分流表单的可选值（接口/分组/协议等）。
// 对应 /Action/call: stream_ipport.show（TYPE=all）。
func (s *IPPortService) Selectors(ctx context.Context) (*IPPortSelectors, error) {
	var out IPPortSelectors
	if err := s.c.CallData(ctx, "stream_ipport", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增端口分流规则，返回固件分配的行 ID。
// 对应 /Action/call: stream_ipport.add（开源 3.x 实现为扁平字段形态）。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *IPPortService) Add(ctx context.Context, rule IPPortRule) (int64, error) {
	resp, err := s.c.Call(ctx, "stream_ipport", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改端口分流规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: stream_ipport.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *IPPortService) Edit(ctx context.Context, rule IPPortRule) error {
	_, err := s.c.Call(ctx, "stream_ipport", "edit", rule)
	return err
}

// Delete 删除端口分流规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}；
// 开源 3.x 实现亦验证了 {"id":"1,2"} 串形态被固件接受）。
// 对应 /Action/call: stream_ipport.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *IPPortService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "stream_ipport", "del", ikuai.IDParam(ids))
	return err
}
