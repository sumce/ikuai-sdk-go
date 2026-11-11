package security

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// connLimitSample 构造 conn_limit.show 的分页响应 Data
// （行字段空表推导，待真机验证）。
func connLimitSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "comment": "限连测试",
			"ip_addr": "172.22.2.0/24", "protocol": "tcp",
			"wan_port": "any", "limits": 500,
		}},
	}
}

// connLimitSelectorsSample 构造 conn_limit.show TYPE=all 的选择器响应
// Data（真机实测 ipgroup 有值）。
func connLimitSelectorsSample() map[string]any {
	return map[string]any{
		"total":   0,
		"data":    []any{},
		"ipgroup": []any{"公共无线", "前台", "监控"},
	}
}

// TestConnLimitList 覆盖 conn_limit.show 的成功路径与参数断言。
func TestConnLimitList(t *testing.T) {
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
			name: "关键字过滤参数透传",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 10, Keywords: "172.22", Finds: "ip_addr,comment"},
			want: map[string]any{"TYPE": "total,data", "limit": "0,10", "FINDS": "ip_addr,comment", "KEYWORDS": "172.22"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("conn_limit", "show", connLimitSample())
			page, err := New(ts.Client()).ConnLimit.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 1 || r.Enabled != EnabledYes || r.Comment != "限连测试" ||
				r.IPAddr != "172.22.2.0/24" || r.Protocol != ConnLimitProtocolTCP ||
				r.WANPort != "any" || r.Limits != 500 {
				t.Fatalf("规则行解码不符: %+v", r)
			}
			if !r.IsEnabled() {
				t.Fatal("IsEnabled 应为 true")
			}
			wantParam(t, paramOf(t, ts.CallsFor("conn_limit")[0]), tt.want)
		})
	}
}

// TestConnLimitSelectors 覆盖 conn_limit.show TYPE=all 选择器解码。
func TestConnLimitSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("conn_limit", "show", connLimitSelectorsSample())
	sel, err := New(ts.Client()).ConnLimit.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 3 || sel.IPGroup[0] != "公共无线" || sel.IPGroup[2] != "监控" {
		t.Fatalf("ipgroup 选择器解码不符: %+v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("conn_limit")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestConnLimitAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestConnLimitAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("conn_limit", "add", map[string]any{})
	ts.SetResponse("conn_limit", "edit", map[string]any{})
	ts.SetResponse("conn_limit", "del", map[string]any{})
	svc := New(ts.Client()).ConnLimit
	ctx := context.Background()

	rule := ConnLimitRule{
		Enabled: EnabledYes, IPAddr: "172.22.2.0/24",
		Protocol: ConnLimitProtocolAny, WANPort: "any", Limits: 500,
		Comment: "ikuai-sdk-test-",
	}
	if _, err := svc.Add(ctx, rule); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	p := paramOf(t, ts.CallsFor("conn_limit")[0])
	if p["ip_addr"] != "172.22.2.0/24" || p["protocol"] != ConnLimitProtocolAny ||
		p["limits"] != float64(500) || p["enabled"] != EnabledYes {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, hasID := p["id"]; hasID {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, ConnLimitRule{ID: 1, Limits: 800}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if err := svc.Delete(ctx, 1, 2); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	recs := ts.CallsFor("conn_limit")
	if recs[1].Action != "edit" || recs[2].Action != "del" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s",
			recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action)
	}
	if _, ok := paramOf(t, recs[3])["ids"]; !ok {
		t.Fatalf("多 ID del 应使用 ids: %v", paramOf(t, recs[3]))
	}
}

// TestConnLimitError 覆盖固件错误路径。
func TestConnLimitError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("conn_limit", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "conn_limit", Action: action}
	})
	svc := New(ts.Client()).ConnLimit
	if _, err := svc.List(context.Background(), nil); err == nil {
		t.Fatal("List 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Selectors(context.Background()); err == nil {
		t.Fatal("Selectors 应失败")
	}
	if _, err := svc.Add(context.Background(), ConnLimitRule{IPAddr: "172.22.1.1"}); err == nil {
		t.Fatal("Add 应失败")
	}
	if err := svc.Edit(context.Background(), ConnLimitRule{ID: 1}); err == nil {
		t.Fatal("Edit 应失败")
	}
	if err := svc.Delete(context.Background(), 1); err == nil {
		t.Fatal("Delete 应失败")
	}
}
