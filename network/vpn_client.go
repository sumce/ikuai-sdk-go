package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// VPN 客户端族共用的通用拨号行形态说明：
// 真机 3.7.24 上 pptp_client/l2tp_client/openvpn-client/ipsec-vpn 的 show
// 返回 {"total":0,"data":[]}（设备未配置任何拨号），行字段无法从响应实证；
// wireguard/ike_client 在此固件版本不存在模块（show 回 30002
// "Not found funcname"，menu.json 中两个菜单项仍存在）。
// 因此下列各类型的行字段由固件菜单语义（i18n 文案 + Web 表单惯例）推导，
// 全部标注待真机验证；列表 List（show）路径已真机实证可正常返回空表。

// PPTPClient 是一条 PPTP 客户端拨号配置（pptp_client.show 行；字段待真机验证）。
type PPTPClient struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 拨号线路名（待真机验证）。
	Interface string `json:"interface"`
	// ServerAddr PPTP 服务器地址（域名或 IP，待真机验证）。
	ServerAddr string `json:"server_addr"`
	// Username 认证用户名（待真机验证）。
	Username string `json:"username"`
	// Password 认证密码（待真机验证）。
	Password string `json:"password"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告拨号是否启用（字段待真机验证）。
func (r PPTPClient) IsEnabled() bool { return r.Enabled == EnabledYes }

// L2TPClient 是一条 L2TP 客户端拨号配置（l2tp_client.show 行；字段待真机验证）。
type L2TPClient struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 拨号线路名（待真机验证）。
	Interface string `json:"interface"`
	// ServerAddr L2TP 服务器地址（待真机验证）。
	ServerAddr string `json:"server_addr"`
	// Username 认证用户名（待真机验证）。
	Username string `json:"username"`
	// Password 认证密码（待真机验证）。
	Password string `json:"password"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告拨号是否启用（字段待真机验证）。
func (r L2TPClient) IsEnabled() bool { return r.Enabled == EnabledYes }

// OpenVPNClient 是一条 OpenVPN 客户端拨号配置（openvpn-client.show 行；
// 字段待真机验证）。
type OpenVPNClient struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 拨号线路名（待真机验证）。
	Interface string `json:"interface"`
	// ServerAddr OpenVPN 服务器地址（待真机验证）。
	ServerAddr string `json:"server_addr"`
	// ServerPort OpenVPN 服务器端口（待真机验证）。
	ServerPort int64 `json:"server_port"`
	// Username 认证用户名（待真机验证）。
	Username string `json:"username"`
	// Password 认证密码（待真机验证）。
	Password string `json:"password"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告拨号是否启用（字段待真机验证）。
func (r OpenVPNClient) IsEnabled() bool { return r.Enabled == EnabledYes }

// IPSecClient 是一条 IPSec VPN 客户端拨号配置（ipsec-vpn.show 行；
// 字段待真机验证）。
type IPSecClient struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 拨号线路名（待真机验证）。
	Interface string `json:"interface"`
	// ServerAddr 对端网关地址（待真机验证）。
	ServerAddr string `json:"server_addr"`
	// Username 认证用户名（待真机验证）。
	Username string `json:"username"`
	// Password 认证密码（待真机验证）。
	Password string `json:"password"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告拨号是否启用（字段待真机验证）。
func (r IPSecClient) IsEnabled() bool { return r.Enabled == EnabledYes }

// WireGuardClient 是一条 WireGuard 客户端拨号配置（wireguard.show 行；
// 字段待真机验证）。
type WireGuardClient struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 拨号线路名（待真机验证）。
	Interface string `json:"interface"`
	// ServerAddr 对端服务器地址（待真机验证）。
	ServerAddr string `json:"server_addr"`
	// ServerPort 对端服务器端口（待真机验证）。
	ServerPort int64 `json:"server_port"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告拨号是否启用（字段待真机验证）。
func (r WireGuardClient) IsEnabled() bool { return r.Enabled == EnabledYes }

// IKEClient 是一条 IKEv2/IPSec 客户端拨号配置（ike_client.show 行；
// 字段待真机验证）。
type IKEClient struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled"`
	// Interface 拨号线路名（待真机验证）。
	Interface string `json:"interface"`
	// ServerAddr 对端服务器地址（待真机验证）。
	ServerAddr string `json:"server_addr"`
	// Username 认证用户名（待真机验证）。
	Username string `json:"username"`
	// Password 认证密码（待真机验证）。
	Password string `json:"password"`
	// Comment 备注（待真机验证）。
	Comment string `json:"comment"`
}

// IsEnabled 报告拨号是否启用（字段待真机验证）。
func (r IKEClient) IsEnabled() bool { return r.Enabled == EnabledYes }

// SDWANStatus 是 SD-WAN（ik_web_sdwan）的连接状态（show 返回 Data.data，
// 真机实证：{"code":"2020","message":"connect server failed"}；code 的
// 完整枚举语义待真机验证）。
type SDWANStatus struct {
	// Code 状态码（真机观测 "2020"=连接云服务器失败）。
	Code string `json:"code"`
	// Message 状态描述（真机观测 "connect server failed"）。
	Message string `json:"message"`
}

// WuyingConfig 是无影云虚拟专线的配置行（wuying.show Data.data[0]，
// 真机实证；单行配置）。Data 顶层另有 routeid（设备路由 ID，真机实证）。
type WuyingConfig struct {
	// ID 行 ID（真机观测恒为 1）。
	ID int64 `json:"id"`
	// Enabled 是否启用（"yes"/"no"，真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// ToCloud 组网方向（真机观测 1；枚举语义待真机验证）。
	ToCloud int64 `json:"tocloud"`
	// WANAddr 本地 WAN 侧网段（如 "10.30.20.0/24"）。
	WANAddr string `json:"wan_addr"`
	// LANAddr 本地 LAN 侧网段（如 "10.31.20.0/24"）。
	LANAddr string `json:"lan_addr"`
	// BindWANAddr 绑定的对端 WAN 网段（如 "10.32.20.0/24"）。
	BindWANAddr string `json:"bind_wan_addr"`
	// BindLANAddr 绑定的对端 LAN 网段（如 "10.33.20.0/24"）。
	BindLANAddr string `json:"bind_lan_addr"`
	// Interface 绑定线路（真机观测空串=自动）。
	Interface string `json:"interface"`
	// VethMAC 虚拟网线 MAC（固件生成，空=未建立）。
	VethMAC string `json:"veth_mac"`
	// Node 云节点编号（真机观测 0；语义待真机验证）。
	Node int64 `json:"node"`
	// Status 连接状态（真机观测 0；枚举语义待真机验证）。
	Status int64 `json:"status"`
	// Message 状态描述。
	Message string `json:"message"`
	// Timestamp 状态时间（Unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// APIToken 绑定令牌（绑定后由云平台下发）。
	APIToken string `json:"api_token"`
	// LibVer 组件版本（空=未安装）。
	LibVer string `json:"libver"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// IsEnabled 报告无影云虚拟专线是否启用。
func (r WuyingConfig) IsEnabled() bool { return r.Enabled == EnabledYes }

// WuyingPage 是 wuying.show 的完整返回形态（真机实证）。
type WuyingPage struct {
	// Routeid 设备路由 ID（云平台侧标识）。
	Routeid string `json:"routeid"`
	// Rows 配置行（单行）。
	Rows []WuyingConfig `json:"data"`
}

// VPNClientService 是 VPN 客户端族（pptp_client/l2tp_client/openvpn-client/
// ipsec-vpn/wireguard/ike_client/ik_web_sdwan/wuying）的子服务。
type VPNClientService struct {
	c *ikuai.Client
}

// PPTPList 获取 PPTP 客户端拨号列表（show 路径真机实证返回空表）。
// 对应 /Action/call: pptp_client.show（TYPE=total,data），即 Web 控制台
// 「VPN 客户端 → PPTP」页面。
func (s *VPNClientService) PPTPList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[PPTPClient], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "pptp_client", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[PPTPClient](resp.Data)
}

// PPTPAdd 新增 PPTP 客户端拨号，返回固件分配的行 ID。
// 对应 /Action/call: pptp_client.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) PPTPAdd(ctx context.Context, row PPTPClient) (int64, error) {
	resp, err := s.c.Call(ctx, "pptp_client", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// PPTPEdit 按 ID 修改 PPTP 客户端拨号。
// 对应 /Action/call: pptp_client.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) PPTPEdit(ctx context.Context, row PPTPClient) error {
	_, err := s.c.Call(ctx, "pptp_client", "edit", row)
	return err
}

// PPTPDelete 删除 PPTP 客户端拨号，支持单个或多个 ID。
// 对应 /Action/call: pptp_client.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) PPTPDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "pptp_client", "del", ikuai.IDParam(ids))
	return err
}

// L2TPList 获取 L2TP 客户端拨号列表（show 路径真机实证返回空表）。
// 对应 /Action/call: l2tp_client.show（TYPE=total,data），即 Web 控制台
// 「VPN 客户端 → L2TP」页面。
func (s *VPNClientService) L2TPList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[L2TPClient], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "l2tp_client", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[L2TPClient](resp.Data)
}

// L2TPAdd 新增 L2TP 客户端拨号，返回固件分配的行 ID。
// 对应 /Action/call: l2tp_client.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) L2TPAdd(ctx context.Context, row L2TPClient) (int64, error) {
	resp, err := s.c.Call(ctx, "l2tp_client", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// L2TPEdit 按 ID 修改 L2TP 客户端拨号。
// 对应 /Action/call: l2tp_client.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) L2TPEdit(ctx context.Context, row L2TPClient) error {
	_, err := s.c.Call(ctx, "l2tp_client", "edit", row)
	return err
}

// L2TPDelete 删除 L2TP 客户端拨号，支持单个或多个 ID。
// 对应 /Action/call: l2tp_client.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) L2TPDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "l2tp_client", "del", ikuai.IDParam(ids))
	return err
}

// OpenVPNList 获取 OpenVPN 客户端拨号列表（show 路径真机实证返回空表）。
// 对应 /Action/call: openvpn-client.show（TYPE=total,data），即 Web 控制台
// 「VPN 客户端 → OpenVPN」页面。
func (s *VPNClientService) OpenVPNList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[OpenVPNClient], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "openvpn-client", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[OpenVPNClient](resp.Data)
}

// OpenVPNAdd 新增 OpenVPN 客户端拨号，返回固件分配的行 ID。
// 对应 /Action/call: openvpn-client.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) OpenVPNAdd(ctx context.Context, row OpenVPNClient) (int64, error) {
	resp, err := s.c.Call(ctx, "openvpn-client", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// OpenVPNEdit 按 ID 修改 OpenVPN 客户端拨号。
// 对应 /Action/call: openvpn-client.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) OpenVPNEdit(ctx context.Context, row OpenVPNClient) error {
	_, err := s.c.Call(ctx, "openvpn-client", "edit", row)
	return err
}

// OpenVPNDelete 删除 OpenVPN 客户端拨号，支持单个或多个 ID。
// 对应 /Action/call: openvpn-client.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) OpenVPNDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "openvpn-client", "del", ikuai.IDParam(ids))
	return err
}

// IPSecList 获取 IPSec VPN 客户端拨号列表（show 路径真机实证返回空表）。
// 对应 /Action/call: ipsec-vpn.show（TYPE=total,data），即 Web 控制台
// 「VPN 客户端 → IPSec VPN」页面。
func (s *VPNClientService) IPSecList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IPSecClient], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ipsec-vpn", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IPSecClient](resp.Data)
}

// IPSecAdd 新增 IPSec VPN 客户端拨号，返回固件分配的行 ID。
// 对应 /Action/call: ipsec-vpn.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) IPSecAdd(ctx context.Context, row IPSecClient) (int64, error) {
	resp, err := s.c.Call(ctx, "ipsec-vpn", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// IPSecEdit 按 ID 修改 IPSec VPN 客户端拨号。
// 对应 /Action/call: ipsec-vpn.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) IPSecEdit(ctx context.Context, row IPSecClient) error {
	_, err := s.c.Call(ctx, "ipsec-vpn", "edit", row)
	return err
}

// IPSecDelete 删除 IPSec VPN 客户端拨号，支持单个或多个 ID。
// 对应 /Action/call: ipsec-vpn.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) IPSecDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ipsec-vpn", "del", ikuai.IDParam(ids))
	return err
}

// WireGuardList 获取 WireGuard 客户端拨号列表。
// 对应 /Action/call: wireguard.show（TYPE=total,data）。
// 注意：真机 3.7.24 无该模块，show 回 30002 "Not found funcname(wireguard)"；
// 错误以 *ikuai.APIError 原样返回。
func (s *VPNClientService) WireGuardList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[WireGuardClient], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "wireguard", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[WireGuardClient](resp.Data)
}

// WireGuardAdd 新增 WireGuard 客户端拨号，返回固件分配的行 ID。
// 对应 /Action/call: wireguard.add。
// ⚠️ 写动作，字段待真机验证（且真机无该模块）：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) WireGuardAdd(ctx context.Context, row WireGuardClient) (int64, error) {
	resp, err := s.c.Call(ctx, "wireguard", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// WireGuardEdit 按 ID 修改 WireGuard 客户端拨号。
// 对应 /Action/call: wireguard.edit。
// ⚠️ 写动作，字段待真机验证（且真机无该模块）：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) WireGuardEdit(ctx context.Context, row WireGuardClient) error {
	_, err := s.c.Call(ctx, "wireguard", "edit", row)
	return err
}

// WireGuardDelete 删除 WireGuard 客户端拨号，支持单个或多个 ID。
// 对应 /Action/call: wireguard.del。
// ⚠️ 写动作，字段待真机验证（且真机无该模块）：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) WireGuardDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "wireguard", "del", ikuai.IDParam(ids))
	return err
}

// IKEList 获取 IKEv2/IPSec 客户端拨号列表。
// 对应 /Action/call: ike_client.show（TYPE=total,data）。
// 注意：真机 3.7.24 无该模块，show 回 30002 "Not found funcname(ike_client)"；
// 错误以 *ikuai.APIError 原样返回。
func (s *VPNClientService) IKEList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IKEClient], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "ike_client", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IKEClient](resp.Data)
}

// IKEAdd 新增 IKEv2/IPSec 客户端拨号，返回固件分配的行 ID。
// 对应 /Action/call: ike_client.add。
// ⚠️ 写动作，字段待真机验证（且真机无该模块）：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) IKEAdd(ctx context.Context, row IKEClient) (int64, error) {
	resp, err := s.c.Call(ctx, "ike_client", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// IKEEdit 按 ID 修改 IKEv2/IPSec 客户端拨号。
// 对应 /Action/call: ike_client.edit。
// ⚠️ 写动作，字段待真机验证（且真机无该模块）：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) IKEEdit(ctx context.Context, row IKEClient) error {
	_, err := s.c.Call(ctx, "ike_client", "edit", row)
	return err
}

// IKEDelete 删除 IKEv2/IPSec 客户端拨号，支持单个或多个 ID。
// 对应 /Action/call: ike_client.del。
// ⚠️ 写动作，字段待真机验证（且真机无该模块）：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) IKEDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ike_client", "del", ikuai.IDParam(ids))
	return err
}

// SDWANStatus 获取 SD-WAN（爱快云异地组网）连接状态。
// 对应 /Action/call: ik_web_sdwan.show（{}，真机实证：
// Data.data={"code":"2020","message":"connect server failed"}）。
func (s *VPNClientService) SDWANStatus(ctx context.Context) (*SDWANStatus, error) {
	var out struct {
		Data SDWANStatus `json:"data"`
	}
	if err := s.c.CallData(ctx, "ik_web_sdwan", "show", bareShowParam(nil), &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// WuyingGet 获取无影云虚拟专线配置（单行）与设备路由 ID。
// 对应 /Action/call: wuying.show（{}，真机实证；TYPE=total,data 回
// 30006 unknown TYPE，故固定空参数）。
func (s *VPNClientService) WuyingGet(ctx context.Context) (*WuyingPage, error) {
	var out WuyingPage
	if err := s.c.CallData(ctx, "wuying", "show", bareShowParam(nil), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WuyingSet 保存无影云虚拟专线配置（写动作）。
// 对应 /Action/call: wuying.set。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNClientService) WuyingSet(ctx context.Context, cfg WuyingConfig) error {
	_, err := s.c.Call(ctx, "wuying", "set", cfg)
	return err
}
