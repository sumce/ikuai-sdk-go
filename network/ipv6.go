package network

import (
	"context"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// dhcp6_lease.show 的 TYPE 选择器常量（真机实证：TYPE=total,data 回
// 30006 unknown TYPE，仅 TYPE=all 可用）。
const (
	// DHCP6LeaseTypeAll DHCPv6 租约全量选择器（真机实证）。
	DHCP6LeaseTypeAll = "all"
)

// IPv6NeighborStatus 的邻居状态枚举（真机观测值，RFC 4861 NUD 状态）。
const (
	// IPv6NeighborStatusStale 旧探测（真机观测）。
	IPv6NeighborStatusStale = "STALE"
	// IPv6NeighborStatusDelay 等待确认（真机观测）。
	IPv6NeighborStatusDelay = "DELAY"
)

// IPv6Config 是一条 IPv6 线路配置（ipv6.show 行，真机实证；每条启用
// IPv6 的线路一行）。dhcp6_* 为 DHCPv6 动态获取模式的运行回显字段。
type IPv6Config struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用该线路 IPv6（"yes"/"no"，真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// Interface 线路名（如 adsl07）。
	Interface string `json:"interface"`
	// Internet IPv6 获取方式（真机观测 "dhcp"=DHCPv6；完整枚举待真机验证）。
	Internet string `json:"internet"`
	// IPv6Addr 静态配置的 IPv6 地址（internet=static 时使用，动态获取为空）。
	IPv6Addr string `json:"ipv6_addr"`
	// IPv6Gateway 静态配置的 IPv6 网关（动态获取为空）。
	IPv6Gateway string `json:"ipv6_gateway"`
	// Prefix 前缀获取方式（真机观测 "auto"；完整枚举待真机验证）。
	Prefix string `json:"prefix"`
	// PrefixHint 前缀提示（真机观测空串；语义待真机验证）。
	PrefixHint string `json:"prefix_hint"`
	// ForcePrefix 是否强制下发前缀（0/1，真机观测 0）。
	ForcePrefix int64 `json:"force_prefix"`
	// ForceGenDUID 是否强制重新生成 DUID（0/1，真机观测 1）。
	ForceGenDUID int64 `json:"force_gen_duid"`
	// LinkAddr 链路本地地址（fe80::，固件计算字段）。
	LinkAddr string `json:"link_addr"`
	// DHCP6IPAddr DHCPv6 获取的地址/前缀长度（如 "2408:...::/64"）。
	DHCP6IPAddr string `json:"dhcp6_ip_addr"`
	// DHCP6IPGateway DHCPv6 获取的网关（fe80::）。
	DHCP6IPGateway string `json:"dhcp6_ip_gateway"`
	// DHCP6DNS1 DHCPv6 下发的首选 DNS。
	DHCP6DNS1 string `json:"dhcp6_dns1"`
	// DHCP6DNS2 DHCPv6 下发的备选 DNS。
	DHCP6DNS2 string `json:"dhcp6_dns2"`
	// DHCP6Prefix1 DHCPv6 获取的前缀（如 "2408:...::/60"）。
	DHCP6Prefix1 string `json:"dhcp6_prefix1"`
	// PreferredLft 首选生存期（秒，固件以字符串返回）。
	PreferredLft string `json:"preferred_lft"`
	// ValidLft 有效生存期（秒，固件以字符串返回）。
	ValidLft string `json:"valid_lft"`
}

// IsEnabled 报告该线路 IPv6 是否启用。
func (r IPv6Config) IsEnabled() bool { return r.Enabled == EnabledYes }

// IPv6Stream 是一条 IPv6 分流统计行（ipv6_stream.show 行，真机实证；
// 运行状态，只读）。total_download/total_upload 为累计字节（固件以
// 字符串返回大数）；download/upload 为当前速率（字节/秒）。
type IPv6Stream struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Interface 线路名（如 adsl07）。
	Interface string `json:"interface"`
	// Conn 当前 IPv6 连接数。
	Conn int64 `json:"conn"`
	// Download 当前下载速率（字节/秒）。
	Download int64 `json:"download"`
	// Upload 当前上传速率（字节/秒）。
	Upload int64 `json:"upload"`
	// TotalDownload 累计下载流量（字节，固件以字符串返回大数）。
	TotalDownload string `json:"total_download"`
	// TotalUpload 累计上传流量（字节，固件以字符串返回大数）。
	TotalUpload string `json:"total_upload"`
}

// DHCP6Lease 是一条 DHCPv6 租约（dhcp6_lease.show TYPE=all 的
// client_data 行，真机实证；运行状态表，只读）。固件 ipv6_addr 值可能
// 带尾随空格（真机观测），用 IPv6AddrTrimmed 取去除空白后的地址。
type DHCP6Lease struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Interface 获取地址的接口。
	Interface string `json:"interface"`
	// LinkAddr 客户端链路本地地址（fe80::）。
	LinkAddr string `json:"link_addr"`
	// DUID 客户端 DUID（如 "00010001322b3bc1d8bbc19a8c8f"）。
	DUID string `json:"duid"`
	// IPv6Addr 分配的 IPv6 地址（固件原样返回，可能带尾随空格）。
	IPv6Addr string `json:"ipv6_addr"`
	// MAC 客户端 MAC 地址。
	MAC string `json:"mac"`
	// Hostname 客户端主机名。
	Hostname string `json:"hostname"`
	// StartTime 租约开始时间（Unix 秒）。
	StartTime int64 `json:"start_time"`
	// Expires 租约到期时间（Unix 秒）。
	Expires int64 `json:"expires"`
	// Timeout 剩余有效时间（秒）。
	Timeout int64 `json:"timeout"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// IPv6AddrTrimmed 返回去除空白后的 IPv6 地址（固件值可能带尾随空格）。
func (r DHCP6Lease) IPv6AddrTrimmed() string { return strings.TrimSpace(r.IPv6Addr) }

// DHCP6LeasePage 是 dhcp6_lease.show TYPE=all 的完整返回形态（真机实证）。
type DHCP6LeasePage struct {
	// ClientTotal 租约总数（真机实证键 client_total）。
	ClientTotal int64 `json:"client_total"`
	// Rows 租约列表（真机实证键 client_data）。
	Rows []DHCP6Lease `json:"client_data"`
}

// IPv6Static 是一条 DHCPv6 静态分配记录（ipv6_static.show 行；真机列表
// 为空，字段名由固件语义推导，待真机验证）。
type IPv6Static struct {
	// ID 行 ID（待真机验证）。
	ID int64 `json:"id"`
	// Interface 绑定接口（待真机验证）。
	Interface string `json:"interface"`
	// DUID 客户端 DUID（待真机验证）。
	DUID string `json:"duid"`
	// IPv6Addr 静态分配的 IPv6 地址（待真机验证）。
	IPv6Addr string `json:"ipv6_addr"`
	// Hostname 主机名称（待真机验证）。
	Hostname string `json:"hostname"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// DHCP6ACLMac 是一条 DHCPv6 黑白名单记录（dhcp6_acl_mac.show 行；真机
// 列表为空，字段名由固件语义推导，待真机验证）。
type DHCP6ACLMac struct {
	// ID 行 ID（待真机验证）。
	ID int64 `json:"id"`
	// MAC 终端 MAC 地址（待真机验证）。
	MAC string `json:"mac"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IPv6Neighbor 是一条 IPv6 邻居表项（ipv6_neighbor.show 行，真机实证；
// 运行状态表，只读）。
type IPv6Neighbor struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Interface 学到邻居的接口（如 vlan1/lan4）。
	Interface string `json:"interface"`
	// IPv6Addr 邻居 IPv6 地址（含 fe80:: 链路本地）。
	IPv6Addr string `json:"ipv6_addr"`
	// MAC 邻居 MAC 地址。
	MAC string `json:"mac"`
	// Status 邻居状态（如 STALE/DELAY，RFC 4861 NUD）。
	Status string `json:"status"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// IPv6Service 是 IPv6 族（ipv6/ipv6_stream/dhcp6_lease/ipv6_static/
// dhcp6_acl_mac/ipv6_neighbor）的子服务。
type IPv6Service struct {
	c *ikuai.Client
}

// List 获取 IPv6 线路配置列表（每条启用 IPv6 的线路一行）。
// 对应 /Action/call: ipv6.show（TYPE=total,data，真机实证），
// 即 Web 控制台「网络设置 → IPv6」页面。
func (s *IPv6Service) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPv6Config], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipv6", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPv6Config](resp.Data)
}

// Set 保存 IPv6 线路配置（写动作）。
// 对应 /Action/call: ipv6.set（行 ID 必填）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) Set(ctx context.Context, cfg IPv6Config) error {
	_, err := s.c.Call(ctx, "ipv6", "set", cfg)
	return err
}

// StreamList 获取 IPv6 分流统计（每条 IPv6 线路一行，运行状态，只读）。
// 对应 /Action/call: ipv6_stream.show（TYPE=total,data，真机实证），
// 即 Web 控制台「网络设置 → IPv6 分流」页面。
func (s *IPv6Service) StreamList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPv6Stream], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipv6_stream", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPv6Stream](resp.Data)
}

// LeaseList 获取 DHCPv6 租约列表（运行状态表，只读）。
// 对应 /Action/call: dhcp6_lease.show（TYPE=all，真机实证：
// TYPE=total/data 回 30006 unknown TYPE，返回 client_data/client_total），
// 即 Web 控制台「网络设置 → DHCPv6 终端」页面。
func (s *IPv6Service) LeaseList(ctx context.Context, opts *ikuai.ListOptions) (*DHCP6LeasePage, error) {
	param := forcedShowParam(opts, DHCP6LeaseTypeAll, nil)
	var out DHCP6LeasePage
	if err := s.c.CallData(ctx, "dhcp6_lease", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StaticList 获取 DHCPv6 静态分配列表。
// 对应 /Action/call: ipv6_static.show（TYPE=total,data；show 路径真机
// 实证返回空表），即 Web 控制台「网络设置 → DHCPv6 静态分配」页面。
func (s *IPv6Service) StaticList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPv6Static], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipv6_static", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPv6Static](resp.Data)
}

// StaticAdd 新增 DHCPv6 静态分配，返回固件分配的行 ID。
// 对应 /Action/call: ipv6_static.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) StaticAdd(ctx context.Context, row IPv6Static) (int64, error) {
	resp, err := s.c.Call(ctx, "ipv6_static", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// StaticEdit 按 ID 修改 DHCPv6 静态分配。
// 对应 /Action/call: ipv6_static.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) StaticEdit(ctx context.Context, row IPv6Static) error {
	_, err := s.c.Call(ctx, "ipv6_static", "edit", row)
	return err
}

// StaticDelete 删除 DHCPv6 静态分配，支持单个或多个 ID。
// 对应 /Action/call: ipv6_static.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) StaticDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ipv6_static", "del", ikuai.IDParam(ids))
	return err
}

// ACLList 获取 DHCPv6 黑白名单列表。
// 对应 /Action/call: dhcp6_acl_mac.show（TYPE=total,data；show 路径真机
// 实证返回空表），即 Web 控制台「网络设置 → DHCPv6 黑白名单」页面。
func (s *IPv6Service) ACLList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DHCP6ACLMac], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dhcp6_acl_mac", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DHCP6ACLMac](resp.Data)
}

// ACLAdd 新增 DHCPv6 黑白名单记录，返回固件分配的行 ID。
// 对应 /Action/call: dhcp6_acl_mac.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) ACLAdd(ctx context.Context, row DHCP6ACLMac) (int64, error) {
	resp, err := s.c.Call(ctx, "dhcp6_acl_mac", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// ACLEdit 按 ID 修改 DHCPv6 黑白名单记录。
// 对应 /Action/call: dhcp6_acl_mac.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) ACLEdit(ctx context.Context, row DHCP6ACLMac) error {
	_, err := s.c.Call(ctx, "dhcp6_acl_mac", "edit", row)
	return err
}

// ACLDelete 删除 DHCPv6 黑白名单记录，支持单个或多个 ID。
// 对应 /Action/call: dhcp6_acl_mac.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *IPv6Service) ACLDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dhcp6_acl_mac", "del", ikuai.IDParam(ids))
	return err
}

// NeighborList 获取 IPv6 邻居表（运行状态表，只读）。
// 对应 /Action/call: ipv6_neighbor.show（TYPE=total,data，真机实证），
// 即 Web 控制台「网络设置 → IPv6 邻居列表」页面。
func (s *IPv6Service) NeighborList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPv6Neighbor], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipv6_neighbor", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPv6Neighbor](resp.Data)
}
