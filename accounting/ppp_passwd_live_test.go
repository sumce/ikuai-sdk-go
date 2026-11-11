package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSelfPasswd 真机只读验证自助密码配置（ppp_passwd.show TYPE=data）。
func TestLiveSelfPasswd(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	cfg, err := svc.SelfPasswd.Get(ctx)
	if err != nil {
		t.Fatalf("ppp_passwd.show 失败: %v", err)
	}
	if cfg == nil {
		t.Log("ppp_passwd 配置未落库")
		return
	}
	t.Logf("enabled=%s allow_web=%d allow_pppoe=%d allow_pptp=%d allow_l2tp=%d allow_ovpn=%d allow_quit=%d",
		cfg.Enabled, cfg.AllowWeb, cfg.AllowPPPoE, cfg.AllowPPTP,
		cfg.AllowL2TP, cfg.AllowOVPN, cfg.AllowQuit)
}
