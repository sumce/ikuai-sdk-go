package sdk

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestRouterLoginAndCall 验证门面登录与 12 大类服务聚合可用。
func TestRouterLoginAndCall(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	router, err := New(&Config{BaseURL: ts.URL, Username: testserver.TestUsername, Password: testserver.TestPassword})
	if err != nil {
		t.Fatalf("构造门面失败: %v", err)
	}
	ctx := context.Background()
	if err := router.Login(ctx); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer router.Logout(ctx)

	// 系统概况走门面。
	ts.SetResponse("homepage", "show", map[string]any{
		"sysstat":   map[string]any{"hostname": "iKuai"},
		"ac_status": map[string]any{"ap_count": 0, "ap_online": 0},
	})
	ov, err := router.Overview.SystemOverview(ctx)
	if err != nil {
		t.Fatalf("门面系统概况失败: %v", err)
	}
	if ov.SysStat.Hostname != "iKuai" {
		t.Fatalf("hostname 解码错误: %q", ov.SysStat.Hostname)
	}

	// 12 个服务字段全部非空。
	for name, svc := range map[string]any{
		"Overview": router.Overview, "Monitoring": router.Monitoring, "System": router.System,
		"Network": router.Network, "FlowControl": router.FlowControl, "AC": router.AC,
		"Accounting": router.Accounting, "Behavior": router.Behavior, "Security": router.Security,
		"Advanced": router.Advanced, "Tools": router.Tools, "LogCenter": router.LogCenter,
	} {
		if svc == nil {
			t.Errorf("门面字段 %s 为 nil", name)
		}
	}

	// 通用兜底通道经门面 Client 可用。
	ts.SetResponse("cflow", "show", map[string]any{"data": map[string]any{}})
	if _, err := router.Client.Call(ctx, "cflow", "show", nil); err != nil {
		t.Fatalf("门面 Client.Call 失败: %v", err)
	}
}

// TestRouterBadConfig 验证非法配置报错。
func TestRouterBadConfig(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("nil Config 应报错")
	}
	if _, err := New(&Config{BaseURL: "ftp://x", Username: "a", Password: "b"}); err == nil {
		t.Fatal("非法 scheme 应报错")
	}
}
