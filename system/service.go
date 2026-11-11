package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是系统设置大类的服务入口，聚合 Web 控制台「系统设置」菜单下
// 的 16 个功能，另含企业微信绑定（wecom：应用绑定 dingtalk 的姊妹 func，
// 与钉钉共用「应用绑定」页面，未单列菜单）。
type Service struct {
	c *ikuai.Client

	// Basic 基础设置（basic）：主机名/NAT/NTP/语言/时区，含 set_language 与 sync_time。
	Basic *BasicService
	// VRRP 主备热备（vrrp_config）：VRRP 单行配置。
	VRRP *VRRPService
	// Disk 磁盘分区（disk_mgmt）：磁盘与分区信息。
	Disk *DiskService
	// File 文件管理（file_mgmt）：目录文件列表。
	File *FileService
	// Raid RAID 管理（raid_mgmt）：RAID 阵列列表。
	Raid *RaidService
	// DingTalk 应用绑定（dingtalk）：钉钉绑定状态。
	DingTalk *DingTalkService
	// WeCom 企业微信绑定（wecom）：应用绑定的姊妹功能，绑定状态查询。
	WeCom *WeComService
	// Register 服务注册（register）：注册节点与网关 ID。
	Register *RegisterService
	// ALG ALG 设置（alg）：应用层网关开关（单行配置）。
	ALG *ALGService
	// IRQBalance 负载监控（irqbalance）：硬/软中断分布与 CPU 调频。
	IRQBalance *IRQBalanceService
	// Sysctl 内核参数（ik_sysctl）：TCP/UDP 超时等内核参数（单行配置）。
	Sysctl *SysctlService
	// WebUser Web 账号管理（webuser）：管理账号与改密标记。
	WebUser *WebUserService
	// RemoteControl 远程控制（remote_control）：Web/SSH/FTP/Telnet 远程管理端口（单行配置）。
	RemoteControl *RemoteControlService
	// Upgrade 升级（upgrade）：固件版本与升级检测。
	Upgrade *UpgradeService
	// Backup 备份（backup）：配置备份文件与磁盘用量。
	Backup *BackupService
	// HardwareInfo 硬件信息（hardwareinfo）：CPU/主板/内存/PCI 设备。
	HardwareInfo *HardwareInfoService
	// Reboots 重启（reboots）：定时重启任务列表。
	Reboots *RebootsService
}

// New 构造系统设置服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:             c,
		Basic:         &BasicService{c: c},
		VRRP:          &VRRPService{c: c},
		Disk:          &DiskService{c: c},
		File:          &FileService{c: c},
		Raid:          &RaidService{c: c},
		DingTalk:      &DingTalkService{c: c},
		WeCom:         &WeComService{c: c},
		Register:      &RegisterService{c: c},
		ALG:           &ALGService{c: c},
		IRQBalance:    &IRQBalanceService{c: c},
		Sysctl:        &SysctlService{c: c},
		WebUser:       &WebUserService{c: c},
		RemoteControl: &RemoteControlService{c: c},
		Upgrade:       &UpgradeService{c: c},
		Backup:        &BackupService{c: c},
		HardwareInfo:  &HardwareInfoService{c: c},
		Reboots:       &RebootsService{c: c},
	}
}

// 固件通用的启用/禁用字符串值（enabled 字段，与固件完全一致）。
const (
	EnabledYes = "yes"
	EnabledNo  = "no"
)

// Raw 以任意 func_name/action/param 调用系统设置大类接口，作为未封装
// action/TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
// 注意：本包多数写动作（重启/升级/改密/备份导入等）未提供强类型封装，
// 只应通过 Raw 在充分知晓后果的前提下调用，且严禁对真机执行未验证的
// 写动作（见 docs/RULES.md §9.2）。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// selectorShowParam 构造带固定默认键的 show 参数：opts == nil 时仅由
// defaults 生成参数；opts 非空时以 opts.BuildParam() 为底，defaults 只补
// 缺失键，用户在 opts.Type / opts.Extra 中显式给出的值优先。
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

// pageShowParam 构造标准分页类 func（raid_mgmt/reboots/backup 等支持
// TYPE=total,data 的 func）的 show 参数：opts == nil 时补
// ikuai.DefaultListOptions()（TYPE=total,data, limit="0,20"）。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// typedShowParam 构造 TYPE 选择器类 func（vrrp_config/disk_mgmt/alg/
// ik_sysctl/remote_control/irqbalance/file_mgmt 等以 TYPE=data 类区块为
// 主的 func）的 show 参数：opts 为 nil 或未显式指定 Type 时强制写入
// defaultType（固件对这类 func 的 TYPE=total 回 30006 unknown TYPE
// (total)，真机 3.7.24 验证）；opts.Type 显式给出的值被尊重（个别 func
// 如 raid_mgmt/backup 同时接受 total,data）。其余键仍尊重 opts
// （limit/ORDER_BY/ORDER 等）。
func typedShowParam(opts *ikuai.ListOptions, defaultType string, defaults map[string]any) map[string]any {
	p := selectorShowParam(opts, defaults)
	if opts == nil || opts.Type == "" {
		p["TYPE"] = defaultType
	}
	return p
}
