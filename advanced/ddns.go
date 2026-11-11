package advanced

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// ddns 的解析 IP 类型常量（固件 location 字段，真机观测 0/1 两种取值，
// 语义对照控制台 i18n：location_v1=接口IP、location_v2=互联网出口IP）。
const (
	// DDNSLocationInterfaceIP 解析到接口 IP。
	DDNSLocationInterfaceIP int64 = 0
	// DDNSLocationWanIP 解析到互联网出口 IP。
	DDNSLocationWanIP int64 = 1
)

// nat_ddns.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// NATDDNSTypeData 内网穿透状态区块。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）；TYPE=all 与 data 等价。
	NATDDNSTypeData = "data"
)

// DDNSRecord 是一条动态域名解析记录（ddns.show total,data 的 data 元素，
// 真机 3.7.24 实证 18 字段、17 行）。
type DDNSRecord struct {
	// ID 记录行 ID（真机观测 1..17；edit/del 必填，add 时无需填写）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// Server 服务商域名（真机观测 "aliyun.com"/"cloudflare.com"）。
	Server string `json:"server"`
	// Username 用户名（服务商 AccessKey ID / 账号）。
	Username string `json:"username"`
	// Passwd 用户密码或 Token（服务商密钥；cf_token=1 时为 Cloudflare
	// Token，真机观测以 "cfut_" 开头）。
	Passwd string `json:"passwd"`
	// Account 账号名（控制台 i18n tokenUsername「账号/Token ID」语义，
	// 真机观测为空串；填用时机待真机验证）。
	Account string `json:"account"`
	// CFToken Cloudflare Token 模式（真机观测 0/1：1 时 passwd 存
	// Cloudflare Token 且 username 为空）。
	CFToken int64 `json:"cf_token"`
	// Interface 解析网卡（真机观测 "adsl07" 等接口名）。
	Interface string `json:"interface"`
	// Line 外网线路（多线分流场景的线路名，真机观测为空串）。
	Line string `json:"line"`
	// Type 记录类型（真机观测 "A"；IPv6 场景为 "AAAA"，控制台 i18n
	// Record_type「记录类型」）。
	Type string `json:"type"`
	// Mode 配置方式（真机观测 0；取值集待真机验证，控制台 i18n
	// config_mode「配置方式」/ Parsing_set「解析设置」）。
	Mode int64 `json:"mode"`
	// Location 解析 IP 类型（0=接口IP，1=互联网出口IP，真机观测 0/1）。
	Location int64 `json:"location"`
	// MAC 终端 MAC（记录类型 AAAA + DHCPv6 场景按终端 MAC 解析，
	// 控制台 i18n Terminal_mac；真机观测为空串）。
	MAC string `json:"mac"`
	// DUID 终端 DUID（DHCPv6 有状态场景按终端 DUID 解析，控制台 i18n
	// Terminal_duid；真机观测为空串）。
	DUID string `json:"duid"`
	// TopDomain 主域名（真机观测 "example.com"/"example.org"）。
	TopDomain string `json:"top_domain"`
	// Domain 完整域名 FQDN（真机观测 "ddns.example.com"）。
	Domain string `json:"domain"`
	// IPAddress 当前解析到的 IP（真机观测公网 IPv4）。
	IPAddress string `json:"ipaddress"`
	// Result 最近一次更新结果（真机观测「成功」；控制台 i18n
	// update_log「更新日志」/ result「更新结果」）。
	Result string `json:"result"`
}

// IsEnabled 记录启用状态为 "yes"。
func (r DDNSRecord) IsEnabled() bool { return r.Enabled == EnabledYes }

// DDNSService 是动态域名（ddns）的子服务。
type DDNSService struct {
	c *ikuai.Client
}

// List 获取动态域名解析记录列表（分页）。
// 对应 /Action/call: ddns.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":17,"data":[...]}），
// 即 Web 控制台「高级应用 → 动态域名」页面。
func (s *DDNSService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DDNSRecord], error) {
	resp, err := s.c.Call(ctx, "ddns", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DDNSRecord](resp.Data)
}

// Add 新增动态域名解析记录，返回固件分配的行 ID。
// 对应 /Action/call: ddns.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *DDNSService) Add(ctx context.Context, rec DDNSRecord) (int64, error) {
	resp, err := s.c.Call(ctx, "ddns", "add", rec)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 修改动态域名解析记录。
// 对应 /Action/call: ddns.edit（rec.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *DDNSService) Edit(ctx context.Context, rec DDNSRecord) error {
	_, err := s.c.Call(ctx, "ddns", "edit", rec)
	return err
}

// Delete 删除动态域名解析记录，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: ddns.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *DDNSService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ddns", "del", ikuai.IDParam(ids))
	return err
}

// NATDDNSStatus 是内网穿透（NAT DDNS）的功能状态（nat_ddns.show 的
// data 区块，真机 3.7.24 实证）。
type NATDDNSStatus struct {
	// Status 功能状态码（真机观测 1 且 errinfo="Function not enabled"，
	// 即未开通；其余取值语义待真机验证）。
	Status int64 `json:"status"`
	// ErrInfo 状态说明（真机观测 "Function not enabled"；对照控制台
	// i18n：未开通/已开启/连接服务器失败/账号过期）。
	ErrInfo string `json:"errinfo"`
	// Username 绑定的爱快云账号（未绑定为空串）。
	Username string `json:"username"`
	// Version 服务版本（真机观测为空串）。
	Version string `json:"version"`
}

// natDDNSDataWrapper 对应 nat_ddns.show 的 Data 外层 {"data": {...}}
// 嵌套形态（真机实证）。
type natDDNSDataWrapper struct {
	Data NATDDNSStatus `json:"data"`
}

// NATDDNSService 是 NAT DDNS（nat_ddns，内网穿透）的子服务。
type NATDDNSService struct {
	c *ikuai.Client
}

// Get 获取内网穿透功能状态（是否开通、绑定账号等）。
// 对应 /Action/call: nat_ddns.show（TYPE=data），
// 即 Web 控制台「高级应用 → NAT DDNS」页面。
// 开通/关闭内网穿透业务需在爱快云平台操作，固件侧无写接口。
func (s *NATDDNSService) Get(ctx context.Context) (*NATDDNSStatus, error) {
	param := typedShowParam(nil, NATDDNSTypeData, nil)
	var w natDDNSDataWrapper
	if err := s.c.CallData(ctx, "nat_ddns", "show", param, &w); err != nil {
		return nil, err
	}
	return &w.Data, nil
}
