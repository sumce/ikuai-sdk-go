package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveNetOptimize 真机只读验证无线网优（ac_net_optimize.show，真机
// 探测实证返回 {"total","data","mode","frequencies"}）。
// ⚠️ 本文件只做只读 show 验证；Set 写动作绝对禁止对真机调用。
func TestLiveNetOptimize(t *testing.T) {
	c := livetest.Client(t)
	info, err := New(c).NetOptimize.Get(context.Background(), nil)
	if err != nil {
		t.Fatalf("ac_net_optimize.show 失败: %v", err)
	}
	t.Logf("无线网优 total=%d mode=%d frequencies=%d data=%d 行",
		info.Total, info.Mode, info.Frequencies, len(info.Data))
	for _, item := range info.Data {
		if m, err := item.Map(); err == nil {
			t.Logf("优化行: %v", m)
		}
	}
}
