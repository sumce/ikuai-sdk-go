// Package system 实现爱快 Web 控制台「系统设置」大类，共 16 个功能：
//
//	basic           基础设置（含 set_language）
//	vrrp_config     主备热备
//	disk_mgmt       磁盘分区
//	file_mgmt       文件管理
//	raid_mgmt       RAID 管理
//	dingtalk        应用绑定（钉钉）
//	register        服务注册
//	alg             ALG 设置
//	irqbalance      负载均衡（中断）
//	ik_sysctl       内核参数
//	webuser         Web 账号管理（含改密）
//	remote_control  远程控制
//	upgrade         升级
//	backup          备份
//	hardwareinfo    硬件信息
//	reboots         重启
//
// 开发状态：已实现（系统设置 大类 17 个功能）。全部通过离线测试与真机只读验证；
// 写动作仅离线覆盖并逐处标注「禁止对真机调用」。
package system
