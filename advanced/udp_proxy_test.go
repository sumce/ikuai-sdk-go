package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// udpProxySample 构造 udp_proxy.show 的响应 Data（顶层字段真机实证；
// 行字段空表推导）。
func udpProxySample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "interface": "adsl01", "port": 8888,
			"renew_time": 30, "udpxy_net_enabled": "no", "enabled": "yes",
		}},
		"interface": []any{[]any{"lan1"}, []any{"adsl01", "移动私网"}},
	}
}

// TestUDPProxyListGet 覆盖 udp_proxy.show 的两种 TYPE 成功路径。
func TestUDPProxyListGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("udp_proxy", "show", udpProxySample())
	svc := New(ts.Client()).UDProxy
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 1 || r.Interface != "adsl01" || r.Port != 8888 ||
		r.RenewTime != 30 || r.NetEnabled != "no" || !r.IsEnabled() {
		t.Fatalf("规则行解码不符: %+v", r)
	}

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if info.Total != 1 || len(info.Rules) != 1 || len(info.Interface) != 2 ||
		info.Interface[1][0] != "adsl01" || info.Interface[1][1] != "移动私网" {
		t.Fatalf("TYPE=all 区块解码不符: %+v", info)
	}

	recs := ts.CallsFor("udp_proxy")
	wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": "total,data", "limit": "0,20"})
	wantParam(t, paramOf(t, recs[1]), map[string]any{"TYPE": "all"})
}

// TestUDPProxyAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestUDPProxyAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, act := range []string{"add", "edit", "del"} {
		ts.SetResponse("udp_proxy", act, map[string]any{})
	}
	svc := New(ts.Client()).UDProxy
	ctx := context.Background()

	rule := UDPProxyRule{Interface: "adsl01", Port: 8899, RenewTime: 60,
		NetEnabled: "yes", Enabled: EnabledYes}
	rowID, err := svc.Add(ctx, rule)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	rule.ID = 3
	rule.Port = 8900
	if err := svc.Edit(ctx, rule); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	recs := ts.CallsFor("udp_proxy")
	if recs[0].Action != "add" || recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action)
	}
	p := paramOf(t, recs[0])
	if p["interface"] != "adsl01" || p["port"] != float64(8899) ||
		p["renew_time"] != float64(60) || p["udpxy_net_enabled"] != "yes" {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}
	p = paramOf(t, recs[1])
	if p["id"] != float64(3) || p["port"] != float64(8900) {
		t.Fatalf("edit 参数不符: %v", p)
	}
	wantParam(t, paramOf(t, recs[2]), map[string]any{"id": float64(3)})
}

// TestUDPProxyError 覆盖固件错误路径。
func TestUDPProxyError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("udp_proxy", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95006, ErrMsg: "失败", FuncName: "udp_proxy", Action: action}
	})
	if _, err := New(ts.Client()).UDProxy.List(context.Background(), nil); err == nil {
		t.Fatal("错误路径应报错")
	} else {
		errIsAPICode(t, err, 95006)
	}
}
