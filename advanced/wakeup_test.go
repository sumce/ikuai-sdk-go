package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// wakeupSample 构造 wakeup.show 的分页响应 Data（行字段空表推导）。
func wakeupSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 3, "mac": "aa:bb:cc:dd:ee:01", "week": "12345",
			"time": "08:30", "comment": "上班前唤醒NAS", "enabled": "yes",
		}},
	}
}

// TestWakeupList 覆盖 wakeup.show 的成功路径与参数断言（真机实证
// TYPE=total,data 形态）。
func TestWakeupList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts 默认分页", nil, map[string]any{"TYPE": "total,data", "limit": "0,20"}},
		{"自定义分页", &ikuai.ListOptions{Offset: 20, Limit: 50}, map[string]any{"TYPE": "total,data", "limit": "20,50"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("wakeup", "show", wakeupSample())
			page, err := New(ts.Client()).Wakeup.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 3 || r.MAC != "aa:bb:cc:dd:ee:01" || r.Week != "12345" ||
				r.Time != "08:30" || r.Comment != "上班前唤醒NAS" || !r.IsEnabled() {
				t.Fatalf("任务行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("wakeup")[0]), tt.want)
		})
	}
}

// TestWakeupAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestWakeupAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("wakeup", "add", map[string]any{})
	ts.SetResponse("wakeup", "edit", map[string]any{})
	ts.SetResponse("wakeup", "del", map[string]any{})
	svc := New(ts.Client()).Wakeup
	ctx := context.Background()

	task := WakeupTask{MAC: "aa:bb:cc:dd:ee:02", Week: "1-7", Time: "07:00", Enabled: EnabledYes}
	rowID, err := svc.Add(ctx, task)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	task.ID = 2
	task.Time = "07:30"
	if err := svc.Edit(ctx, task); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 2); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	recs := ts.CallsFor("wakeup")
	if recs[0].Action != "add" || recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action)
	}
	p := paramOf(t, recs[0])
	if p["mac"] != "aa:bb:cc:dd:ee:02" || p["enabled"] != "yes" {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}
	p = paramOf(t, recs[1])
	if p["id"] != float64(2) || p["time"] != "07:30" {
		t.Fatalf("edit 参数不符: %v", p)
	}
	wantParam(t, paramOf(t, recs[2]), map[string]any{"id": float64(2)})
}

// TestWakeupWake 覆盖立即唤醒（写动作，仅离线；action 名待真机验证）。
func TestWakeupWake(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("wakeup", "wake", map[string]any{})
	if err := New(ts.Client()).Wakeup.Wake(context.Background(), "aa:bb:cc:dd:ee:03"); err != nil {
		t.Fatalf("Wake 失败: %v", err)
	}
	rec := ts.CallsFor("wakeup")[0]
	if rec.Action != "wake" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"mac": "aa:bb:cc:dd:ee:03"})
}

// TestWakeupError 覆盖固件错误路径。
func TestWakeupError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("wakeup", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95004, ErrMsg: "失败", FuncName: "wakeup", Action: action}
	})
	if _, err := New(ts.Client()).Wakeup.List(context.Background(), nil); err == nil {
		t.Fatal("错误路径应报错")
	} else {
		errIsAPICode(t, err, 95004)
	}
}
