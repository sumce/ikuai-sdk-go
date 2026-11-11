package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWebAuth 真机只读验证 Web 认证服务配置（webauth.show TYPE=data）。
func TestLiveWebAuth(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	cfg, err := svc.WebAuth.Get(ctx)
	if err != nil {
		t.Fatalf("webauth.show 失败: %v", err)
	}
	if cfg == nil {
		t.Log("webauth 配置未落库（data 为空）")
		return
	}
	// 只记录非敏感字段，避免泄漏 radius_key/passwd。
	t.Logf("enabled=%s interface=%s popups=%d idle_time=%d weixin=%d tryout_time=%d",
		cfg.Enabled, cfg.Interface, cfg.Popups, cfg.IdleTime, cfg.Weixin, cfg.TryoutTime)
}
