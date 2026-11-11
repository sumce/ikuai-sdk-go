package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveBasic 真机只读验证基础设置（basic.show TYPE=timestamp,data,interface,support_fast）。
// 写动作 set_language/sync_time 严禁对真机调用。
func TestLiveBasic(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	info, err := New(c).Basic.Get(ctx)
	if err != nil {
		t.Fatalf("basic.show 失败: %v", err)
	}
	row := info.Row()
	if row == nil {
		t.Fatal("basic.show data 不应为空")
	}
	if row.Hostname == "" {
		t.Errorf("hostname 为空: %+v", row)
	}
	if len(info.Interface) == 0 {
		t.Error("接口列表不应为空")
	}
	if info.Timestamp <= 0 {
		t.Errorf("timestamp 异常: %d", info.Timestamp)
	}
	t.Logf("主机名=%s 语言=%d 时区=%d NAT开关=%d 接口数=%d 快速转发支持=%d",
		row.Hostname, row.Language, row.TimeZone, row.SwitchNat, len(info.Interface), info.SupportFast)
}
