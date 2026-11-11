package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// portMirrorSample 构造 port_mirror.show 的分页响应 Data（行字段空表推导）。
func portMirrorSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 2, "target_port": "lan4", "mirror_port": "lan1,lan2",
			"dir": "all", "comment": "抓包调试",
		}},
	}
}

// portMirrorAllSample 构造 port_mirror.show TYPE=all 的响应 Data（真机
// 3.7.24 探测实证顶层字段）。
func portMirrorAllSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 2, "target_port": "lan4", "mirror_port": "lan1,lan2",
			"dir": "all", "comment": "抓包调试",
		}},
		"interface": []any{[]any{"lan1"}, []any{"lan2", "内网接入三层交换机"}},
	}
}

// TestPortMirrorListGet 覆盖 port_mirror.show 的两种 TYPE 成功路径。
func TestPortMirrorListGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	// 同一响应同时包含 data 行与 interface 区块：List 只解码 total/data，
	// Get 额外解码 interface，互不影响。
	ts.SetResponse("port_mirror", "show", portMirrorAllSample())
	svc := New(ts.Client()).PortMirror
	ctx := context.Background()

	page, err := svc.List(ctx, &ikuai.ListOptions{Offset: 0, Limit: 5})
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 2 || r.TargetPort != "lan4" || r.MirrorPort != "lan1,lan2" ||
		r.Dir != PortMirrorDirAll || r.Comment != "抓包调试" {
		t.Fatalf("规则行解码不符: %+v", r)
	}

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if info.Total != 1 || len(info.Rules) != 1 || len(info.Interface) != 2 ||
		info.Interface[1][0] != "lan2" || info.Interface[1][1] != "内网接入三层交换机" {
		t.Fatalf("TYPE=all 区块解码不符: %+v", info)
	}

	recs := ts.CallsFor("port_mirror")
	wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": "total,data", "limit": "0,5"})
	wantParam(t, paramOf(t, recs[1]), map[string]any{"TYPE": "all"})
}

// TestPortMirrorAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestPortMirrorAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("port_mirror", "add", map[string]any{})
	ts.SetResponse("port_mirror", "edit", map[string]any{})
	ts.SetResponse("port_mirror", "del", map[string]any{})
	svc := New(ts.Client()).PortMirror
	ctx := context.Background()

	rule := PortMirrorRule{TargetPort: "lan4", MirrorPort: "lan1", Dir: PortMirrorDirIn}
	rowID, err := svc.Add(ctx, rule)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	rule.ID = 4
	rule.Dir = PortMirrorDirOut
	if err := svc.Edit(ctx, rule); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 4, 5); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	recs := ts.CallsFor("port_mirror")
	if recs[0].Action != "add" || recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action)
	}
	p := paramOf(t, recs[0])
	if p["target_port"] != "lan4" || p["mirror_port"] != "lan1" || p["dir"] != "in" {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}
	p = paramOf(t, recs[1])
	if p["id"] != float64(4) || p["dir"] != "out" {
		t.Fatalf("edit 参数不符: %v", p)
	}
	if p := paramOf(t, recs[2]); len(p["ids"].([]any)) != 2 {
		t.Fatalf("del 多 ID 参数不符: %v", p)
	}
}

// TestPortMirrorError 覆盖固件错误路径。
func TestPortMirrorError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("port_mirror", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30003, ErrMsg: "参数错误", FuncName: "port_mirror", Action: action}
	})
	if _, err := New(ts.Client()).PortMirror.List(context.Background(), nil); err == nil {
		t.Fatal("错误路径应报错")
	} else {
		errIsAPICode(t, err, 30003)
	}
}
