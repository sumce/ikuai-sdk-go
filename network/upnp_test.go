package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// upnpSample 构造真机形态的 upnpd.show 响应 Data。
func upnpSample() map[string]any {
	return map[string]any{"data": []any{map[string]any{
		"enabled": "no", "exclude_port": "1-1024",
		"lan_ip": "0.0.0.0-255.255.255.255", "interface": "any",
		"check_link": 0, "check_interval": 5, "rst_switch": 0,
		"rst_week": "1234567", "rst_time": "00:00", "id": 1,
	}}}
}

// TestUPnPGetSet 覆盖 upnpd.show（空参数，真机实证）与 upnpd.set（仅离线）。
func TestUPnPGetSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("upnpd", "show", upnpSample())
	ts.SetResponse("upnpd", "set", map[string]any{})
	svc := New(ts.Client()).UPnP
	ctx := context.Background()

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.ExcludePort != "1-1024" ||
		cfg.LANIP != "0.0.0.0-255.255.255.255" || cfg.Interface != "any" ||
		cfg.CheckInterval != 5 || cfg.RstWeek != "1234567" ||
		cfg.RstTime != "00:00" || cfg.IsEnabled() {
		t.Fatalf("UPnP 配置解码不符: %+v", cfg)
	}
	// 真机实证：upnpd.show 不带 TYPE/limit（TYPE=total,data 回 30006）。
	wantParam(t, paramOf(t, ts.CallsFor("upnpd")[0]), map[string]any{})

	if err := svc.Set(ctx, *cfg); err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("upnpd")[1]
	if rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
	if p := paramOf(t, rec); p["exclude_port"] != "1-1024" {
		t.Fatalf("set 参数不符: %v", p)
	}
}

// TestUPnPLeaseList 覆盖 upnpd_leases.show（TYPE=total,data，真机实证空表；
// 行字段待真机验证）。
func TestUPnPLeaseList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("upnpd_leases", "show", map[string]any{"total": 0, "data": []any{}})
	page, err := New(ts.Client()).UPnP.LeaseList(context.Background(), nil)
	if err != nil {
		t.Fatalf("LeaseList 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 0 {
		t.Fatalf("租约列表解码不符: %+v", page)
	}
	rec := ts.CallsFor("upnpd_leases")[0]
	if rec.FuncName != "upnpd_leases" || rec.Action != "show" {
		t.Fatalf("调用记录不符: %s.%s", rec.FuncName, rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestUPnPError 覆盖固件错误路径（模拟真机 TYPE 不接受的错误形态）。
func TestUPnPError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("upnpd", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "upnpd", Action: action}
	})
	_, err := New(ts.Client()).UPnP.Get(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
