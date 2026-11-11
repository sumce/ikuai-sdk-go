package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveTCPDumpStatus 真机只读验证抓包任务状态查询（tcpdump.show
// 空参，真机实测空闲时 filename 为空、size 为 "0"）。
func TestLiveTCPDumpStatus(t *testing.T) {
	c := livetest.Client(t)
	state, err := New(c).TCPDump.Status(context.Background())
	if err != nil {
		t.Fatalf("tcpdump.show 失败: %v", err)
	}
	if state == nil {
		t.Fatal("状态行不应为 nil（真机实测恒返回一行占位）")
	}
	t.Logf("interface=%q mac=%q hit_num=%q proto=%q ip_addr=%q port=%q filename=%q size=%q last_time=%q",
		state.Interface, state.MAC, state.HitNum, state.Proto,
		state.IPAddr, state.Port, state.Filename, state.Size, state.LastTime)
	if state.HitNum == "" {
		t.Error("hit_num 不应为空（真机实测默认 \"100\"）")
	}
}
