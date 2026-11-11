package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// nat_rule.show 的 TYPE 选择器与 action 枚举（抓包实证）。
const (
	// NATRuleTypeDataTotal NAT 规则列表（TYPE=total,data，抓包实证）。
	NATRuleTypeDataTotal = "total,data"
	// NATRuleTypeInterfaceAction 下拉候选数据 TYPE（抓包实证：
	// interface 为 [名称, 备注] 嵌套数组，action 为动作列表）。
	NATRuleTypeInterfaceAction = "interface,action"
)

// nat_rule.action 的动作枚举值（抓包实证 ["filter","snat","dnat"]）。
const (
	NATRuleActionFilter = "filter"
	NATRuleActionSNAT   = "snat"
	NATRuleActionDNAT   = "dnat"
)

// NATRule 是一条 NAT 规则（nat_rule.show 行；真机列表为空，行字段由
// 固件语义推导，待真机验证；action 枚举已抓包实证）。
type NATRule struct {
	// ID 行 ID（待真机验证）。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 生效线路（候选见 Selectors，待真机验证）。
	Interface string `json:"interface"`
	// Action 动作（filter/snat/dnat，枚举抓包实证）。
	Action string `json:"action"`
	// SrcAddr 源地址（待真机验证）。
	SrcAddr string `json:"src_addr"`
	// DstAddr 目的地址（待真机验证）。
	DstAddr string `json:"dst_addr"`
	// ToAddr 转换后地址（待真机验证）。
	ToAddr string `json:"to_addr"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告规则是否启用（字段待真机验证）。
func (r NATRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// NATRuleSelectors 是 nat_rule.show TYPE=interface,action 的下拉候选数据
// （抓包实证），供构造 NATRule 时选择取值。
type NATRuleSelectors struct {
	// Interface 接口候选（[名称] 或 [名称, 备注]）。
	Interface [][]string `json:"interface"`
	// Action 动作候选（["filter","snat","dnat"]）。
	Action []string `json:"action"`
}

// Netmap 是一条 DMZ 主机映射（netmap.show 行；真机列表为空，行字段由
// 固件语义推导，待真机验证）。
type Netmap struct {
	// ID 行 ID（待真机验证）。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 外网线路（待真机验证）。
	Interface string `json:"interface"`
	// WANAddr 外网地址（待真机验证；多 WAN 指定公网 IP 时使用）。
	WANAddr string `json:"wan_addr"`
	// LANAddr 内网 DMZ 主机地址（待真机验证）。
	LANAddr string `json:"lan_addr"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告 DMZ 映射是否启用（字段待真机验证）。
func (r Netmap) IsEnabled() bool { return r.Enabled == EnabledYes }

// NATService 是 NAT 规则族（nat_rule/netmap）的子服务。
type NATService struct {
	c *ikuai.Client
}

// List 获取 NAT 规则列表。
// 对应 /Action/call: nat_rule.show（TYPE=total,data，抓包实证），
// 即 Web 控制台「网络设置 → NAT 规则」页面。
func (s *NATService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[NATRule], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "nat_rule", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[NATRule](resp.Data)
}

// Selectors 获取 NAT 规则的接口与动作候选（抓包实证）。
// 对应 /Action/call: nat_rule.show（TYPE=interface,action）。
func (s *NATService) Selectors(ctx context.Context) (*NATRuleSelectors, error) {
	var out NATRuleSelectors
	if err := s.c.CallData(ctx, "nat_rule", "show",
		map[string]any{"TYPE": NATRuleTypeInterfaceAction}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增 NAT 规则，返回固件分配的行 ID。
// 对应 /Action/call: nat_rule.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NATService) Add(ctx context.Context, rule NATRule) (int64, error) {
	resp, err := s.c.Call(ctx, "nat_rule", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 NAT 规则。
// 对应 /Action/call: nat_rule.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NATService) Edit(ctx context.Context, rule NATRule) error {
	_, err := s.c.Call(ctx, "nat_rule", "edit", rule)
	return err
}

// Delete 删除 NAT 规则，支持单个或多个 ID。
// 对应 /Action/call: nat_rule.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NATService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "nat_rule", "del", ikuai.IDParam(ids))
	return err
}

// List 获取 DMZ 主机映射列表。
// 对应 /Action/call: netmap.show（TYPE=total,data；show 路径真机实证
// 返回空表），即 Web 控制台「服务应用 → 端口映射 → DMZ 主机」页面。
func (s *NATService) NetmapList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[Netmap], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "netmap", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[Netmap](resp.Data)
}

// NetmapAdd 新增 DMZ 主机映射，返回固件分配的行 ID。
// 对应 /Action/call: netmap.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NATService) NetmapAdd(ctx context.Context, row Netmap) (int64, error) {
	resp, err := s.c.Call(ctx, "netmap", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// NetmapEdit 按 ID 修改 DMZ 主机映射。
// 对应 /Action/call: netmap.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NATService) NetmapEdit(ctx context.Context, row Netmap) error {
	_, err := s.c.Call(ctx, "netmap", "edit", row)
	return err
}

// NetmapDelete 删除 DMZ 主机映射，支持单个或多个 ID。
// 对应 /Action/call: netmap.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *NATService) NetmapDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "netmap", "del", ikuai.IDParam(ids))
	return err
}
