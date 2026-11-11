package security

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// ARP 绑定类型/绑定状态的固件枚举值（i18n firewall.arp 区块
// bind_type_data/bind_state_data；真机 3.7.24 实测行值为字符串）。
const (
	// ARPBindTypeNormal 普通：IP 和 MAC 的对应关系状态（默认配置）。
	ARPBindTypeNormal = "0"
	// ARPBindTypeUnique 唯一：IP 和 MAC 对应关系绑定，匹配失败不能上网。
	ARPBindTypeUnique = "1"

	// ARPBindStateUnbound 未绑定。
	ARPBindStateUnbound = "0"
	// ARPBindStateBound 已绑定。
	ARPBindStateBound = "1"
)

// ARPEntry 是一条 ARP 表条目（arp.show total,data 的 data 元素，
// 真机 3.7.24 实测 21 行；字段为实测值）。
type ARPEntry struct {
	// ID 行 ID（真机实测值）。
	ID int64 `json:"id"`
	// IPAddr IP 地址（实测 "172.22.2.3"）。
	IPAddr string `json:"ip_addr"`
	// IPAddrInt IP 地址的整数表示（固件 show 返回的计算字段）。
	IPAddrInt int64 `json:"ip_addr_int"`
	// MAC MAC 地址（实测 "00:0c:29:57:23:25"）。
	MAC string `json:"mac"`
	// Interface 所属网卡（实测 "vlan1"/"lan5" 等）。
	Interface string `json:"interface"`
	// BindType 绑定类型（"0"=普通 / "1"=唯一，枚举 ARPBindType*）。
	BindType string `json:"bind_type"`
	// BindState 绑定状态（"0"=未绑定 / "1"=已绑定，枚举 ARPBindState*）。
	BindState string `json:"bind_state"`
	// Comment 备注。
	Comment string `json:"comment"`
	// OfflineTime 最后下线时间（i18n arp.offline_time；需开启「终端上下线
	// 记录」才出现，本机未观测到该字段，待真机验证）。
	OfflineTime *string `json:"offline_time,omitempty"`
}

// IsBound 该条目是否已绑定（bind_state == "1"）。
func (e ARPEntry) IsBound() bool { return e.BindState == ARPBindStateBound }

// IsUnique 绑定类型是否为「唯一」（bind_type == "1"，匹配失败不能上网）。
func (e ARPEntry) IsUnique() bool { return e.BindType == ARPBindTypeUnique }

// ARPSelectors 是 ARP 绑定页的选择器与全局开关区块（arp.show TYPE=all
// 的区块，真机 3.7.24 实测）。
type ARPSelectors struct {
	// Interface 可选所属网卡列表，每项 [接口名, 备注]（全部接口）。
	Interface [][]string `json:"interface"`
	// LANInterface 可选内网（LAN）接口列表（i18n arp.association_interface
	// 「关联接口」的取值来源）。
	LANInterface [][]string `json:"lan_interface"`
	// ARPFilter 「非绑定MAC不允许上网」全局开关（i18n arp.binding_mac；
	// 实测 0=关；1=开的语义待真机验证）。
	ARPFilter int64 `json:"arp_filter"`
	// ARPFilterIface 「关联接口」（i18n arp.association_interface；实测
	// "any"=全部线路）。
	ARPFilterIface string `json:"arp_filter_iface"`
	// DHCPDARP 「兼容已绑定为 DHCP 静态分配」开关（i18n arp.binding_arp；
	// 实测 0=关；1=开的语义待真机验证）。
	DHCPDARP int64 `json:"dhcpd_arp"`
}

// ARPService 是 ARP 表（arp）的子服务。arp 为运行时表，本子服务只读，
// 不提供任何写方法。
type ARPService struct {
	c *ikuai.Client
}

// List 获取 ARP 表（分页）。
// 对应 /Action/call: arp.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「安全设置 → ARP绑定」页面。
func (s *ARPService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ARPEntry], error) {
	resp, err := s.c.Call(ctx, "arp", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ARPEntry](resp.Data)
}

// Selectors 获取 ARP 绑定页的接口选择器与全局开关（TYPE=all 区块）。
// 对应 /Action/call: arp.show（TYPE=all）。
func (s *ARPService) Selectors(ctx context.Context) (*ARPSelectors, error) {
	var out ARPSelectors
	if err := s.c.CallData(ctx, "arp", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Search 按关键字检索 ARP 表（分页），等价于控制台列表页的搜索框：
// KEYWORDS=keywords，FINDS 默认限定 ip_addr,mac,comment,interface 四字段
// （真机 3.7.24 实测可用）。opts.Extra["FINDS"] 可显式覆盖检索字段。
// 对应 /Action/call: arp.show（TYPE=total,data, FINDS, KEYWORDS）。
func (s *ARPService) Search(ctx context.Context, keywords string, opts *ikuai.ListOptions) (*ikuai.Page[ARPEntry], error) {
	defaults := map[string]any{
		"FINDS":    "ip_addr,mac,comment,interface",
		"KEYWORDS": keywords,
	}
	resp, err := s.c.Call(ctx, "arp", "show", pageShowParam(opts, defaults))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ARPEntry](resp.Data)
}
