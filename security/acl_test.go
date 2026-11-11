package security

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// aclSample 构造 acl.show 的分页响应 Data（行字段空表推导，待真机验证）。
func aclSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 3, "enabled": "yes", "comment": "封禁测试",
			"src_addr": "172.22.2.0/24", "dst_addr": "", "src_port": "",
			"dst_port": "443", "iinterface": "any", "ointerface": "adsl07",
			"protocol": "tcp_udp", "action": "drop", "dir": "forward",
			"dir_matching": "close", "week": "1234567", "time": "00:00-23:59",
		}},
	}
}

// aclSelectorsSample 构造 acl.show TYPE=all 的选择器响应 Data（真机实测）。
func aclSelectorsSample() map[string]any {
	return map[string]any{
		"total":        0,
		"data":         []any{},
		"interface":    []any{[]any{"any"}, []any{"lan1", "内网"}, []any{"adsl07", "联通静公网"}},
		"ipgroup":      []any{"办公区", "访客"},
		"macgroup":     []any{},
		"dtgroup":      []any{"工作时间"},
		"iplib_status": 0,
	}
}

// TestACLList 覆盖 acl.show 的成功路径与参数断言。
func TestACLList(t *testing.T) {
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
			name: "分页+排序",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 10, OrderBy: "id", Order: "asc"},
			want: map[string]any{"TYPE": "total,data", "limit": "20,10", "ORDER_BY": "id", "ORDER": "asc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("acl", "show", aclSample())
			page, err := New(ts.Client()).ACL.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 3 || r.Enabled != "yes" || r.Comment != "封禁测试" ||
				r.SrcAddr != "172.22.2.0/24" || r.DstPort != "443" ||
				r.IInterface != "any" || r.OInterface != "adsl07" ||
				r.Protocol != ACLProtocolTCPUDP || r.Action != ACLActionDrop ||
				r.Dir != ACLDirForward || r.DirMatching != ACLDirMatchingClose ||
				r.Week != "1234567" || r.Time != "00:00-23:59" {
				t.Fatalf("规则行解码不符: %+v", r)
			}
			if !r.IsEnabled() {
				t.Fatal("IsEnabled 应为 true")
			}
			wantParam(t, paramOf(t, ts.CallsFor("acl")[0]), tt.want)
		})
	}
}

// TestACLSelectors 覆盖 acl.show TYPE=all 选择器解码。
func TestACLSelectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl", "show", aclSelectorsSample())
	sel, err := New(ts.Client()).ACL.Selectors(context.Background())
	if err != nil {
		t.Fatalf("Selectors 失败: %v", err)
	}
	if len(sel.Interface) != 3 || sel.Interface[0][0] != "any" || sel.Interface[2][1] != "联通静公网" {
		t.Fatalf("接口选择器解码不符: %v", sel.Interface)
	}
	if len(sel.IPGroup) != 2 || sel.IPGroup[0] != "办公区" || len(sel.DTGroup) != 1 ||
		len(sel.MACGroup) != 0 || sel.IPLibStatus != 0 {
		t.Fatalf("选择器解码不符: %+v", sel)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestACLAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestACLAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl", "add", map[string]any{})
	ts.SetResponse("acl", "edit", map[string]any{})
	ts.SetResponse("acl", "del", map[string]any{})
	svc := New(ts.Client()).ACL
	ctx := context.Background()

	rule := ACLRule{
		Enabled: EnabledYes, SrcAddr: "172.22.2.0/24", DstPort: "443",
		IInterface: "any", OInterface: "adsl07",
		Protocol: ACLProtocolTCPUDP, Action: ACLActionDrop, Dir: ACLDirForward,
		Week: "1234567", Time: "00:00-23:59", Comment: "ikuai-sdk-test-",
	}
	if _, err := svc.Add(ctx, rule); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	p := paramOf(t, ts.CallsFor("acl")[0])
	if p["action"] != ACLActionDrop || p["protocol"] != ACLProtocolTCPUDP ||
		p["enabled"] != EnabledYes || p["dst_port"] != "443" {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, hasID := p["id"]; hasID {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, ACLRule{ID: 3, Action: ACLActionAccept}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if err := svc.Delete(ctx, 3, 4, 5); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	recs := ts.CallsFor("acl")
	if recs[1].Action != "edit" || recs[2].Action != "del" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s",
			recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action)
	}
	delP := paramOf(t, recs[2])
	if delP["id"] != float64(3) {
		t.Fatalf("单 ID del 应使用 id: %v", delP)
	}
	if _, ok := paramOf(t, recs[3])["ids"]; !ok {
		t.Fatalf("多 ID del 应使用 ids: %v", paramOf(t, recs[3]))
	}
}

// TestACLAddValidation 覆盖 Add 的参数校验（action 必须为 accept/drop）。
func TestACLAddValidation(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client()).ACL
	cases := []string{"", "block", "ACCEPT", "forward"}
	for _, action := range cases {
		if _, err := svc.Add(context.Background(), ACLRule{Action: action}); !errors.Is(err, ikuai.ErrInvalidParam) {
			t.Fatalf("非法 action %q 应返回 ErrInvalidParam, got %v", action, err)
		}
	}
	if len(ts.CallsFor("acl")) != 0 {
		t.Fatal("参数校验失败不应发起调用")
	}
}

// TestACLError 覆盖固件错误路径。
func TestACLError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("acl", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "acl", Action: action}
	})
	svc := New(ts.Client()).ACL
	if _, err := svc.List(context.Background(), nil); err == nil {
		t.Fatal("List 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Selectors(context.Background()); err == nil {
		t.Fatal("Selectors 应失败")
	}
	if _, err := svc.Add(context.Background(), ACLRule{Action: ACLActionAccept}); err == nil {
		t.Fatal("Add 应失败")
	}
	if err := svc.Edit(context.Background(), ACLRule{ID: 1, Action: ACLActionDrop}); err == nil {
		t.Fatal("Edit 应失败")
	}
	if err := svc.Delete(context.Background(), 1); err == nil {
		t.Fatal("Delete 应失败")
	}
}
