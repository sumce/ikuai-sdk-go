package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestVLANCRUD 覆盖 vlan 的 show/add/edit/del（show 行形态真机实证）。
func TestVLANCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("vlan", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 2, "enabled": "yes", "vlan_id": "1211", "vlan_name": "vlan1211",
			"interface": "lan3", "ip_addr": "172.22.2.1", "ip_addr_int": 2887123201,
			"netmask": "255.255.255.0", "ip_mask": "172.22.2.1/255.255.255.0",
			"mac": "", "comment": "lan3接口对端服务器出租数据",
		}},
	})
	ts.SetResponse("vlan", "add", map[string]any{})
	ts.SetResponse("vlan", "edit", map[string]any{})
	ts.SetResponse("vlan", "del", map[string]any{})
	svc := New(ts.Client()).VLAN

	page, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || page.Rows[0].VLANID != "1211" ||
		page.Rows[0].VLANName != "vlan1211" || page.Rows[0].Interface != "lan3" ||
		!page.Rows[0].IsEnabled() {
		t.Fatalf("VLAN 行解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("vlan")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	if _, err := svc.Add(context.Background(), VLAN{Enabled: EnabledYes, VLANID: "100", Interface: "lan3"}); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if err := svc.Edit(context.Background(), VLAN{ID: 2, Comment: "x"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(context.Background(), 2); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	recs := ts.CallsFor("vlan")
	if len(recs) != 4 {
		t.Fatalf("应有 4 次调用, got %d", len(recs))
	}
	for i, want := range []string{"show", "add", "edit", "del"} {
		if recs[i].Action != want {
			t.Fatalf("第 %d 次调用 action = %q, 期望 %q", i+1, recs[i].Action, want)
		}
	}
}

// TestVLANError 覆盖固件错误路径。
func TestVLANError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("vlan", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "vlan", Action: action}
	})
	_, err := New(ts.Client()).VLAN.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 || ae.Action != "show" {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
