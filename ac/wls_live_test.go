package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveBlacklistList 真机只读验证无线黑白名单（wls_black.show；本机
// 无 AP，列表预期为空）。
func TestLiveBlacklistList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Blacklist.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("wls_black.show 失败: %v", err)
	}
	t.Logf("无线黑白名单 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("规则: id=%d mode=%s mac=%s object=%s week=%s time=%s",
			r.ID, r.ControlMode, r.MAC, r.Object, r.Week, r.Time)
		if r.MAC == "" {
			t.Error("规则行 mac 为空（行字段空表推导，待真机验证）")
		}
	}
}

// TestLiveMvlanList 真机只读验证无线终端 VLAN（wls_mvlan.show；本机
// 无 AP，列表预期为空）。
func TestLiveMvlanList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Mvlan.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("wls_mvlan.show 失败: %v", err)
	}
	t.Logf("无线终端 VLAN total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("绑定: id=%d mac=%s ssid=%s vlan_id=%d", r.ID, r.MAC, r.SSID, r.VlanID)
		if r.VlanID <= 0 {
			t.Error("绑定行 vlan_id 非法（行字段空表推导，待真机验证）")
		}
	}
}
