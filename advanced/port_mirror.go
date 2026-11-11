package advanced

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// port_mirror 的镜像方向常量（固件 dir 字段，取值对照控制台 i18n
// dropdownOptions：all=双向流量、in=进流量、out=出流量）。
const (
	PortMirrorDirAll = "all"
	PortMirrorDirIn  = "in"
	PortMirrorDirOut = "out"
)

// port_mirror.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// PortMirrorTypeAll 规则列表 + 可选端口区块。
	PortMirrorTypeAll = "all"
)

// PortMirrorRule 是一条端口镜像规则（port_mirror.show data 数组元素）。
//
// 行字段空表推导，待真机验证：真机 3.7.24 未配置镜像规则（data 恒为
// 空数组），字段名按控制台 i18n service.port_mirror.header_title
// （target_port=镜像端口、mirror_port=被镜像端口、dir=进出流量）+ 开源
// 实现核对推导。
type PortMirrorRule struct {
	// ID 规则行 ID（edit/del 必填；add 时无需填写）。
	ID int64 `json:"id,omitempty"`
	// TargetPort 镜像端口（目的端口，把指定端口镜像到此端口）。
	TargetPort string `json:"target_port"`
	// MirrorPort 被镜像端口（源端口，可为多端口串，形态待真机验证）。
	MirrorPort string `json:"mirror_port"`
	// Dir 镜像方向（all=双向流量 / in=进流量 / out=出流量）。
	Dir string `json:"dir"`
	// Comment 备注（空表推导，待真机验证）。
	Comment string `json:"comment"`
}

// PortMirrorInfo 是端口镜像的规则列表与可选端口区块
// （port_mirror.show TYPE=all，真机 3.7.24 实证顶层字段）。
type PortMirrorInfo struct {
	// Total 规则数（真机观测 0）。
	Total int64 `json:"total"`
	// Rules 规则行列表。
	Rules []PortMirrorRule `json:"data"`
	// Interface 可选端口列表，每项 [端口名] 或 [端口名, 备注]
	// （真机实证，如 ["lan1"]、["lan2","内网接入三层交换机"]）。
	Interface [][]string `json:"interface"`
}

// PortMirrorService 是端口镜像（port_mirror）的子服务。
type PortMirrorService struct {
	c *ikuai.Client
}

// List 获取端口镜像规则列表（分页）。
// 对应 /Action/call: port_mirror.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → 端口镜像」页面。
func (s *PortMirrorService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[PortMirrorRule], error) {
	resp, err := s.c.Call(ctx, "port_mirror", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[PortMirrorRule](resp.Data)
}

// Get 获取端口镜像规则列表与可选端口区块。
// 对应 /Action/call: port_mirror.show（TYPE=all，真机探测实证）。
func (s *PortMirrorService) Get(ctx context.Context) (*PortMirrorInfo, error) {
	var info PortMirrorInfo
	if err := s.c.CallData(ctx, "port_mirror", "show",
		map[string]any{"TYPE": PortMirrorTypeAll}, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Add 新增端口镜像规则，返回固件分配的行 ID。
// 对应 /Action/call: port_mirror.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *PortMirrorService) Add(ctx context.Context, rule PortMirrorRule) (int64, error) {
	resp, err := s.c.Call(ctx, "port_mirror", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 修改端口镜像规则。
// 对应 /Action/call: port_mirror.edit（rule.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *PortMirrorService) Edit(ctx context.Context, rule PortMirrorRule) error {
	_, err := s.c.Call(ctx, "port_mirror", "edit", rule)
	return err
}

// Delete 删除端口镜像规则，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: port_mirror.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *PortMirrorService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "port_mirror", "del", ikuai.IDParam(ids))
	return err
}
