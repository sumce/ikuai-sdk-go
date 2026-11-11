package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// dhcp_static/dhcp_lease 共用的 status 状态枚举（真机观测 + 固件 i18n）。
const (
	// DHCPStatusScan 扫描分配（i18n s-1）。
	DHCPStatusScan = int64(-1)
	// DHCPStatusDynamic 动态分配（i18n s0，真机观测）。
	DHCPStatusDynamic = int64(0)
	// DHCPStatusStatic 静态分配（i18n s1）。
	DHCPStatusStatic = int64(1)
)

// DHCPServer 是一个 DHCP 服务端的配置行（dhcp_server.show 行，真机实证）。
// 固件自定义选项成对出现：opt_typeNN（0=十六进制 1=IPv4地址 2=字符串）为
// 值类型开关，optNN 为选项值。
type DHCPServer struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// Interface 服务接口（如 lan4）。
	Interface string `json:"interface"`
	// PHYIfnames 关联物理接口（"all" 或逗号分隔的网卡名）。
	PHYIfnames string `json:"phy_ifnames"`
	// AddrPool 客户端地址池（如 "192.168.2.1-192.168.2.10"）。
	AddrPool string `json:"addr_pool"`
	// ExcludePool 排除地址（空串表示无）。
	ExcludePool string `json:"exclude_pool"`
	// Netmask 子网掩码。
	Netmask string `json:"netmask"`
	// Gateway 网关。
	Gateway string `json:"gateway"`
	// DNS1 首选 DNS。
	DNS1 string `json:"dns1"`
	// DNS2 备选 DNS。
	DNS2 string `json:"dns2"`
	// Domain 域名。
	Domain string `json:"domain"`
	// NextServer 下一服务器（PXE 引导；语义待真机验证）。
	NextServer string `json:"next_server"`
	// Wins1 主 WINS 服务器。
	Wins1 string `json:"wins1"`
	// Wins2 辅助 WINS 服务器。
	Wins2 string `json:"wins2"`
	// Lease 租期（分钟）。
	Lease int64 `json:"lease"`
	// Delay 过期地址保留时间（分钟，0=到期立即回收）。
	Delay int64 `json:"delay"`
	// Status 服务端状态（1=服务正常；真机观测）。
	Status int64 `json:"status"`
	// Available 地址池剩余地址数（固件计算字段，add/edit 无需填写）。
	Available int64 `json:"available"`
	// CheckAddrValid 检查接口 IP 有效性（1=开启）。
	CheckAddrValid int64 `json:"check_addr_valid"`
	// CheckRelayOnly 只应用于 DHCP 中继（0/1）。
	CheckRelayOnly int64 `json:"check_relay_only"`
	// OptType15 自定义选项 15 的值类型开关（0/1/2，成对 Opt15）。
	OptType15 int64 `json:"opt_type15"`
	// Opt15 自定义选项 15 的值。
	Opt15 string `json:"opt15"`
	// OptType28 自定义选项 28 的值类型开关（成对 Opt28）。
	OptType28 int64 `json:"opt_type28"`
	// Opt28 自定义选项 28 的值。
	Opt28 string `json:"opt28"`
	// OptType43 自定义选项 43 的值类型开关（成对 Opt43）。
	OptType43 int64 `json:"opt_type43"`
	// Opt43 自定义选项 43 的值。
	Opt43 string `json:"opt43"`
	// OptType60 自定义选项 60 的值类型开关（成对 Opt60）。
	OptType60 int64 `json:"opt_type60"`
	// Opt60 自定义选项 60 的值。
	Opt60 string `json:"opt60"`
	// OptType66 自定义选项 66 的值类型开关（成对 Opt66）。
	OptType66 int64 `json:"opt_type66"`
	// Opt66 自定义选项 66 的值。
	Opt66 string `json:"opt66"`
	// OptType67 自定义选项 67 的值类型开关（成对 Opt67）。
	OptType67 int64 `json:"opt_type67"`
	// Opt67 自定义选项 67 的值。
	Opt67 string `json:"opt67"`
	// OptType80 自定义选项 80 的值类型开关（成对 Opt80）。
	OptType80 int64 `json:"opt_type80"`
	// Opt80 自定义选项 80 的值。
	Opt80 string `json:"opt80"`
	// OptType119 自定义选项 119 的值类型开关（成对 Opt119）。
	OptType119 int64 `json:"opt_type119"`
	// Opt119 自定义选项 119 的值。
	Opt119 string `json:"opt119"`
	// OptType121 自定义选项 121 的值类型开关（真机观测 2=字符串；成对 Opt121）。
	OptType121 int64 `json:"opt_type121"`
	// Opt121 自定义选项 121 的值（静态路由表）。
	Opt121 string `json:"opt121"`
	// OptType125 自定义选项 125 的值类型开关（成对 Opt125）。
	OptType125 int64 `json:"opt_type125"`
	// Opt125 自定义选项 125 的值。
	Opt125 string `json:"opt125"`
	// OptType128 自定义选项 128 的值类型开关（成对 Opt128）。
	OptType128 int64 `json:"opt_type128"`
	// Opt128 自定义选项 128 的值。
	Opt128 string `json:"opt128"`
	// OptType138 自定义选项 138 的值类型开关（成对 Opt138）。
	OptType138 int64 `json:"opt_type138"`
	// Opt138 自定义选项 138 的值。
	Opt138 string `json:"opt138"`
}

// IsEnabled 报告 DHCP 服务端是否启用。
func (r DHCPServer) IsEnabled() bool { return r.Enabled == EnabledYes }

// DHCPStatic 是一条 DHCP 静态分配记录（dhcp_static.show 行，真机实证）。
type DHCPStatic struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Interface 绑定接口（如 lan4）。
	Interface string `json:"interface"`
	// IPAddr 终端 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt IPv4 地址的整型表示（排序键，固件计算字段）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// MAC 终端 MAC 地址。
	MAC string `json:"mac"`
	// Hostname 主机名称。
	Hostname string `json:"hostname"`
	// Status 分配状态（-1=扫描分配 0=动态分配 1=静态分配）。
	Status int64 `json:"status"`
	// StartTime 分配开始时间（Unix 秒）。
	StartTime int64 `json:"start_time"`
	// EndTime 分配结束时间（Unix 秒）。
	EndTime int64 `json:"end_time"`
	// Timeout 剩余有效时间（秒）。
	Timeout int64 `json:"timeout"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// DHCPLease 是一条 DHCP 客户端租约（dhcp_lease.show 行，真机实证；
// 运行状态表，字段与 dhcp_static 同形但只读）。
type DHCPLease struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Interface 获取地址的接口。
	Interface string `json:"interface"`
	// IPAddr 分配的 IPv4 地址。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt IPv4 地址的整型表示（排序键）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// MAC 终端 MAC 地址。
	MAC string `json:"mac"`
	// Hostname 主机名称。
	Hostname string `json:"hostname"`
	// Status 分配状态（-1=扫描分配 0=动态分配 1=静态分配）。
	Status int64 `json:"status"`
	// StartTime 租约开始时间（Unix 秒）。
	StartTime int64 `json:"start_time"`
	// EndTime 租约到期时间（Unix 秒）。
	EndTime int64 `json:"end_time"`
	// Timeout 剩余有效时间（秒）。
	Timeout int64 `json:"timeout"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// DHCPACLMac 是一条 DHCP 黑白名单记录（dhcp_acl_mac.show 行；真机列表为空，
// 字段名待真机验证）。
type DHCPACLMac struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// MAC 终端 MAC 地址。
	MAC string `json:"mac"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// DHCPService 是 DHCP 服务族（dhcp_server/dhcp_static/dhcp_lease/
// dhcp_acl_mac）的子服务。
type DHCPService struct {
	c *ikuai.Client
}

// ServerList 获取 DHCP 服务端列表。
// 对应 /Action/call: dhcp_server.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → DHCP 设置 → DHCP 服务端」页面。
func (s *DHCPService) ServerList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DHCPServer], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dhcp_server", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DHCPServer](resp.Data)
}

// ServerAdd 新增 DHCP 服务端，返回固件分配的行 ID。
// 对应 /Action/call: dhcp_server.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) ServerAdd(ctx context.Context, row DHCPServer) (int64, error) {
	resp, err := s.c.Call(ctx, "dhcp_server", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// ServerEdit 按 ID 修改 DHCP 服务端。
// 对应 /Action/call: dhcp_server.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) ServerEdit(ctx context.Context, row DHCPServer) error {
	_, err := s.c.Call(ctx, "dhcp_server", "edit", row)
	return err
}

// ServerDelete 删除 DHCP 服务端，支持单个或多个 ID。
// 对应 /Action/call: dhcp_server.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) ServerDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dhcp_server", "del", ikuai.IDParam(ids))
	return err
}

// StaticList 获取 DHCP 静态分配列表。
// 对应 /Action/call: dhcp_static.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → DHCP 设置 → 静态分配」页面。
func (s *DHCPService) StaticList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DHCPStatic], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dhcp_static", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DHCPStatic](resp.Data)
}

// StaticAdd 新增静态分配，返回固件分配的行 ID。
// 对应 /Action/call: dhcp_static.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) StaticAdd(ctx context.Context, row DHCPStatic) (int64, error) {
	resp, err := s.c.Call(ctx, "dhcp_static", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// StaticEdit 按 ID 修改静态分配。
// 对应 /Action/call: dhcp_static.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) StaticEdit(ctx context.Context, row DHCPStatic) error {
	_, err := s.c.Call(ctx, "dhcp_static", "edit", row)
	return err
}

// StaticDelete 删除静态分配，支持单个或多个 ID。
// 对应 /Action/call: dhcp_static.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) StaticDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dhcp_static", "del", ikuai.IDParam(ids))
	return err
}

// LeaseList 获取 DHCP 客户端租约列表（运行状态表，只读）。
// 对应 /Action/call: dhcp_lease.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → DHCP 设置 → 客户端列表」页面。
func (s *DHCPService) LeaseList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DHCPLease], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dhcp_lease", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DHCPLease](resp.Data)
}

// ACLList 获取 DHCP 黑白名单列表。
// 对应 /Action/call: dhcp_acl_mac.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → DHCP 设置 → 黑白名单」页面。
func (s *DHCPService) ACLList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DHCPACLMac], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dhcp_acl_mac", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DHCPACLMac](resp.Data)
}

// ACLAdd 新增 DHCP 黑白名单记录，返回固件分配的行 ID。
// 对应 /Action/call: dhcp_acl_mac.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) ACLAdd(ctx context.Context, row DHCPACLMac) (int64, error) {
	resp, err := s.c.Call(ctx, "dhcp_acl_mac", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// ACLEdit 按 ID 修改 DHCP 黑白名单记录。
// 对应 /Action/call: dhcp_acl_mac.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) ACLEdit(ctx context.Context, row DHCPACLMac) error {
	_, err := s.c.Call(ctx, "dhcp_acl_mac", "edit", row)
	return err
}

// ACLDelete 删除 DHCP 黑白名单记录，支持单个或多个 ID。
// 对应 /Action/call: dhcp_acl_mac.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DHCPService) ACLDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dhcp_acl_mac", "del", ikuai.IDParam(ids))
	return err
}
