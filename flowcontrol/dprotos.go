package flowcontrol

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// CustomProto 是一条自定义协议定义行（dprotos.show total,data 的 data
// 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，行字段无法从真机数据
// 核对；dprotos.show TYPE=all 实测仅携带 ipgroup 选择器区块。字段名按
// 「自定义协议」功能语义（协议名 + 端口特征）保守推断，除 id 外全部
// 标注待真机验证。
type CustomProto struct {
	// ID 协议行 ID。
	ID int64 `json:"id,omitempty"`
	// Name 协议名称。字段语义待真机验证。
	Name string `json:"name"`
	// Protocol 传输层协议（tcp/udp）。字段语义待真机验证。
	Protocol string `json:"protocol"`
	// Ports 端口特征串（如 "1000-2000,3000"）。字段语义待真机验证。
	Ports string `json:"ports"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
}

// CustomProtoSelectors 是自定义协议表单的可选值区块（dprotos.show TYPE=all
// 的区块，真机 3.7.24 实测仅 ipgroup）。
type CustomProtoSelectors struct {
	// IPGroup 可选 IPv4 分组名列表。
	IPGroup []string `json:"ipgroup"`
}

// CustomProtosService 是自定义协议（dprotos）的子服务。
type CustomProtosService struct {
	c *ikuai.Client
}

// List 获取自定义协议列表（分页）。
// 对应 /Action/call: dprotos.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「流控分流/行为管控」引用的自定义协议库页面。
func (s *CustomProtosService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[CustomProto], error) {
	resp, err := s.c.Call(ctx, "dprotos", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[CustomProto](resp.Data)
}

// Selectors 获取自定义协议表单的可选值（IP 分组）。
// 对应 /Action/call: dprotos.show（TYPE=all）。
func (s *CustomProtosService) Selectors(ctx context.Context) (*CustomProtoSelectors, error) {
	var out CustomProtoSelectors
	if err := s.c.CallData(ctx, "dprotos", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增自定义协议，返回固件分配的行 ID。
// 对应 /Action/call: dprotos.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *CustomProtosService) Add(ctx context.Context, proto CustomProto) (int64, error) {
	resp, err := s.c.Call(ctx, "dprotos", "add", proto)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改自定义协议（proto.ID 需携带目标行 ID）。
// 对应 /Action/call: dprotos.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *CustomProtosService) Edit(ctx context.Context, proto CustomProto) error {
	_, err := s.c.Call(ctx, "dprotos", "edit", proto)
	return err
}

// Delete 删除自定义协议，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: dprotos.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *CustomProtosService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dprotos", "del", ikuai.IDParam(ids))
	return err
}

// AdvancedProto 是一条高级自定义协议定义行（dprotos_l7.show total,data
// 的 data 元素）。
//
// ⚠️ 行结构说明：探测真机（3.7.24）上该列表为空，且 dprotos_l7.show
// TYPE=all 实测不含任何选择器区块（仅 total/data），行字段完全无法从
// 真机数据核对；字段名按「高级自定义协议」功能语义保守推断，除 id 外
// 全部标注待真机验证。
type AdvancedProto struct {
	// ID 协议行 ID。
	ID int64 `json:"id,omitempty"`
	// Name 协议名称。字段语义待真机验证。
	Name string `json:"name"`
	// Comment 备注。字段语义待真机验证。
	Comment string `json:"comment"`
}

// AdvancedProtosService 是高级自定义协议（dprotos_l7）的子服务。
type AdvancedProtosService struct {
	c *ikuai.Client
}

// List 获取高级自定义协议列表（分页）。
// 对应 /Action/call: dprotos_l7.show（TYPE=total,data, limit="offset,rows"），
// 即 Web 控制台「高级自定义协议」页面。
func (s *AdvancedProtosService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[AdvancedProto], error) {
	resp, err := s.c.Call(ctx, "dprotos_l7", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[AdvancedProto](resp.Data)
}

// Add 新增高级自定义协议，返回固件分配的行 ID。
// 对应 /Action/call: dprotos_l7.add。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *AdvancedProtosService) Add(ctx context.Context, proto AdvancedProto) (int64, error) {
	resp, err := s.c.Call(ctx, "dprotos_l7", "add", proto)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改高级自定义协议（proto.ID 需携带目标行 ID）。
// 对应 /Action/call: dprotos_l7.edit。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用；字段待真机验证。
func (s *AdvancedProtosService) Edit(ctx context.Context, proto AdvancedProto) error {
	_, err := s.c.Call(ctx, "dprotos_l7", "edit", proto)
	return err
}

// Delete 删除高级自定义协议，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: dprotos_l7.del。
// ⚠️ 写动作，仅离线测试覆盖，禁止对真机调用。
func (s *AdvancedProtosService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dprotos_l7", "del", ikuai.IDParam(ids))
	return err
}
