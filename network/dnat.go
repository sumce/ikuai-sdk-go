package network

import (
	"context"
	"strconv"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// dnat.show 的 TYPE 选择器与协议枚举常量（抓包实证）。
const (
	// DNATTypeDataTotal 列表页抓包使用的 TYPE（total,data）。
	DNATTypeDataTotal = "total,data"
	// DNATTypeInterfaceProtocol 下拉候选数据 TYPE（interface,protocol，
	// 抓包实证：interface 为 [名称, 备注] 嵌套数组，protocol 为协议列表）。
	DNATTypeInterfaceProtocol = "interface,protocol"
	// DNATTypeIPGroup 源地址分组候选 TYPE（ipgroup，抓包实证：分组名数组）。
	DNATTypeIPGroup = "ipgroup"
)

// dnat.protocol 的协议枚举值（抓包实证）。
const (
	DNATProtocolTCP    = "tcp"
	DNATProtocolUDP    = "udp"
	DNATProtocolTCPUDP = "tcp+udp"
)

// DNATRule 是一条端口映射规则（dnat.show 行 / dnat.add 参数，抓包实证）。
type DNATRule struct {
	// ID 行 ID（add 时无需填写；edit/del 必填）。omitempty 保证 add 参数
	// 与抓包一致（抓包实证 add 不携带 id）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// Interface 外网接口（如 adsl07，"all" 表示全部外网口；抓包实证）。
	Interface string `json:"interface"`
	// WANPort 外网端口串，逗号分隔的端口/端口段（如 "7422,30501-45000"）。
	WANPort string `json:"wan_port"`
	// LANAddr 内网目标地址（如 "172.22.1.2"）。
	LANAddr string `json:"lan_addr"`
	// LANAddrInt 内网地址的整型表示（固件 show 返回的计算字段，add/edit
	// 无需填写；omitempty 保证 add 参数与抓包一致）。
	LANAddrInt int64 `json:"lan_addr_int,omitempty"`
	// LANPort 内网端口串，与 WANPort 端口数一一对应（如 "22,30501-45000"）。
	LANPort string `json:"lan_port"`
	// Protocol 协议（tcp/udp/tcp+udp）。
	Protocol string `json:"protocol"`
	// SrcAddr 允许访问的源地址（空串表示不限）。
	SrcAddr string `json:"src_addr"`
	// Comment 备注。
	Comment string `json:"comment"`
}

// IsEnabled 报告规则是否启用（固件 enabled 字段为 "yes"/"no"）。
func (r DNATRule) IsEnabled() bool { return r.Enabled == EnabledYes }

// WANPorts 解析固件的外网端口串（如 "7422,30501-45000"）为端口段列表；
// 非法片段被跳过。
func (r DNATRule) WANPorts() []PortRange { return parsePortRanges(r.WANPort) }

// LANPorts 解析内网端口串为端口段列表（语义同 WANPorts）。
func (r DNATRule) LANPorts() []PortRange { return parsePortRanges(r.LANPort) }

// PortRange 是一个端口或端口段（Min==Max 表示单端口）。
type PortRange struct {
	// Min 起始端口。
	Min int64
	// Max 结束端口（单端口时与 Min 相等）。
	Max int64
}

// parsePortRanges 解析固件的端口串 "80,443,30501-45000"。
func parsePortRanges(s string) []PortRange {
	var out []PortRange
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if lo, hi, ok := strings.Cut(part, "-"); ok {
			min, err1 := strconv.ParseInt(strings.TrimSpace(lo), 10, 64)
			max, err2 := strconv.ParseInt(strings.TrimSpace(hi), 10, 64)
			if err1 == nil && err2 == nil && min <= max {
				out = append(out, PortRange{Min: min, Max: max})
			}
			continue
		}
		if v, err := strconv.ParseInt(part, 10, 64); err == nil {
			out = append(out, PortRange{Min: v, Max: v})
		}
	}
	return out
}

// DNATSelectors 是 dnat.show TYPE=interface,protocol 的下拉候选数据
// （抓包实证），供构造 DNATRule 时选择取值。
type DNATSelectors struct {
	// Interface 外网接口候选（[接口名] 或 [接口名, 备注]，无 "all" 项）。
	Interface [][]string `json:"interface"`
	// Protocol 协议候选（["tcp","udp","tcp+udp"]，抓包实证）。
	Protocol []string `json:"protocol"`
}

// IPGroups 返回 dnat.show TYPE=ipgroup 的源地址分组名列表（抓包实证）。
// 对应 /Action/call: dnat.show（TYPE=ipgroup）。
func (s *DNATService) IPGroups(ctx context.Context) ([]string, error) {
	var out struct {
		IPGroup []string `json:"ipgroup"`
	}
	if err := s.c.CallData(ctx, "dnat", "show",
		map[string]any{"TYPE": DNATTypeIPGroup}, &out); err != nil {
		return nil, err
	}
	return out.IPGroup, nil
}

// DNATService 是端口映射（dnat）的子服务。
type DNATService struct {
	c *ikuai.Client
}

// List 获取端口映射列表。
// 对应 /Action/call: dnat.show（TYPE=total,data, limit="offset,rows"，抓包实证），
// 即 Web 控制台「服务应用 → 端口映射」页面。
func (s *DNATService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DNATRule], error) {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	resp, err := s.c.Call(ctx, "dnat", "show", opts.BuildParam())
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DNATRule](resp.Data)
}

// Selectors 获取端口映射的外网接口与协议候选（抓包实证）。
// 对应 /Action/call: dnat.show（TYPE=interface,protocol）。
func (s *DNATService) Selectors(ctx context.Context) (*DNATSelectors, error) {
	var out DNATSelectors
	if err := s.c.CallData(ctx, "dnat", "show",
		map[string]any{"TYPE": DNATTypeInterfaceProtocol}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Add 新增端口映射，返回固件分配的行 ID。
// 对应 /Action/call: dnat.add（抓包实证：enabled/comment/interface/lan_addr/
// protocol/wan_port/lan_port/src_addr）。
func (s *DNATService) Add(ctx context.Context, rule DNATRule) (int64, error) {
	resp, err := s.c.Call(ctx, "dnat", "add", rule)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改端口映射。
// 对应 /Action/call: dnat.edit（rule.ID 随 JSON 自动携带）。
func (s *DNATService) Edit(ctx context.Context, rule DNATRule) error {
	_, err := s.c.Call(ctx, "dnat", "edit", rule)
	return err
}

// Delete 删除端口映射，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}，
// del 参数形态抓包实证 {"id":11}）。
// 对应 /Action/call: dnat.del。
func (s *DNATService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "dnat", "del", ikuai.IDParam(ids))
	return err
}
