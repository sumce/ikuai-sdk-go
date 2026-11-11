package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveServerList 真机只读验证 AP 列表（ac_server.show，真机探测实证
// 返回 {"total","data"}；本机无 AP，列表预期为空）。
func TestLiveServerList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Server.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("ac_server.show 失败: %v", err)
	}
	t.Logf("AP 列表 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("AP: id=%d mac=%s ip=%s model=%s ver=%s status=%s",
			r.ID, r.MAC, r.IP, r.Model, r.Version, r.Status)
		if r.MAC == "" {
			t.Error("AP 行 mac 为空（行字段空表推导，待真机验证）")
		}
	}
}
