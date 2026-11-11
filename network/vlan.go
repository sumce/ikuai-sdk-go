package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// VLAN 是一个 VLAN 子接口的配置行（vlan.show 行，真机实证）。
type VLAN struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// VLANID VLAN ID（固件以字符串返回，支持终结模式如 "100-200"）。
	VLANID string `json:"vlan_id"`
	// VLANName VLAN 接口名（如 "vlan1211"）。
	VLANName string `json:"vlan_name"`
	// Interface 所属线路（物理接口对应的接口名，如 lan3）。
	Interface string `json:"interface"`
	// IPAddr 接口 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt IPv4 地址的整型表示（排序键，固件计算字段）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
	// IPMask "IP/掩码" 复合串（未配置时为空串）。
	IPMask string `json:"ip_mask"`
	// MAC VLAN 接口 MAC（未配置为空串）。
	MAC string `json:"mac"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// IsEnabled 报告 VLAN 子接口是否启用。
func (r VLAN) IsEnabled() bool { return r.Enabled == EnabledYes }

// VLANService 是 VLAN 设置（vlan）的子服务。
type VLANService struct {
	c *ikuai.Client
}

// List 获取 VLAN 子接口列表。
// 对应 /Action/call: vlan.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → VLAN 设置」页面。
func (s *VLANService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[VLAN], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "vlan", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[VLAN](resp.Data)
}

// Add 新增 VLAN 子接口，返回固件分配的行 ID。
// 对应 /Action/call: vlan.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VLANService) Add(ctx context.Context, row VLAN) (int64, error) {
	resp, err := s.c.Call(ctx, "vlan", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 VLAN 子接口。
// 对应 /Action/call: vlan.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VLANService) Edit(ctx context.Context, row VLAN) error {
	_, err := s.c.Call(ctx, "vlan", "edit", row)
	return err
}

// Delete 删除 VLAN 子接口，支持单个或多个 ID。
// 对应 /Action/call: vlan.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VLANService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "vlan", "del", ikuai.IDParam(ids))
	return err
}
