package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestRebootsList 覆盖 reboots.show 的成功路径与参数断言。
func TestRebootsList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts 默认分页", nil,
			map[string]any{"TYPE": RebootsTypeDataTotal, "limit": "0,20"}},
		{"第二页", &ikuai.ListOptions{Offset: 20, Limit: 20},
			map[string]any{"TYPE": RebootsTypeDataTotal, "limit": "20,20"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("reboots", "show", map[string]any{"total": 0, "data": []any{}})
			page, err := New(ts.Client()).Reboots.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 0 || len(page.Rows) != 0 {
				t.Fatalf("空任务列表应为 0 行: %+v", page)
			}
			wantParam(t, paramOf(t, ts.CallsFor("reboots")[0]), tt.want)
		})
	}
}

// TestRebootsListRows 覆盖行 Raw 解码（行字段未实证）。
func TestRebootsListRows(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("reboots", "show", map[string]any{
		"total": 1,
		"data":  []any{map[string]any{"id": 1, "week": "1", "time": "03:00", "enabled": "yes"}},
	})
	page, err := New(ts.Client()).Reboots.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	m, err := page.Rows[0].Map()
	if err != nil {
		t.Fatalf("Map 失败: %v", err)
	}
	if m["time"] != "03:00" || m["enabled"] != "yes" {
		t.Fatalf("行 map 解码不符: %v", m)
	}
}

// TestRebootsAddDelete 覆盖 add/del（写动作，action 按标准列表惯例推断，离线）。
func TestRebootsAddDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("reboots", "add", map[string]any{})
	ts.SetResponse("reboots", "del", map[string]any{})
	svc := New(ts.Client()).Reboots

	if _, err := svc.Add(context.Background(),
		RebootTask{Enabled: EnabledYes, Week: "1", Time: "03:00"}); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	rec := ts.CallsFor("reboots")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if p["enabled"] != "yes" || p["week"] != "1" || p["time"] != "03:00" {
		t.Fatalf("add 参数不符: %v", p)
	}

	if err := svc.Delete(context.Background(), 7); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	recDel := ts.CallsFor("reboots")[1]
	if recDel.Action != "del" {
		t.Fatalf("action = %q, 期望 del", recDel.Action)
	}
	wantParam(t, paramOf(t, recDel), map[string]any{"id": float64(7)})
}

// TestRebootTaskIsEnabled 覆盖 enabled 辅助方法。
func TestRebootTaskIsEnabled(t *testing.T) {
	if !(RebootTask{Enabled: EnabledYes}).IsEnabled() {
		t.Fatal("enabled=yes 应判为启用")
	}
	if (RebootTask{Enabled: EnabledNo}).IsEnabled() {
		t.Fatal("enabled=no 不应判为启用")
	}
}

// TestRebootsError 覆盖固件错误路径。
func TestRebootsError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("reboots", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95011, ErrMsg: "失败", FuncName: "reboots", Action: action}
	})
	if _, err := New(ts.Client()).Reboots.List(context.Background(), nil); err == nil {
		t.Fatal("list 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95011 {
		t.Fatalf("期望 APIError(95011), got %v", err)
	}
	if _, err := New(ts.Client()).Reboots.Add(context.Background(), RebootTask{}); err == nil {
		t.Fatal("add 错误应上抛")
	}
	if err := New(ts.Client()).Reboots.Delete(context.Background(), 1); err == nil {
		t.Fatal("del 错误应上抛")
	}
}
