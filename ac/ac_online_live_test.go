package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveOnlineClients 真机只读验证无线终端列表（ac_online_clt.show；
// 列名经真机 ORDER_BY/FINDS 探测实证）。
func TestLiveOnlineClients(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).OnlineClients

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("ac_online_clt.show 失败: %v", err)
	}
	t.Logf("在线无线终端 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("终端: mac=%s ip=%s ssid=%s signal=%s up=%d down=%d ap=%s",
			r.MAC, r.IPAddr, r.SSID, r.Signal, r.Upload, r.Download, r.APMAC)
		if r.MAC == "" {
			t.Error("终端行 mac 为空")
		}
	}

	// 只读搜索（FINDS 字段集真机探测实证合法）。
	found, err := svc.Search(ctx, "aa", nil)
	if err != nil {
		t.Fatalf("ac_online_clt.show KEYWORDS 失败: %v", err)
	}
	t.Logf("关键字 aa 命中 total=%d", found.Total)
}
