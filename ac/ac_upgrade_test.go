package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// acUpgradeSample 构造 ac_upgrade.show 的分页响应 Data（行字段空表推导）。
func acUpgradeSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 9, "mac": "aa:bb:cc:dd:ee:01", "model": "ikuai AP-300",
			"version": "1.2.0", "new_version": "1.3.2",
		}},
	}
}

// TestUpgradeList 覆盖 ac_upgrade.show 的成功路径与参数断言。
func TestUpgradeList(t *testing.T) {
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
			name: "自定义分页",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 10},
			want: map[string]any{"TYPE": "total,data", "limit": "20,10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ac_upgrade", "show", acUpgradeSample())
			page, err := New(ts.Client()).Upgrade.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 9 || r.MAC != "aa:bb:cc:dd:ee:01" || r.Model != "ikuai AP-300" ||
				r.Version != "1.2.0" || r.NewVersion != "1.3.2" {
				t.Fatalf("升级行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ac_upgrade")[0]), tt.want)
		})
	}
}

// TestUpgradeAction 覆盖 Upgrade 写动作参数形态（⚠️ 写动作，仅离线；
// 真机 AP 升级动作绝对禁止调用）。
func TestUpgradeAction(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_upgrade", "set", map[string]any{})
	svc := New(ts.Client()).Upgrade
	ctx := context.Background()

	if err := svc.Upgrade(ctx, 9); err != nil {
		t.Fatalf("Upgrade 单 ID 失败: %v", err)
	}
	if err := svc.Upgrade(ctx, 9, 10); err != nil {
		t.Fatalf("Upgrade 多 ID 失败: %v", err)
	}
	recs := ts.CallsFor("ac_upgrade")
	if len(recs) != 2 || recs[0].Action != "set" || recs[1].Action != "set" {
		t.Fatalf("action 序列不符: %v", recs)
	}
	if p := paramOf(t, recs[0]); p["id"] != float64(9) {
		t.Fatalf("单 ID 参数应为 id: %v", p)
	}
	if _, ok := paramOf(t, recs[1])["ids"]; !ok {
		t.Fatalf("多 ID 参数应为 ids: %v", paramOf(t, recs[1]))
	}
}

// TestUpgradeError 覆盖固件错误路径。
func TestUpgradeError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ac_upgrade", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "ac_upgrade", Action: action}
	})
	_, err := New(ts.Client()).Upgrade.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
