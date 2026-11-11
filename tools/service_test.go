package tools

import (
	"context"
	"encoding/json"
	"testing"

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
// 之外的键（len 相等），避免实现偷偷多发参数。数值经 JSON 往返为
// float64，want 侧用 float64 书写。
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
	if svc.Ping == nil || svc.Traceroute == nil || svc.TCPDump == nil ||
		svc.IPerf == nil || svc.Subnet == nil || svc.SpeedTest == nil ||
		svc.SysCheck == nil || svc.Watchdog == nil {
		t.Fatal("New 应聚合全部 8 个子服务")
	}
}

// TestServiceRaw 验证逃生口 Raw 原样透传 func_name/action/param。
func TestServiceRaw(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("watchdog", "show", map[string]any{"data": []any{map[string]any{"watchdog": 1, "watchdog_thresh": 30}}})
	svc := New(ts.Client())
	resp, err := svc.Raw(context.Background(), "watchdog", "show", map[string]any{})
	if err != nil {
		t.Fatalf("Raw 失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("Raw 应成功, got %+v", resp)
	}
	rec := ts.CallsFor("watchdog")[0]
	if rec.FuncName != "watchdog" || rec.Action != "show" {
		t.Fatalf("Raw 透传不符: %s.%s", rec.FuncName, rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}
