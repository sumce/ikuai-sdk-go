package system

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
	if svc.Basic == nil || svc.VRRP == nil || svc.Disk == nil || svc.File == nil ||
		svc.Raid == nil || svc.DingTalk == nil || svc.WeCom == nil ||
		svc.Register == nil || svc.ALG == nil ||
		svc.IRQBalance == nil || svc.Sysctl == nil || svc.WebUser == nil ||
		svc.RemoteControl == nil || svc.Upgrade == nil || svc.Backup == nil ||
		svc.HardwareInfo == nil || svc.Reboots == nil {
		t.Fatal("New 应聚合全部 17 个子服务")
	}
}

// TestServiceRaw 验证逃生口 Raw 原样透传 func_name/action/param。
func TestServiceRaw(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("hardwareinfo", "show", map[string]any{
		"hardwareinfo": map[string]any{"cpumodel": "test"},
	})
	svc := New(ts.Client())
	resp, err := svc.Raw(context.Background(), "hardwareinfo", "show", map[string]any{})
	if err != nil {
		t.Fatalf("Raw 失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("Raw 应成功, got %+v", resp)
	}
	rec := ts.CallsFor("hardwareinfo")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestTypedShowParam 单元验证 typedShowParam 的 TYPE 归一逻辑。
func TestTypedShowParam(t *testing.T) {
	// opts=nil → 仅默认 TYPE。
	p := typedShowParam(nil, "data", nil)
	wantParam(t, p, map[string]any{"TYPE": "data"})
	// opts 带分页 → BuildParam 产生 total,data，被归一为默认 TYPE。
	p = typedShowParam(&ikuai.ListOptions{Offset: 20, Limit: 20}, "data", nil)
	wantParam(t, p, map[string]any{"TYPE": "data", "limit": "20,20"})
	// defaults 补缺失键且不覆盖归一后的 TYPE。
	p = typedShowParam(nil, "hardirq", map[string]any{"limit": "0,20"})
	wantParam(t, p, map[string]any{"TYPE": "hardirq", "limit": "0,20"})
	// 用户显式 TYPE 优先。
	p = typedShowParam(&ikuai.ListOptions{Type: "softirq"}, "hardirq", nil)
	wantParam(t, p, map[string]any{"TYPE": "softirq"})
}
