package advanced

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// WakeupTask 是一条网络唤醒定时任务（wakeup.show data 数组元素）。
//
// 行字段空表推导，待真机验证：真机 3.7.24 未配置定时唤醒任务（data 恒
// 为空数组），字段名按控制台 i18n tools.wakeup（MacAddr=MAC地址、
// planned_task=计划任务、chooseDate=请选择日期）+ 固件列表行惯例
// （week/time，对照 simple_qos 等实测行）推导。
type WakeupTask struct {
	// ID 任务行 ID（edit/del 必填；add 时无需填写）。
	ID int64 `json:"id,omitempty"`
	// MAC 目标终端 MAC 地址（控制台 i18n MacAddr「MAC地址」）。
	MAC string `json:"mac"`
	// Week 重复唤醒的星期（固件列表行惯例 "1234567"=周一至周日，
	// 空表推导，待真机验证）。
	Week string `json:"week"`
	// Time 唤醒时刻（空表推导，格式待真机验证，推测 "HH:MM"）。
	Time string `json:"time"`
	// Comment 备注（空表推导，待真机验证）。
	Comment string `json:"comment"`
	// Enabled 是否启用（"yes"/"no"，空表推导，待真机验证）。
	Enabled string `json:"enabled"`
}

// IsEnabled 任务启用状态为 "yes"。
func (t WakeupTask) IsEnabled() bool { return t.Enabled == EnabledYes }

// WakeupService 是网络唤醒（wakeup，WOL）的子服务。
type WakeupService struct {
	c *ikuai.Client
}

// List 获取网络唤醒定时任务列表（分页）。
// 对应 /Action/call: wakeup.show（TYPE=total,data，真机探测实证返回
// {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → 网络唤醒」页面的「定期唤醒列表」。
func (s *WakeupService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[WakeupTask], error) {
	resp, err := s.c.Call(ctx, "wakeup", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[WakeupTask](resp.Data)
}

// Add 新增定时唤醒任务，返回固件分配的行 ID。
// 对应 /Action/call: wakeup.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *WakeupService) Add(ctx context.Context, task WakeupTask) (int64, error) {
	resp, err := s.c.Call(ctx, "wakeup", "add", task)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 修改定时唤醒任务。
// 对应 /Action/call: wakeup.edit（task.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *WakeupService) Edit(ctx context.Context, task WakeupTask) error {
	_, err := s.c.Call(ctx, "wakeup", "edit", task)
	return err
}

// Delete 删除定时唤醒任务，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: wakeup.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *WakeupService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "wakeup", "del", ikuai.IDParam(ids))
	return err
}

// Wake 立即向目标终端发送 WOL 唤醒包（控制台 i18n「立即执行唤醒」）。
// 对应 /Action/call: wakeup.wake（param={"mac": mac}）。
//
// ⚠️⚠️ 写动作，且动作名与参数形态系按控制台语义推导（真机 3.7.24 探测
// 无法只读验证）：绝对禁止对本 SDK 使用方未明确指示的真机随意调用，
// 仅离线测试覆盖；动作名（wake）与参数键（mac）待真机验证，若固件报
// unknown action 请改走 Service.Raw 逃生口实测。
func (s *WakeupService) Wake(ctx context.Context, mac string) error {
	_, err := s.c.Call(ctx, "wakeup", "wake", map[string]any{"mac": mac})
	return err
}
