package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// webFilterSelectorsSample 构造真机探测形态的 url_black.show TYPE=all 区块
// （url_black 实测含 url_white_refer；url_redirect/url_keywords/url_replace
// 实测含 pattern）。
func webFilterSelectorsSample(withPattern bool) map[string]any {
	data := map[string]any{
		"total": 0, "data": []any{},
		"ipgroup":  []any{"公共无线", "前台"},
		"macgroup": []any{}, "dtgroup": []any{},
		"url_white_refer": 0,
	}
	if withPattern {
		data["pattern"] = []any{"exact", "vague"}
	}
	return data
}

// TestURLBlackList 覆盖 url_black.show 分页（行字段空表推导，待真机验证）。
func TestURLBlackList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
		},
		{
			name: "第二页",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 20},
			want: map[string]any{"TYPE": "total,data", "limit": "20,20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("url_black", "show", map[string]any{
				"total": 1,
				"data": []any{
					map[string]any{"id": 1, "enabled": "yes", "mode": URLBlackModeBlack,
						"domain": "www.example.com", "ip_addr": "172.22.1.0/24",
						"time": "00:00-23:59", "week": "1234567", "comment": "黑名单"},
				},
			})
			page, err := New(ts.Client()).WebFilter.BlackList(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("BlackList 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || !r.IsEnabled() || r.Mode != URLBlackModeBlack ||
				r.Domain != "www.example.com" || r.IPAddr != "172.22.1.0/24" {
				t.Fatalf("黑白名单行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("url_black")[0]), tt.want)
		})
	}
}

// TestURLBlackSelectors 覆盖 url_black.show TYPE=all（真机实测区块）。
func TestURLBlackSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("url_black", "show", webFilterSelectorsSample(false))
	sel, err := New(ts.Client()).WebFilter.BlackSelectors(context.Background())
	if err != nil {
		t.Fatalf("BlackSelectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 2 || sel.IPGroup[0] != "公共无线" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	if sel.URLWhiteRefer != 0 {
		t.Fatalf("URLWhiteRefer = %d, 期望 0", sel.URLWhiteRefer)
	}
	wantParam(t, paramOf(t, ts.CallsFor("url_black")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestURLBlackWrites 覆盖 url_black add/edit/del（写动作，离线）。
func TestURLBlackWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("url_black", "add", map[string]any{})
	ts.SetResponse("url_black", "edit", map[string]any{})
	ts.SetResponse("url_black", "del", map[string]any{})
	svc := New(ts.Client()).WebFilter
	ctx := context.Background()

	rowID, err := svc.BlackAdd(ctx, URLBlackRule{Enabled: EnabledYes,
		Mode: URLBlackModeBlack, Domain: "www.example.com",
		Time: "00:00-23:59", Week: "1234567"})
	if err != nil {
		t.Fatalf("BlackAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("url_black")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "mode", "domain", "time", "week"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.BlackEdit(ctx, URLBlackRule{ID: 1, Mode: URLBlackModeWhite}); err != nil {
		t.Fatalf("BlackEdit 失败: %v", err)
	}
	rec = ts.CallsFor("url_black")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) || p["mode"] != float64(URLBlackModeWhite) {
		t.Fatalf("edit 参数不符: %v", p)
	}

	if err := svc.BlackDelete(ctx, 1, 2); err != nil {
		t.Fatalf("BlackDelete 失败: %v", err)
	}
	rec = ts.CallsFor("url_black")[2]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{1, 2}})
}

// TestURLRedirect 覆盖 url_redirect 的 List/Selectors/Add/Edit/Delete
// （行字段空表推导，待真机验证）。
func TestURLRedirect(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("url_redirect", "show", map[string]any{
		"total": 1,
		"data": []any{
			map[string]any{"id": 1, "enabled": "yes", "comment": "跳转测试",
				"mode": URLMatchVague, "src_url": "www.old.com", "dst_url": "www.new.com",
				"excluded": "ads", "hit_rate": 100, "prio": 1,
				"time": "00:00-23:59", "week": "1234567"},
		},
		"ipgroup": []any{"监控"}, "macgroup": []any{}, "dtgroup": []any{},
		"pattern": []any{"exact", "vague"},
	})
	svc := New(ts.Client()).WebFilter
	ctx := context.Background()

	page, err := svc.RedirectList(ctx, nil)
	if err != nil {
		t.Fatalf("RedirectList 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.Comment != "跳转测试" || r.Mode != URLMatchVague ||
		r.SrcURL != "www.old.com" || r.DstURL != "www.new.com" ||
		r.Excluded != "ads" || r.HitRate != 100 || r.Prio != 1 {
		t.Fatalf("重定向规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("url_redirect")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	sel, err := svc.RedirectSelectors(ctx)
	if err != nil {
		t.Fatalf("RedirectSelectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 1 || sel.IPGroup[0] != "监控" {
		t.Fatalf("分组选择器解码不符: %v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("url_redirect")[1]), map[string]any{"TYPE": SelectorsTypeAll})

	ts.Reset()
	ts.SetResponse("url_redirect", "add", map[string]any{})
	ts.SetResponse("url_redirect", "edit", map[string]any{})
	ts.SetResponse("url_redirect", "del", map[string]any{})

	rowID, err := svc.RedirectAdd(ctx, URLRedirectRule{Enabled: EnabledYes,
		Comment: "跳转", Mode: URLMatchExact, SrcURL: "www.old.com",
		DstURL: "www.new.com", HitRate: 50, Prio: 10})
	if err != nil {
		t.Fatalf("RedirectAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("url_redirect")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "comment", "mode", "src_url", "dst_url",
		"hit_rate", "prio"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.RedirectEdit(ctx, URLRedirectRule{ID: 1, DstURL: "www.v2.com"}); err != nil {
		t.Fatalf("RedirectEdit 失败: %v", err)
	}
	rec = ts.CallsFor("url_redirect")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.RedirectDelete(ctx, 1); err != nil {
		t.Fatalf("RedirectDelete 失败: %v", err)
	}
	rec = ts.CallsFor("url_redirect")[2]
	wantParam(t, paramOf(t, rec), map[string]any{"id": 1})
}

// TestURLKeywords 覆盖 url_keywords 的 List/Add/Edit/Delete（行字段空表
// 推导，待真机验证）。
func TestURLKeywords(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("url_keywords", "show", map[string]any{
		"total": 1,
		"data": []any{
			map[string]any{"id": 1, "enabled": "yes", "comment": "替换测试",
				"mode": URLMatchExact, "src_url": "www.example.com",
				"ori_keyword": "广告", "rep_keyword": "",
				"excluded": "", "hit_rate": 100, "prio": 1,
				"time": "00:00-23:59", "week": "1234567"},
		},
	})
	svc := New(ts.Client()).WebFilter
	ctx := context.Background()

	page, err := svc.KeywordsList(ctx, nil)
	if err != nil {
		t.Fatalf("KeywordsList 失败: %v", err)
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.OriKeyword != "广告" || r.RepKeyword != "" ||
		r.Mode != URLMatchExact || r.HitRate != 100 {
		t.Fatalf("关键字规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("url_keywords")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	// Selectors 真机实测含 pattern 区块。
	ts.Reset()
	ts.SetResponse("url_keywords", "show", webFilterSelectorsSample(true))
	sel, err := svc.KeywordsSelectors(ctx)
	if err != nil {
		t.Fatalf("KeywordsSelectors 失败: %v", err)
	}
	_ = sel
	wantParam(t, paramOf(t, ts.CallsFor("url_keywords")[0]), map[string]any{"TYPE": SelectorsTypeAll})

	ts.Reset()
	ts.SetResponse("url_keywords", "add", map[string]any{})
	ts.SetResponse("url_keywords", "edit", map[string]any{})
	ts.SetResponse("url_keywords", "del", map[string]any{})

	rowID, err := svc.KeywordsAdd(ctx, URLKeywordRule{Enabled: EnabledYes,
		Mode: URLMatchVague, OriKeyword: "广告", RepKeyword: "内容"})
	if err != nil {
		t.Fatalf("KeywordsAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("url_keywords")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if _, ok := p["ori_keyword"]; !ok {
		t.Fatalf("add 参数缺少 ori_keyword: %v", p)
	}

	if err := svc.KeywordsEdit(ctx, URLKeywordRule{ID: 1, RepKeyword: "新内容"}); err != nil {
		t.Fatalf("KeywordsEdit 失败: %v", err)
	}
	rec = ts.CallsFor("url_keywords")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.KeywordsDelete(ctx, 1, 2); err != nil {
		t.Fatalf("KeywordsDelete 失败: %v", err)
	}
	rec = ts.CallsFor("url_keywords")[2]
	wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{1, 2}})
}

// TestURLReplace 覆盖 url_replace 的 List/Add/Edit/Delete（行字段空表
// 推导，待真机验证）。
func TestURLReplace(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("url_replace", "show", map[string]any{
		"total": 1,
		"data": []any{
			map[string]any{"id": 1, "enabled": "yes", "comment": "参数替换",
				"mode": URLMatchExact, "src_url": "www.example.com",
				"param_keyword": "page", "rep_keyword": "2",
				"excluded": "", "hit_rate": 100, "prio": 1,
				"time": "00:00-23:59", "week": "1234567"},
		},
	})
	svc := New(ts.Client()).WebFilter
	ctx := context.Background()

	page, err := svc.ReplaceList(ctx, nil)
	if err != nil {
		t.Fatalf("ReplaceList 失败: %v", err)
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.ParamKeyword != "page" || r.RepKeyword != "2" {
		t.Fatalf("参数替换规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("url_replace")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	ts.Reset()
	ts.SetResponse("url_replace", "show", webFilterSelectorsSample(true))
	if _, err = svc.ReplaceSelectors(ctx); err != nil {
		t.Fatalf("ReplaceSelectors 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("url_replace")[0]), map[string]any{"TYPE": SelectorsTypeAll})

	ts.Reset()
	ts.SetResponse("url_replace", "add", map[string]any{})
	ts.SetResponse("url_replace", "edit", map[string]any{})
	ts.SetResponse("url_replace", "del", map[string]any{})

	rowID, err := svc.ReplaceAdd(ctx, URLReplaceRule{Enabled: EnabledYes,
		Mode: URLMatchVague, ParamKeyword: "page", RepKeyword: "1"})
	if err != nil {
		t.Fatalf("ReplaceAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("url_replace")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	if p := paramOf(t, rec); p["param_keyword"] != "page" {
		t.Fatalf("add 参数不符: %v", p)
	}

	if err := svc.ReplaceEdit(ctx, URLReplaceRule{ID: 1, RepKeyword: "3"}); err != nil {
		t.Fatalf("ReplaceEdit 失败: %v", err)
	}
	rec = ts.CallsFor("url_replace")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.ReplaceDelete(ctx, 1); err != nil {
		t.Fatalf("ReplaceDelete 失败: %v", err)
	}
	rec = ts.CallsFor("url_replace")[2]
	wantParam(t, paramOf(t, rec), map[string]any{"id": 1})
}

// TestURLMatchMode 常量与固件值一致。
func TestURLMatchMode(t *testing.T) {
	if URLMatchExact != "exact" || URLMatchVague != "vague" {
		t.Fatalf("匹配模式常量不符: %q %q", URLMatchExact, URLMatchVague)
	}
	if URLBlackModeBlack != 0 || URLBlackModeWhite != 1 {
		t.Fatalf("黑白名单模式常量不符: %d %d", URLBlackModeBlack, URLBlackModeWhite)
	}
}

// TestWebFilterError 覆盖固件错误路径。
func TestWebFilterError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"url_black", "url_redirect", "url_keywords", "url_replace"} {
		fname := fn
		ts.HandleFuncName(fname, func(action string, param json.RawMessage) (any, error) {
			return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: fname, Action: action}
		})
	}
	svc := New(ts.Client()).WebFilter
	ctx := context.Background()

	if _, err := svc.BlackList(ctx, nil); err == nil {
		t.Fatal("url_black.show 错误应上抛")
	}
	if _, err := svc.BlackSelectors(ctx); err == nil {
		t.Fatal("url_black.show TYPE=all 错误应上抛")
	}
	if _, err := svc.BlackAdd(ctx, URLBlackRule{}); err == nil {
		t.Fatal("url_black.add 错误应上抛")
	}
	if err := svc.BlackDelete(ctx, 1); err == nil {
		t.Fatal("url_black.del 错误应上抛")
	}
	if _, err := svc.RedirectList(ctx, nil); err == nil {
		t.Fatal("url_redirect.show 错误应上抛")
	}
	if _, err := svc.RedirectSelectors(ctx); err == nil {
		t.Fatal("url_redirect.show TYPE=all 错误应上抛")
	}
	if _, err := svc.KeywordsList(ctx, nil); err == nil {
		t.Fatal("url_keywords.show 错误应上抛")
	}
	if _, err := svc.KeywordsSelectors(ctx); err == nil {
		t.Fatal("url_keywords.show TYPE=all 错误应上抛")
	}
	if _, err := svc.ReplaceList(ctx, nil); err == nil {
		t.Fatal("url_replace.show 错误应上抛")
	}
	if _, err := svc.ReplaceSelectors(ctx); err == nil {
		t.Fatal("url_replace.show TYPE=all 错误应上抛")
	}
	if err := svc.ReplaceDelete(ctx, 1); err == nil {
		t.Fatal("url_replace.del 错误应上抛")
	}
}
