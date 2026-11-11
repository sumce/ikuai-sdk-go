package ac

import (
	"context"
	"fmt"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// APGroup 是一个 AP 分组（ac_group.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 无 AP 设备，分组列表恒为空数组，
// 行字段无法实证。字段名按固件 i18n（ap-group 区块 name/ap_count/
// channel/max_pc）推导。
type APGroup struct {
	// ID 分组 ID（add 时无需填写；edit/del 必填）。
	ID int64 `json:"id,omitempty"`
	// Name 分组名称（i18n ap-group.name）。
	Name string `json:"name"`
	// APCount 已分配 AP 数（固件 show 返回的计算字段，add/edit 无需填写）。
	APCount int64 `json:"ap_count"`
	// Channel 分组信道配置（i18n ap-group.channel；取值格式待真机验证）。
	Channel string `json:"channel"`
	// MaxPC 分组最大带机量（i18n ap-group.max_pc；0 表示不限制）。
	MaxPC int64 `json:"max_pc"`
}

// GroupService 是 AP 分组（ac_group）的子服务。
type GroupService struct {
	c *ikuai.Client
}

// List 获取 AP 分组列表（分页）。
// 对应 /Action/call: ac_group.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「AC管理 → AP分组」页面。
func (s *GroupService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[APGroup], error) {
	resp, err := s.c.Call(ctx, "ac_group", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[APGroup](resp.Data)
}

// Add 新增 AP 分组，返回固件分配的行 ID（写动作）。
// 对应 /Action/call: ac_group.add。name 为必填（i18n ap-group.name），
// 为空（或全空白）返回 ikuai.ErrInvalidParam。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *GroupService) Add(ctx context.Context, group APGroup) (int64, error) {
	if strings.TrimSpace(group.Name) == "" {
		return 0, fmt.Errorf("ac: 分组名称不能为空: %w", ikuai.ErrInvalidParam)
	}
	resp, err := s.c.Call(ctx, "ac_group", "add", group)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 AP 分组（写动作）。
// 对应 /Action/call: ac_group.edit（group.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *GroupService) Edit(ctx context.Context, group APGroup) error {
	_, err := s.c.Call(ctx, "ac_group", "edit", group)
	return err
}

// Delete 删除 AP 分组（组内 AP 移入「未分组」，写动作），支持单个或
// 多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: ac_group.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *GroupService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ac_group", "del", ikuai.IDParam(ids))
	return err
}
