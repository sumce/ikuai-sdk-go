package ac

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// ac_net_optimize.show 返回的 mode/frequencies 配置值（真机 3.7.24
// 观测值 0；枚举语义待真机验证——推测 mode 为优化开关/模式、
// frequencies 为频段选择）。
const (
	// NetOptimizeModeOff 无线网优关闭（真机观测值）。
	NetOptimizeModeOff = int64(0)
	// NetOptimizeFreqAll 频段选择观测值 0（语义待真机验证）。
	NetOptimizeFreqAll = int64(0)
)

// NetOptimizeInfo 是无线网优页的聚合返回（ac_net_optimize.show）。
type NetOptimizeInfo struct {
	// Total 优化结果行总数。
	Total int64 `json:"total"`
	// Data 优化结果行（真机无 AP 恒为空数组，行结构待真机验证）。
	Data []ACOptimizeItem `json:"data"`
	// Mode 无线网优模式（真机观测 0；固件对该值为持久化回显——请求参数
	// 传 mode/frequencies 不影响返回值，真机探测实证；枚举语义待真机验证）。
	Mode int64 `json:"mode"`
	// Frequencies 频段选择（真机观测 0；语义待真机验证）。
	Frequencies int64 `json:"frequencies"`
}

// ACOptimizeItem 是无线网优的一行体检结果（ac_net_optimize.show data
// 数组元素）。
//
// 真机 3.7.24 无 AP 设备，data 恒为空数组，行字段无法实证，故行对象
// 以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type ACOptimizeItem struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (i *ACOptimizeItem) UnmarshalJSON(b []byte) error {
	i.Raw = append(i.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (i ACOptimizeItem) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(i.Raw, &m); err != nil {
		return nil, fmt.Errorf("ac: 解码 ac_net_optimize 行失败: %w", err)
	}
	return m, nil
}

// NetOptimizeConfig 是无线网优的保存配置（ac_net_optimize.set 参数）。
type NetOptimizeConfig struct {
	// Mode 无线网优模式（真机观测值 0；枚举语义待真机验证）。
	Mode int64 `json:"mode"`
	// Frequencies 频段选择（真机观测值 0；语义待真机验证）。
	Frequencies int64 `json:"frequencies"`
}

// NetOptimizeService 是无线网优（ac_net_optimize）的子服务。
type NetOptimizeService struct {
	c *ikuai.Client
}

// Get 获取无线网优配置与体检结果（分页）。
// 对应 /Action/call: ac_net_optimize.show（TYPE=total,data, limit=
// "offset,rows"，真机探测实证返回 {"total","data","mode","frequencies"}，
// mode/frequencies 为持久化配置回显），
// 即 Web 控制台「AC管理 → 无线网优」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *NetOptimizeService) Get(ctx context.Context, opts *ikuai.ListOptions) (*NetOptimizeInfo, error) {
	var out NetOptimizeInfo
	if err := s.c.CallData(ctx, "ac_net_optimize", "show",
		pageShowParam(opts, nil), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Set 保存无线网优配置（写动作）。
// 对应 /Action/call: ac_net_optimize.set（action 名待真机验证；param 仅含
// mode/frequencies 两个 show 回显键）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段语义待真机验证。
func (s *NetOptimizeService) Set(ctx context.Context, cfg NetOptimizeConfig) error {
	_, err := s.c.Call(ctx, "ac_net_optimize", "set", cfg)
	return err
}
