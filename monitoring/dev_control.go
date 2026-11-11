package monitoring

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// dev_control.show 的 TYPE 选择器常量（抓包实证）。
const (
	// DevControlTypeData 控制台「周边设备监控」列表页 TYPE。
	DevControlTypeData = "total,data"
)

// DevControlDevice 是智能设备管控的一行（dev_control.show data 数组
// 元素）。
//
// 抓包与真机（3.7.24，未接入智能设备）的 data 均为空数组，行字段无法
// 实证，故行对象以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type DevControlDevice struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (d *DevControlDevice) UnmarshalJSON(b []byte) error {
	d.Raw = append(d.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (d DevControlDevice) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(d.Raw, &m); err != nil {
		return nil, fmt.Errorf("monitoring: 解码 dev_control 行失败: %w", err)
	}
	return m, nil
}

// DevControlService 是周边设备监控（dev_control）的子服务。
type DevControlService struct {
	c *ikuai.Client
}

// List 获取智能设备管控列表（分页）。
// 对应 /Action/call: dev_control.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「状态监控 → 周边设备监控」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *DevControlService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DevControlDevice], error) {
	param := pageShowParam(opts, map[string]any{"TYPE": DevControlTypeData})
	resp, err := s.c.Call(ctx, "dev_control", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DevControlDevice](resp.Data)
}

// CheckOnlineStatus 触发一次智能设备在线状态检测（重新探测周边设备的
// 在线情况，结果经 List 查询）。对应 /Action/call:
// dev_control.check_online_status（burp 抓包实证 param 为空对象，成功
// 响应无 Data）。该动作虽为 POST 调用，但语义为只读检测，不改变任何
// 设备或配置数据。
func (s *DevControlService) CheckOnlineStatus(ctx context.Context) error {
	_, err := s.c.Call(ctx, "dev_control", "check_online_status", map[string]any{})
	return err
}
