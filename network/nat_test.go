package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestNATRuleListSelectors 覆盖 nat_rule.show 的列表与候选路径
// （TYPE=interface,action 候选为抓包实证）。
func TestNATRuleListSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	// 列表（抓包实证 total=0 空表；行字段待真机验证）。
	ts.SetResponse("nat_rule", "show", map[string]any{"total": 0, "data": []any{}})
	svc := New(ts.Client()).NAT
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 0 {
		t.Fatalf("列表解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("nat_rule")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	// 候选（抓包实证）。
	ts.SetResponse("nat_rule", "show", map[string]any{
		"interface": []any{[]any{"lan2", "内网接入三层交换机"}, []any{"adsl07"}},
		"action":    []any{"filter", "snat", "dnat"},
	})
	sel, err := svc.Selectors(ctx)
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 2 || sel.Interface[0][1] != "内网接入三层交换机" ||
		len(sel.Action) != 3 || sel.Action[0] != NATRuleActionFilter {
		t.Fatalf("候选解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("nat_rule")[1]),
		map[string]any{"TYPE": NATRuleTypeInterfaceAction})
}

// TestNATRuleCRUD 覆盖 nat_rule 的 add/edit/del（写动作仅离线）。
func TestNATRuleCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("nat_rule", "add", map[string]any{})
	ts.SetResponse("nat_rule", "edit", map[string]any{})
	ts.SetResponse("nat_rule", "del", map[string]any{})
	svc := New(ts.Client()).NAT
	ctx := context.Background()

	if _, err := svc.Add(ctx, NATRule{Enabled: EnabledYes, Interface: "adsl07", Action: NATRuleActionSNAT}); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if err := svc.Edit(ctx, NATRule{ID: 1}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	assertActionSeq(t, ts.CallsFor("nat_rule"), "add", "edit", "del")
	if p := paramOf(t, ts.CallsFor("nat_rule")[0]); p["action"] != "snat" {
		t.Fatalf("add 参数不符: %v", p)
	}
}

// TestNetmapCRUD 覆盖 netmap 的 show/add/edit/del（show 真机实证空表；
// 行字段待真机验证）。
func TestNetmapCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, a := range []string{"show", "add", "edit", "del"} {
		ts.SetResponse("netmap", a, map[string]any{})
	}
	ts.SetResponse("netmap", "show", map[string]any{"total": 0, "data": []any{}})
	svc := New(ts.Client()).NAT
	ctx := context.Background()

	page, err := svc.NetmapList(ctx, nil)
	if err != nil {
		t.Fatalf("NetmapList 失败: %v", err)
	}
	if page.Total != 0 {
		t.Fatalf("DMZ 列表解码不符: %+v", page)
	}
	if _, err := svc.NetmapAdd(ctx, Netmap{Enabled: EnabledYes, LANAddr: "192.168.9.100"}); err != nil {
		t.Fatalf("NetmapAdd 失败: %v", err)
	}
	if err := svc.NetmapEdit(ctx, Netmap{ID: 1}); err != nil {
		t.Fatalf("NetmapEdit 失败: %v", err)
	}
	if err := svc.NetmapDelete(ctx, 1); err != nil {
		t.Fatalf("NetmapDelete 失败: %v", err)
	}
	assertActionSeq(t, ts.CallsFor("netmap"), "show", "add", "edit", "del")
	if p := paramOf(t, ts.CallsFor("netmap")[1]); p["lan_addr"] != "192.168.9.100" {
		t.Fatalf("add 参数不符: %v", p)
	}
}

// TestNATError 覆盖固件错误路径。
func TestNATError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("nat_rule", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "nat_rule", Action: action}
	})
	_, err := New(ts.Client()).NAT.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
