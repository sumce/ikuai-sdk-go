package ac

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// errIsInvalidParam 断言 err 链上携带 ikuai.ErrInvalidParam。
func errIsInvalidParam(err error) bool {
	return errors.Is(err, ikuai.ErrInvalidParam)
}

// paramOf 解析一条调用记录的 param 为通用 map。
func paramOf(t *testing.T, rec testserver.CallRecord) map[string]any {
	t.Helper()
	var p map[string]any
	if err := json.Unmarshal(rec.Param, &p); err != nil {
		t.Fatalf("解析 param 失败: %v (param=%s)", err, rec.Param)
	}
	return p
}

// wantParam 断言 p 中与 want 相同的键逐键相等，并断言参数中不含 want
// 之外的键（len 相等），避免实现偷偷多发参数。
func wantParam(t *testing.T, p, want map[string]any) {
	t.Helper()
	for k, v := range want {
		got, ok := p[k]
		if !ok {
			t.Fatalf("param 缺少键 %q: %v", k, p)
		}
		if got != v {
			t.Fatalf("param[%q] = %v, 期望 %v", k, got, v)
		}
	}
	if len(p) != len(want) {
		t.Fatalf("param 键数 = %d (%v), 期望 %d (%v)", len(p), p, len(want), want)
	}
}

// TestNewWiring 验证 New 聚合的 8 个子服务均已挂接。
func TestNewWiring(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client())
	if svc.Status == nil || svc.Server == nil || svc.Group == nil || svc.Upgrade == nil ||
		svc.Blacklist == nil || svc.Mvlan == nil || svc.OnlineClients == nil || svc.NetOptimize == nil {
		t.Fatal("New 应聚合全部 8 个子服务")
	}
}

// TestServiceRaw 验证逃生口 Raw 原样透传 func_name/action/param。
func TestServiceRaw(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("wls_black", "show", map[string]any{"total": 0, "data": []any{}})
	svc := New(ts.Client())
	resp, err := svc.Raw(context.Background(), "wls_black", "show",
		map[string]any{"TYPE": "total,data"})
	if err != nil {
		t.Fatalf("Raw 失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("Raw 应成功, got %+v", resp)
	}
	rec := ts.CallsFor("wls_black")[0]
	if rec.FuncName != "wls_black" || rec.Action != "show" {
		t.Fatalf("Raw 透传不符: %s.%s", rec.FuncName, rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "total,data"})
}

// TestPageShowParamDefaults 覆盖 pageShowParam/selectorShowParam 的参数
// 合并语义（opts 显式键优先，defaults 只补缺失）。
func TestPageShowParamDefaults(t *testing.T) {
	// opts=nil 时 pageShowParam 补默认分页。
	p := pageShowParam(nil, map[string]any{"ORDER_BY": "id"})
	if p["TYPE"] != "total,data" || p["limit"] != "0,20" || p["ORDER_BY"] != "id" {
		t.Fatalf("pageShowParam(nil) 不符: %v", p)
	}
	// opts 显式 Type 覆盖默认。
	p = pageShowParam(&ikuai.ListOptions{Type: "data"}, map[string]any{"TYPE": "total,data"})
	if p["TYPE"] != "data" {
		t.Fatalf("opts.Type 应优先: %v", p)
	}
	// selectorShowParam(nil) 不带分页。
	p = selectorShowParam(nil, map[string]any{"TYPE": ACStatusTypeAll})
	wantParam(t, p, map[string]any{"TYPE": ACStatusTypeAll})
	// selectorShowParam(nil, nil) → 空参数。
	if p := selectorShowParam(nil, nil); len(p) != 0 {
		t.Fatalf("selectorShowParam(nil,nil) 应为空: %v", p)
	}
}
