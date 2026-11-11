package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveOnlineUsers 真机只读验证在线用户（ppp_online.show）。
func TestLiveOnlineUsers(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.OnlineUsers.List(ctx, nil)
	if err != nil {
		t.Fatalf("ppp_online.show 失败: %v", err)
	}
	t.Logf("在线用户总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码在线用户行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}
}
