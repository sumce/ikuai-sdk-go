// Package network 实现爱快 Web 控制台「网络设置」大类，共 35 个功能：
//
//	lan              内外网设置
//	dhcp_server      DHCP 服务端
//	dhcp_static      DHCP 静态分配
//	dhcp_lease       DHCP 客户端列表
//	dhcp_acl_mac     DHCP 黑白名单
//	dns              DNS 设置
//	dns_replace      多线 DNS 分流
//	vlan             VLAN 设置
//	ipgroup          IPv4 分组
//	ipv6group        IPv6 分组
//	macgroup         MAC 分组
//	dtgroup          时间分组
//	static_rt        静态路由
//	static_rt_table  当前路由表
//	pptp_client      PPTP 客户端
//	l2tp_client      L2TP 客户端
//	openvpn-client   OpenVPN 客户端
//	ipsec-vpn        IPSec 客户端
//	wireguard        WireGuard 客户端
//	ike_client       IPSec IKEv2 客户端
//	ik_web_sdwan     SD-WAN
//	wuying           无影云
//	upnpd            UPnP 状态
//	upnpd_leases     UPnP 设置/租约
//	nat_rule         NAT 规则
//	dnat             端口映射
//	netmap           DMZ 主机
//	ipv6             IPv6 设置
//	ipv6_stream      IPv6 分流
//	dhcp6_lease      DHCPv6 终端
//	ipv6_static      DHCPv6 静态分配
//	dhcp6_acl_mac    DHCPv6 黑白名单
//	ipv6_neighbor    IPv6 邻居表
//	igmp_proxy       IGMP 代理
//	iptv             IPTV
//
// 开发状态：已实现（网络设置 大类 35 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package network
