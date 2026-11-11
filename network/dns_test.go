package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// dnsSample 构造真机形态的 dns.show TYPE=dns_config 响应 Data。
func dnsSample() map[string]any {
	return map[string]any{"data": []any{map[string]any{
		"id": 1, "enabled": "yes", "dns1": "223.5.5.5", "dns2": "114.114.114.114",
		"cachemode": 0, "cache_ttl": 300, "forbid_dns_4a": 0, "proxy_force": 0,
		"proxy_force_dns": "", "network": 0, "defense": 0,
		"query": "", "query_head_ip": "", "query_args_ip": "",
	}}}
}

// TestDNSGetSet 覆盖 dns.show TYPE=dns_config 与 dns.set（写动作仅离线）。
func TestDNSGetSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dns", "show", dnsSample())
	ts.SetResponse("dns", "set", map[string]any{})
	svc := New(ts.Client()).DNS

	cfg, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.DNS1 != "223.5.5.5" ||
		cfg.CacheTTL != 300 || !cfg.IsEnabled() {
		t.Fatalf("DNS 配置解码不符: %+v", cfg)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dns")[0]),
		map[string]any{"TYPE": DNSTypeConfig})

	if err := svc.Set(context.Background(), *cfg); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("dns")[1]
	if rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
	if p := paramOf(t, rec); p["dns1"] != "223.5.5.5" {
		t.Fatalf("set 参数不符: %v", p)
	}
}

// TestDNSReplaceCRUD 覆盖 dns_replace 的 show/add/edit/del。
func TestDNSReplaceCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dns_replace", "show", map[string]any{"total": 0, "data": []any{}})
	ts.SetResponse("dns_replace", "add", map[string]any{})
	ts.SetResponse("dns_replace", "edit", map[string]any{})
	ts.SetResponse("dns_replace", "del", map[string]any{})
	svc := New(ts.Client()).DNS

	page, err := svc.ReplaceList(context.Background(), nil)
	if err != nil {
		t.Fatalf("ReplaceList 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 0 {
		t.Fatalf("列表解码不符: %+v", page)
	}
	if _, err := svc.ReplaceAdd(context.Background(), DNSReplace{Interface: "adsl01", DNS1: "223.5.5.5"}); err != nil {
		t.Fatalf("ReplaceAdd 失败: %v", err)
	}
	if err := svc.ReplaceEdit(context.Background(), DNSReplace{ID: 1}); err != nil {
		t.Fatalf("ReplaceEdit 失败: %v", err)
	}
	if err := svc.ReplaceDelete(context.Background(), 1); err != nil {
		t.Fatalf("ReplaceDelete 失败: %v", err)
	}
	recs := ts.CallsFor("dns_replace")
	if len(recs) != 4 {
		t.Fatalf("应有 4 次调用, got %d", len(recs))
	}
	for i, want := range []string{"show", "add", "edit", "del"} {
		if recs[i].Action != want {
			t.Fatalf("第 %d 次调用 action = %q, 期望 %q", i+1, recs[i].Action, want)
		}
		if recs[i].FuncName != "dns_replace" {
			t.Fatalf("func_name = %q, 期望 dns_replace", recs[i].FuncName)
		}
	}
}

// TestDNSError 覆盖固件错误路径。
func TestDNSError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dns", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "dns", Action: action}
	})
	if _, err := New(ts.Client()).DNS.Get(context.Background()); err == nil {
		t.Fatal("固件错误应上抛")
	}
}
