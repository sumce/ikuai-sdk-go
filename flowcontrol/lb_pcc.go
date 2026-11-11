package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// MultiWanRule 是一条多线负载规则行（lb_pcc.show total,data 的 data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；字段名按表单选择器（interface/isplist，TYPE=all 实测）与同族
// 分流规则行（stream_ipport 实测）保守推断，除 id 外全部标注待真机验证。
type MultiWanRule struct {
	// ID 规则行 ID。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。字段语义待真机验证。
	Enabled string `json:"enabled"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
	// Interface 出接口名（多接口逗号分隔）。字段语义待真机验证。
	Interface string `json:"interface"`
	// ISP 运营商（值域参考 TYPE=all 的 isplist：all/chinatelecom/
	// chinaunicom/chinamobile/chinacernet）。字段语义待真机验证。
	ISP string `json:"isp"`
	// Weight 负载比重（控制台界面按 20M:10M:5M 等比例展示）。
	// 字段语义待真机验证。
	Weight int64 `json:"weight"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r MultiWanRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// MultiWanSelectors 是多线负载表单的可选值区块（lb_pcc.show TYPE=all 的
// interface/isplist 区块，真机 3.7.24 实测）。
type MultiWanSelectors struct {
	// Interface 可选接口列表，每项 [接口名, 备注]，如 ["adsl01","移动私网"]。
	Interface [][]string `json:"interface"`
	// ISPList 可选运营商列表（实测 ["all","chinatelecom","chinaunicom",
	// "chinamobile","chinacernet"]）。
	ISPList []string `json:"isplist"`
}

// MultiWanService 是多线负载（lb_pcc）的子服务。
type MultiWanService struct {
	c *ikuai.Client
}

// List 获取多线负载规则列表（分页）。
// 对应 /Action/call: lb_pcc.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「网络设置 → 多线负载」页面。
func (s *MultiWanService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[MultiWanRule], error) {
	resp, err := s.c.Call(ctx, "lb_pcc", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[MultiWanRule](resp.Data)
}

// Selectors 获取多线负载表单的可选值（接口列表与运营商列表）。
// 对应 /Action/call: lb_pcc.show（TYPE=all）。
func (s *MultiWanService) Selectors(ctx context.Context) (*MultiWanSelectors, error) {
	var out MultiWanSelectors
	if err := s.c.CallData(ctx, "lb_pcc", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增多线负载规则，返回固件分配的行 ID。
// 对应 /Action/call: lb_pcc.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *MultiWanService) Add(ctx context.Context, rule MultiWanRule) (int64, error) {
	resp, err := s.c.Call(ctx, "lb_pcc", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改多线负载规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: lb_pcc.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *MultiWanService) Edit(ctx context.Context, rule MultiWanRule) error {
	_, err := s.c.Call(ctx, "lb_pcc", "edit", rule)
	return err
}

// Delete 删除多线负载规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: lb_pcc.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *MultiWanService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "lb_pcc", "del", ikuai.IDParam(ids))
	return err
}
