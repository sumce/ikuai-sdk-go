package behavior

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// MACComment 是一条 MAC 备注行（mac_comment.show total,data 的 data 元素；
// 行字段为真机 3.7.24 实测：mac/comment/id）。
//
// MAC 备注会覆盖终端 IP 信息里的备注展示（监控/审计日志的 comment 字段
// 即联动本表），即 Web 控制台「行为管控 → 终端信息管理」页面。
type MACComment struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// MAC 终端 MAC 地址（如 "e0:d5:5e:cd:5b:20"）。
	MAC string `json:"mac"`
	// Comment 备注/终端名称。
	Comment string `json:"comment"`
}

// MACCommentService 是 MAC 备注（mac_comment）的子服务。
type MACCommentService struct {
	c *ikuai.Client
}

// List 获取 MAC 备注列表（分页）。
// 对应 /Action/call: mac_comment.show（TYPE=total,data, limit="offset,rows"）。
func (s *MACCommentService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[MACComment], error) {
	resp, err := s.c.Call(ctx, "mac_comment", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[MACComment](resp.Data)
}

// Add 新增 MAC 备注，返回固件分配的行 ID。
// 对应 /Action/call: mac_comment.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MACCommentService) Add(ctx context.Context, row MACComment) (int64, error) {
	resp, err := s.c.Call(ctx, "mac_comment", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 MAC 备注（row.ID 需携带目标行 ID）。
// 对应 /Action/call: mac_comment.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MACCommentService) Edit(ctx context.Context, row MACComment) error {
	_, err := s.c.Call(ctx, "mac_comment", "edit", row)
	return err
}

// Delete 删除 MAC 备注，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: mac_comment.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MACCommentService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "mac_comment", "del", ikuai.IDParam(ids))
	return err
}
