package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveRoute 真机只读验证路由族（static_rt.show 与
// static_rt_table.show 运行状态表）。
func TestLiveRoute(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Route

	statics, err := svc.StaticList(ctx, nil)
	if err != nil {
		t.Fatalf("static_rt.show 失败: %v", err)
	}
	t.Logf("静态路由 total=%d", statics.Total)

	table, err := svc.TableList(ctx, nil)
	if err != nil {
		t.Fatalf("static_rt_table.show 失败: %v", err)
	}
	t.Logf("当前路由表 total=%d", table.Total)
	if table.Total == 0 || len(table.Rows) == 0 {
		t.Fatal("真机路由表不应为空")
	}
	if table.Rows[0].DstAddr == "" {
		t.Error("路由表首行 dst_addr 为空")
	}
}
