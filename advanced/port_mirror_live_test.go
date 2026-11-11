package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLivePortMirror 真机只读验证端口镜像规则与可选端口（port_mirror.show
// TYPE=total,data / TYPE=all，真机探测实证）。只读，不做任何写动作。
func TestLivePortMirror(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).PortMirror
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("port_mirror.show 失败: %v", err)
	}
	t.Logf("端口镜像规则 total=%d", page.Total)
	for _, r := range page.Rows {
		t.Logf("规则: id=%d target=%s mirror=%s dir=%s comment=%s",
			r.ID, r.TargetPort, r.MirrorPort, r.Dir, r.Comment)
		if r.TargetPort == "" || r.MirrorPort == "" {
			t.Error("镜像规则 target_port/mirror_port 为空（行字段空表推导，待真机验证）")
		}
	}

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("port_mirror.show TYPE=all 失败: %v", err)
	}
	t.Logf("可选端口: %d 项", len(info.Interface))
}
