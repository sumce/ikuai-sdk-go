package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// ddnsSample 构造 ddns.show 的分页响应 Data（真机 3.7.24 探测实证，
// 18 字段行样例取自真机并脱敏）。
func ddnsSample() map[string]any {
	return map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{
				"passwd": "secret", "cf_token": 0, "account": "", "line": "",
				"interface": "adsl07", "result": "成功", "ipaddress": "203.0.113.10",
				"id": 1, "enabled": "yes", "type": "A", "mode": 0, "location": 0,
				"mac": "", "duid": "", "server": "aliyun.com", "top_domain": "example.com",
				"username": "LTAIxxxx", "domain": "ddns.example.com",
			},
			map[string]any{
				"passwd": "cfut_token", "cf_token": 1, "account": "token-id", "line": "wan1",
				"interface": "adsl03", "result": "成功", "ipaddress": "203.0.113.11",
				"id": 2, "enabled": "no", "type": "AAAA", "mode": 1, "location": 1,
				"mac": "aa:bb:cc:dd:ee:ff", "duid": "00030001...", "server": "cloudflare.com",
				"top_domain": "example.org", "username": "", "domain": "03.adsl.example.org",
			},
		},
	}
}

// TestDDNSList 覆盖 ddns.show 的成功路径与参数断言（真机实证
// TYPE=total,data 形态）。
func TestDDNSList(t *testing.T) {
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
			name: "自定义分页+关键字",
			opts: &ikuai.ListOptions{Offset: 5, Limit: 10, Keywords: "aoxea",
				Finds: "domain,top_domain"},
			want: map[string]any{"TYPE": "total,data", "limit": "5,10",
				"KEYWORDS": "aoxea", "FINDS": "domain,top_domain"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ddns", "show", ddnsSample())
			page, err := New(ts.Client()).DDNS.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 2 || len(page.Rows) != 2 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 1 || !r.IsEnabled() || r.Server != "aliyun.com" ||
				r.Username != "LTAIxxxx" || r.Passwd != "secret" || r.CFToken != 0 ||
				r.Interface != "adsl07" || r.Line != "" || r.Type != "A" ||
				r.Mode != 0 || r.Location != DDNSLocationInterfaceIP ||
				r.MAC != "" || r.DUID != "" || r.TopDomain != "example.com" ||
				r.Domain != "ddns.example.com" || r.IPAddress != "203.0.113.10" ||
				r.Result != "成功" || r.Account != "" {
				t.Fatalf("DDNS 行解码不符: %+v", r)
			}
			r2 := page.Rows[1]
			if r2.CFToken != 1 || r2.Type != "AAAA" ||
				r2.Location != DDNSLocationWanIP || r2.IsEnabled() ||
				r2.MAC != "aa:bb:cc:dd:ee:ff" || r2.DUID != "00030001..." ||
				r2.Account != "token-id" || r2.Line != "wan1" || r2.Mode != 1 {
				t.Fatalf("DDNS 行2解码不符: %+v", r2)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ddns")[0]), tt.want)
		})
	}
}

// TestDDNSAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestDDNSAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ddns", "add", map[string]any{})
	ts.SetResponse("ddns", "edit", map[string]any{})
	ts.SetResponse("ddns", "del", map[string]any{})
	svc := New(ts.Client()).DDNS
	ctx := context.Background()

	rec := DDNSRecord{Server: "aliyun.com", Username: "ak", Passwd: "sk",
		Interface: "adsl01", Type: "A", Location: DDNSLocationInterfaceIP,
		TopDomain: "example.com", Domain: "home.example.com", Enabled: EnabledYes}
	rowID, err := svc.Add(ctx, rec)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	rec.ID = 7
	if err := svc.Edit(ctx, rec); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 7); err != nil {
		t.Fatalf("Delete 单个失败: %v", err)
	}
	if err := svc.Delete(ctx, 7, 8, 9); err != nil {
		t.Fatalf("Delete 多个失败: %v", err)
	}

	recs := ts.CallsFor("ddns")
	if len(recs) != 4 {
		t.Fatalf("调用数不符: %d", len(recs))
	}
	if recs[0].Action != "add" || recs[1].Action != "edit" || recs[2].Action != "del" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action)
	}
	p := paramOf(t, recs[0])
	if p["server"] != "aliyun.com" || p["username"] != "ak" || p["enabled"] != "yes" || p["location"] != float64(0) {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}
	p = paramOf(t, recs[1])
	if p["id"] != float64(7) || p["domain"] != "home.example.com" {
		t.Fatalf("edit 参数不符: %v", p)
	}
	wantParam(t, paramOf(t, recs[2]), map[string]any{"id": float64(7)})
	if p := paramOf(t, recs[3]); len(p["ids"].([]any)) != 3 {
		t.Fatalf("del 多 ID 参数不符: %v", p)
	}
}

// TestDDNSError 覆盖固件错误路径。
func TestDDNSError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ddns", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "无此模块", FuncName: "ddns", Action: action}
	})
	_, err := New(ts.Client()).DDNS.List(context.Background(), nil)
	errIsAPICode(t, err, 30002)
}

// natDDNSSample 构造 nat_ddns.show 的响应 Data（真机 3.7.24 探测实证，
// 未开通形态）。
func natDDNSSample() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"status": 1, "errinfo": "Function not enabled",
			"username": "", "version": "",
		},
	}
}

// TestNATDDNSGet 覆盖 nat_ddns.show TYPE=data 的成功路径与参数断言
// （真机实证：本 func 不支持 TYPE=total）。
func TestNATDDNSGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("nat_ddns", "show", natDDNSSample())
	st, err := New(ts.Client()).NATDDNS.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if st.Status != 1 || st.ErrInfo != "Function not enabled" ||
		st.Username != "" || st.Version != "" {
		t.Fatalf("状态解码不符: %+v", st)
	}
	rec := ts.CallsFor("nat_ddns")[0]
	if rec.Action != "show" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "data"})
}

// TestNATDDNSError 覆盖固件错误路径。
func TestNATDDNSError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("nat_ddns", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "nat_ddns", Action: action}
	})
	_, err := New(ts.Client()).NATDDNS.Get(context.Background())
	errIsAPICode(t, err, 30006)
}
