package monitoring

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveCFlow 真机只读验证分流监控（cflow.show）。
func TestLiveCFlow(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	out, err := svc.CFlow.Summary(ctx)
	if err != nil {
		t.Fatalf("cflow.show 失败: %v", err)
	}
	todayConn, err := out.Data.Today.ConnCnt.Int64()
	if err != nil {
		t.Fatalf("today.conn_cnt 解析失败: %v", err)
	}
	weekConn, err := out.Data.Week.ConnCnt.Int64()
	if err != nil {
		t.Fatalf("week.conn_cnt 解析失败: %v", err)
	}
	t.Logf("今日连接=%d 本周连接=%d 昨日连接=%q clean_time=%d",
		todayConn, weekConn, out.Data.Yesterday.ConnCnt, out.CleanTime)
}
