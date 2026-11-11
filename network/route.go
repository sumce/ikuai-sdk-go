package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// static_rt.ip_type 的地址族枚举（真机观测 "4"=IPv4）。
const (
	// StaticRouteIPTypeV4 IPv4 路由（真机观测值）。
	StaticRouteIPTypeV4 = "4"
	// StaticRouteIPTypeV6 IPv6 路由（推导值，待真机验证）。
	StaticRouteIPTypeV6 = "6"
)

// StaticRoute 是一条静态路由（static_rt.show 行，真机实证）。
type StaticRoute struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// DstAddr 目的地址（如 "172.20.0.0"）。
	DstAddr string `json:"dst_addr"`
	// DstAddrInt 目的地址的整型表示（排序键，固件计算字段）。
	DstAddrInt int64 `json:"dst_addr_int"`
	// Netmask 子网掩码（如 "255.255.255.0"）。
	Netmask string `json:"netmask"`
	// Gateway 网关（下一跳）。
	Gateway string `json:"gateway"`
	// GatewayInt 网关地址的整型表示（固件计算字段）。
	GatewayInt int64 `json:"gateway_int"`
	// Interface 出接口/线路（如 lan2）。
	Interface string `json:"interface"`
	// Prio 优先级（数值越小优先级越高）。
	Prio int64 `json:"prio"`
	// IPType 地址族（"4"=IPv4）。
	IPType string `json:"ip_type"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// IsEnabled 报告静态路由是否启用。
func (r StaticRoute) IsEnabled() bool { return r.Enabled == EnabledYes }

// RouteTableEntry 是当前路由表中的一条路由（static_rt_table.show 行，
// 真机实证；运行状态表，只读）。
type RouteTableEntry struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// DstAddr 目的地址（如 "0.0.0.0"）。
	DstAddr string `json:"dst_addr"`
	// DstAddrInt 目的地址的整型表示（排序键）。
	DstAddrInt int64 `json:"dst_addr_int"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
	// Gateway 网关（下一跳）。
	Gateway string `json:"gateway"`
	// Interface 出接口/线路（如 adsl02）。
	Interface string `json:"interface"`
	// Prio 优先级（数值越小优先级越高）。
	Prio int64 `json:"prio"`
}

// RouteService 是路由族（static_rt/static_rt_table）的子服务。
type RouteService struct {
	c *ikuai.Client
}

// StaticList 获取静态路由列表。
// 对应 /Action/call: static_rt.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → 静态路由」页面。
func (s *RouteService) StaticList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[StaticRoute], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "static_rt", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[StaticRoute](resp.Data)
}

// StaticAdd 新增静态路由，返回固件分配的行 ID。
// 对应 /Action/call: static_rt.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *RouteService) StaticAdd(ctx context.Context, row StaticRoute) (int64, error) {
	resp, err := s.c.Call(ctx, "static_rt", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// StaticEdit 按 ID 修改静态路由。
// 对应 /Action/call: static_rt.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *RouteService) StaticEdit(ctx context.Context, row StaticRoute) error {
	_, err := s.c.Call(ctx, "static_rt", "edit", row)
	return err
}

// StaticDelete 删除静态路由，支持单个或多个 ID。
// 对应 /Action/call: static_rt.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *RouteService) StaticDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "static_rt", "del", ikuai.IDParam(ids))
	return err
}

// TableList 获取当前路由表（内核路由，运行状态，只读）。
// 对应 /Action/call: static_rt_table.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → 静态路由 → 当前路由表」页面。
func (s *RouteService) TableList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[RouteTableEntry], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "static_rt_table", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[RouteTableEntry](resp.Data)
}
