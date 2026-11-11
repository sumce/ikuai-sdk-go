package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveRemoteControl 真机只读验证远程控制（remote_control.show）。
// Save 为写动作，严禁对真机调用。
func TestLiveRemoteControl(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).RemoteControl

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("remote_control.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("remote_control.show data 不应为空（固件恒返回单行）")
	}
	if cfg.HTTPPort <= 0 || cfg.HTTPSPort <= 0 {
		t.Errorf("管理端口异常: %+v", cfg)
	}
	t.Logf("http=%d https=%d 强制HTTPS=%d 外网Web=%d ssh=%d:%d ftp=%d:%d telnetd=%d",
		cfg.HTTPPort, cfg.HTTPSPort, cfg.ForceHTTPS, cfg.OpenWANWeb,
		cfg.OpenSSHD, cfg.SSHDPort, cfg.OpenFTP, cfg.FTPPort, cfg.OpenTelnetd)

	page, err := svc.List(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("remote_control.show 分页形态失败: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Errorf("分页形态应返回单行: %d", len(page.Rows))
	}
}
