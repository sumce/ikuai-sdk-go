package logcenter

import (
	"context"
	"encoding/json"
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

// wantParam 断言 p 中与 want 相同的键逐键相等（值经 JSON 规范化比较，
// 兼容数字类型差异），并断言参数中不含 want 之外的键（len 相等），
// 避免实现偷偷多发参数。
func wantParam(t *testing.T, p, want map[string]any) {
	t.Helper()
	for k, v := range want {
		got, ok := p[k]
		if !ok {
			t.Fatalf("param 缺少键 %q: %v", k, p)
		}
		if !jsonEqual(got, v) {
			t.Fatalf("param[%q] = %v, 期望 %v", k, got, v)
		}
	}
	if len(p) != len(want) {
		t.Fatalf("param 键数 = %d (%v), 期望 %d (%v)", len(p), p, len(want), want)
	}
}

// jsonEqual 比较两个值 JSON 序列化后是否等价（float64(1) == int64(1)）。
func jsonEqual(a, b any) bool {
	ab, aerr := json.Marshal(a)
	bb, berr := json.Marshal(b)
	if aerr != nil || berr != nil {
		return false
	}
	return string(ab) == string(bb)
}

// TestNewWiring 验证 New 聚合的子服务均已挂接。
func TestNewWiring(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client())
	if svc.SystemLog == nil || svc.Messages == nil || svc.RTCollect == nil {
		t.Fatal("New 应聚合全部 3 个子服务")
	}
}

// TestServiceRaw 验证逃生口 Raw 原样透传 func_name/action/param。
func TestServiceRaw(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("syslog-arp", "show", map[string]any{"total": 0, "data": []any{}})
	resp, err := New(ts.Client()).Raw(context.Background(), "syslog-arp", "show", map[string]any{})
	if err != nil {
		t.Fatalf("Raw 失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("Raw 应成功, got %+v", resp)
	}
	rec := ts.CallsFor("syslog-arp")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestSyslogFuncConstants 9 张 syslog 表的 func_name 与固件 menu.json 一致。
func TestSyslogFuncConstants(t *testing.T) {
	want := map[SyslogFunc]string{
		SyslogPPPAuth:  "syslog-pppauth",
		SyslogARP:      "syslog-arp",
		SyslogAPAction: "syslog-apaction",
		SyslogDHCPd:    "syslog-dhcpd",
		SyslogDDNS:     "syslog-ddns",
		SyslogWANPPPoE: "syslog-wanpppoe",
		SyslogNotice:   "syslog-notice",
		SyslogSysEvent: "syslog-sysevent",
		SyslogWebAdmin: "syslog-webadmin",
	}
	if len(want) != 9 {
		t.Fatalf("syslog func 常量应共 9 个, got %d", len(want))
	}
	for got, exp := range want {
		if string(got) != exp {
			t.Fatalf("SyslogFunc 常量 %q != %q", got, exp)
		}
	}
}

// TestTypeConstants show 动作 TYPE 选择器常量与固件值一致。
func TestTypeConstants(t *testing.T) {
	if SyslogTypeAll != "all" {
		t.Fatalf("SyslogTypeAll = %q, 期望 all", SyslogTypeAll)
	}
	if MessagesTypeTotal != "total" {
		t.Fatalf("MessagesTypeTotal = %q, 期望 total", MessagesTypeTotal)
	}
	if RTCollectTypeAll != "all" {
		t.Fatalf("RTCollectTypeAll = %q, 期望 all", RTCollectTypeAll)
	}
	if RTCollectTypeWarningNotify != "warning_notify" {
		t.Fatalf("RTCollectTypeWarningNotify = %q, 期望 warning_notify", RTCollectTypeWarningNotify)
	}
}

// TestPageShowParam 验证标准分页参数的构造逻辑（与 behavior 包同形态）。
func TestPageShowParam(t *testing.T) {
	if got := pageShowParam(nil, nil); len(got) != 2 ||
		got["TYPE"] != "total,data" || got["limit"] != "0,20" {
		t.Fatalf("nil opts 应得默认分页 {TYPE:total,data, limit:0,20}, got %v", got)
	}
	got := pageShowParam(&ikuai.ListOptions{Offset: 40, Limit: 10,
		Keywords: "DHCPACK", Finds: "msgtype,event"}, nil)
	if got["limit"] != "40,10" || got["KEYWORDS"] != "DHCPACK" || got["FINDS"] != "msgtype,event" {
		t.Fatalf("opts 键应透传: %v", got)
	}
	// defaults 只补缺失键，用户显式值优先。
	got = pageShowParam(&ikuai.ListOptions{Limit: 5, Type: "data"},
		map[string]any{"TYPE": "all"})
	if got["TYPE"] != "data" {
		t.Fatalf("opts.Type 显式值应优先: %v", got)
	}
}
