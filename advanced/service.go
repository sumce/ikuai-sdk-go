package advanced

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是高级应用大类的服务入口，聚合 Web 控制台
// 「高级应用」菜单下的全部 13 个功能。
type Service struct {
	c *ikuai.Client

	// WebCache 网页缓存（ik_web_cache）：缓存线路/服务配置与缓存状态。
	WebCache *WebCacheService
	// DDNS 动态域名（ddns）：多线 DDNS 解析记录。
	DDNS *DDNSService
	// NATDDNS NAT DDNS（nat_ddns）：内网穿透功能状态。
	NATDDNS *NATDDNSService
	// SNMP SNMP 服务（netsnmp）：路由器自身作为 SNMP Agent 的配置。
	SNMP *SNMPService
	// SNMPClient SNMP 客户端（netsnmpc）：向网管平台上报的采集配置。
	SNMPClient *SNMPClientService
	// PortMirror 端口镜像（port_mirror）：交换机端口镜像规则。
	PortMirror *PortMirrorService
	// QEMU 虚拟机（qemu）：KVM 虚拟机列表与宿主机资源信息。
	QEMU *QEMUService
	// Plugins 插件（plugins）：已安装插件列表。
	Plugins *PluginsService
	// Wakeup 网络唤醒（wakeup）：WOL 定时唤醒任务与立即唤醒。
	Wakeup *WakeupService
	// FileServices 文件服务（ftp_server/smbd/http_server）：FTP/Samba/HTTP 服务。
	FileServices *FileServicesService
	// UDProxy UDP 代理（udp_proxy）：udpxy IPTV 单线复用代理。
	UDProxy *UDPProxyService
}

// New 构造高级应用服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:            c,
		WebCache:     &WebCacheService{c: c},
		DDNS:         &DDNSService{c: c},
		NATDDNS:      &NATDDNSService{c: c},
		SNMP:         &SNMPService{c: c},
		SNMPClient:   &SNMPClientService{c: c},
		PortMirror:   &PortMirrorService{c: c},
		QEMU:         &QEMUService{c: c},
		Plugins:      &PluginsService{c: c},
		Wakeup:       &WakeupService{c: c},
		FileServices: &FileServicesService{c: c},
		UDProxy:      &UDPProxyService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用高级应用大类接口，作为未封装
// TYPE 组合与未建模动作（如插件安装、虚拟机电源操作）的逃生口
// （每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// 固件通用的启用/禁用字符串值（enabled 字段，与固件完全一致）。
const (
	EnabledYes = "yes"
	EnabledNo  = "no"
)

// selectorShowParam 构造 TYPE 选择器类 func（ik_web_cache/nat_ddns/plugins/
// smbd 等）的 show 参数。
//
//	opts == nil 时仅由 defaults 生成参数（与控制台抓包一致，不带分页
//	limit）；opts 非空时以 opts.BuildParam() 为底，defaults 只补缺失键，
//	用户在 opts.Type / opts.Extra 中显式给出的值优先。
func selectorShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	var p map[string]any
	if opts != nil {
		p = opts.BuildParam()
	}
	if p == nil {
		p = map[string]any{}
	}
	for k, v := range defaults {
		if _, ok := p[k]; !ok {
			p[k] = v
		}
	}
	return p
}

// pageShowParam 构造分页类 func（ddns/netsnmp/netsnmpc/port_mirror/qemu/
// wakeup/ftp_server/http_server/udp_proxy 等支持 TYPE=total,data 的 func）
// 的 show 参数：opts == nil 时补 ikuai.DefaultListOptions()
// （TYPE=total,data, limit="0,20"），defaults 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// typedShowParam 构造不接受 TYPE=total 的 func（ik_web_cache/nat_ddns/
// plugins/smbd，真机 3.7.24 对 TYPE=total,data 回 30006 unknown TYPE
// (total)）的 show 参数：opts 为 nil 或未显式指定 Type 时强制写入
// defaultType；opts.Type 显式给出的值被尊重。其余键仍尊重 opts
// （limit/ORDER_BY/ORDER 等）。
func typedShowParam(opts *ikuai.ListOptions, defaultType string, defaults map[string]any) map[string]any {
	p := selectorShowParam(opts, defaults)
	if opts == nil || opts.Type == "" {
		p["TYPE"] = defaultType
	}
	return p
}
