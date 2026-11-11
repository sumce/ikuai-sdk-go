package advanced

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// plugins.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// PluginsTypeData 已安装插件列表。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）；TYPE=all 与 data 等价。
	PluginsTypeData = "data"
)

// Plugin 是一个已安装插件（plugins.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 未安装任何插件（data 恒为空数组），
// 行字段无法实证（控制台 i18n system-setting.plugin-management 页面语义：
// 插件名/版本/状态/描述等），故行对象以 Raw 原样保留，字段语义待真机
// 验证后再细化强类型。
type Plugin struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (p *Plugin) UnmarshalJSON(b []byte) error {
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

// MarshalJSON 把行对象原样序列化（写动作提交参数与原始行逐字段一致）。
func (p Plugin) MarshalJSON() ([]byte, error) {
	if len(p.Raw) == 0 {
		return []byte("{}"), nil
	}
	return p.Raw, nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (p Plugin) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(p.Raw, &m); err != nil {
		return nil, fmt.Errorf("advanced: 解码 plugins 行失败: %w", err)
	}
	return m, nil
}

// PluginsService 是插件（plugins）的子服务。
type PluginsService struct {
	c *ikuai.Client
}

// List 获取已安装插件列表（分页）。
// 对应 /Action/call: plugins.show（TYPE=data，真机探测实证；注意本 func
// 不支持 TYPE=total,data），
// 即 Web 控制台「高级应用 → 插件管理」页面。
// 插件的安装/卸载/启停为未建模的写动作（动作名与字段待真机验证且影响
// 系统状态），需要时请走 Service.Raw 逃生口。
func (s *PluginsService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[Plugin], error) {
	param := typedShowParam(opts, PluginsTypeData, nil)
	resp, err := s.c.Call(ctx, "plugins", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[Plugin](resp.Data)
}
