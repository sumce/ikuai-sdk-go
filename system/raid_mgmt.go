package system

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// raid_mgmt.show 的 TYPE 选择器常量（真机 3.7.24 验证）。
const (
	// RaidTypeData 控制台列表页使用的 TYPE（burp 抓包实证，data 形态）。
	RaidTypeData = "data"
	// RaidTypeDataTotal 标准分页形态（真机验证固件同样接受）。
	RaidTypeDataTotal = "total,data"
)

// RAIDDevice 是一个 RAID 阵列的信息行（raid_mgmt.show data 数组元素）。
//
// 真机（3.7.24，未配置 RAID）的 data 为空数组，行字段无法实证，故行
// 对象以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type RAIDDevice struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (d *RAIDDevice) UnmarshalJSON(b []byte) error {
	d.Raw = append(d.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (d RAIDDevice) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(d.Raw, &m); err != nil {
		return nil, fmt.Errorf("system: 解码 raid_mgmt 行失败: %w", err)
	}
	return m, nil
}

// RaidService 是 RAID 管理（raid_mgmt）的子服务。
type RaidService struct {
	c *ikuai.Client
}

// List 获取 RAID 阵列列表。
// 对应 /Action/call: raid_mgmt.show（TYPE=data, limit="offset,rows"，
// burp 抓包实证；TYPE=total,data 真机同样接受），
// 即 Web 控制台「系统设置 → RAID」页面。
// opts 为 nil 时仅发送 {"TYPE":"data"}，分页/排序经 opts 指定
// （固件接受 limit，但 TYPE=total 会被归一为 data）。
func (s *RaidService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[RAIDDevice], error) {
	param := typedShowParam(opts, RaidTypeData, nil)
	resp, err := s.c.Call(ctx, "raid_mgmt", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[RAIDDevice](resp.Data)
}
