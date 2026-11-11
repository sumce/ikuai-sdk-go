package accounting

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// PeriodUnit 枚举候选值：套餐有效期周期类型（ppp_package.period_unit，
// 固件 i18n combo-management 提供 hour/day/month 三个选项，json 取值
// 形态未实证，待真机验证）。
const (
	// PeriodUnitHour 按小时计费。
	PeriodUnitHour = "hour"
	// PeriodUnitDay 按天计费。
	PeriodUnitDay = "day"
	// PeriodUnitMonth 按月计费。
	PeriodUnitMonth = "month"
)

// Package 是一条计费套餐（ppp_package.show 行；菜单「套餐管理」，
// combo-management）。真机（3.7.24，认证计费未启用）data 为空数组，
// 行字段由固件 i18n combo-management 表单项推导（packname 与 ppp_paylog
// 实证行、pppuser.packages 交叉印证），字段语义待真机验证。
type Package struct {
	// ID 行 ID（add 时无需填写；edit/del 必填）。omitempty 保证 add 参数
	// 不携带 id。
	ID int64 `json:"id,omitempty"`
	// PackName 套餐名称（与 ppp_paylog.packname 同名实证）。
	PackName string `json:"packname"`
	// PackTime 有效期时长（数值，配合 PeriodUnit）。
	PackTime int64 `json:"packtime"`
	// PeriodUnit 周期类型（hour/day/month，见 PeriodUnit* 常量）。
	PeriodUnit string `json:"period_unit"`
	// Price 套餐价格（元）。
	Price float64 `json:"price"`
	// UpSpeed 上行带宽（KB/s，0=不限）。
	UpSpeed int64 `json:"up_speed"`
	// DownSpeed 下行带宽（KB/s，0=不限）。
	DownSpeed int64 `json:"down_speed"`
}

// PackagesService 是套餐管理（ppp_package）的子服务。
type PackagesService struct {
	c *ikuai.Client
}

// List 获取套餐列表（分页）。
// 对应 /Action/call: ppp_package.show（TYPE=total,data, limit="offset,rows"，
// 真机实证），即 Web 控制台「认证计费 → 认证账号 → 套餐管理」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *PackagesService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[Package], error) {
	resp, err := s.c.Call(ctx, "ppp_package", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[Package](resp.Data)
}

// Add 新增套餐，返回固件分配的行 ID。
// 对应 /Action/call: ppp_package.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *PackagesService) Add(ctx context.Context, pkg Package) (int64, error) {
	resp, err := s.c.Call(ctx, "ppp_package", "add", pkg)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改套餐。
// 对应 /Action/call: ppp_package.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *PackagesService) Edit(ctx context.Context, pkg Package) error {
	_, err := s.c.Call(ctx, "ppp_package", "edit", pkg)
	return err
}

// Delete 删除套餐，支持单个或多个 ID。
// 对应 /Action/call: ppp_package.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *PackagesService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ppp_package", "del", ikuai.IDParam(ids))
	return err
}
