package monitoring

import (
	"context"
	"strings"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// partialTypes 构造指定 TYPE 组合的 ListOptions（真机测试辅助）。
func partialTypes(types ...string) *ikuai.ListOptions {
	return &ikuai.ListOptions{Type: strings.Join(types, ",")}
}

// TestLiveInterfaceMonitor 真机只读验证线路监控（monitor_iface.show）。
func TestLiveInterfaceMonitor(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	m, err := svc.Interface.Monitor(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_iface.show 全量 TYPE 失败: %v", err)
	}
	t.Logf("iface_check=%d iface_stream=%d ether_info=%d snapshoot_lan=%d snapshoot_wan=%d",
		len(m.IfaceCheck), len(m.IfaceStream), len(m.EtherInfo), len(m.SnapshootLAN), len(m.SnapshootWAN))
	if len(m.IfaceCheck) == 0 || len(m.IfaceStream) == 0 {
		t.Error("真机应有 iface_check/iface_stream 数据")
	}
	if len(m.EtherInfo) == 0 || len(m.SnapshootLAN) == 0 || len(m.SnapshootWAN) == 0 {
		t.Error("真机应有 ether_info/snapshoot 数据")
	}

	// 部分区块组合。
	m2, err := svc.Interface.Monitor(ctx, partialTypes(IfaceTypeCheck, IfaceTypeStream))
	if err != nil {
		t.Fatalf("monitor_iface.show 部分 TYPE 失败: %v", err)
	}
	if len(m2.IfaceCheck) == 0 || m2.EtherInfo != nil {
		t.Fatalf("部分 TYPE 返回不符: iface_check=%d ether_info=%v", len(m2.IfaceCheck), m2.EtherInfo)
	}
}
