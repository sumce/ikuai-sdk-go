package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveVLAN 真机只读验证 VLAN 设置（vlan.show）。
func TestLiveVLAN(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	page, err := New(c).VLAN.List(ctx, nil)
	if err != nil {
		t.Fatalf("vlan.show 失败: %v", err)
	}
	t.Logf("VLAN total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("vlan: id=%d name=%s vlan_id=%s if=%s", r.ID, r.VLANName, r.VLANID, r.Interface)
	}
}
