package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// pppuserFixture 是真机 pppuser.show 行的全字段（真机 3.7.24 实证）。
func pppuserFixture() map[string]any {
	return map[string]any{
		"auto_vlanid": 1, "bind_vlanid": "0", "bind_ifname": "any",
		"last_conntime": 0, "cardid": "", "id": 1, "comment": "",
		"username": "admin", "duration": 0, "passwd": "A151211058Ba@",
		"last_offtime": 0, "expires": 1892484059, "start_time": 1791741629,
		"create_time": 1791741676, "ppptype": "any", "enabled": "yes",
		"pppname": "", "share": 1, "auto_mac": 0, "upload": 999999,
		"download": 999999, "ip_type": 0, "mac": "", "address": "",
		"name": "", "phone": "", "packages": 0, "proxy_username": "",
		"pppoev6_wan": "",
	}
}

// TestUsersList 覆盖 pppuser.show 的成功路径、真机行解码与参数断言。
func TestUsersList(t *testing.T) {
	tests := []struct {
		name  string
		opts  *ikuai.ListOptions
		wantP map[string]any
	}{
		{name: "nil-opts 默认分页", opts: nil,
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,20"}},
		{name: "关键字搜索", opts: &ikuai.ListOptions{Offset: 0, Limit: 5,
			Finds: "username,name,address,phone,comment", Keywords: "admin"},
			wantP: map[string]any{"TYPE": "total,data", "limit": "0,5",
				"FINDS": "username,name,address,phone,comment", "KEYWORDS": "admin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("pppuser", "show", map[string]any{
				"total": 1, "data": []any{pppuserFixture()},
			})
			page, err := New(ts.Client()).Users.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("行数不符: %+v", page)
			}
			u := page.Rows[0]
			if u.ID != 1 || u.Username != "admin" || u.Passwd != "A151211058Ba@" ||
				!u.IsEnabled() || u.PPPType != PPPTypeAny || u.Share != 1 ||
				!u.UnlimitedSpeed() || u.IPType != IPTypeNone ||
				u.BindVlanID != "0" || u.BindIfname != "any" || u.AutoVlanID != 1 ||
				u.ExpiresTime().Unix() != 1892484059 || u.CreateTime != 1791741676 ||
				u.StartTime != 1791741629 {
				t.Fatalf("真机行字段解码不符: %+v", u)
			}
			wantParam(t, paramOf(t, ts.CallsFor("pppuser")[0]), tt.wantP)
		})
	}
}

// TestUsersCRUD 覆盖 add/edit/del 的参数断言（写动作，仅离线）。
func TestUsersCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("pppuser", "add", map[string]any{})
	ts.SetResponse("pppuser", "edit", map[string]any{})
	ts.SetResponse("pppuser", "del", map[string]any{})
	svc := New(ts.Client()).Users
	ctx := context.Background()

	if _, err := svc.Add(ctx, PPPUser{
		Username: "ikuai-sdk-test-1", Passwd: "pw123", Enabled: EnabledYes,
		PPPType: PPPTypePPPoE, Share: 1, Upload: 999999, Download: 999999,
	}); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	addRec := ts.CallsFor("pppuser")[0]
	if addRec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", addRec.Action)
	}
	p := paramOf(t, addRec)
	if p["username"] != "ikuai-sdk-test-1" || p["passwd"] != "pw123" ||
		p["enabled"] != EnabledYes || p["ppptype"] != PPPTypePPPoE {
		t.Fatalf("add param 字段不符: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, PPPUser{ID: 9, Username: "u9", Share: 2}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	editRec := ts.CallsFor("pppuser")[1]
	if editRec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", editRec.Action)
	}
	p = paramOf(t, editRec)
	if p["id"] != float64(9) || p["username"] != "u9" || p["share"] != float64(2) {
		t.Fatalf("edit param 字段不符: %v", p)
	}

	if err := svc.Delete(ctx, 9, 10); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	delRec := ts.CallsFor("pppuser")[2]
	if delRec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", delRec.Action)
	}
	wantParam(t, paramOf(t, delRec), map[string]any{"ids": []any{float64(9), float64(10)}})
}

// TestUsersError 覆盖固件错误路径。
func TestUsersError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("pppuser", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95000, ErrMsg: "失败", FuncName: "pppuser", Action: action}
	})
	svc := New(ts.Client()).Users
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("List 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95000 {
		t.Fatalf("期望 APIError(95000), got %v", err)
	}
	if _, err := svc.Add(ctx, PPPUser{}); err == nil {
		t.Fatal("Add 应失败")
	}
	if err := svc.Edit(ctx, PPPUser{ID: 1}); err == nil {
		t.Fatal("Edit 应失败")
	}
	if err := svc.Delete(ctx, 1); err == nil {
		t.Fatal("Delete 应失败")
	}
}
