package accounting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// Authmode 枚举常量：PPPoE 服务端认证方式（pppoe_server.authmode，真机
// 观测 0；取值语义按固件 i18n pppoe_server 选项顺序推导，待真机验证）。
const (
	// AuthmodeLocalAccount 本地账号认证。
	AuthmodeLocalAccount int64 = 0
	// AuthmodeLocalEmptyPassword 本地账号+空密码认证。
	AuthmodeLocalEmptyPassword int64 = 1
	// AuthmodeAnyUser 任意用户拨号认证。
	AuthmodeAnyUser int64 = 2
	// AuthmodeRadius 第三方 Radius 认证。
	AuthmodeRadius int64 = 3
)

// PPPoEServerConfig 是 PPPoE 服务端单行配置（pppoe_server.show 行，
// 真机 3.7.24 实证 33 字段；语义按固件 i18n pppoe_server 页面标注）。
type PPPoEServerConfig struct {
	// ID 行 ID（恒为 1，save 时随行携带）。
	ID int64 `json:"id"`
	// Enabled PPPoE 服务端开关（"yes"/"no"；真机观测 "no"）。
	Enabled string `json:"enabled"`
	// ServerName PPPoE 服务名称（真机观测 "iKuai"）。
	ServerName string `json:"server_name"`
	// ServerIP 服务器地址（真机观测 "10.1.1.1"）。
	ServerIP string `json:"server_ip"`
	// Interface 内网线路（真机观测 "lan1"）。
	Interface string `json:"interface"`
	// AddrPool 客户端地址池（如 "10.1.1.2-10.1.1.254"，每线路仅一条）。
	AddrPool string `json:"addr_pool"`
	// DNS1 首选 DNS（真机观测 "114.114.114.114"）。
	DNS1 string `json:"dns1"`
	// DNS2 备选 DNS（真机观测 "119.29.29.29"）。
	DNS2 string `json:"dns2"`
	// Authmode 认证方式（0=本地账号认证，见 Authmode* 常量）。
	Authmode int64 `json:"authmode"`
	// ShareDeny 超出账号共享数动作（1=限制登录，0=替换当前用户；语义待真机验证）。
	ShareDeny int64 `json:"share_deny"`
	// NasIdentifier NAS 标识（真机观测 "iKuai"）。
	NasIdentifier string `json:"nas_identifier"`
	// NasIPAddress NAS IP（Radius 认证时使用）。
	NasIPAddress string `json:"nas_ip_address"`
	// RadiusIP Radius 服务端 IP（真机观测 "127.0.0.1"）。
	RadiusIP string `json:"radius_ip"`
	// Secret Radius 共享密钥（真机观测 "123"）。
	Secret string `json:"secret"`
	// Authport Radius 认证端口（真机观测 1812）。
	Authport int64 `json:"authport"`
	// Accountport Radius 计费端口（真机观测 1813）。
	Accountport int64 `json:"accountport"`
	// Comment 备注。
	Comment string `json:"comment"`
	// RateLimitLan 客户端 LAN 互访限速（0/1）。
	RateLimitLan int64 `json:"rate_limit_lan"`
	// DropClient 禁止客户端互访（0/1；真机观测 1）。
	DropClient int64 `json:"drop_client"`
	// ForcePPPoE 强制客户端拨号上网（0/1）。
	ForcePPPoE int64 `json:"force_pppoe"`
	// EnhanceCheck 加强断线检测（0/1；真机观测 1）。
	EnhanceCheck int64 `json:"enhance_check"`
	// BindVlan 支持账号绑定 VLAN（0/1；单层 VLAN 与 QinQ）。
	BindVlan int64 `json:"bind_vlan"`
	// VerifyVlan 启用账号 VLAN 校验（0/1；真机观测 1）。
	VerifyVlan int64 `json:"verify_vlan"`
	// BindIface 支持账号绑定网卡（0/1）。
	BindIface int64 `json:"bind_iface"`
	// ForceVerifyName 强制校验服务名称（0/1）。
	ForceVerifyName int64 `json:"force_verify_name"`
	// MTU 最大传输单元（真机观测 1480）。
	MTU int64 `json:"mtu"`
	// MRU 最大接收单元（真机观测 1480）。
	MRU int64 `json:"mru"`
	// LCPEchoInterval LCP 探测间隔时间（秒，真机观测 10）。
	LCPEchoInterval int64 `json:"lcp_echo_interval"`
	// LCPEchoFailure LCP 探测失败次数（真机观测 3）。
	LCPEchoFailure int64 `json:"lcp_echo_failure"`
	// Maxconnect 定时断开用户连接的时长（秒，0=无限制）。
	Maxconnect int64 `json:"maxconnect"`
	// RestartTimer 定时断开重启服务器（0/1）。
	RestartTimer int64 `json:"restart_timer"`
	// RestartWeek 定时重启星期（真机观测 1234567，整型）。
	RestartWeek int64 `json:"restart_week"`
	// RestartTime 定时重启时刻（真机观测 "06:00"）。
	RestartTime string `json:"restart_time"`
}

// IsEnabled 报告 PPPoE 服务端是否开启。
func (c PPPoEServerConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// PPTPServerConfig 是 PPTP 服务端单行配置（pptp_server.show 行，
// 真机 3.7.24 实证 10 字段）。
type PPTPServerConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Enabled PPTP 服务端开关（"yes"/"no"；真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// ServerIP 服务器地址（真机观测 "10.0.0.1"）。
	ServerIP string `json:"server_ip"`
	// ServerPort 服务端口（真机观测 1723）。
	ServerPort int64 `json:"server_port"`
	// AddrPool 客户端地址池（真机观测 "10.0.0.2-10.0.0.254"）。
	AddrPool string `json:"addr_pool"`
	// DNS1 首选 DNS（真机观测 "223.5.5.5"）。
	DNS1 string `json:"dns1"`
	// DNS2 备选 DNS（真机观测 "119.29.29.29"）。
	DNS2 string `json:"dns2"`
	// OpenMPPE MPPE 加密（真机观测 2；枚举语义待真机验证，推测 0=关闭）。
	OpenMPPE int64 `json:"open_mppe"`
	// MTU 最大传输单元（真机观测 1400）。
	MTU int64 `json:"mtu"`
	// MRU 最大接收单元（真机观测 1400）。
	MRU int64 `json:"mru"`
}

// IsEnabled 报告 PPTP 服务端是否开启。
func (c PPTPServerConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// L2TPServerConfig 是 L2TP 服务端单行配置（l2tp_server.show 行，
// 真机 3.7.24 实证 13 字段）。
type L2TPServerConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Enabled L2TP 服务端开关（"yes"/"no"；真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// ServerIP 服务器地址（真机观测 "10.1.0.1"）。
	ServerIP string `json:"server_ip"`
	// ServerPort 服务端口（真机观测 1701）。
	ServerPort int64 `json:"server_port"`
	// AddrPool 客户端地址池（真机观测 "10.1.0.2-10.1.0.254"）。
	AddrPool string `json:"addr_pool"`
	// DNS1 首选 DNS（真机观测 "223.5.5.5"）。
	DNS1 string `json:"dns1"`
	// DNS2 备选 DNS（真机观测 "119.29.29.29"）。
	DNS2 string `json:"dns2"`
	// ForceIPsec 强制 IPsec 加密（0/1）。
	ForceIPsec int64 `json:"force_ipsec"`
	// IPsecSecret IPsec 预共享密钥（ForceIPsec=1 时使用）。
	IPsecSecret string `json:"ipsec_secret"`
	// LeftID 本端标识（IPsec 证书/PSK 标识）。
	LeftID string `json:"leftid"`
	// RightID 对端标识（IPsec 证书/PSK 标识）。
	RightID string `json:"rightid"`
	// MTU 最大传输单元（真机观测 1400）。
	MTU int64 `json:"mtu"`
	// MRU 最大接收单元（真机观测 1400）。
	MRU int64 `json:"mru"`
}

// IsEnabled 报告 L2TP 服务端是否开启。
func (c L2TPServerConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// OpenVPNServerConfig 是 OpenVPN 服务端单行配置（openvpn-server.show 行，
// 真机 3.7.24 实证 21 字段）。证书/密钥类字段为固件自有编码：PEM 内容
// 换行以 "@" 分隔、空格以 "#" 替代（真机实证），辅助方法见
// DecodePEMField。
type OpenVPNServerConfig struct {
	// Port 服务端口（固件以字符串承载，真机观测 "1194"）。
	Port string `json:"port"`
	// Subnet VPN 网段（真机观测 "10.7.7.0"）。
	Subnet string `json:"subnet"`
	// Mask 网段掩码（真机观测 "255.255.255.0"）。
	Mask string `json:"mask"`
	// TunMTU 隧道 MTU（固件以字符串承载，真机观测 "1400"）。
	TunMTU string `json:"tun_mtu"`
	// Proto 隧道协议（真机观测 "udp"；tcp/udp）。
	Proto string `json:"proto"`
	// Cipher 加密算法（真机观测 "BF-CBC"）。
	Cipher string `json:"cipher"`
	// Auth 摘要算法（空=默认；语义待真机验证）。
	Auth string `json:"auth"`
	// CompLZO LZO 压缩（固件以字符串承载，"1"/"0"）。
	CompLZO string `json:"comp_lzo"`
	// DevType 隧道类型（真机观测 "tun"）。
	DevType string `json:"dev_type"`
	// Topology 拓扑类型（真机观测 "subnet"）。
	Topology string `json:"topology"`
	// Method 配置方法（真机观测 0；枚举语义待真机验证）。
	Method int64 `json:"method"`
	// TLSAuth TLS-Auth 静态密钥（固件编码，见 DecodePEMField）。
	TLSAuth string `json:"tls_auth"`
	// CA CA 证书（固件编码，见 DecodePEMField）。
	CA string `json:"ca"`
	// Cert 服务器证书（固件编码，见 DecodePEMField）。
	Cert string `json:"cert"`
	// Key 服务器私钥（固件编码，见 DecodePEMField）。
	Key string `json:"key"`
	// PushRoute 推送路由（真机观测 "10.7.0.0/16"）。
	PushRoute string `json:"push_route"`
	// PushRouteComment 推送路由备注。
	PushRouteComment string `json:"push_route_comment"`
	// PushDNS 推送 DNS。
	PushDNS string `json:"push_dns"`
	// PushGateway 推送网关（固件以字符串承载，"0"=不推送）。
	PushGateway string `json:"push_gateway"`
	// ExtraConfig 附加配置（多行）。
	ExtraConfig string `json:"extra_config"`
	// Enabled OpenVPN 服务端开关（"yes"/"no"；真机观测 "no"）。
	Enabled string `json:"enabled"`
}

// IsEnabled 报告 OpenVPN 服务端是否开启。
func (c OpenVPNServerConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// DecodePEMField 将固件编码的证书/密钥字段还原为标准 PEM 文本：
// 真机实证编码规则为 PEM 换行以 "@" 分隔、空格以 "#" 替代。
func DecodePEMField(encoded string) string {
	if encoded == "" {
		return ""
	}
	return strings.ReplaceAll(strings.ReplaceAll(encoded, "@", "\n"), "#", " ")
}

// IKEv2Server 是 IPSec IKEv2 服务端配置行（ike_server.show 行）。
//
// 真机 3.7.24（iKuaiOS x64 Build202608141210）对该 func_name 一律返回
// 30002 "Not found funcname(ike_server)"（控制台菜单含此页，推测需认证
// 计费/VPN 授权激活后才下发 func），行字段无法实证，故以 Raw 原样保留；
// 字段语义待可用的固件验证（参考同族 ike_client 与固件 i18n ipsec_ikev2）。
type IKEv2Server struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (s *IKEv2Server) UnmarshalJSON(b []byte) error {
	s.Raw = append(s.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (s IKEv2Server) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(s.Raw, &m); err != nil {
		return nil, fmt.Errorf("accounting: 解码 ike_server 行失败: %w", err)
	}
	return m, nil
}

// VPNServersService 是 VPN 服务端族（pppoe_server/pptp_server/l2tp_server/
// openvpn-server/ike_server）的子服务。五个 func 的 show 均为单行配置
// （真机实证 TYPE=data；TYPE=total,data 回 30006 unknown TYPE）。
type VPNServersService struct {
	c *ikuai.Client
}

// getRow 以 TYPE=data 调用指定 func 的 show 并取第一行（配置类通用形态）。
func (s *VPNServersService) getRow(ctx context.Context, funcName string, out any) error {
	return s.c.CallData(ctx, funcName, "show", map[string]any{"TYPE": cfgTypeData}, out)
}

// PPPoE 获取 PPPoE 服务端配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: pppoe_server.show（TYPE=data，真机实证），即 Web
// 控制台「认证计费 → VPN 服务端 → PPPoE 服务端」页面。
func (s *VPNServersService) PPPoE(ctx context.Context) (*PPPoEServerConfig, error) {
	var out struct {
		Data []PPPoEServerConfig `json:"data"`
	}
	if err := s.getRow(ctx, "pppoe_server", &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetPPPoE 保存 PPPoE 服务端配置（整行保存，随行携带 ID）。
// 对应 /Action/call: pppoe_server.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNServersService) SetPPPoE(ctx context.Context, cfg PPPoEServerConfig) error {
	_, err := s.c.Call(ctx, "pppoe_server", "save", cfg)
	return err
}

// PPTP 获取 PPTP 服务端配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: pptp_server.show（TYPE=data，真机实证）。
func (s *VPNServersService) PPTP(ctx context.Context) (*PPTPServerConfig, error) {
	var out struct {
		Data []PPTPServerConfig `json:"data"`
	}
	if err := s.getRow(ctx, "pptp_server", &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetPPTP 保存 PPTP 服务端配置（整行保存，随行携带 ID）。
// 对应 /Action/call: pptp_server.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNServersService) SetPPTP(ctx context.Context, cfg PPTPServerConfig) error {
	_, err := s.c.Call(ctx, "pptp_server", "save", cfg)
	return err
}

// L2TP 获取 L2TP 服务端配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: l2tp_server.show（TYPE=data，真机实证）。
func (s *VPNServersService) L2TP(ctx context.Context) (*L2TPServerConfig, error) {
	var out struct {
		Data []L2TPServerConfig `json:"data"`
	}
	if err := s.getRow(ctx, "l2tp_server", &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetL2TP 保存 L2TP 服务端配置（整行保存，随行携带 ID）。
// 对应 /Action/call: l2tp_server.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNServersService) SetL2TP(ctx context.Context, cfg L2TPServerConfig) error {
	_, err := s.c.Call(ctx, "l2tp_server", "save", cfg)
	return err
}

// OpenVPN 获取 OpenVPN 服务端配置。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: openvpn-server.show（TYPE=data，真机实证）。
func (s *VPNServersService) OpenVPN(ctx context.Context) (*OpenVPNServerConfig, error) {
	var out struct {
		Data []OpenVPNServerConfig `json:"data"`
	}
	if err := s.getRow(ctx, "openvpn-server", &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// SetOpenVPN 保存 OpenVPN 服务端配置（整行保存，随行携带 ID）。
// 对应 /Action/call: openvpn-server.save（save 动作参照同构配置页
// alg.save 抓包）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *VPNServersService) SetOpenVPN(ctx context.Context, cfg OpenVPNServerConfig) error {
	_, err := s.c.Call(ctx, "openvpn-server", "save", cfg)
	return err
}

// IKEv2 获取 IPSec IKEv2 服务端配置。
// 对应 /Action/call: ike_server.show（TYPE=data）。
// 真机 3.7.24 未授权时返回 30002 "Not found funcname(ike_server)"（真机
// 实证），调用方应预期 *ikuai.APIError；func 可用时按行解码。
func (s *VPNServersService) IKEv2(ctx context.Context) (*IKEv2Server, error) {
	var out struct {
		Data []IKEv2Server `json:"data"`
	}
	if err := s.getRow(ctx, "ike_server", &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}
