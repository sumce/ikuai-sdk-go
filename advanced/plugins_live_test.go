package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLivePlugins 真机只读验证插件列表（plugins.show TYPE=data，真机
// 探测实证：本 func 不支持 TYPE=total,data）。只读；安装/卸载/启停等
// 插件操作绝对禁止对真机调用。
func TestLivePlugins(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).Plugins.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("plugins.show 失败: %v", err)
	}
	t.Logf("插件 total=%d", page.Total)
	for _, p := range page.Rows {
		if m, merr := p.Map(); merr == nil {
			t.Logf("插件: %v", m)
		}
	}
}
