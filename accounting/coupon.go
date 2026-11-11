package accounting

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// CouponType 枚举候选值：上网码字符类型（coupon 批量生成表单
// coupon_type；固件 i18n coupon 提供纯数字/纯字母/数字+字母三个选项，
// json 取值形态未实证，待真机验证）。
const (
	// CouponTypePureNumber 纯数字。
	CouponTypePureNumber = "pure_number"
	// CouponTypePureLetters 纯字母。
	CouponTypePureLetters = "pure_letters"
	// CouponTypeNumberLetters 数字+字母。
	CouponTypeNumberLetters = "number_letters"
)

// Coupon 是一条上网码（coupon.show 行）。
//
// 真机（3.7.24，认证计费未启用）data 为空数组，行字段无法实证；控制台
// 列（固件 i18n coupon）为：上网码 coupon_code / 过期时间 expires /
// 限时 timeout / 使用记录 used。故行对象以 Raw 原样保留，字段语义待
// 真机验证后再细化强类型。
type Coupon struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (c *Coupon) UnmarshalJSON(b []byte) error {
	c.Raw = append(c.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (c Coupon) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(c.Raw, &m); err != nil {
		return nil, fmt.Errorf("accounting: 解码 coupon 行失败: %w", err)
	}
	return m, nil
}

// CouponBuildParam 是批量生成上网码的表单参数（固件 i18n coupon 生成
// 对话框推导；真机空表无法实证，字段语义待真机验证）。
type CouponBuildParam struct {
	// Number 上网码个数（i18n coupon_number）。
	Number int64 `json:"coupon_number"`
	// Length 单个上网码长度（i18n coupon_length）。
	Length int64 `json:"coupon_length"`
	// Type 上网码字符类型（i18n coupon_type，见 CouponType* 常量）。
	Type string `json:"coupon_type"`
	// ValidPeriod 有效期（i18n valid_period；单位待真机验证）。
	ValidPeriod int64 `json:"valid_period"`
	// Timeout 限时上网时长（i18n timeout；0=不限，单位待真机验证）。
	Timeout int64 `json:"timeout"`
}

// CouponService 是优惠券管理（coupon）的子服务。
type CouponService struct {
	c *ikuai.Client
}

// List 获取上网码列表（分页，支持关键字/过滤）。
// 对应 /Action/call: coupon.show（TYPE=total,data, limit="offset,rows"，
// 真机实证），即 Web 控制台「认证计费 → 认证账号 → 优惠券管理」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *CouponService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[Coupon], error) {
	resp, err := s.c.Call(ctx, "coupon", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[Coupon](resp.Data)
}

// Generate 批量生成上网码。
// 对应 /Action/call: coupon.add（批量生成走标准 add 通道，参照
// i18n batch_build/build_coupon；action 名待真机验证）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *CouponService) Generate(ctx context.Context, p CouponBuildParam) (int64, error) {
	resp, err := s.c.Call(ctx, "coupon", "add", p)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Delete 删除上网码，支持单个或多个 ID。
// 对应 /Action/call: coupon.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *CouponService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "coupon", "del", ikuai.IDParam(ids))
	return err
}
