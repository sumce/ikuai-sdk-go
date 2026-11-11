package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveGroupList 真机只读验证 AP 分组（ac_group.show；本机无 AP，
// 列表预期为空）。
func TestLiveGroupList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Group.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("ac_group.show 失败: %v", err)
	}
	t.Logf("AP 分组 total=%d", page.Total)
	for _, g := range page.Rows {
		t.Logf("分组: id=%d name=%s ap_count=%d channel=%s max_pc=%d",
			g.ID, g.Name, g.APCount, g.Channel, g.MaxPC)
		if g.Name == "" {
			t.Error("分组行 name 为空（行字段空表推导，待真机验证）")
		}
	}
}
