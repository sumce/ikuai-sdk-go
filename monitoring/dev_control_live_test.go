package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDevControl 真机只读验证周边设备监控（dev_control.show 与
// check_online_status，后者为语义只读的在线状态检测）。
func TestLiveDevControl(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.DevControl.List(ctx, nil)
	if err != nil {
		t.Fatalf("dev_control.show 失败: %v", err)
	}
	t.Logf("智能设备总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码设备行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}

	if err := svc.DevControl.CheckOnlineStatus(ctx); err != nil {
		t.Fatalf("dev_control.check_online_status 失败: %v", err)
	}
	t.Log("check_online_status 触发成功")
}
