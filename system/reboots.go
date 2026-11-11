package system

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// reboots.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
const (
	// RebootsTypeDataTotal 定时重启任务列表的标准分页形态（burp 抓包
	// 实证 TYPE=total,data；本 func 与 raid/backup 一致接受该形态）。
	RebootsTypeDataTotal = "total,data"
)

// RebootTask 是一个定时重启任务的行（reboots.show data 数组元素）。
//
// 真机（3.7.24）未配置定时重启，show 的 data 为空数组，行字段无法实证。
// 以下字段按固件定时任务惯例建模：写动作，字段待真机验证。
type RebootTask struct {
	// ID 行 ID（add 时忽略，edit/del 时使用）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，待真机验证）。
	Enabled string `json:"enabled,omitempty"`
	// Week 生效日（星期列表，格式待真机验证）。
	Week string `json:"week,omitempty"`
	// Time 重启时刻（推测 "HH:MM"，待真机验证）。
	Time string `json:"time,omitempty"`
	// Comment 任务备注。
	Comment string `json:"comment,omitempty"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (t RebootTask) IsEnabled() bool { return t.Enabled == EnabledYes }

// RebootsRow 是定时重启任务列表的行对象（行字段未实证时的 Raw 形态）。
//
// reboots.show 的 data 为空数组时行字段无法实证；Add/Edit 以 RebootTask
// 建模，List 查询到的行以本类型原样保留，字段语义待真机验证。
type RebootsRow struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (r *RebootsRow) UnmarshalJSON(b []byte) error {
	r.Raw = append(r.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (r RebootsRow) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(r.Raw, &m); err != nil {
		return nil, fmt.Errorf("system: 解码 reboots 行失败: %w", err)
	}
	return m, nil
}

// RebootsService 是重启（reboots）的子服务。
type RebootsService struct {
	c *ikuai.Client
}

// List 获取定时重启任务列表。
// 对应 /Action/call: reboots.show（TYPE=total,data, limit="offset,rows"，
// burp 抓包实证），即 Web 控制台「系统设置 → 重启」页面的定时重启表。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *RebootsService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[RebootsRow], error) {
	param := pageShowParam(opts, nil)
	resp, err := s.c.Call(ctx, "reboots", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[RebootsRow](resp.Data)
}

// Add 新增定时重启任务，返回固件分配的行 ID（写动作）。
// 对应 /Action/call: reboots.add（action 名按标准列表 func 惯例推断）。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
// 立即重启（非定时）动作未在抓包/探测中出现，未提供封装；
// 需要时经 Service.Raw 显式调用并自行承担风险。
func (s *RebootsService) Add(ctx context.Context, task RebootTask) (int64, error) {
	resp, err := s.c.Call(ctx, "reboots", "add", task)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Delete 删除定时重启任务，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: reboots.del（action 名按标准列表 func 惯例推断）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *RebootsService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "reboots", "del", ikuai.IDParam(ids))
	return err
}
