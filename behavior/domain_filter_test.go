package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestDomainBlacklistList 覆盖 domain_blacklist.show 分页（行字段空表推导，
// 待真机验证；ipaddr/weekdays 拼写按开源生产实证）。
func TestDomainBlacklistList(t *testing.T) {
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
			ts.SetResponse("domain_blacklist", "show", map[string]any{
				"total": 1,
				"data": []any{
					map[string]any{"id": 1, "enabled": "yes",
						"domain_group": "休闲娱乐-游戏网站", "ipaddr": "监控",
						"time": "00:00-23:59", "weekdays": "1234567", "comment": "禁游戏"},
				},
			})
			page, err := New(ts.Client()).DomainFilter.BlacklistList(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("BlacklistList 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || !r.IsEnabled() || r.DomainGroup != "休闲娱乐-游戏网站" ||
				r.IPAddr != "监控" || r.Weekdays != "1234567" || r.Comment != "禁游戏" {
				t.Fatalf("域名黑名单行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("domain_blacklist")[0]), tt.want)
		})
	}
}

// TestDomainBlacklistSelectors 覆盖 domain_blacklist.show TYPE=all。
func TestDomainBlacklistSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("domain_blacklist", "show", map[string]any{
		"total": 0, "data": []any{},
		"ipgroup":  []any{"监控"},
		"macgroup": []any{}, "dtgroup": []any{},
	})
	sel, err := New(ts.Client()).DomainFilter.BlacklistSelectors(context.Background())
	if err != nil {
		t.Fatalf("BlacklistSelectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 1 || sel.IPGroup[0] != "监控" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("domain_blacklist")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestDomainBlacklistWrites 覆盖 add/edit/del（写动作，离线）。
func TestDomainBlacklistWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("domain_blacklist", "add", map[string]any{})
	ts.SetResponse("domain_blacklist", "edit", map[string]any{})
	ts.SetResponse("domain_blacklist", "del", map[string]any{})
	svc := New(ts.Client()).DomainFilter
	ctx := context.Background()

	rowID, err := svc.BlacklistAdd(ctx, DomainBlacklistRule{Enabled: EnabledYes,
		DomainGroup: "休闲娱乐", IPAddr: "172.22.1.0/24",
		Time: "00:00-23:59", Weekdays: "1234567"})
	if err != nil {
		t.Fatalf("BlacklistAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("domain_blacklist")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "domain_group", "ipaddr", "time", "weekdays"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.BlacklistEdit(ctx, DomainBlacklistRule{ID: 1, Enabled: EnabledNo}); err != nil {
		t.Fatalf("BlacklistEdit 失败: %v", err)
	}
	rec = ts.CallsFor("domain_blacklist")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.BlacklistDelete(ctx, 1); err != nil {
		t.Fatalf("BlacklistDelete 失败: %v", err)
	}
	rec = ts.CallsFor("domain_blacklist")[2]
	wantParam(t, paramOf(t, rec), map[string]any{"id": 1})

	if err := svc.BlacklistDelete(ctx, 1, 2, 3); err != nil {
		t.Fatalf("BlacklistDelete 多 ID 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("domain_blacklist")[3]), map[string]any{"ids": []any{1, 2, 3}})
}

// domainGroupSample 构造真机探测形态的 domain_group.show TYPE=all 响应
// （真机 3.7.24 实测 url_type 树）。
func domainGroupSample() map[string]any {
	return map[string]any{
		"total": 0, "data": []any{},
		"url_type": []any{
			map[string]any{"name": "交通旅游", "id": "交通旅游", "children": []any{
				map[string]any{"name": "旅游网站", "id": "交通旅游-旅游网站"},
				map[string]any{"name": "酒店宾馆", "id": "交通旅游-酒店宾馆"},
			}},
			map[string]any{"name": "休闲娱乐", "id": "休闲娱乐", "children": []any{
				map[string]any{"name": "动漫网站", "id": "休闲娱乐-动漫网站"},
			}},
		},
		"sys_url": []any{},
	}
}

// TestDomainGroupSelectors 覆盖 domain_group.show TYPE=all 的类型树解码
// （真机实测形态）。
func TestDomainGroupSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("domain_group", "show", domainGroupSample())
	sel, err := New(ts.Client()).DomainFilter.GroupSelectors(context.Background())
	if err != nil {
		t.Fatalf("GroupSelectors 失败: %v", err)
	}
	if len(sel.URLType) != 2 {
		t.Fatalf("url_type 解码不符: %d 类", len(sel.URLType))
	}
	first := sel.URLType[0]
	if first.Name != "交通旅游" || first.ID != "交通旅游" || len(first.Children) != 2 {
		t.Fatalf("类型树首节点解码不符: %+v", first)
	}
	if first.Children[0].ID != "交通旅游-旅游网站" || first.Children[1].Name != "酒店宾馆" {
		t.Fatalf("子类型解码不符: %+v", first.Children)
	}
	if len(sel.SysURL) != 0 {
		t.Fatalf("sys_url 应为空数组: %v", sel.SysURL)
	}
	wantParam(t, paramOf(t, ts.CallsFor("domain_group")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestDomainGroupList 覆盖 domain_group.show 分页（行字段空表推导，待真机
// 验证）。
func TestDomainGroupList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("domain_group", "show", map[string]any{
		"total": 1,
		"data": []any{
			map[string]any{"id": 1, "group_name": "公司白名单",
				"domain": "www.corp.com,cms.corp.com", "comment": "办公"},
		},
	})
	page, err := New(ts.Client()).DomainFilter.GroupList(context.Background(), nil)
	if err != nil {
		t.Fatalf("GroupList 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Name != "公司白名单" || r.Domain != "www.corp.com,cms.corp.com" {
		t.Fatalf("域名分组行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("domain_group")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestDomainGroupWrites 覆盖 add/edit/del（写动作，离线）。
func TestDomainGroupWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("domain_group", "add", map[string]any{})
	ts.SetResponse("domain_group", "edit", map[string]any{})
	ts.SetResponse("domain_group", "del", map[string]any{})
	svc := New(ts.Client()).DomainFilter
	ctx := context.Background()

	rowID, err := svc.GroupAdd(ctx, DomainGroupRow{Name: "自定义组", Domain: "www.a.com"})
	if err != nil {
		t.Fatalf("GroupAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("domain_group")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if _, ok := p["group_name"]; !ok {
		t.Fatalf("add 参数缺少 group_name: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应含 id: %v", p)
	}

	if err := svc.GroupEdit(ctx, DomainGroupRow{ID: 1, Name: "改名组"}); err != nil {
		t.Fatalf("GroupEdit 失败: %v", err)
	}
	rec = ts.CallsFor("domain_group")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.GroupDelete(ctx, 1, 2); err != nil {
		t.Fatalf("GroupDelete 失败: %v", err)
	}
	rec = ts.CallsFor("domain_group")[2]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{1, 2}})
}

// TestDomainFilterError 覆盖固件错误路径。
func TestDomainFilterError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"domain_blacklist", "domain_group"} {
		fname := fn
		ts.HandleFuncName(fname, func(action string, param json.RawMessage) (any, error) {
			return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: fname, Action: action}
		})
	}
	svc := New(ts.Client()).DomainFilter
	ctx := context.Background()

	if _, err := svc.BlacklistList(ctx, nil); err == nil {
		t.Fatal("domain_blacklist.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.BlacklistSelectors(ctx); err == nil {
		t.Fatal("domain_blacklist.show TYPE=all 错误应上抛")
	}
	if _, err := svc.BlacklistAdd(ctx, DomainBlacklistRule{}); err == nil {
		t.Fatal("domain_blacklist.add 错误应上抛")
	}
	if err := svc.BlacklistDelete(ctx, 1); err == nil {
		t.Fatal("domain_blacklist.del 错误应上抛")
	}
	if _, err := svc.GroupList(ctx, nil); err == nil {
		t.Fatal("domain_group.show 错误应上抛")
	}
	if _, err := svc.GroupSelectors(ctx); err == nil {
		t.Fatal("domain_group.show TYPE=all 错误应上抛")
	}
	if _, err := svc.GroupAdd(ctx, DomainGroupRow{}); err == nil {
		t.Fatal("domain_group.add 错误应上抛")
	}
	if err := svc.GroupDelete(ctx, 1); err == nil {
		t.Fatal("domain_group.del 错误应上抛")
	}
}
