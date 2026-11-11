package overview

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是系统概况大类的服务入口，聚合 Web 控制台
// 「系统概况」首页涉及的两个 func_name：homepage（首页仪表盘）
// 与 sysstat（系统实时状态）。
type Service struct {
	c *ikuai.Client

	// Homepage 首页仪表盘（homepage.show）：系统概况/接口快照/应用流量/
	// 每日统计/DHCP 可分配地址数等数据块。
	Homepage *HomepageService
	// SysStat 系统实时状态（sysstat.show）：CPU/内存/连接/版本信息。
	SysStat *SysStatService
}

// New 构造系统概况服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:        c,
		Homepage: &HomepageService{c: c},
		SysStat:  &SysStatService{c: c},
	}
}

// Raw 以任意 func_name/action/param 调用系统概况大类接口，作为未封装
// TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}
