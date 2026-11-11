package monitoring

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// camera.show 的 TYPE 选择器常量（抓包实证）。
const (
	// CameraTypeDataTotal 控制台「摄像头监控」列表页 TYPE。
	CameraTypeDataTotal = "total,data"
)

// CameraDevice 是摄像头探测的一行（camera.show data 数组元素）。
//
// 抓包与真机（3.7.24，未接入摄像头）的 data 均为空数组，行字段无法
// 实证，故行对象以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type CameraDevice struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (d *CameraDevice) UnmarshalJSON(b []byte) error {
	d.Raw = append(d.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (d CameraDevice) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(d.Raw, &m); err != nil {
		return nil, fmt.Errorf("monitoring: 解码 camera 行失败: %w", err)
	}
	return m, nil
}

// CameraService 是摄像头监控（camera）的子服务。
type CameraService struct {
	c *ikuai.Client
}

// List 获取摄像头探测列表（分页）。
// 对应 /Action/call: camera.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「状态监控 → 摄像头监控」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *CameraService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[CameraDevice], error) {
	param := pageShowParam(opts, map[string]any{"TYPE": CameraTypeDataTotal})
	resp, err := s.c.Call(ctx, "camera", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[CameraDevice](resp.Data)
}
