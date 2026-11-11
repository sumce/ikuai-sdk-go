package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestGroupsIPCRUD 覆盖 ipgroup 的 show/add/edit/del（show 行形态真机实证）。
func TestGroupsIPCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ipgroup", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 1, "group_name": "出租300", "addr_pool": "172.20.0.1/24",
			"type": 0, "comment": "",
		}},
	})
	ts.SetResponse("ipgroup", "add", map[string]any{})
	ts.SetResponse("ipgroup", "edit", map[string]any{})
	ts.SetResponse("ipgroup", "del", map[string]any{})
	svc := New(ts.Client()).Groups

	page, err := svc.IPList(context.Background(), nil)
	if err != nil {
		t.Fatalf("IPList 失败: %v", err)
	}
	if page.Total != 1 || page.Rows[0].GroupName != "出租300" ||
		page.Rows[0].Type != IPGroupTypeAddr || page.Rows[0].Comment != "" {
		t.Fatalf("IPGroup 解码不符: %+v", page)
	}
	if pools := page.Rows[0].AddrPools(); len(pools) != 1 || pools[0] != "172.20.0.1/24" {
		t.Fatalf("AddrPools 解析不符: %v", pools)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ipgroup")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	if _, err := svc.IPAdd(context.Background(), IPGroup{GroupName: "g", AddrPool: "172.20.0.0/24"}); err != nil {
		t.Fatalf("IPAdd 失败: %v", err)
	}
	if err := svc.IPEdit(context.Background(), IPGroup{ID: 1}); err != nil {
		t.Fatalf("IPEdit 失败: %v", err)
	}
	if err := svc.IPDelete(context.Background(), 1); err != nil {
		t.Fatalf("IPDelete 失败: %v", err)
	}
	recs := ts.CallsFor("ipgroup")
	if len(recs) != 4 {
		t.Fatalf("应有 4 次调用, got %d", len(recs))
	}
	for i, want := range []string{"show", "add", "edit", "del"} {
		if recs[i].Action != want {
			t.Fatalf("第 %d 次调用 action = %q, 期望 %q", i+1, recs[i].Action, want)
		}
	}
}

// TestGroupsIPv6MACDT 覆盖 ipv6group/macgroup/dtgroup 的四类调用。
func TestGroupsIPv6MACDT(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"ipv6group", "macgroup", "dtgroup"} {
		ts.SetResponse(fn, "show", map[string]any{"total": 0, "data": []any{}})
		ts.SetResponse(fn, "add", map[string]any{})
		ts.SetResponse(fn, "edit", map[string]any{})
		ts.SetResponse(fn, "del", map[string]any{})
	}
	svc := New(ts.Client()).Groups
	ctx := context.Background()

	if _, err := svc.IPv6List(ctx, nil); err != nil {
		t.Fatalf("IPv6List 失败: %v", err)
	}
	if _, err := svc.IPv6Add(ctx, IPv6Group{GroupName: "g6", AddrPool: "2408::/64"}); err != nil {
		t.Fatalf("IPv6Add 失败: %v", err)
	}
	if err := svc.IPv6Edit(ctx, IPv6Group{ID: 1}); err != nil {
		t.Fatalf("IPv6Edit 失败: %v", err)
	}
	if err := svc.IPv6Delete(ctx, 1); err != nil {
		t.Fatalf("IPv6Delete 失败: %v", err)
	}

	if _, err := svc.MACList(ctx, nil); err != nil {
		t.Fatalf("MACList 失败: %v", err)
	}
	if _, err := svc.MACAdd(ctx, MACGroup{GroupName: "gm", MACs: "aa:bb:cc:dd:ee:ff,11:22:33:44:55:66"}); err != nil {
		t.Fatalf("MACAdd 失败: %v", err)
	}
	if err := svc.MACEdit(ctx, MACGroup{ID: 1}); err != nil {
		t.Fatalf("MACEdit 失败: %v", err)
	}
	if err := svc.MACDelete(ctx, 1); err != nil {
		t.Fatalf("MACDelete 失败: %v", err)
	}

	if _, err := svc.DTList(ctx, nil); err != nil {
		t.Fatalf("DTList 失败: %v", err)
	}
	if _, err := svc.DTAdd(ctx, DTGroup{GroupName: "gt", Week: "12345", Time: "00:00-08:00"}); err != nil {
		t.Fatalf("DTAdd 失败: %v", err)
	}
	if err := svc.DTEdit(ctx, DTGroup{ID: 1}); err != nil {
		t.Fatalf("DTEdit 失败: %v", err)
	}
	if err := svc.DTDelete(ctx, 1); err != nil {
		t.Fatalf("DTDelete 失败: %v", err)
	}

	// 每类 func 均应产生 show/add/edit/del 四次调用。
	for _, fn := range []string{"ipv6group", "macgroup", "dtgroup"} {
		recs := ts.CallsFor(fn)
		if len(recs) != 4 {
			t.Fatalf("%s 应有 4 次调用, got %d", fn, len(recs))
		}
		for i, want := range []string{"show", "add", "edit", "del"} {
			if recs[i].Action != want {
				t.Fatalf("%s 第 %d 次调用 action = %q, 期望 %q", fn, i+1, recs[i].Action, want)
			}
		}
	}
}

// TestGroupsMACListHelper 覆盖 MAC 分组列表解析辅助。
func TestGroupsMACListHelper(t *testing.T) {
	g := MACGroup{MACs: " aa:bb:cc:dd:ee:ff ,, 11:22:33:44:55:66 "}
	got := g.MACList()
	if len(got) != 2 || got[0] != "aa:bb:cc:dd:ee:ff" || got[1] != "11:22:33:44:55:66" {
		t.Fatalf("MACList 解析不符: %v", got)
	}
}

// TestGroupsError 覆盖固件错误路径。
func TestGroupsError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ipgroup", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "ipgroup", Action: action}
	})
	_, err := New(ts.Client()).Groups.IPList(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
