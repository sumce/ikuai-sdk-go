package sdk

import (
	"context"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/ac"
	"github.com/sumce/ikuai-sdk-go/accounting"
	"github.com/sumce/ikuai-sdk-go/advanced"
	"github.com/sumce/ikuai-sdk-go/behavior"
	"github.com/sumce/ikuai-sdk-go/flowcontrol"
	"github.com/sumce/ikuai-sdk-go/logcenter"
	"github.com/sumce/ikuai-sdk-go/monitoring"
	"github.com/sumce/ikuai-sdk-go/network"
	"github.com/sumce/ikuai-sdk-go/overview"
	"github.com/sumce/ikuai-sdk-go/security"
	"github.com/sumce/ikuai-sdk-go/system"
	"github.com/sumce/ikuai-sdk-go/tools"
)

// Config 是门面构造参数。
type Config struct {
	// BaseURL 路由器管理地址，如 http://192.168.1.1（支持 https 与端口）。
	BaseURL string
	// Username / Password 为 Web 控制台账号；Password 将以 md5 提交。
	Username string
	Password string
	// Options 透传核心包函数式选项（WithTimeout/WithTLSInsecureSkipVerify/...）。
	Options []ikuai.Option
}

// Router 聚合全部 12 个业务服务的门面。字段与 Web 控制台菜单大类一一对应。
type Router struct {
	// Client 是底层核心客户端（通用 Call/CRUD 兜底、会话管理）。
	Client *ikuai.Client

	// Overview 系统概况（homepage/sysstat）。
	Overview *overview.Service
	// Monitoring 状态监控（线路/终端/协议/策略/负载/分流/摄像头/交换机/周边）。
	Monitoring *monitoring.Service
	// System 系统设置（基础/热备/磁盘/文件/RAID/升级/备份/账号/企业微信…）。
	System *system.Service
	// Network 网络设置（LAN/DHCP/DNS/VLAN/分组/路由/VPN客户端/UPnP/NAT/IPv6/组播）。
	Network *network.Service
	// FlowControl 流控分流（多线负载/分流族/智能流控/限速/自定义协议）。
	FlowControl *flowcontrol.Service
	// AC AC管理（无线控制器）。
	AC *ac.Service
	// Accounting 认证计费（Web认证/服务端/账号/套餐/优惠券/通知）。
	Accounting *accounting.Service
	// Behavior 行为管控（审计/访问控制/网页与域名过滤）。
	Behavior *behavior.Service
	// Security 安全设置（ACL/ARP/连接数/高级/防火墙）。
	Security *security.Service
	// Advanced 高级应用（缓存/DDNS/SNMP/镜像/虚拟机/插件/唤醒/文件服务）。
	Advanced *advanced.Service
	// Tools 应用工具（Ping/Traceroute/抓包/iperf/子网/测速/诊断/看门狗）。
	Tools *tools.Service
	// LogCenter 日志中心（syslog 族/站内信/实时采集）。
	LogCenter *logcenter.Service
}

// New 构造门面路由器（不登录；随后调用 Login）。
func New(cfg *Config) (*Router, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sdk: Config 不能为空: %w", ikuai.ErrInvalidParam)
	}
	c, err := ikuai.NewClient(cfg.BaseURL, cfg.Username, cfg.Password, cfg.Options...)
	if err != nil {
		return nil, err
	}
	return &Router{
		Client:      c,
		Overview:    overview.New(c),
		Monitoring:  monitoring.New(c),
		System:      system.New(c),
		Network:     network.New(c),
		FlowControl: flowcontrol.New(c),
		AC:          ac.New(c),
		Accounting:  accounting.New(c),
		Behavior:    behavior.New(c),
		Security:    security.New(c),
		Advanced:    advanced.New(c),
		Tools:       tools.New(c),
		LogCenter:   logcenter.New(c),
	}, nil
}

// Login 登录并建立会话（其余调用电前必须成功登录一次；
// 会话过期由核心包自动重登）。
func (r *Router) Login(ctx context.Context) error { return r.Client.Login(ctx) }

// Logout 清除本地会话。
func (r *Router) Logout(ctx context.Context) { r.Client.Logout(ctx) }

// KeepAlive 保活会话（等价控制台首页轮询）。
func (r *Router) KeepAlive(ctx context.Context) error { return r.Client.KeepAlive(ctx) }
