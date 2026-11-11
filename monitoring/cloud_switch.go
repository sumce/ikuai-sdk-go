package monitoring

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// cloud_switch.show 的 TYPE 选择器与排序常量（抓包实证）。
const (
	// CloudSwitchTypeData 控制台「交换机监控」列表页 TYPE（抓包原文
	// 含尾逗号 "total,data,"，为固件怪癖，SDK 发送等价的干净形态）。
	CloudSwitchTypeData = "total,data"
	// CloudSwitchOrderByID 控制台默认按 id 排序。
	CloudSwitchOrderByID = "id"
	// CloudSwitchOrderDesc 控制台默认降序。
	CloudSwitchOrderDesc = "desc"
)

// CloudSwitchDevice 是一台云管理交换机的一行（cloud_switch.show data
// 数组元素）。
//
// 抓包与真机（3.7.24，未接入云交换机）的 data 均为空数组，行字段无法
// 实证，故行对象以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type CloudSwitchDevice struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (d *CloudSwitchDevice) UnmarshalJSON(b []byte) error {
	d.Raw = append(d.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (d CloudSwitchDevice) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(d.Raw, &m); err != nil {
		return nil, fmt.Errorf("monitoring: 解码 cloud_switch 行失败: %w", err)
	}
	return m, nil
}

// CloudSwitchService 是交换机监控（cloud_switch）的子服务。
type CloudSwitchService struct {
	c *ikuai.Client
}

// List 获取云管理交换机列表（分页，默认按 id 降序）。
// 对应 /Action/call: cloud_switch.show（TYPE=total,data, ORDER_BY=id,
// ORDER=desc, limit="offset,rows"），
// 即 Web 控制台「状态监控 → 交换机监控」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions() 并按控制台默认排序。
func (s *CloudSwitchService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[CloudSwitchDevice], error) {
	param := pageShowParam(opts, map[string]any{
		"TYPE":     CloudSwitchTypeData,
		"ORDER_BY": CloudSwitchOrderByID,
		"ORDER":    CloudSwitchOrderDesc,
	})
	resp, err := s.c.Call(ctx, "cloud_switch", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[CloudSwitchDevice](resp.Data)
}
