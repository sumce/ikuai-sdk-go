package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveUpgradeList 真机只读验证 AP 可升级列表（ac_upgrade.show；
// 本机无 AP，列表预期为空）。
// ⚠️ 本文件只做只读 show 验证；Upgrade 写动作绝对禁止对真机调用。
func TestLiveUpgradeList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Upgrade.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("ac_upgrade.show 失败: %v", err)
	}
	t.Logf("AP 可升级列表 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("可升级: id=%d mac=%s model=%s %s -> %s",
			r.ID, r.MAC, r.Model, r.Version, r.NewVersion)
		if r.MAC == "" {
			t.Error("升级行 mac 为空（行字段空表推导，待真机验证）")
		}
	}
}
