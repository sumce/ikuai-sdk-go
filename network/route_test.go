package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestRouteStaticCRUD 覆盖 static_rt 的 show/add/edit/del（行形态真机实证）。
func TestRouteStaticCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("static_rt", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "enabled": "yes", "dst_addr": "172.20.0.0",
			"dst_addr_int": 2886992644, "netmask": "255.255.255.0",
			"gateway": "192.168.9.254", "gateway_int": 3232236286,
			"interface": "lan2", "prio": 1, "ip_type": "4", "comment": "",
		}},
	})
	ts.SetResponse("static_rt", "add", map[string]any{})
	ts.SetResponse("static_rt", "edit", map[string]any{})
	ts.SetResponse("static_rt", "del", map[string]any{})
	svc := New(ts.Client()).Route

	page, err := svc.StaticList(context.Background(), nil)
	if err != nil {
		t.Fatalf("StaticList 失败: %v", err)
	}
	r := page.Rows[0]
	if page.Total != 1 || r.DstAddr != "172.20.0.0" || r.Netmask != "255.255.255.0" ||
		r.Gateway != "192.168.9.254" || r.Interface != "lan2" ||
		r.IPType != StaticRouteIPTypeV4 || !r.IsEnabled() {
		t.Fatalf("静态路由解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("static_rt")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	if _, err := svc.StaticAdd(context.Background(), StaticRoute{Enabled: EnabledYes, DstAddr: "10.0.0.0", Netmask: "255.0.0.0", Gateway: "192.168.9.254", Interface: "lan2", IPType: StaticRouteIPTypeV4}); err != nil {
		t.Fatalf("StaticAdd 失败: %v", err)
	}
	if err := svc.StaticEdit(context.Background(), StaticRoute{ID: 1}); err != nil {
		t.Fatalf("StaticEdit 失败: %v", err)
	}
	if err := svc.StaticDelete(context.Background(), 1); err != nil {
		t.Fatalf("StaticDelete 失败: %v", err)
	}
	recs := ts.CallsFor("static_rt")
	if len(recs) != 4 {
		t.Fatalf("应有 4 次调用, got %d", len(recs))
	}
	for i, want := range []string{"show", "add", "edit", "del"} {
		if recs[i].Action != want {
			t.Fatalf("第 %d 次调用 action = %q, 期望 %q", i+1, recs[i].Action, want)
		}
	}
}

// TestRouteTableList 覆盖 static_rt_table.show（运行状态表，只读）。
func TestRouteTableList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("static_rt_table", "show", map[string]any{
		"total": 2,
		"data": []any{
			map[string]any{"id": 1, "dst_addr": "0.0.0.0", "dst_addr_int": 0,
				"netmask": "0.0.0.0", "gateway": "10.32.20.254", "interface": "adsl02", "prio": 1},
			map[string]any{"id": 2, "dst_addr": "172.22.1.0", "dst_addr_int": 2887122944,
				"netmask": "255.255.255.0", "gateway": "0.0.0.0", "interface": "lan1", "prio": 1},
		},
	})
	page, err := New(ts.Client()).Route.TableList(context.Background(), nil)
	if err != nil {
		t.Fatalf("TableList 失败: %v", err)
	}
	if page.Total != 2 || page.Rows[0].Gateway != "10.32.20.254" ||
		page.Rows[1].Interface != "lan1" {
		t.Fatalf("路由表解码不符: %+v", page)
	}
	rec := ts.CallsFor("static_rt_table")[0]
	if rec.FuncName != "static_rt_table" || rec.Action != "show" {
		t.Fatalf("调用记录不符: %s.%s", rec.FuncName, rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestRouteError 覆盖固件错误路径。
func TestRouteError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("static_rt_table", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "static_rt_table", Action: action}
	})
	_, err := New(ts.Client()).Route.TableList(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
