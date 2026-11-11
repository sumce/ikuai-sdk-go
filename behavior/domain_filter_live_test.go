package behavior

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDomainFilter 真机只读验证域名过滤两 func 的 show。
// BlacklistAdd/BlacklistEdit/BlacklistDelete/GroupAdd/GroupEdit/GroupDelete
// 为写动作，严禁对真机调用。
func TestLiveDomainFilter(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).DomainFilter

	// 域名黑名单 + 选择器。
	page, err := svc.BlacklistList(ctx, nil)
	if err != nil {
		t.Fatalf("domain_blacklist.show 失败: %v", err)
	}
	t.Logf("域名黑名单: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d domain_group=%q ipaddr=%q weekdays=%s enabled=%s",
			r.ID, r.DomainGroup, r.IPAddr, r.Weekdays, r.Enabled)
	}

	sel, err := svc.BlacklistSelectors(ctx)
	if err != nil {
		t.Fatalf("domain_blacklist.show TYPE=all 失败: %v", err)
	}
	t.Logf("域名黑名单可选 IP 分组 %d 个: %v", len(sel.IPGroup), sel.IPGroup)
	if len(sel.IPGroup) == 0 {
		t.Error("domain_blacklist TYPE=all 的 ipgroup 不应为空（探测真机实测 7 组）")
	}

	// 域名分组（网址库）+ 系统网址类型树。
	groups, err := svc.GroupList(ctx, nil)
	if err != nil {
		t.Fatalf("domain_group.show 失败: %v", err)
	}
	t.Logf("自定义域名分组: total=%d 本页行数=%d", groups.Total, len(groups.Rows))
	for i, r := range groups.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d name=%q domain=%q comment=%q", r.ID, r.Name, r.Domain, r.Comment)
	}

	gsel, err := svc.GroupSelectors(ctx)
	if err != nil {
		t.Fatalf("domain_group.show TYPE=all 失败: %v", err)
	}
	if len(gsel.URLType) != 9 {
		t.Errorf("url_type 应为 9 大类（探测真机实测）, got %d", len(gsel.URLType))
	}
	if len(gsel.URLType) > 0 {
		first := gsel.URLType[0]
		t.Logf("网址类型树: %d 大类, 首类 %q 有 %d 子类",
			len(gsel.URLType), first.Name, len(first.Children))
	}
}
