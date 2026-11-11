package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestBlacklistList 覆盖 wls_black.show 的成功路径与参数断言
// （行字段空表推导）。
func TestBlacklistList(t *testing.T) {
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
			opts: &ikuai.ListOptions{Offset: 0, Limit: 50, OrderBy: "id", Order: "desc"},
			want: map[string]any{"TYPE": "total,data", "limit": "0,50", "ORDER_BY": "id", "ORDER": "desc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("wls_black", "show", map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"id": 4, "control_mode": WlsModeBlack, "mac": "aa:bb:cc:dd:ee:ff",
					"object": "Office-WiFi", "week": "1111100", "time": "00:00-23:59",
					"comment": "访客设备",
				}},
			})
			page, err := New(ts.Client()).Blacklist.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 4 || r.ControlMode != WlsModeBlack || r.MAC != "aa:bb:cc:dd:ee:ff" ||
				r.Object != "Office-WiFi" || r.Week != "1111100" ||
				r.Time != "00:00-23:59" || r.Comment != "访客设备" {
				t.Fatalf("黑白名单行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("wls_black")[0]), tt.want)
		})
	}
}

// TestBlacklistAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestBlacklistAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("wls_black", "add", map[string]any{})
	ts.SetResponse("wls_black", "edit", map[string]any{})
	ts.SetResponse("wls_black", "del", map[string]any{})
	svc := New(ts.Client()).Blacklist
	ctx := context.Background()

	if _, err := svc.Add(ctx, WlsBlackRule{
		ControlMode: WlsModeBlack, MAC: "aa:bb:cc:dd:ee:ff",
		Object: "Office-WiFi", Week: "1111100", Time: "00:00-23:59",
	}); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	p := paramOf(t, ts.CallsFor("wls_black")[0])
	if p["control_mode"] != WlsModeBlack || p["mac"] != "aa:bb:cc:dd:ee:ff" ||
		p["object"] != "Office-WiFi" {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, hasID := p["id"]; hasID {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, WlsBlackRule{ID: 4, Comment: "x"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 4); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	if err := svc.Delete(ctx, 4, 5); err != nil {
		t.Fatalf("Delete 多 ID 失败: %v", err)
	}
	recs := ts.CallsFor("wls_black")
	if recs[1].Action != "edit" || recs[2].Action != "del" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action)
	}
	if _, ok := paramOf(t, recs[3])["ids"]; !ok {
		t.Fatalf("多 ID del 应使用 ids: %v", paramOf(t, recs[3]))
	}
}

// TestBlacklistAddValidation 覆盖 Add 的参数校验（mac 必填）。
func TestBlacklistAddValidation(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client()).Blacklist
	if _, err := svc.Add(context.Background(), WlsBlackRule{ControlMode: WlsModeBlack}); !errIsInvalidParam(err) {
		t.Fatalf("空 mac 应返回 ErrInvalidParam, got %v", err)
	}
	if len(ts.CallsFor("wls_black")) != 0 {
		t.Fatal("参数校验失败不应发起调用")
	}
}

// TestBlacklistError 覆盖固件错误路径。
func TestBlacklistError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("wls_black", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "wls_black", Action: action}
	})
	_, err := New(ts.Client()).Blacklist.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}

// TestMvlanList 覆盖 wls_mvlan.show 的成功路径与参数断言
// （行字段空表推导）。
func TestMvlanList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("wls_mvlan", "show", map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 7, "mac": "11:22:33:44:55:66", "ssid": "Office-WiFi",
			"vlan_id": 100, "comment": "办公终端",
		}},
	})
	page, err := New(ts.Client()).Mvlan.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 7 || r.MAC != "11:22:33:44:55:66" || r.SSID != "Office-WiFi" ||
		r.VlanID != 100 || r.Comment != "办公终端" {
		t.Fatalf("VLAN 绑定行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("wls_mvlan")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestMvlanAddEditDelete 覆盖 add/edit/del（写动作，仅离线）。
func TestMvlanAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("wls_mvlan", "add", map[string]any{})
	ts.SetResponse("wls_mvlan", "edit", map[string]any{})
	ts.SetResponse("wls_mvlan", "del", map[string]any{})
	svc := New(ts.Client()).Mvlan
	ctx := context.Background()

	if _, err := svc.Add(ctx, WlsMvlanRule{MAC: "11:22:33:44:55:66", VlanID: 100}); err != nil {
		t.Fatalf("Add 失败: %v", err)
	}
	p := paramOf(t, ts.CallsFor("wls_mvlan")[0])
	if p["mac"] != "11:22:33:44:55:66" || p["vlan_id"] != float64(100) {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, hasID := p["id"]; hasID {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}

	if err := svc.Edit(ctx, WlsMvlanRule{ID: 7, VlanID: 200}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 7); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	recs := ts.CallsFor("wls_mvlan")
	if recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action)
	}
	if p := paramOf(t, recs[1]); p["id"] != float64(7) || p["vlan_id"] != float64(200) {
		t.Fatalf("edit 参数不符: %v", p)
	}
}

// TestMvlanAddValidation 覆盖 Add 的参数校验（mac/ssid 至少一项、vlanid 范围）。
func TestMvlanAddValidation(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	svc := New(ts.Client()).Mvlan
	if _, err := svc.Add(context.Background(), WlsMvlanRule{VlanID: 100}); !errIsInvalidParam(err) {
		t.Fatalf("mac/ssid 全空应返回 ErrInvalidParam, got %v", err)
	}
	if _, err := svc.Add(context.Background(), WlsMvlanRule{MAC: "aa", VlanID: 0}); !errIsInvalidParam(err) {
		t.Fatalf("vlan_id=0 应返回 ErrInvalidParam, got %v", err)
	}
	if _, err := svc.Add(context.Background(), WlsMvlanRule{MAC: "aa", VlanID: 4091}); !errIsInvalidParam(err) {
		t.Fatalf("vlan_id=4091 应返回 ErrInvalidParam, got %v", err)
	}
	if len(ts.CallsFor("wls_mvlan")) != 0 {
		t.Fatal("参数校验失败不应发起调用")
	}
}

// TestMvlanError 覆盖固件错误路径。
func TestMvlanError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("wls_mvlan", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "wls_mvlan", Action: action}
	})
	_, err := New(ts.Client()).Mvlan.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
