package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWebCache 真机只读验证网页缓存配置与状态（ik_web_cache.show
// TYPE=data / TYPE=status，真机探测实证）。只读，不做任何写动作。
func TestLiveWebCache(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).WebCache
	ctx := context.Background()

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("ik_web_cache.show TYPE=data 失败: %v", err)
	}
	t.Logf("网页缓存: port=%d line_count=%d service_count=%d",
		cfg.Port, cfg.LineCount, cfg.ServiceCount)
	if len(cfg.Lines) != int(cfg.LineCount) {
		t.Errorf("line_count=%d 与 line 行数 %d 不符", cfg.LineCount, len(cfg.Lines))
	}
	for _, l := range cfg.Lines {
		t.Logf("线路: id=%d name=%s enabled=%s status=%s", l.ID, l.Name, l.Enabled, l.Status)
		if l.Name == "" {
			t.Error("线路行 name 为空")
		}
	}
	for _, s := range cfg.Services {
		t.Logf("缓存服务: type=%s enabled=%s max_size=%d inactive_time=%d min_uses=%d",
			s.Type, s.Enabled, s.MaxSize, s.InactiveTime, s.CacheMinUses)
	}

	rows, err := svc.Status(ctx)
	if err != nil {
		t.Fatalf("ik_web_cache.show TYPE=status 失败: %v", err)
	}
	for _, r := range rows {
		t.Logf("缓存状态: type=%s status=%s err=%s", r.ServiceType, r.CacheStatus, r.Err)
	}
}
