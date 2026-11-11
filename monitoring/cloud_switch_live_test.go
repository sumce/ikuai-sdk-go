package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveCloudSwitch 真机只读验证交换机监控（cloud_switch.show）。
func TestLiveCloudSwitch(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.CloudSwitch.List(ctx, nil)
	if err != nil {
		t.Fatalf("cloud_switch.show 失败: %v", err)
	}
	t.Logf("云交换机总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码交换机行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}
}
