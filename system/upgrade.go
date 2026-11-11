package system

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/sumce/ikuai-sdk-go"
)

// upgrade.show 的 TYPE 选择器常量（抓包 + 真机 3.7.24 验证）。
const (
	// UpgradeTypeData 升级信息区块（版本/特征库版本/升级开关）。
	UpgradeTypeData = "data"
	// UpgradeTypeFileinfo 本地上传固件文件信息区块（未上传时为空对象，
	// 语义待真机验证）。
	UpgradeTypeFileinfo = "fileinfo"
	// UpgradeTypeCloudSync 云端配置同步开关区块。
	UpgradeTypeCloudSync = "cloud_sync"
)

// upgrade.show 控制台实际使用的 TYPE 组合（burp 抓包实证）。
const (
	// UpgradeTypeDataFileinfo 是「系统升级」页的组合（data,fileinfo），
	// Upgrade.Status 使用。
	UpgradeTypeDataFileinfo = UpgradeTypeData + "," + UpgradeTypeFileinfo
)

// UpgradeData 是升级信息区块（upgrade.show TYPE=data 的 data 键，
// 注意固件此处 data 为对象而非数组）。
type UpgradeData struct {
	// AutoUpgradeLibDPI 应用协议特征库自动升级开关（1=开启）。
	AutoUpgradeLibDPI int64 `json:"auto_upgrade_lib_dpi"`
	// AutoUpgradeLibIM 应用识别特征库自动升级开关（1=开启）。
	AutoUpgradeLibIM int64 `json:"auto_upgrade_lib_im"`
	// AutoUpgradeLibDomain 应用域名特征库自动升级开关（1=开启）。
	AutoUpgradeLibDomain int64 `json:"auto_upgrade_lib_domain"`
	// AutoUpgradeSec 安全特征库自动升级开关（1=开启）。
	AutoUpgradeSec int64 `json:"auto_upgrade_sec"`
	// LibprotoVer 当前应用协议特征库版本（如 "2.0.486"）。
	LibprotoVer string `json:"libproto_ver"`
	// LibauditVer 当前审计特征库版本（如 "2.1.17"）。
	LibauditVer string `json:"libaudit_ver"`
	// LibdomainVer 当前域名特征库版本（如 "2.1.0"）。
	LibdomainVer string `json:"libdomain_ver"`
	// SystemVer 当前系统版本号（如 "3.7.24"）。
	SystemVer string `json:"system_ver"`
	// BuildDate 当前系统构建时间（固件以字符串表示，数值形如 YYYYMMDDHHMM）。
	BuildDate string `json:"build_date"`
	// NewLibprotoVer 云端最新应用协议特征库版本。
	NewLibprotoVer string `json:"new_libproto_ver"`
	// NewLibauditVer 云端最新审计特征库版本。
	NewLibauditVer string `json:"new_libaudit_ver"`
	// NewLibdomainVer 云端最新域名特征库版本。
	NewLibdomainVer string `json:"new_libdomain_ver"`
	// NewSystemVer 云端最新系统版本号（无新版本时与 SystemVer 相同）。
	NewSystemVer string `json:"new_system_ver"`
	// NewBuildDate 云端最新系统构建时间（字符串 YYYYMMDDHHMM）。
	NewBuildDate string `json:"new_build_date"`
	// UpdateContent 新版本更新说明（多行文本）。
	UpdateContent string `json:"update_content"`
	// IgnoreUpgradeVer 用户选择忽略的升级版本号（未忽略为空）。
	IgnoreUpgradeVer string `json:"ignore_upgrade_ver"`
	// Bootguide 启动介质（观测值 "hd"）。
	Bootguide string `json:"bootguide"`
	// FirmwareChannel 固件更新通道（观测值 0；枚举语义待真机验证）。
	FirmwareChannel int64 `json:"firmware_channel"`
}

// HasNewVersion 报告云端是否存在比当前更新的系统版本。
func (d UpgradeData) HasNewVersion() bool {
	return d.NewSystemVer != "" && d.NewSystemVer != d.SystemVer
}

// BuildTime 解析当前系统构建时间（BuildDate，YYYYMMDDHHMM）为本地时间；
// 空值或解析失败返回零值 time.Time 与错误。
func (d UpgradeData) BuildTime() (time.Time, error) {
	return parseBuildDate(d.BuildDate)
}

// NewBuildTime 解析云端最新系统构建时间（NewBuildDate）为本地时间；
// 空值或解析失败返回零值 time.Time 与错误。
func (d UpgradeData) NewBuildTime() (time.Time, error) {
	return parseBuildDate(d.NewBuildDate)
}

// parseBuildDate 解析固件的构建时间字符串（YYYYMMDDHHMM）。
func parseBuildDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if _, err := strconv.ParseInt(s, 10, 64); err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("200601021504", s, time.Local)
}

// UpgradeStatus 是系统升级页的聚合结果（upgrade.show
// TYPE=data,fileinfo）。
type UpgradeStatus struct {
	// Data 升级信息区块（TYPE 含 data）。
	Data UpgradeData `json:"data"`
	// Fileinfo 本地上传固件文件信息（TYPE 含 fileinfo；未上传时固件返回
	// 空对象，原始 JSON 原样保留，字段待真机验证）。
	Fileinfo json.RawMessage `json:"fileinfo"`
}

// UpgradeService 是升级（upgrade）的子服务。
type UpgradeService struct {
	c *ikuai.Client
}

// Status 获取系统升级信息（当前/最新版本、特征库版本、更新说明）。
// 对应 /Action/call: upgrade.show（TYPE=data,fileinfo，抓包实证），
// 即 Web 控制台「系统设置 → 系统升级」页面。
// 判断是否有新版本用 UpgradeData.HasNewVersion。
func (s *UpgradeService) Status(ctx context.Context) (*UpgradeStatus, error) {
	var out UpgradeStatus
	if err := s.c.CallData(ctx, "upgrade", "show",
		map[string]any{"TYPE": UpgradeTypeDataFileinfo}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CloudSync 获取云端配置同步开关（观测值 1=开启；语义待真机验证）。
// 对应 /Action/call: upgrade.show（TYPE=cloud_sync）。
func (s *UpgradeService) CloudSync(ctx context.Context) (int64, error) {
	var out struct {
		CloudSync int64 `json:"cloud_sync"`
	}
	if err := s.c.CallData(ctx, "upgrade", "show",
		map[string]any{"TYPE": UpgradeTypeCloudSync}, &out); err != nil {
		return 0, err
	}
	return out.CloudSync, nil
}
