package flowcontrol

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// domainSample 构造真机探测形态的 stream_domain.show 响应 Data
// （探测真机列表为空，行字段取自开源 3.x 实现，见 DomainRule 注释）。
func domainSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "comment": "旁路由",
			"interface": "adsl03", "domain": "*.bypass.local",
			"src_addr": "192.168.2.0/24", "week": "1234567",
			"time": "00:00-23:59",
		}},
	}
}

// TestDomainList 覆盖 stream_domain.show 分页列表的成功路径与参数断言。
func TestDomainList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_domain", "show", domainSample())
	svc := New(ts.Client()).Domain

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.Domain != "*.bypass.local" ||
		r.Interface != "adsl03" || r.SrcAddr != "192.168.2.0/24" {
		t.Fatalf("规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_domain")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestDomainSelectors 覆盖 TYPE=all 表单选择器路径（真机实测区块）。
func TestDomainSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_domain", "show", map[string]any{
		"total": 0, "data": []any{},
		"interface": []any{
			[]any{"adsl01", "移动私网"},
			[]any{"adsl03", "联通外网"},
		},
		"ipgroup":  []any{"公共无线", "前台"},
		"macgroup": []any{},
		"dtgroup":  []any{},
	})
	sel, err := New(ts.Client()).Domain.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 2 || sel.Interface[1][1] != "联通外网" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.IPGroup) != 2 || sel.MACGroup == nil || sel.DTGroup == nil {
		t.Fatalf("其余区块解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_domain")[0]),
		map[string]any{"TYPE": SelectorsTypeAll})
}

// TestDomainWrites 覆盖 add/edit/del 的参数形态（写动作，离线）。
func TestDomainWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("stream_domain", "add", map[string]any{})
	ts.SetResponse("stream_domain", "edit", map[string]any{})
	ts.SetResponse("stream_domain", "del", map[string]any{})
	svc := New(ts.Client()).Domain
	ctx := context.Background()

	rowID, err := svc.Add(ctx, DomainRule{Enabled: EnabledYes,
		Interface: "wan1", Domain: "example.com", Comment: "测试"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("stream_domain")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "interface", "domain", "comment"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.Edit(ctx, DomainRule{ID: 4, Enabled: EnabledNo}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if rec = ts.CallsFor("stream_domain")[1]; rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(4) {
		t.Fatalf("edit 参数应携带 id=4: %v", p)
	}

	if err := svc.Delete(ctx, 4, 5); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("stream_domain")[2]),
		map[string]any{"ids": []any{4, 5}})
}

// TestDomainError 覆盖固件错误路径。
func TestDomainError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("stream_domain", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "stream_domain", Action: action}
	})
	svc := New(ts.Client()).Domain
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Selectors(ctx); err == nil {
		t.Fatal("TYPE=all show 错误应上抛")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}
