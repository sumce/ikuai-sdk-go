package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveRaid 真机只读验证 RAID 管理（raid_mgmt.show）。
func TestLiveRaid(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Raid

	// 控制台形态（TYPE=data）。
	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("raid_mgmt.show TYPE=data 失败: %v", err)
	}
	t.Logf("RAID 阵列数=%d", len(page.Rows))

	// 标准分页形态（TYPE=total,data，真机同样接受）。
	page2, err := svc.List(ctx, &ikuai.ListOptions{Type: RaidTypeDataTotal, Limit: 20})
	if err != nil {
		t.Fatalf("raid_mgmt.show TYPE=total,data 失败: %v", err)
	}
	if page2.Total != int64(len(page2.Rows)) {
		t.Errorf("total=%d 与行数=%d 不一致", page2.Total, len(page2.Rows))
	}
	for _, r := range page2.Rows {
		m, err := r.Map()
		if err != nil {
			t.Errorf("行 map 解码失败: %v", err)
		}
		t.Logf("RAID 行: %v", m)
	}
}
