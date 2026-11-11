package behavior

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveACL 真机只读验证访问控制族四 func 的 show。
// MACAdd/MACEdit/MACDelete/L7Add/L7Edit/L7Delete/QQAdd/QQEdit/QQDelete/
// L2RouteSet 为写动作，严禁对真机调用。
func TestLiveACL(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).ACL

	// MAC 访问控制（真机实测 7 行，含 comment:1/expires 特殊字段）。
	page, err := svc.MACList(ctx, nil)
	if err != nil {
		t.Fatalf("acl_mac.show 失败: %v", err)
	}
	t.Logf("MAC 访问控制: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d mac=%s comment=%q enabled=%s week=%s time=%s expires=%d",
			r.ID, r.MAC, r.Comment, r.Enabled, r.Week, r.Time, r.Expires)
	}
	if len(page.Rows) == 0 {
		t.Error("acl_mac 列表不应为空（探测真机实测 7 行）")
	} else {
		r := page.Rows[0]
		if r.MAC == "" || r.Enabled == "" {
			t.Errorf("MAC 规则行字段与探测形态不符: %+v", r)
		}
	}

	// 七层协议控制 + 选择器。
	l7, err := svc.L7List(ctx, nil)
	if err != nil {
		t.Fatalf("acl_l7.show 失败: %v", err)
	}
	t.Logf("七层协议控制: total=%d 本页行数=%d", l7.Total, len(l7.Rows))

	sel, err := svc.L7Selectors(ctx)
	if err != nil {
		t.Fatalf("acl_l7.show TYPE=all 失败: %v", err)
	}
	t.Logf("L7 可选 IP 分组 %d 个: %v", len(sel.IPGroup), sel.IPGroup)
	if len(sel.IPGroup) == 0 {
		t.Error("acl_l7 TYPE=all 的 ipgroup 不应为空（探测真机实测 7 组）")
	}

	// 二层访问控制（真机实测单行）。
	cfg, err := svc.L2RouteGet(ctx)
	if err != nil {
		t.Fatalf("acl_l2route.show 失败: %v", err)
	}
	if cfg != nil {
		t.Logf("二层访问控制: id=%d nol2rt=%d 例外IP=%q times=%s ttl=%d",
			cfg.ID, cfg.Nol2rt, cfg.Nol2rtIP, cfg.Times, cfg.TTLNum)
	} else {
		t.Log("acl_l2route.show 返回空配置行")
	}

	sel2, err := svc.L2RouteSelectors(ctx)
	if err != nil {
		t.Fatalf("acl_l2route.show TYPE=all 失败: %v", err)
	}
	t.Logf("L2Route 可选 IP 分组 %d 个", len(sel2.IPGroup))

	// QQ 准入控制 + 选择器。
	qq, err := svc.QQList(ctx, nil)
	if err != nil {
		t.Fatalf("acl_qq.show 失败: %v", err)
	}
	t.Logf("QQ 准入控制: total=%d 本页行数=%d", qq.Total, len(qq.Rows))

	qsel, err := svc.QQSelectors(ctx)
	if err != nil {
		t.Fatalf("acl_qq.show TYPE=all 失败: %v", err)
	}
	t.Logf("QQ 可选 IP 分组 %d 个", len(qsel.IPGroup))
}
