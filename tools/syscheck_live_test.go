package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSysCheckResult 真机只读验证系统体检结果查询（iksyscheck.show
// 空参；真机未体检时 7 项 status 均为 "-1"；不发起任何新体检）。
func TestLiveSysCheckResult(t *testing.T) {
	c := livetest.Client(t)
	state, err := New(c).SysCheck.Result(context.Background())
	if err != nil {
		t.Fatalf("iksyscheck.show 失败: %v", err)
	}
	sections := []struct {
		name   string
		status string
		res    int
	}{
		{"disk", state.DiskStatus, len(state.DiskRes)},
		{"memory", state.MemoryStatus, len(state.MemoryRes)},
		{"internet", state.InternetStatus, len(state.InternetRes)},
		{"dhcp_discover", state.DHCPDiscoverStatus, len(state.DHCPDiscoverRes)},
		{"pppoe_discover", state.PPPoEDiscoverStatus, len(state.PPPoEDiscoverRes)},
		{"arp", state.ARPStatus, len(state.ARPRes)},
		{"cloud_link", state.CloudLinkStatus, len(state.CloudLinkRes)},
	}
	for _, sec := range sections {
		if sec.status == "" {
			t.Errorf("检测项 %s status 不应为空", sec.name)
		}
		t.Logf("  %s: status=%q res=%d 条", sec.name, sec.status, sec.res)
		for i, item := range state.InternetRes {
			if sec.name != "internet" {
				continue
			}
			m, err := item.Map()
			if err != nil {
				t.Errorf("internet 体检项解码失败: %v", err)
				continue
			}
			t.Logf("    internet[%d]: %v", i, m)
		}
	}
}
