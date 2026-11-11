package network

import (
	"context"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// lan.show 的 TYPE 选择器常量（抓包实证）。
const (
	// LANTypeSupportWispDataTotal 控制台「内外网设置」页使用的 TYPE
	// （support_wisp,total,data，抓包实证），support_wisp 为无线中继支持标志。
	LANTypeSupportWispDataTotal = "support_wisp,total,data"
	// LANTypeDataTotal 仅取接口列表（真机实证，等价于 total,data）。
	LANTypeDataTotal = "total,data"
	// LANTypeAll 全量（真机实证）。
	LANTypeAll = "all"
)

// LANInterface 是一个内网接口的配置行（lan.show 行；抓包 + 真机实证）。
type LANInterface struct {
	// ID 接口行 ID（lan.edit 时必填）。
	ID int64 `json:"id"`
	// Name 接口名（如 lan1）。
	Name string `json:"name"`
	// Comment 接口备注。
	Comment string `json:"comment"`
	// IPMask "IP/掩码" 复合串（如 "192.168.9.1/255.255.255.0"，固件原样返回）。
	IPMask string `json:"ip_mask"`
	// BandETH 绑定的物理网卡名（如 eth8）。
	BandETH string `json:"bandeth"`
	// BandIF 绑定网卡的 MAC（如 "44:a9:2c:21:c1:56"）。
	BandIF string `json:"bandif"`
	// BandFlag 带宽设置标志（抓包中为字符串 "0"；字段语义待真机验证）。
	BandFlag string `json:"bandflag"`
	// Bandmode 带宽模式（0=不限制）。
	Bandmode int64 `json:"bandmode"`
	// MAC 接口 MAC（克隆后与物理 MAC 不同；未克隆为空串）。
	MAC string `json:"mac"`
	// Speed 协商速率（Mbps，0=协商中/未知）。
	Speed int64 `json:"speed"`
	// Duplex 双工模式（1=全双工）。
	Duplex int64 `json:"duplex"`
	// Linkmode 链路模式（真机观测 4；枚举语义待真机验证）。
	Linkmode int64 `json:"linkmode"`
	// LANVisit LAN 互访控制（1=允许其他 LAN 访问此 LAN）。
	LANVisit int64 `json:"lan_visit"`
	// Policy 策略标志（真机观测 0；字段语义待真机验证）。
	Policy int64 `json:"policy"`
}

// LANPage 是 lan.show 的完整返回形态（含 support_wisp 顶层标志，抓包实证）。
type LANPage struct {
	// SupportWisp 无线中继（WISP）支持标志（真机观测 0；语义待真机验证）。
	SupportWisp int64 `json:"support_wisp"`
	// Total 接口总数。
	Total int64 `json:"total"`
	// Rows 内网接口列表。
	Rows []LANInterface `json:"data"`
}

// LANService 是内外网设置（lan）的子服务。
type LANService struct {
	c *ikuai.Client
}

// List 获取内网接口列表（含 support_wisp 顶层标志）。
// 对应 /Action/call: lan.show（TYPE=support_wisp,total,data，抓包实证），
// 即 Web 控制台「网络设置 → 内外网设置」的内网侧列表。
// opts 为 nil 或未指定 Type 时使用 support_wisp,total,data。
// 注意：外网（WAN）线路配置不在 lan.show 的 total,data 返回中（真机验证），
// WAN 行编辑参数形态待真机验证，暂不提供封装。
func (s *LANService) List(ctx context.Context, opts *ikuai.ListOptions) (*LANPage, error) {
	param := forcedShowParam(opts, LANTypeSupportWispDataTotal, nil)
	var out LANPage
	if err := s.c.CallData(ctx, "lan", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Interface 获取单个内网接口（编辑表单回显，抓包实证 {"id":"N",
// TYPE=support_wisp,total,data}）。接口不存在时返回 (nil, nil)。
func (s *LANService) Interface(ctx context.Context, id int64) (*LANInterface, error) {
	if id <= 0 {
		return nil, fmt.Errorf("network: lan.Interface: id 必须为正整数: %w", ikuai.ErrInvalidParam)
	}
	param := map[string]any{"id": id, "TYPE": LANTypeSupportWispDataTotal}
	var out LANPage
	if err := s.c.CallData(ctx, "lan", "show", param, &out); err != nil {
		return nil, err
	}
	if len(out.Rows) == 0 {
		return nil, nil
	}
	row := out.Rows[0]
	return &row, nil
}

// Set 保存内网接口配置（写动作）。
// 对应 /Action/call: lan.edit（iface.ID 必填，随 JSON 自动携带）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *LANService) Set(ctx context.Context, iface LANInterface) error {
	_, err := s.c.Call(ctx, "lan", "edit", iface)
	return err
}
