package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestWebUserPasswdFlag 覆盖 webuser.show TYPE=mod_passwd 的成功路径与参数断言。
func TestWebUserPasswdFlag(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webuser", "show", map[string]any{"mod_passwd": 0})
	flag, err := New(ts.Client()).WebUser.PasswdFlag(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("PasswdFlag 失败: %v", err)
	}
	if flag != 0 {
		t.Fatalf("mod_passwd 解码不符: %d", flag)
	}
	wantParam(t, paramOf(t, ts.CallsFor("webuser")[0]),
		map[string]any{"TYPE": WebUserTypeModPasswd, "username": "testuser"})
}

// TestWebUserPasswdFlagEmptyUsername 空用户名返回 ErrInvalidParam。
func TestWebUserPasswdFlagEmptyUsername(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	_, err := New(ts.Client()).WebUser.PasswdFlag(context.Background(), "")
	if err == nil {
		t.Fatal("空 username 应报错")
	}
}

// TestWebUserAdd 覆盖 webuser.add（写动作，离线）。
func TestWebUserAdd(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webuser", "add", map[string]any{})
	rowID, err := New(ts.Client()).WebUser.Add(context.Background(),
		WebUserAccount{Username: "op", Passwd: "pw"})
	if err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("webuser")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if p["username"] != "op" || p["passwd"] != "pw" {
		t.Fatalf("add 参数不符: %v", p)
	}
}

// TestWebUserEditDelete 覆盖 edit/del（写动作，离线）。
func TestWebUserEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webuser", "edit", map[string]any{})
	ts.SetResponse("webuser", "del", map[string]any{})
	svc := New(ts.Client()).WebUser

	// Edit 空 ID 报参数错误。
	if err := svc.Edit(context.Background(), WebUserAccount{Username: "op"}); err == nil {
		t.Fatal("空 ID 的 edit 应报错")
	}
	if err := svc.Edit(context.Background(), WebUserAccount{ID: 2, Username: "op2"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	recEdit := ts.CallsFor("webuser")[0]
	if recEdit.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", recEdit.Action)
	}
	p := paramOf(t, recEdit)
	if p["id"].(float64) != 2 || p["username"] != "op2" {
		t.Fatalf("edit 参数不符: %v", p)
	}

	// 单个删除 → {"id":3}；多个删除 → {"ids":[...]}。
	if err := svc.Delete(context.Background(), 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if err := svc.Delete(context.Background(), 4, 5); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	recs := ts.CallsFor("webuser")
	if recs[1].Action != "del" {
		t.Fatalf("action = %q, 期望 del", recs[1].Action)
	}
	wantParam(t, paramOf(t, recs[1]), map[string]any{"id": float64(3)})
	got := paramOf(t, recs[2])
	ids, ok := got["ids"].([]any)
	if !ok || len(ids) != 2 || ids[0].(float64) != 4 || ids[1].(float64) != 5 {
		t.Fatalf("多 ID del 参数不符: %v", got)
	}
}

// TestWebUserModPasswd 覆盖 webuser.mod_passwd（写动作，离线）。
func TestWebUserModPasswd(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("webuser", "mod_passwd", map[string]any{})
	if err := New(ts.Client()).WebUser.ModPasswd(context.Background(),
		WebUserPasswdMod{Username: "testuser", Passwd: "new-pw"}); err != nil {
		t.Fatalf("ModPasswd 失败: %v", err)
	}
	rec := ts.CallsFor("webuser")[0]
	if rec.Action != "mod_passwd" {
		t.Fatalf("action = %q, 期望 mod_passwd", rec.Action)
	}
	wantParam(t, paramOf(t, rec),
		map[string]any{"username": "testuser", "passwd": "new-pw"})
}

// TestWebUserError 覆盖固件错误路径。
func TestWebUserError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("webuser", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95008, ErrMsg: "失败", FuncName: "webuser", Action: action}
	})
	if _, err := New(ts.Client()).WebUser.PasswdFlag(context.Background(), "testuser"); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95008 {
		t.Fatalf("期望 APIError(95008), got %v", err)
	}
	if _, err := New(ts.Client()).WebUser.Add(context.Background(), WebUserAccount{Username: "op"}); err == nil {
		t.Fatal("add 错误应上抛")
	}
	if err := New(ts.Client()).WebUser.ModPasswd(context.Background(), WebUserPasswdMod{Username: "u", Passwd: "p"}); err == nil {
		t.Fatal("mod_passwd 错误应上抛")
	}
}
