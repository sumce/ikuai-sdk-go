package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// SimpleQoSRule 是一条 IP 限速规则行（simple_qos.show total,data 的
// data 元素，真机 3.7.24 实测 2 行）。
type SimpleQoSRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，实测值）。
	Enabled string `json:"enabled"`
	// Comment 备注。
	Comment string `json:"comment"`
	// Interface 绑定的 WAN 接口名串（逗号分隔多接口，实测样例
	// "adsl01,adsl02,..."）。
	Interface string `json:"interface"`
	// IPAddr 限速对象地址（实测为单 IP，如 "172.22.1.2"；网段/分组
	// 形态待真机验证）。
	IPAddr string `json:"ip_addr"`
	// Upload 上行限速值（实测 18560；单位 Kb/s 待真机验证）。
	Upload int64 `json:"upload"`
	// Download 下行限速值（实测 18560；单位 Kb/s 待真机验证）。
	Download int64 `json:"download"`
	// Type 限速类型（实测值 1；编码语义待真机验证）。
	Type int64 `json:"type"`
	// Attr 限速属性（实测值 0；编码语义待真机验证）。
	Attr int64 `json:"attr"`
	// Protocol 协议（实测 "any"）。
	Protocol string `json:"protocol"`
	// SrcPort 源端口串（实测为空表示不限制）。
	SrcPort string `json:"src_port"`
	// DstPort 目的端口串（实测为空表示不限制）。
	DstPort string `json:"dst_port"`
	// Week 生效日（"1234567" = 周一至周日，实测值）。
	Week string `json:"week"`
	// Time 生效时段（实测 "00:00-23:59"）。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r SimpleQoSRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// SimpleQoSSelectors 是 IP 限速表单的可选值区块（simple_qos.show TYPE=all
// 的区块，真机 3.7.24 实测）。
type SimpleQoSSelectors struct {
	// Interface 可选接口列表，每项 [接口名, 备注]。
	Interface [][]string `json:"interface"`
	// IPGroup 可选 IPv4 分组名列表。
	IPGroup []string `json:"ipgroup"`
}

// SimpleQoSService 是 IP 限速（simple_qos）的子服务。
type SimpleQoSService struct {
	c *ikuai.Client
}

// List 获取 IP 限速规则列表（分页）。
// 对应 /Action/call: simple_qos.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「流控设置 → 限速 → IP 限速」页面。
func (s *SimpleQoSService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SimpleQoSRule], error) {
	resp, err := s.c.Call(ctx, "simple_qos", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[SimpleQoSRule](resp.Data)
}

// Selectors 获取 IP 限速表单的可选值（接口/IP 分组）。
// 对应 /Action/call: simple_qos.show（TYPE=all）。
func (s *SimpleQoSService) Selectors(ctx context.Context) (*SimpleQoSSelectors, error) {
	var out SimpleQoSSelectors
	if err := s.c.CallData(ctx, "simple_qos", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增 IP 限速规则，返回固件分配的行 ID。
// 对应 /Action/call: simple_qos.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *SimpleQoSService) Add(ctx context.Context, rule SimpleQoSRule) (int64, error) {
	resp, err := s.c.Call(ctx, "simple_qos", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 IP 限速规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: simple_qos.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *SimpleQoSService) Edit(ctx context.Context, rule SimpleQoSRule) error {
	_, err := s.c.Call(ctx, "simple_qos", "edit", rule)
	return err
}

// Delete 删除 IP 限速规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: simple_qos.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *SimpleQoSService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "simple_qos", "del", ikuai.IDParam(ids))
	return err
}

// MacQoSRule 是一条 MAC 限速规则行（mac_qos.show total,data 的 data 元素，
// 真机 3.7.24 实测 1 行）。
type MacQoSRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，实测值）。
	Enabled string `json:"enabled"`
	// Comment 备注（实测样例 "快速添加"）。
	Comment string `json:"comment"`
	// Interface 绑定的 WAN 接口名串（实测为空串）。
	Interface string `json:"interface"`
	// MACAddr 终端 MAC 地址（实测 "00:0c:29:0e:10:5f"）。
	MACAddr string `json:"mac_addr"`
	// IPType IP 协议族（实测字符串 "4" = IPv4；"6" 语义待真机验证）。
	IPType string `json:"ip_type"`
	// Upload 上行限速值（实测 25000；单位 Kb/s 待真机验证）。
	Upload int64 `json:"upload"`
	// Download 下行限速值（实测 25000；单位 Kb/s 待真机验证）。
	Download int64 `json:"download"`
	// Type 限速类型（实测值 0；编码语义待真机验证）。
	Type int64 `json:"type"`
	// Attr 限速属性（实测值 3；编码语义待真机验证）。
	Attr int64 `json:"attr"`
	// Week 生效日（"1234567" = 周一至周日，实测值）。
	Week string `json:"week"`
	// Time 生效时段（实测 "00:00-23:59"）。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r MacQoSRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// MacQoSSelectors 是 MAC 限速表单的可选值区块（mac_qos.show TYPE=all 的
// 区块，真机 3.7.24 实测）。
type MacQoSSelectors struct {
	// Interface 可选接口列表，每项 [接口名, 备注]。
	Interface [][]string `json:"interface"`
	// MACGroup 可选 MAC 分组名列表。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表。
	DTGroup []string `json:"dtgroup"`
}

// MacQoSService 是 MAC 限速（mac_qos）的子服务。
type MacQoSService struct {
	c *ikuai.Client
}

// List 获取 MAC 限速规则列表（分页）。
// 对应 /Action/call: mac_qos.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「流控设置 → 限速 → MAC 限速」页面。
func (s *MacQoSService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[MacQoSRule], error) {
	resp, err := s.c.Call(ctx, "mac_qos", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[MacQoSRule](resp.Data)
}

// Selectors 获取 MAC 限速表单的可选值（接口/MAC 分组/时间分组）。
// 对应 /Action/call: mac_qos.show（TYPE=all）。
func (s *MacQoSService) Selectors(ctx context.Context) (*MacQoSSelectors, error) {
	var out MacQoSSelectors
	if err := s.c.CallData(ctx, "mac_qos", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增 MAC 限速规则，返回固件分配的行 ID。
// 对应 /Action/call: mac_qos.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MacQoSService) Add(ctx context.Context, rule MacQoSRule) (int64, error) {
	resp, err := s.c.Call(ctx, "mac_qos", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 MAC 限速规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: mac_qos.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MacQoSService) Edit(ctx context.Context, rule MacQoSRule) error {
	_, err := s.c.Call(ctx, "mac_qos", "edit", rule)
	return err
}

// Delete 删除 MAC 限速规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: mac_qos.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MacQoSService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "mac_qos", "del", ikuai.IDParam(ids))
	return err
}
