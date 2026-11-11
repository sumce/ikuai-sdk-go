package advanced

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// ik_web_cache.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// WebCacheTypeData 缓存配置区块（线路/服务/端口）。本 func 不接受
	// TYPE=total（真机回 30006 unknown TYPE (total)）；TYPE=all 与 data
	// 等价。
	WebCacheTypeData = "data"
	// WebCacheTypeStatus 缓存状态区块（各服务的运行状态与错误信息）。
	WebCacheTypeStatus = "status"
)

// WebCacheLine 是网页缓存的一条线路配置（ik_web_cache.show data.line
// 数组元素，真机 3.7.24 实证 15 行）。
type WebCacheLine struct {
	// ID 线路行 ID（真机观测 1..15，与线路一一对应）。
	ID int64 `json:"id"`
	// Name 线路名（真机观测 "adsl03"/"vwan1" 等）。
	Name string `json:"name"`
	// Up 线路上行带宽（0 表示不限速；单位待真机验证，控制台提示
	// 「上下行带宽为０，表示不限速」）。
	Up int64 `json:"up"`
	// Down 线路下行带宽（0 表示不限速；单位待真机验证）。
	Down int64 `json:"down"`
	// Weight 权重（值越大，该网口被使用的概率越大，真机观测 1）。
	Weight int64 `json:"weight"`
	// Comment 备注（真机观测为空串）。
	Comment string `json:"comment"`
	// Enabled 线路缓存开关（yes/no，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// Status 线路状态（真机观测 "idle"/"invalid"，"invalid" 为无效线路）。
	Status string `json:"status"`
}

// IsEnabled 线路缓存开关为 "yes"。
func (l WebCacheLine) IsEnabled() bool { return l.Enabled == EnabledYes }

// WebCacheServiceCfg 是一类缓存服务的配置（ik_web_cache.show
// data.service 数组元素，真机观测 video/data 两项）。
type WebCacheServiceCfg struct {
	// Type 缓存类型（"video"=视频缓存，"data"=应用缓存，真机实证）。
	Type string `json:"type"`
	// Enabled 缓存开关（yes/no，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// MaxSize 允许缓存使用的最大硬盘空间（真机观测 10；单位待真机验证，
	// 控制台语义为「缓冲空间」）。
	MaxSize int64 `json:"max_size"`
	// InactiveTime 缓存非活跃时间（真机观测 video=10080 / data=43200；
	// 单位待真机验证，语义为缓存文件在该时长内未被访问时自动清理）。
	InactiveTime int64 `json:"inactive_time"`
	// CacheMinUses 最小访问阈值——请求多少次开始缓存（真机观测 1）。
	CacheMinUses int64 `json:"cache_min_uses"`
}

// IsEnabled 缓存开关为 "yes"。
func (c WebCacheServiceCfg) IsEnabled() bool { return c.Enabled == EnabledYes }

// WebCacheConfig 是网页缓存的完整配置（ik_web_cache.show 的 data 区块
// 整体，真机 3.7.24 实证字段）。
type WebCacheConfig struct {
	// LineCount 线路条数（真机观测 15）。
	LineCount int64 `json:"line_count"`
	// Lines 线路配置列表。
	Lines []WebCacheLine `json:"line"`
	// Port 缓存服务监听端口（真机观测 650）。
	Port int64 `json:"port"`
	// ServiceCount 缓存服务类数（真机观测 2：video/data）。
	ServiceCount int64 `json:"service_count"`
	// Services 各类缓存服务配置。
	Services []WebCacheServiceCfg `json:"service"`
}

// WebCacheStatusRow 是一类缓存服务的运行状态（ik_web_cache.show
// TYPE=status 的 status 数组元素，真机 3.7.24 实证）。
type WebCacheStatusRow struct {
	// ServiceType 缓存类型（"video"/"data"，真机实证）。
	ServiceType string `json:"service_type"`
	// CacheStatus 缓存状态（真机观测 "idle"；"running"=运行中、
	// "running_stop"=停止运行，其余取值待真机验证）。
	CacheStatus string `json:"cache_status"`
	// Err 状态说明（真机观测「没有挂载外部存储(Ext文件系统)」，
	// 正常时为空）。
	Err string `json:"err"`
}

// webCacheDataWrapper 对应 ik_web_cache.show 的 Data 外层
// {"data": {...配置...}} 嵌套形态（真机实证）。
type webCacheDataWrapper struct {
	Data WebCacheConfig `json:"data"`
}

// webCacheStatusWrapper 对应 ik_web_cache.show TYPE=status 的 Data 外层
// {"status": [...]} 形态（真机实证）。
type webCacheStatusWrapper struct {
	Status []WebCacheStatusRow `json:"status"`
}

// WebCacheService 是网页缓存（ik_web_cache）的子服务。
type WebCacheService struct {
	c *ikuai.Client
}

// Get 获取网页缓存配置（线路列表、监听端口、视频/应用缓存服务配置）。
// 对应 /Action/call: ik_web_cache.show（TYPE=data），
// 即 Web 控制台「高级应用 → 网页缓存 → 缓存设置」页面。
func (s *WebCacheService) Get(ctx context.Context) (*WebCacheConfig, error) {
	param := typedShowParam(nil, WebCacheTypeData, nil)
	var w webCacheDataWrapper
	if err := s.c.CallData(ctx, "ik_web_cache", "show", param, &w); err != nil {
		return nil, err
	}
	return &w.Data, nil
}

// Status 获取网页缓存运行状态（视频/应用缓存的缓存状态与错误信息）。
// 对应 /Action/call: ik_web_cache.show（TYPE=status），
// 即 Web 控制台「高级应用 → 网页缓存 → 缓存状态」页面。
func (s *WebCacheService) Status(ctx context.Context) ([]WebCacheStatusRow, error) {
	param := typedShowParam(nil, WebCacheTypeStatus, nil)
	var w webCacheStatusWrapper
	if err := s.c.CallData(ctx, "ik_web_cache", "show", param, &w); err != nil {
		return nil, err
	}
	return w.Status, nil
}

// Save 保存网页缓存配置（线路开关/权重、监听端口、各缓存服务参数）。
// 对应 /Action/call: ik_web_cache.edit（cfg 为 Get 返回并修改后的完整
// 配置，整块提交）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；action 名（edit）与字段
// 集待真机验证。
func (s *WebCacheService) Save(ctx context.Context, cfg WebCacheConfig) error {
	_, err := s.c.Call(ctx, "ik_web_cache", "edit", cfg)
	return err
}
