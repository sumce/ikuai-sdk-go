package security

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveConnLimitList 真机只读验证连接数限制列表（conn_limit.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveConnLimitList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).ConnLimit.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("conn_limit.show 失败: %v", err)
	}
	t.Logf("连接数限制 total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 5 {
			break
		}
		t.Logf("  行: id=%d enabled=%s ip=%q proto=%s wan_port=%q limits=%d",
			r.ID, r.Enabled, r.IPAddr, r.Protocol, r.WANPort, r.Limits)
		if r.Limits <= 0 {
			t.Errorf("第 %d 行 limits 异常: %d（行字段空表推导，待真机验证）", i, r.Limits)
		}
	}
}

// TestLiveConnLimitSelectors 真机只读验证连接数限制表单选择器
// （conn_limit.show TYPE=all；真机实测 ipgroup 有值）。
func TestLiveConnLimitSelectors(t *testing.T) {
	c := livetest.Client(t)
	sel, err := New(c).ConnLimit.Selectors(context.Background())
	if err != nil {
		t.Fatalf("conn_limit.show TYPE=all 失败: %v", err)
	}
	t.Logf("ipgroup=%v", sel.IPGroup)
}
