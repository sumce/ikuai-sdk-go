package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSysctl 真机只读验证内核参数（ik_sysctl.show TYPE=data）。
// Save 为写动作，严禁对真机调用。
func TestLiveSysctl(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Sysctl

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("ik_sysctl.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("ik_sysctl.show data 不应为空（固件恒返回单行）")
	}
	if cfg.EstablishedTimeout <= 0 || cfg.UDPTimeout <= 0 {
		t.Errorf("超时参数异常: %+v", cfg)
	}
	t.Logf("BBR=%d established=%ds time_wait=%ds udp=%ds udp_stream=%ds icmp=%ds",
		cfg.BBR, cfg.EstablishedTimeout, cfg.TimeWaitTimeout, cfg.UDPTimeout,
		cfg.UDPStreamTimeout, cfg.ICMPTimeout)

	page, err := svc.List(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("ik_sysctl.show 分页形态失败: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Errorf("分页形态应返回单行: %d", len(page.Rows))
	}
}
