package accounting

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// ProxyAccount 是一条代拨（代理）账号（pppoe_proxy.show 行，菜单
// 「代理账号」子页）。
//
// 真机 3.7.24（iKuaiOS x64 Build202608141210）对 func_name pppoe_proxy
// 一律返回 30002 "Not found funcname(pppoe_proxy)"（控制台菜单含「代理
// 账号」「在线账号」两个子页，推测需认证计费授权激活后才下发 func），
// 行字段无法实证。控制台列（固件 i18n agent_service）为：代拨线路名称
// agent_line_name / 代拨账号 agent_account / 代拨密码 agent_password /
// 代拨接口 agent_interface / 被代拨账号 account_to_be_dialed / 子网掩码
// netmask / 网关 gateway / 备注 comment。故行对象以 Raw 原样保留，
// 字段语义待真机验证后再细化强类型。
type ProxyAccount struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (a *ProxyAccount) UnmarshalJSON(b []byte) error {
	a.Raw = append(a.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (a ProxyAccount) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(a.Raw, &m); err != nil {
		return nil, fmt.Errorf("accounting: 解码 pppoe_proxy 行失败: %w", err)
	}
	return m, nil
}

// ProxyOnlineAccount 是一条代理在线账号（pppoe_proxy.show 行，菜单
// 「在线账号」子页）。与代理账号共用 func_name（固件 menu.json 中
// pppoe_proxy 出现两次），TYPE 区分方式与行字段均无法实证（真机返回
// 30002），故以 Raw 原样保留，字段语义待真机验证。
type ProxyOnlineAccount struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (a *ProxyOnlineAccount) UnmarshalJSON(b []byte) error {
	a.Raw = append(a.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (a ProxyOnlineAccount) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(a.Raw, &m); err != nil {
		return nil, fmt.Errorf("accounting: 解码 pppoe_proxy 在线账号行失败: %w", err)
	}
	return m, nil
}

// ProxyService 是代理账号/在线账号（pppoe_proxy）的子服务。
//
// 固件 menu.json 中 pppoe_proxy 挂在「认证计费 → 代理服务」下两个子页
// （agent-account 代理账号列表、online-account 在线账号列表）。真机
// 3.7.24 未授权时两个方法均返回 30002 "Not found funcname(pppoe_proxy)"
// （真机实证），调用方应预期 *ikuai.APIError；func 可用时按标准分页解码。
type ProxyService struct {
	c *ikuai.Client
}

// AgentAccounts 获取代理（代拨）账号列表（分页）。
// 对应 /Action/call: pppoe_proxy.show（TYPE=total,data, limit="offset,rows"；
// 两子页的 TYPE 区分方式待真机验证，func 可用后按行结构核对）。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *ProxyService) AgentAccounts(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ProxyAccount], error) {
	resp, err := s.c.Call(ctx, "pppoe_proxy", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ProxyAccount](resp.Data)
}

// OnlineAccounts 获取代理在线账号列表（分页）。
// 对应 /Action/call: pppoe_proxy.show（与 AgentAccounts 共用 func_name；
// 两子页的 TYPE 区分方式待真机验证）。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *ProxyService) OnlineAccounts(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ProxyOnlineAccount], error) {
	resp, err := s.c.Call(ctx, "pppoe_proxy", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ProxyOnlineAccount](resp.Data)
}
