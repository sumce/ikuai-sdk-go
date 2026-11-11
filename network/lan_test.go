package network

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// lanSample 构造真机形态的 lan.show 响应 Data（TYPE=support_wisp,total,data）。
func lanSample() map[string]any {
	return map[string]any{
		"support_wisp": 0,
		"total":        1,
		"data": []any{map[string]any{
			"id": 1, "name": "lan4", "comment": "李冰卧室",
			"ip_mask": "192.168.9.1/255.255.255.0", "bandeth": "eth3",
			"bandif": "44:a9:2c:21:c1:51", "bandflag": "0", "bandmode": 0,
			"mac": "", "speed": 0, "duplex": 1, "linkmode": 4,
			"lan_visit": 1, "policy": 0,
		}},
	}
}

// TestLANList 覆盖 lan.show 的成功路径与参数断言。
func TestLANList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认 support_wisp,total,data",
			opts: nil,
			want: map[string]any{"TYPE": LANTypeSupportWispDataTotal, "limit": "0,20"},
		},
		{
			name: "显式 TYPE 被尊重",
			opts: &ikuai.ListOptions{Type: LANTypeDataTotal},
			want: map[string]any{"TYPE": LANTypeDataTotal},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("lan", "show", lanSample())
			page, err := New(ts.Client()).LAN.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.SupportWisp != 0 || page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.Name != "lan4" || r.IPMask != "192.168.9.1/255.255.255.0" ||
				r.BandETH != "eth3" || r.Linkmode != 4 || r.LANVisit != 1 {
				t.Fatalf("接口行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("lan")[0]), tt.want)
		})
	}
}

// TestLANInterface 覆盖单接口回显路径（id + TYPE 参数断言）。
func TestLANInterface(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("lan", "show", lanSample())
	iface, err := New(ts.Client()).LAN.Interface(context.Background(), 1)
	if err != nil {
		t.Fatalf("Interface 失败: %v", err)
	}
	if iface == nil || iface.ID != 1 || iface.Name != "lan4" {
		t.Fatalf("单接口解码不符: %+v", iface)
	}
	wantParam(t, paramOf(t, ts.CallsFor("lan")[0]),
		map[string]any{"id": float64(1), "TYPE": LANTypeSupportWispDataTotal})

	// 非法 id 返回 ErrInvalidParam。
	if _, err := New(ts.Client()).LAN.Interface(context.Background(), 0); err == nil {
		t.Fatal("id<=0 应报错")
	}
}

// TestLANSet 覆盖写动作参数形态（仅离线，禁止对真机调用）。
func TestLANSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("lan", "edit", map[string]any{})
	err := New(ts.Client()).LAN.Set(context.Background(), LANInterface{ID: 1, Name: "lan4", IPMask: "192.168.9.1/255.255.255.0"})
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}
	rec := ts.CallsFor("lan")[0]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"] != float64(1) || p["name"] != "lan4" {
		t.Fatalf("edit 参数不符: %v", p)
	}
}

// TestLANError 覆盖固件错误路径。
func TestLANError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("lan", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "lan", Action: action}
	})
	_, err := New(ts.Client()).LAN.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 || ae.FuncName != "lan" {
		t.Fatalf("期望 APIError(95002,lan), got %v", err)
	}
}
