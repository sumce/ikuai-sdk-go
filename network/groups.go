package network

import (
	"context"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// ipgroup.type 的地址池类型枚举（真机观测 0；完整枚举待真机验证）。
const (
	// IPGroupTypeAddr 地址池类型 0（真机观测值；语义待真机验证）。
	IPGroupTypeAddr = int64(0)
)

// IPGroup 是一个 IPv4 分组（ipgroup.show 行，真机实证）。
type IPGroup struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// GroupName 分组名称。
	GroupName string `json:"group_name"`
	// AddrPool 地址池（单 IP/网段/IP 段，如 "172.20.0.1/24"）。
	AddrPool string `json:"addr_pool"`
	// Type 地址池类型（真机观测 0；枚举语义待真机验证）。
	Type int64 `json:"type"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// AddrPools 按逗号拆分地址池串为单个地址项。
func (r IPGroup) AddrPools() []string {
	return splitList(r.AddrPool)
}

// IPv6Group 是一个 IPv6 分组（ipv6group.show 行；真机列表为空，字段名
// 由固件 i18n 推导，待真机验证）。
type IPv6Group struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// GroupName 分组名称。
	GroupName string `json:"group_name"`
	// AddrPool IPv6 地址池（网段/前缀）。
	AddrPool string `json:"addr_pool"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// MACGroup 是一个 MAC 分组（macgroup.show 行；真机列表为空，字段名由
// 固件 i18n 推导，待真机验证）。
type MACGroup struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// GroupName 分组名称。
	GroupName string `json:"group_name"`
	// MACs MAC 列表（逗号分隔）。
	MACs string `json:"macs"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// MACList 按逗号拆分 MAC 列表串。
func (r MACGroup) MACList() []string { return splitList(r.MACs) }

// DTGroup 是一个时间分组（dtgroup.show 行；真机列表为空，字段名由固件
// i18n 推导，待真机验证）。
type DTGroup struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// GroupName 分组名称。
	GroupName string `json:"group_name"`
	// Week 生效周期（"0"-"6" 的组合串，0=周日；待真机验证）。
	Week string `json:"week"`
	// Time 生效时间段（如 "00:00-08:00,09:00-11:00"）。
	Time string `json:"time"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// splitList 按逗号拆分固件的列表串，去掉空项与首尾空白。
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// GroupsService 是分组族（ipgroup/ipv6group/macgroup/dtgroup）的子服务。
type GroupsService struct {
	c *ikuai.Client
}

// IPList 获取 IPv4 分组列表。
// 对应 /Action/call: ipgroup.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → 客户端分组 → IP 分组」页面。
func (s *GroupsService) IPList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPGroup], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipgroup", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPGroup](resp.Data)
}

// IPAdd 新增 IPv4 分组，返回固件分配的行 ID。
// 对应 /Action/call: ipgroup.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) IPAdd(ctx context.Context, row IPGroup) (int64, error) {
	resp, err := s.c.Call(ctx, "ipgroup", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// IPEdit 按 ID 修改 IPv4 分组。
// 对应 /Action/call: ipgroup.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) IPEdit(ctx context.Context, row IPGroup) error {
	_, err := s.c.Call(ctx, "ipgroup", "edit", row)
	return err
}

// IPDelete 删除 IPv4 分组，支持单个或多个 ID。
// 对应 /Action/call: ipgroup.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) IPDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ipgroup", "del", ikuai.IDParam(ids))
	return err
}

// IPv6List 获取 IPv6 分组列表。
// 对应 /Action/call: ipv6group.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → 客户端分组 → IPv6 分组」页面。
func (s *GroupsService) IPv6List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPv6Group], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipv6group", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPv6Group](resp.Data)
}

// IPv6Add 新增 IPv6 分组，返回固件分配的行 ID。
// 对应 /Action/call: ipv6group.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) IPv6Add(ctx context.Context, row IPv6Group) (int64, error) {
	resp, err := s.c.Call(ctx, "ipv6group", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// IPv6Edit 按 ID 修改 IPv6 分组。
// 对应 /Action/call: ipv6group.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) IPv6Edit(ctx context.Context, row IPv6Group) error {
	_, err := s.c.Call(ctx, "ipv6group", "edit", row)
	return err
}

// IPv6Delete 删除 IPv6 分组，支持单个或多个 ID。
// 对应 /Action/call: ipv6group.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) IPv6Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ipv6group", "del", ikuai.IDParam(ids))
	return err
}

// MACList 获取 MAC 分组列表。
// 对应 /Action/call: macgroup.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → 客户端分组 → MAC 分组」页面。
func (s *GroupsService) MACList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[MACGroup], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "macgroup", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[MACGroup](resp.Data)
}

// MACAdd 新增 MAC 分组，返回固件分配的行 ID。
// 对应 /Action/call: macgroup.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) MACAdd(ctx context.Context, row MACGroup) (int64, error) {
	resp, err := s.c.Call(ctx, "macgroup", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// MACEdit 按 ID 修改 MAC 分组。
// 对应 /Action/call: macgroup.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) MACEdit(ctx context.Context, row MACGroup) error {
	_, err := s.c.Call(ctx, "macgroup", "edit", row)
	return err
}

// MACDelete 删除 MAC 分组，支持单个或多个 ID。
// 对应 /Action/call: macgroup.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) MACDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "macgroup", "del", ikuai.IDParam(ids))
	return err
}

// DTList 获取时间分组列表。
// 对应 /Action/call: dtgroup.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → 客户端分组 → 时间分组」页面。
func (s *GroupsService) DTList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DTGroup], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dtgroup", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DTGroup](resp.Data)
}

// DTAdd 新增时间分组，返回固件分配的行 ID。
// 对应 /Action/call: dtgroup.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) DTAdd(ctx context.Context, row DTGroup) (int64, error) {
	resp, err := s.c.Call(ctx, "dtgroup", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// DTEdit 按 ID 修改时间分组。
// 对应 /Action/call: dtgroup.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) DTEdit(ctx context.Context, row DTGroup) error {
	_, err := s.c.Call(ctx, "dtgroup", "edit", row)
	return err
}

// DTDelete 删除时间分组，支持单个或多个 ID。
// 对应 /Action/call: dtgroup.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *GroupsService) DTDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dtgroup", "del", ikuai.IDParam(ids))
	return err
}
