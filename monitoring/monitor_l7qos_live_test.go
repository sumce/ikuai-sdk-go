package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveL7QoS 真机只读验证策略监控（monitor_l7qos.show）。
func TestLiveL7QoS(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	st, err := svc.L7QoS.Status(ctx, nil)
	if err != nil {
		t.Fatalf("monitor_l7qos.show 失败: %v", err)
	}
	t.Logf("流控策略行数=%d 接口数=%d", len(st.Data), len(st.Interface))
}
