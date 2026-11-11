package advanced

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// udp_proxy.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// UDPProxyTypeAll 规则列表 + 可选接口区块。
	UDPProxyTypeAll = "all"
)

// UDPProxyRule 是一条 UDP 代理（udpxy/IPTV 单线复用）规则
// （udp_proxy.show data 数组元素）。
//
// 行字段空表推导，待真机验证：真机 3.7.24 未配置 UDP 代理（data 恒为
// 空数组），字段名按控制台 i18n service.iptv.header_title
// （Signal_interface=信号源接口、Service_port=服务端口、
// renew_time=订阅周期(秒)、udpxy_net_enabled=外网访问）推导。
type UDPProxyRule struct {
	// ID 规则行 ID（edit/del 必填；add 时无需填写）。
	ID int64 `json:"id,omitempty"`
	// Interface 信号源接口（IPTV 组播信号所在接口，可选值见
	// UDPProxyInfo.Interface）。
	Interface string `json:"interface"`
	// Port 服务端口（udpxy HTTP 监听端口）。
	Port int64 `json:"port"`
	// RenewTime 订阅周期（秒，组播订阅刷新间隔，控制台 i18n
	// renew_time「订阅周期(秒)」）。
	RenewTime int64 `json:"renew_time"`
	// NetEnabled 允许外网访问（控制台 i18n udpxy_net_enabled，取值
	// yes/no 待真机验证）。
	NetEnabled string `json:"udpxy_net_enabled"`
	// Enabled 是否启用（"yes"/"no"，空表推导，待真机验证）。
	Enabled string `json:"enabled"`
}

// IsEnabled 规则启用状态为 "yes"。
func (r UDPProxyRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// UDPProxyInfo 是 UDP 代理的规则列表与可选接口区块
// （udp_proxy.show TYPE=all，真机 3.7.24 实证顶层字段）。
type UDPProxyInfo struct {
	// Total 规则数（真机观测 0）。
	Total int64 `json:"total"`
	// Rules 规则行列表。
	Rules []UDPProxyRule `json:"data"`
	// Interface 可选接口列表，每项 [接口名] 或 [接口名, 备注]
	// （真机实证，含内网/VLAN/外网接口）。
	Interface [][]string `json:"interface"`
}

// UDPProxyService 是 UDP 代理（udp_proxy，udpxy/IPTV 单线复用）的子
// 服务。
type UDPProxyService struct {
	c *ikuai.Client
}

// List 获取 UDP 代理规则列表（分页）。
// 对应 /Action/call: udp_proxy.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → UDP 代理（IPTV）」页面。
func (s *UDPProxyService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[UDPProxyRule], error) {
	resp, err := s.c.Call(ctx, "udp_proxy", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[UDPProxyRule](resp.Data)
}

// Get 获取 UDP 代理规则列表与可选接口区块。
// 对应 /Action/call: udp_proxy.show（TYPE=all，真机探测实证）。
func (s *UDPProxyService) Get(ctx context.Context) (*UDPProxyInfo, error) {
	var info UDPProxyInfo
	if err := s.c.CallData(ctx, "udp_proxy", "show",
		map[string]any{"TYPE": UDPProxyTypeAll}, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Add 新增 UDP 代理规则，返回固件分配的行 ID。
// 对应 /Action/call: udp_proxy.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *UDPProxyService) Add(ctx context.Context, rule UDPProxyRule) (int64, error) {
	resp, err := s.c.Call(ctx, "udp_proxy", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 修改 UDP 代理规则。
// 对应 /Action/call: udp_proxy.edit（rule.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *UDPProxyService) Edit(ctx context.Context, rule UDPProxyRule) error {
	_, err := s.c.Call(ctx, "udp_proxy", "edit", rule)
	return err
}

// Delete 删除 UDP 代理规则，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: udp_proxy.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *UDPProxyService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "udp_proxy", "del", ikuai.IDParam(ids))
	return err
}
