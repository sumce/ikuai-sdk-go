package security

import (
	"context"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// ACL 规则的固件枚举值（i18n firewall.acl 区块 action_data/dir_data/
// dir_matching/interface_map，真机 3.7.24）。
const (
	// ACLActionAccept 允许（i18n acl.action_data.accept）。
	ACLActionAccept = "accept"
	// ACLActionDrop 阻断（i18n acl.action_data.drop）。
	ACLActionDrop = "drop"

	// ACLDirForward 转发（i18n acl.dir_data.forward）。
	ACLDirForward = "forward"
	// ACLDirInput 进（i18n acl.dir_data.input）。
	ACLDirInput = "input"

	// ACLDirMatchingClose 连接方向匹配：关闭（i18n acl.dir_matching.close）。
	ACLDirMatchingClose = "close"
	// ACLDirMatchingOriginal 连接方向匹配：原始方向（i18n acl.dir_matching.original_direction）。
	ACLDirMatchingOriginal = "original_direction"
	// ACLDirMatchingResponse 连接方向匹配：应答方向（i18n acl.dir_matching.response_direction）。
	ACLDirMatchingResponse = "response_direction"

	// ACLProtocolAny 任意协议（i18n acl.interface_map.any）。
	ACLProtocolAny = "any"
	// ACLProtocolTCP TCP（i18n acl.interface_map.tcp）。
	ACLProtocolTCP = "tcp"
	// ACLProtocolUDP UDP（i18n acl.interface_map.udp）。
	ACLProtocolUDP = "udp"
	// ACLProtocolTCPUDP TCP+UDP（i18n acl.interface_map.tcp_udp）。
	ACLProtocolTCPUDP = "tcp_udp"
	// ACLProtocolICMP ICMP（i18n acl.interface_map.icmp）。
	ACLProtocolICMP = "icmp"
	// ACLProtocolGRE GRE（i18n acl.interface_map.gre）。
	ACLProtocolGRE = "gre"
)

// ACLRule 是一条访问控制规则行（acl.show total,data 的 data 元素）。
//
// 空表推导，待真机验证：真机 3.7.24 的 ACL 规则列表为空（total=0），
// 行字段无法实证。字段名按固件 i18n（firewall.acl 区块：src_addr/dst_addr/
// src_port/dst_port/iinterface/ointerface/protocol/action/dir/week/time）
// 与表单语义推导。固件另有「协议栈」「IPv6 后缀匹配」（i18n
// protocol_stack/suffix_matching/matching_content）等表单项，其行字段名
// 待真机验证，可通过 Service.Raw 补充提交。
type ACLRule struct {
	// ID 规则行 ID（add 时无需填写；edit/del 必填）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"；推导值，待真机验证）。
	Enabled string `json:"enabled"`
	// Comment 备注（推导字段，待真机验证）。
	Comment string `json:"comment"`
	// SrcAddr 源地址（i18n acl.src_addr；可为单 IP/网段/IP 分组/MAC 分组名，
	// 分组可选项见 acl.show TYPE=all 的 ipgroup/macgroup 区块；空串=任意）。
	SrcAddr string `json:"src_addr"`
	// DstAddr 目的地址（i18n acl.dst_addr；取值语义同 SrcAddr）。
	DstAddr string `json:"dst_addr"`
	// SrcPort 源端口串（i18n acl.src_port；如 "80" 或 "1000-2000,3000"，空=任意）。
	SrcPort string `json:"src_port"`
	// DstPort 目的端口串（i18n acl.dst_port；格式同 SrcPort）。
	DstPort string `json:"dst_port"`
	// IInterface 进接口（i18n acl.iinterface「进接口」；可选项来自
	// acl.show TYPE=all 的 interface 区块，"any"=任意）。
	IInterface string `json:"iinterface"`
	// OInterface 出接口（i18n acl.ointerface「出接口」；取值同 IInterface）。
	OInterface string `json:"ointerface"`
	// Protocol 协议（i18n acl.protocol；枚举 ACLProtocol*，空/any=任意）。
	Protocol string `json:"protocol"`
	// Action 动作（i18n acl.action；枚举 ACLActionAccept / ACLActionDrop）。
	Action string `json:"action"`
	// Dir 方向（i18n acl.dir_data；枚举 ACLDirForward / ACLDirInput）。
	Dir string `json:"dir"`
	// DirMatching 连接方向匹配（i18n acl.dir_matching；枚举
	// ACLDirMatching*；json 字段名待真机验证）。
	DirMatching string `json:"dir_matching"`
	// Week 周期（i18n acl.week；形如 "1234567" 的星期串，推导，待真机验证）。
	Week string `json:"week"`
	// Time 时段（i18n acl.time；格式 "00:00-09:00"，推导，待真机验证）。
	Time string `json:"time"`
}

// IsEnabled 固件 enabled 字段为 "yes"/"no"（推导，待真机验证）。
func (r ACLRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// ACLSelectors 是 ACL 规则表单的可选值区块（acl.show TYPE=all 的区块，
// 真机 3.7.24 实测）。
type ACLSelectors struct {
	// Interface 可选进/出接口列表，每项 [接口名, 备注]（首项 ["any"]=任意）。
	Interface [][]string `json:"interface"`
	// IPGroup 可选 IPv4 分组名列表（真机实测有值）。
	IPGroup []string `json:"ipgroup"`
	// MACGroup 可选 MAC 分组名列表。
	MACGroup []string `json:"macgroup"`
	// DTGroup 可选时间分组名列表。
	DTGroup []string `json:"dtgroup"`
	// IPLibStatus IP 库状态（实测 0；枚举语义待真机验证）。
	IPLibStatus int64 `json:"iplib_status"`
}

// ACLService 是访问控制（acl）的子服务。
type ACLService struct {
	c *ikuai.Client
}

// List 获取 ACL 规则列表（分页）。
// 对应 /Action/call: acl.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「安全设置 → ACL规则」页面。
func (s *ACLService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ACLRule], error) {
	resp, err := s.c.Call(ctx, "acl", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ACLRule](resp.Data)
}

// Selectors 获取 ACL 规则表单的可选值（接口/IP 分组/MAC 分组/时间分组）。
// 对应 /Action/call: acl.show（TYPE=all）。
func (s *ACLService) Selectors(ctx context.Context) (*ACLSelectors, error) {
	var out ACLSelectors
	if err := s.c.CallData(ctx, "acl", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增 ACL 规则，返回固件分配的行 ID。
// 对应 /Action/call: acl.add。rule.Action 必须为 ACLActionAccept /
// ACLActionDrop（i18n acl.action_data 仅此两值），否则返回
// ikuai.ErrInvalidParam。
// ⚠️ 写动作，字段集待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) Add(ctx context.Context, rule ACLRule) (int64, error) {
	if rule.Action != ACLActionAccept && rule.Action != ACLActionDrop {
		return 0, fmt.Errorf("security: ACL 动作必须为 %q 或 %q: %w",
			ACLActionAccept, ACLActionDrop, ikuai.ErrInvalidParam)
	}
	resp, err := s.c.Call(ctx, "acl", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 ACL 规则（rule.ID 需携带目标行 ID）。
// 对应 /Action/call: acl.edit。
// ⚠️ 写动作，字段集待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) Edit(ctx context.Context, rule ACLRule) error {
	_, err := s.c.Call(ctx, "acl", "edit", rule)
	return err
}

// Delete 删除 ACL 规则，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: acl.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *ACLService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "acl", "del", ikuai.IDParam(ids))
	return err
}
