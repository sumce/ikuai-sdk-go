package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// 系统体检各检测项状态枚举值（*_status 字段；"-1" 为真机实测未检测
// 形态，其余取值语义待真机验证）。
const (
	// SysCheckStatusNone 未检测（真机实测 "-1"，res 为空数组）。
	SysCheckStatusNone = "-1"
)

// SysCheckItem 是一个检测项的单条结果（*_res 数组元素）。
//
// 真机 3.7.24 未进行过体检，各 res 数组为空，元素字段无法实证，
// 行结构以 Raw 形态原样保留，字段语义待真机验证。
type SysCheckItem struct {
	// Raw 元素的原始 JSON。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留元素的原始 JSON（元素结构未知时不报错）。
func (i *SysCheckItem) UnmarshalJSON(b []byte) error {
	i.Raw = append(i.Raw[:0], b...)
	return nil
}

// Map 将元素原始 JSON 解码为通用 map（字段语义待真机验证时的通道）。
func (i SysCheckItem) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(i.Raw, &m); err != nil {
		return nil, fmt.Errorf("tools: 解码体检项失败: %w", err)
	}
	return m, nil
}

// SysCheckState 是系统体检的当前状态（iksyscheck.show 的 Data，真机
// 3.7.24 实测：7 个检测项各有一组 res/status 字段，未体检时 res 为
// 空数组、status 为 "-1"）。
type SysCheckState struct {
	// DiskRes 硬盘检测结果明细（i18n route_test.disk_test）。
	DiskRes []SysCheckItem `json:"disk_res"`
	// DiskStatus 硬盘检测状态（枚举 SysCheckStatus*）。
	DiskStatus string `json:"disk_status"`
	// MemoryRes 内存检测结果明细（i18n route_test.mem_test）。
	MemoryRes []SysCheckItem `json:"memory_res"`
	// MemoryStatus 内存检测状态。
	MemoryStatus string `json:"memory_status"`
	// InternetRes 线路检测结果明细（i18n route_test.wan_test）。
	InternetRes []SysCheckItem `json:"internet_res"`
	// InternetStatus 线路检测状态。
	InternetStatus string `json:"internet_status"`
	// DHCPDiscoverRes 内网 DHCP 检测结果明细（i18n route_test.dhcp_test）。
	DHCPDiscoverRes []SysCheckItem `json:"dhcp_discover_res"`
	// DHCPDiscoverStatus 内网 DHCP 检测状态。
	DHCPDiscoverStatus string `json:"dhcp_discover_status"`
	// PPPoEDiscoverRes 内网 PPPoE 检测结果明细（i18n route_test.pppoe_test）。
	PPPoEDiscoverRes []SysCheckItem `json:"pppoe_discover_res"`
	// PPPoEDiscoverStatus 内网 PPPoE 检测状态。
	PPPoEDiscoverStatus string `json:"pppoe_discover_status"`
	// ARPRes 网关冲突检测结果明细（i18n route_test.arp_test）。
	ARPRes []SysCheckItem `json:"arp_res"`
	// ARPStatus 网关冲突检测状态。
	ARPStatus string `json:"arp_status"`
	// CloudLinkRes 云平台连接体检结果明细（i18n route_test.cloud）。
	CloudLinkRes []SysCheckItem `json:"cloud_link_res"`
	// CloudLinkStatus 云平台连接体检状态。
	CloudLinkStatus string `json:"cloud_link_status"`
}

// SysCheckService 是系统诊断（iksyscheck）的子服务
// （tools.route-test「路由体检」页面）。
type SysCheckService struct {
	c *ikuai.Client
}

// Result 查询最近一次系统体检的结果（只读）。
// 对应 /Action/call: iksyscheck.show（空参数 {}，真机 3.7.24 实测返回
// 7 个检测项的 res/status 区块；未体检时 status 为 "-1"），
// 即 Web 控制台「应用工具 → 路由体检」页面。
func (s *SysCheckService) Result(ctx context.Context) (*SysCheckState, error) {
	var out SysCheckState
	if err := s.c.CallData(ctx, "iksyscheck", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Execute 发起一次全面路由体检（诊断动作）。
// 对应 /Action/call: iksyscheck.begin_test（param {}——action 名按
// i18n route_test.begin_test「开始体检」按钮推断，未对真机调用，
// 待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用：体检包含硬盘/内存/线路/DHCP/
// PPPoE/ARP/云平台连接等真实检测，仅离线测试覆盖。
func (s *SysCheckService) Execute(ctx context.Context) error {
	_, err := s.c.Call(ctx, "iksyscheck", "begin_test", map[string]any{})
	return err
}

// Stop 停止进行中的系统体检（诊断动作）。
// 对应 /Action/call: iksyscheck.stop_test（action 名按 i18n
// route_test.stop_test「停止体检」按钮推断，未对真机调用，待真机
// 验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用，仅离线测试覆盖。
func (s *SysCheckService) Stop(ctx context.Context) error {
	_, err := s.c.Call(ctx, "iksyscheck", "stop_test", map[string]any{})
	return err
}
