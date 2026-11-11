package network

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDNAT 真机只读验证端口映射（dnat.show 三种 TYPE，抓包实证）。
func TestLiveDNAT(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).DNAT

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("dnat.show 失败: %v", err)
	}
	t.Logf("端口映射 total=%d", page.Total)
	if len(page.Rows) > 0 {
		r := page.Rows[0]
		t.Logf("首行: id=%d if=%s wan_port=%s lan_addr=%s proto=%s",
			r.ID, r.Interface, r.WANPort, r.LANAddr, r.Protocol)
		if r.Interface == "" || r.Protocol == "" {
			t.Error("首行 interface/protocol 为空")
		}
		if ports := r.WANPorts(); len(ports) == 0 {
			t.Error("WANPorts 解析为空")
		}
	}

	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("dnat.show TYPE=interface,protocol 失败: %v", err)
	}
	if len(sel.Interface) == 0 || len(sel.Protocol) == 0 {
		t.Fatal("接口/协议候选不应为空")
	}
	t.Logf("接口候选 %d 个, 协议候选 %v", len(sel.Interface), sel.Protocol)

	groups, err := svc.IPGroups(ctx)
	if err != nil {
		t.Fatalf("dnat.show TYPE=ipgroup 失败: %v", err)
	}
	t.Logf("源地址分组候选: %v", groups)
}
