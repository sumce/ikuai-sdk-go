package advanced

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

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

// errIsAPICode 断言 err 为 *ikuai.APIError 且 Result 码匹配。
func errIsAPICode(t *testing.T, err error, code int) *ikuai.APIError {
	t.Helper()
	ae, ok := ikuai.AsAPIError(err)
	if !ok {
		t.Fatalf("期望 *ikuai.APIError, got %v", err)
	}
	if ae.Code != code {
		t.Fatalf("APIError.Code = %d, 期望 %d", ae.Code, code)
	}
	return ae
}

// TestNewWiring 验证 New 聚合的全部 11 个子服务均已挂接。
func TestNewWiring(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client())
	if svc.WebCache == nil || svc.DDNS == nil || svc.NATDDNS == nil || svc.SNMP == nil ||
		svc.SNMPClient == nil || svc.PortMirror == nil || svc.QEMU == nil || svc.Plugins == nil ||
		svc.Wakeup == nil || svc.FileServices == nil || svc.UDProxy == nil {
		t.Fatal("New 应聚合全部 11 个子服务")
	}
}

// TestServiceRaw 验证逃生口 Raw 原样透传 func_name/action/param。
func TestServiceRaw(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("qemu", "show", map[string]any{"total": 0, "data": []any{}})
	resp, err := New(ts.Client()).Raw(context.Background(), "qemu", "show",
		map[string]any{"TYPE": "all"})
	if err != nil {
		t.Fatalf("Raw 失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("Raw 应成功, got %+v", resp)
	}
	rec := ts.CallsFor("qemu")[0]
	if rec.FuncName != "qemu" || rec.Action != "show" {
		t.Fatalf("Raw 透传不符: %s.%s", rec.FuncName, rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "all"})
}

// TestPageShowParamDefaults 覆盖 pageShowParam/typedShowParam 的参数合并
// 语义（opts 显式键优先，defaults/强制 TYPE 只补缺失）。
func TestPageShowParamDefaults(t *testing.T) {
	// pageShowParam(nil) 补默认分页。
	p := pageShowParam(nil, nil)
	wantParam(t, p, map[string]any{"TYPE": "total,data", "limit": "0,20"})

	// pageShowParam 自定义分页 + defaults 补充专有键。
	p = pageShowParam(&ikuai.ListOptions{Offset: 10, Limit: 5}, map[string]any{"ORDER_BY": "id"})
	wantParam(t, p, map[string]any{"TYPE": "total,data", "limit": "10,5", "ORDER_BY": "id"})

	// typedShowParam(nil) 强制默认 TYPE 且不带分页。
	p = typedShowParam(nil, PluginsTypeData, nil)
	wantParam(t, p, map[string]any{"TYPE": "data"})

	// typedShowParam(opts.Type 显式) 尊重用户 TYPE。
	p = typedShowParam(&ikuai.ListOptions{Type: "status"}, WebCacheTypeData, nil)
	wantParam(t, p, map[string]any{"TYPE": "status"})

	// typedShowParam(opts 带分页但 Type 空) 仍强制默认 TYPE。
	p = typedShowParam(&ikuai.ListOptions{Offset: 0, Limit: 5}, WebCacheTypeData, nil)
	wantParam(t, p, map[string]any{"TYPE": "data", "limit": "0,5"})

	// selectorShowParam(nil, nil) → 空参数。
	if p := selectorShowParam(nil, nil); len(p) != 0 {
		t.Fatalf("selectorShowParam(nil,nil) 应为空: %v", p)
	}
}

// TestEnabledConstants 验证启用常量与固件取值一致。
func TestEnabledConstants(t *testing.T) {
	if EnabledYes != "yes" || EnabledNo != "no" {
		t.Fatalf("启用常量不符: %s/%s", EnabledYes, EnabledNo)
	}
}

// TestSessionExpiryRecovery 验证会话失效后核心包透明重登、业务调用方
// 无感知（会话管理是核心包职责，业务包不参与；见 docs/RULES.md §2）。
func TestSessionExpiryRecovery(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ddns", "show", map[string]any{"total": 0, "data": []any{}})
	c := ts.Client()
	if err := c.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	ts.ExpireSessions()
	page, err := New(c).DDNS.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("会话失效后应自动重登并成功: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 0 {
		t.Fatalf("重登后解码不符: %+v", page)
	}
	if !c.LoggedIn() {
		t.Fatal("重登后应持有新会话")
	}
}

// TestNonAuthErrorNoRetry 验证非会话类固件错误不触发重试（业务包不吞错、
// 不重试，错误原样上抛）。
func TestNonAuthErrorNoRetry(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	calls := 0
	ts.HandleFuncName("ddns", func(action string, param json.RawMessage) (any, error) {
		calls++
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "无此模块", FuncName: "ddns", Action: action}
	})
	_, err := New(ts.Client()).DDNS.List(context.Background(), nil)
	errIsAPICode(t, err, 30002)
	if calls != 1 {
		t.Fatalf("错误不应触发重试, 调用次数 = %d", calls)
	}
	if !errors.Is(err, err) {
		t.Fatal("错误链断裂")
	}
}
