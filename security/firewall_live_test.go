package security

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveFirewallGet 真机只读验证云防火墙状态（firewall.show TYPE=all）。
// Save 为写动作（action/字段待真机验证），严禁对真机调用。
func TestLiveFirewallGet(t *testing.T) {
	c := livetest.Client(t)
	st, err := New(c).Firewall.Get(context.Background())
	if err != nil {
		t.Fatalf("firewall.show TYPE=all 失败: %v", err)
	}
	if st == nil {
		t.Fatal("firewall.show 应返回状态")
	}
	t.Logf("云防火墙: status=%d cpu_cores=%d", st.Status, st.CPUCores)
	if st.CPUCores <= 0 {
		t.Errorf("cpu_cores 应为正数, got %d", st.CPUCores)
	}
}

// TestLiveFirewallStatus 真机只读验证云防火墙状态区块
// （firewall.show TYPE=status）。
func TestLiveFirewallStatus(t *testing.T) {
	c := livetest.Client(t)
	n, err := New(c).Firewall.Status(context.Background())
	if err != nil {
		t.Fatalf("firewall.show TYPE=status 失败: %v", err)
	}
	t.Logf("云防火墙插件 status=%d", n)
}
