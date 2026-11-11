package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// registerSample 构造抓包形态的 register.show 响应 Data（data,gwid）。
func registerSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"id": 1, "node": 0, "code": "5f63cbf7f14879619e76f6bfabbaac04",
			"comment": "核心路由",
		}},
		"gwid": "d072dadeaacc4feb98a22dd0878b867a",
	}
}

// TestRegisterGet 覆盖 TYPE=data,gwid 的成功路径与参数断言。
func TestRegisterGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("register", "show", registerSample())
	info, err := New(ts.Client()).Register.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if info.GWID != "d072dadeaacc4feb98a22dd0878b867a" {
		t.Fatalf("gwid 解码不符: %q", info.GWID)
	}
	if len(info.Data) != 1 {
		t.Fatalf("注册节点数不符: %+v", info.Data)
	}
	n := info.Data[0]
	if n.ID != 1 || n.Node != 0 || n.Code != "5f63cbf7f14879619e76f6bfabbaac04" || n.Comment != "核心路由" {
		t.Fatalf("节点行解码不符: %+v", n)
	}
	if info.SCode != nil {
		t.Fatalf("未请求 scode 时应为 nil: %+v", info.SCode)
	}
	wantParam(t, paramOf(t, ts.CallsFor("register")[0]),
		map[string]any{"TYPE": RegisterTypeDataGWID})
}

// TestRegisterGetWithServiceCode 覆盖 TYPE=data,gwid,scode 路径。
func TestRegisterGetWithServiceCode(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	data := registerSample()
	data["scode"] = map[string]any{
		"status": 0, "serve_errcode": "", "serve_errmsg": "",
		"serve_timeout": "", "serve_number": "",
	}
	ts.SetResponse("register", "show", data)
	info, err := New(ts.Client()).Register.GetWithServiceCode(context.Background())
	if err != nil {
		t.Fatalf("GetWithServiceCode 失败: %v", err)
	}
	if info.SCode == nil || info.SCode.Status != 0 {
		t.Fatalf("scode 解码不符: %+v", info.SCode)
	}
	if info.GWID != "d072dadeaacc4feb98a22dd0878b867a" || len(info.Data) != 1 {
		t.Fatalf("data/gwid 解码不符: %+v", info)
	}
	wantParam(t, paramOf(t, ts.CallsFor("register")[0]),
		map[string]any{"TYPE": RegisterTypeDataGWIDSCode})
}

// TestRegisterError 覆盖固件错误路径。
func TestRegisterError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("register", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95006, ErrMsg: "失败", FuncName: "register", Action: action}
	})
	if _, err := New(ts.Client()).Register.Get(context.Background()); err == nil {
		t.Fatal("Get 错误应上抛")
	}
	if _, err := New(ts.Client()).Register.GetWithServiceCode(context.Background()); err == nil {
		t.Fatal("GetWithServiceCode 错误应上抛")
	}
}
