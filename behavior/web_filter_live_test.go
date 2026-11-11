package behavior

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWebFilter 真机只读验证网页过滤四 func 的 show。
// BlackAdd/BlackEdit/BlackDelete/RedirectAdd/RedirectEdit/RedirectDelete/
// KeywordsAdd/KeywordsEdit/KeywordsDelete/ReplaceAdd/ReplaceEdit/
// ReplaceDelete 为写动作，严禁对真机调用。
func TestLiveWebFilter(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).WebFilter

	// 网页黑白名单 + 选择器。
	page, err := svc.BlackList(ctx, nil)
	if err != nil {
		t.Fatalf("url_black.show 失败: %v", err)
	}
	t.Logf("网页黑白名单: total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 3 {
			break
		}
		t.Logf("  行: id=%d mode=%d domain=%q ip=%q enabled=%s",
			r.ID, r.Mode, r.Domain, r.IPAddr, r.Enabled)
	}

	sel, err := svc.BlackSelectors(ctx)
	if err != nil {
		t.Fatalf("url_black.show TYPE=all 失败: %v", err)
	}
	t.Logf("黑白名单可选 IP 分组 %d 个, url_white_refer=%d",
		len(sel.IPGroup), sel.URLWhiteRefer)
	if len(sel.IPGroup) == 0 {
		t.Error("url_black TYPE=all 的 ipgroup 不应为空（探测真机实测 7 组）")
	}

	// 网页重定向 + 选择器。
	rd, err := svc.RedirectList(ctx, nil)
	if err != nil {
		t.Fatalf("url_redirect.show 失败: %v", err)
	}
	t.Logf("网页重定向: total=%d 本页行数=%d", rd.Total, len(rd.Rows))

	rsel, err := svc.RedirectSelectors(ctx)
	if err != nil {
		t.Fatalf("url_redirect.show TYPE=all 失败: %v", err)
	}
	t.Logf("重定向可选 IP 分组 %d 个", len(rsel.IPGroup))

	// 关键字替换 + 选择器。
	kw, err := svc.KeywordsList(ctx, nil)
	if err != nil {
		t.Fatalf("url_keywords.show 失败: %v", err)
	}
	t.Logf("关键字替换: total=%d 本页行数=%d", kw.Total, len(kw.Rows))

	ksel, err := svc.KeywordsSelectors(ctx)
	if err != nil {
		t.Fatalf("url_keywords.show TYPE=all 失败: %v", err)
	}
	t.Logf("关键字替换可选 IP 分组 %d 个", len(ksel.IPGroup))

	// 参数替换 + 选择器。
	rp, err := svc.ReplaceList(ctx, nil)
	if err != nil {
		t.Fatalf("url_replace.show 失败: %v", err)
	}
	t.Logf("参数替换: total=%d 本页行数=%d", rp.Total, len(rp.Rows))

	psel, err := svc.ReplaceSelectors(ctx)
	if err != nil {
		t.Fatalf("url_replace.show TYPE=all 失败: %v", err)
	}
	t.Logf("参数替换可选 IP 分组 %d 个", len(psel.IPGroup))
}
