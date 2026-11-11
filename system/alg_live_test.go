package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveALG 真机只读验证 ALG 设置（alg.show TYPE=data）。
// Save 为写动作，严禁对真机调用。
func TestLiveALG(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).ALG

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("alg.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("alg.show data 不应为空（固件恒返回单行）")
	}
	if cfg.ID != 1 {
		t.Errorf("行 ID 应为 1: %+v", cfg)
	}
	t.Logf("FTP=%d TFTP=%d SIP=%d H.323=%d ftp_ports=%q sip_ports=%q tftp_ports=%q",
		cfg.SupportFTP, cfg.SupportTFTP, cfg.SupportSIP, cfg.SupportH323,
		cfg.FTPPorts, cfg.SIPPorts, cfg.TFTPPorts)

	page, err := svc.List(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("alg.show 分页形态失败: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Errorf("分页形态应返回单行: %d", len(page.Rows))
	}
}
