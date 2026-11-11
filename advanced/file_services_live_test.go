package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveFileServices 真机只读验证 FTP/Samba/HTTP 服务（ftp_server/
// smbd/http_server.show，真机探测实证）。只读，不做任何写动作。
func TestLiveFileServices(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).FileServices
	ctx := context.Background()

	// FTP：用户行 + 全局配置（真机实证 TYPE=all 顶层字段）。
	page, err := svc.FTPList(ctx, nil)
	if err != nil {
		t.Fatalf("ftp_server.show 失败: %v", err)
	}
	t.Logf("FTP 用户 total=%d", page.Total)
	for _, u := range page.Rows {
		t.Logf("FTP 用户: id=%d username=%s permission=%s home=%s enabled=%s",
			u.ID, u.Username, u.Permission, u.HomeDir, u.Enabled)
		if u.Username == "" {
			t.Error("FTP 用户行 username 为空（行字段空表推导，待真机验证）")
		}
	}
	cfg, err := svc.FTPGetConfig(ctx)
	if err != nil {
		t.Fatalf("ftp_server.show TYPE=all 失败: %v", err)
	}
	t.Logf("FTP 配置: open_ftp=%d port=%d access=%d partname=%v",
		cfg.OpenFTP, cfg.FTPPort, cfg.FTPAccess, cfg.Partname)

	// SMB：配置区块（真机实证单行）。
	info, err := svc.SMBGet(ctx)
	if err != nil {
		t.Fatalf("smbd.show 失败: %v", err)
	}
	smb, err := svc.SMBGetConfig(ctx)
	if err != nil {
		t.Fatalf("smbd 配置行获取失败: %v", err)
	}
	if smb != nil {
		t.Logf("SMB: enabled=%s workgroup=%s wsdd2=%d interface=%s access=%d dir_total=%d",
			smb.Enabled, smb.Workgroup, smb.WSDD2, smb.Interface, smb.Access, info.DirTotal)
		if smb.Workgroup == "" {
			t.Error("SMB workgroup 为空")
		}
	} else {
		t.Log("SMB: 无配置行")
	}

	// HTTP：实例行 + 可选分区。
	hp, err := svc.HTTPList(ctx, nil)
	if err != nil {
		t.Fatalf("http_server.show 失败: %v", err)
	}
	t.Logf("HTTP 实例 total=%d", hp.Total)
	for _, r := range hp.Rows {
		t.Logf("HTTP: id=%d tagname=%s port=%d domain=%s home=%s",
			r.ID, r.Tagname, r.HTTPPort, r.ServerName, r.HomeDir)
	}
	hopts, err := svc.HTTPGetOptions(ctx)
	if err != nil {
		t.Fatalf("http_server.show TYPE=all 失败: %v", err)
	}
	t.Logf("HTTP 可选分区: %v", hopts.Partname)
}
