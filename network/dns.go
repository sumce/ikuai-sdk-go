package network

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// dns.show 的 TYPE 选择器常量（真机实证：其他 TYPE 一律 30006 unknown TYPE）。
const (
	// DNSTypeConfig DNS 设置单行数据（真机实证，3.7.24）。
	DNSTypeConfig = "dns_config"
)

// DNSConfig 是 DNS 设置单行数据（dns.show TYPE=dns_config 行，真机实证）。
type DNSConfig struct {
	// ID 行 ID（恒为 1）。
	ID int64 `json:"id"`
	// Enabled DNS 代理开关（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// DNS1 首选 DNS。
	DNS1 string `json:"dns1"`
	// DNS2 备选 DNS。
	DNS2 string `json:"dns2"`
	// Cachemode DNS 缓存模式（0=关闭；枚举语义待真机验证）。
	Cachemode int64 `json:"cachemode"`
	// CacheTTL DNS 缓存 TTL（秒，真机观测 300）。
	CacheTTL int64 `json:"cache_ttl"`
	// ForbidDNS4A 禁止 AAAA 记录(IPv6)解析（0/1）。
	ForbidDNS4A int64 `json:"forbid_dns_4a"`
	// ProxyForce 强制客户端 DNS 代理（0/1）。
	ProxyForce int64 `json:"proxy_force"`
	// ProxyForceDNS 强制代理时下发的 DNS 地址（ProxyForce=1 时使用）。
	ProxyForceDNS string `json:"proxy_force_dns"`
	// Network DNS 反向代理开关（0/1；语义待真机验证）。
	Network int64 `json:"network"`
	// Defense DNS 防御开关（0/1；语义待真机验证）。
	Defense int64 `json:"defense"`
	// Query DoH 请求地址（如 "https://doh.pub/dns-query"，空=未启用）。
	Query string `json:"query"`
	// QueryHeadIP DoH 请求头 IP（字段语义待真机验证）。
	QueryHeadIP string `json:"query_head_ip"`
	// QueryArgsIP DoH 请求参数 IP（字段语义待真机验证）。
	QueryArgsIP string `json:"query_args_ip"`
}

// IsEnabled 报告 DNS 代理是否开启。
func (r DNSConfig) IsEnabled() bool { return r.Enabled == EnabledYes }

// DNSReplace 是一条多线 DNS 分流规则（dns_replace.show 行；真机列表为空，
// 字段名由固件 i18n 推导，待真机验证）。
type DNSReplace struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Interface 策略线路（匹配该线路的 DNS 请求走指定服务器）。
	Interface string `json:"interface"`
	// DNS1 首选 DNS。
	DNS1 string `json:"dns1"`
	// DNS2 备选 DNS。
	DNS2 string `json:"dns2"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// DNSService 是 DNS 设置族（dns/dns_replace）的子服务。
type DNSService struct {
	c *ikuai.Client
}

// Get 获取 DNS 设置（DNS 代理/缓存/DoH/强制代理等）。
// 对应 /Action/call: dns.show（TYPE=dns_config，真机 3.7.24 实证；
// 其他 TYPE 值固件回 30006 unknown TYPE），即 Web 控制台
// 「网络设置 → DNS 设置」页面。
func (s *DNSService) Get(ctx context.Context) (*DNSConfig, error) {
	var out struct {
		Data []DNSConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "dns", "show",
		map[string]any{"TYPE": DNSTypeConfig}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// Set 保存 DNS 设置（写动作）。
// 对应 /Action/call: dns.set（基础配置类保存动作，见 docs/RULES.md §2）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DNSService) Set(ctx context.Context, cfg DNSConfig) error {
	_, err := s.c.Call(ctx, "dns", "set", cfg)
	return err
}

// ReplaceList 获取多线 DNS 分流规则列表。
// 对应 /Action/call: dns_replace.show（TYPE=total,data），即 Web 控制台
// 「网络设置 → DNS 设置 → 多线 DNS」页面。
func (s *DNSService) ReplaceList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DNSReplace], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dns_replace", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DNSReplace](resp.Data)
}

// ReplaceAdd 新增多线 DNS 分流规则，返回固件分配的行 ID。
// 对应 /Action/call: dns_replace.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DNSService) ReplaceAdd(ctx context.Context, row DNSReplace) (int64, error) {
	resp, err := s.c.Call(ctx, "dns_replace", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// ReplaceEdit 按 ID 修改多线 DNS 分流规则。
// 对应 /Action/call: dns_replace.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DNSService) ReplaceEdit(ctx context.Context, row DNSReplace) error {
	_, err := s.c.Call(ctx, "dns_replace", "edit", row)
	return err
}

// ReplaceDelete 删除多线 DNS 分流规则，支持单个或多个 ID。
// 对应 /Action/call: dns_replace.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *DNSService) ReplaceDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dns_replace", "del", ikuai.IDParam(ids))
	return err
}
