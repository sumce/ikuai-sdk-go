package security

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveARPList 真机只读验证 ARP 表（arp.show；真机实测 21 行）。
func TestLiveARPList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).ARP.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("arp.show 失败: %v", err)
	}
	t.Logf("ARP 表 total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, e := range page.Rows {
		if i >= 5 {
			break
		}
		if e.IPAddr == "" || e.MAC == "" || e.Interface == "" {
			t.Errorf("第 %d 行关键字段为空: %+v", i, e)
		}
		t.Logf("  行: id=%d ip=%s mac=%s iface=%s bind_type=%s bind_state=%s comment=%q",
			e.ID, e.IPAddr, e.MAC, e.Interface, e.BindType, e.BindState, e.Comment)
		if e.BindType != ARPBindTypeNormal && e.BindType != ARPBindTypeUnique {
			t.Errorf("第 %d 行 bind_type 非法: %q", i, e.BindType)
		}
		if e.BindState != ARPBindStateUnbound && e.BindState != ARPBindStateBound {
			t.Errorf("第 %d 行 bind_state 非法: %q", i, e.BindState)
		}
	}
}

// TestLiveARPSearch 真机只读验证 ARP 关键字检索（arp.show FINDS/KEYWORDS）。
func TestLiveARPSearch(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).ARP.Search(context.Background(), "172.22", nil)
	if err != nil {
		t.Fatalf("arp.show KEYWORDS 检索失败: %v", err)
	}
	t.Logf("关键字 172.22 命中 total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, e := range page.Rows {
		if i >= 5 {
			break
		}
		t.Logf("  行: ip=%s mac=%s iface=%s", e.IPAddr, e.MAC, e.Interface)
	}
}

// TestLiveARPSelectors 真机只读验证 ARP 绑定页选择器与全局开关
// （arp.show TYPE=all）。
func TestLiveARPSelectors(t *testing.T) {
	c := livetest.Client(t)
	sel, err := New(c).ARP.Selectors(context.Background())
	if err != nil {
		t.Fatalf("arp.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 || len(sel.LANInterface) == 0 {
		t.Error("接口选择器不应为空")
	}
	t.Logf("interface=%d lan_interface=%d arp_filter=%d arp_filter_iface=%s dhcpd_arp=%d",
		len(sel.Interface), len(sel.LANInterface), sel.ARPFilter, sel.ARPFilterIface, sel.DHCPDARP)
}

// TestLiveARPPaging 真机只读验证 ARP 表第二页（total 充足时）。
func TestLiveARPPaging(t *testing.T) {
	c := livetest.Client(t)
	p2, err := New(c).ARP.List(context.Background(), &ikuai.ListOptions{Offset: 20, Limit: 10})
	if err != nil {
		t.Fatalf("arp.show 第二页失败: %v", err)
	}
	t.Logf("第二页行数=%d", len(p2.Rows))
}
