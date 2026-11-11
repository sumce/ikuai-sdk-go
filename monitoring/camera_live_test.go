package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveCamera 真机只读验证摄像头监控（camera.show）。
func TestLiveCamera(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.Camera.List(ctx, nil)
	if err != nil {
		t.Fatalf("camera.show 失败: %v", err)
	}
	t.Logf("摄像头总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码摄像头行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}
}
