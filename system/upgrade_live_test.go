package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveUpgrade 真机只读验证升级信息（upgrade.show 两种 TYPE）。
// ⚠️ upgrade 的一切写动作（触发升级/重启等）严禁对真机调用。
func TestLiveUpgrade(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Upgrade

	st, err := svc.Status(ctx)
	if err != nil {
		t.Fatalf("upgrade.show TYPE=data,fileinfo 失败: %v", err)
	}
	d := st.Data
	if d.SystemVer == "" {
		t.Error("system_ver 不应为空")
	}
	if _, err := d.BuildTime(); err != nil {
		t.Errorf("build_date 解析失败: %v", err)
	}
	if d.NewSystemVer != "" {
		if _, err := d.NewBuildTime(); err != nil {
			t.Errorf("new_build_date 解析失败: %v", err)
		}
	}
	t.Logf("当前版本=%s(%s) 云端版本=%s(%s) 有新版本=%v 特征库 proto=%s audit=%s domain=%s",
		d.SystemVer, d.BuildDate, d.NewSystemVer, d.NewBuildDate, d.HasNewVersion(),
		d.LibprotoVer, d.LibauditVer, d.LibdomainVer)
	if d.UpdateContent != "" {
		t.Logf("更新说明: %s", d.UpdateContent)
	}
	t.Logf("fileinfo 原始 JSON: %s", st.Fileinfo)

	v, err := svc.CloudSync(ctx)
	if err != nil {
		t.Fatalf("upgrade.show TYPE=cloud_sync 失败: %v", err)
	}
	t.Logf("cloud_sync=%d", v)
}
