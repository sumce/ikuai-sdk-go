package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// vrrpSample 构造真机探测形态的 vrrp_config.show 响应 Data（TYPE=data）。
func vrrpSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"remote_addr": "", "type": 1, "id": 1, "enabled": "no", "prio": 150,
			"method": 2, "domain": "www.baidu.com", "dns": "114.114.114.114",
			"gateway": "www.baidu.com", "ifnames": "", "interval": 3,
			"auto_sync": 0, "single_line": 0, "ignore_wanstatus": 1,
			"interfaces": "", "virtual_ips": "", "ht_iface": "",
		}},
		"version": "3.7.24",
	}
}

// TestVRRPGet 覆盖 vrrp_config.show TYPE=data 的成功路径与参数断言。
func TestVRRPGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("vrrp_config", "show", vrrpSample())
	info, err := New(ts.Client()).VRRP.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if info.Version != "3.7.24" {
		t.Fatalf("version 解码不符: %q", info.Version)
	}
	row := info.Row()
	if row == nil || row.ID != 1 || row.Type != 1 || row.Enabled != "no" ||
		row.Prio != 150 || row.Method != 2 || row.Interval != 3 ||
		row.IgnoreWanstatus != 1 || row.DNS != "114.114.114.114" {
		t.Fatalf("配置行解码不符: %+v", row)
	}
	if row.IsEnabled() {
		t.Fatal("enabled=no 不应判为启用")
	}
	wantParam(t, paramOf(t, ts.CallsFor("vrrp_config")[0]),
		map[string]any{"TYPE": VRRPTypeData})
}

// TestVRRPList 覆盖 List 的默认与自定义分页参数（TYPE 归一）。
func TestVRRPList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts",
			opts: nil,
			want: map[string]any{"TYPE": VRRPTypeData},
		},
		{
			name: "带分页但归一 TYPE",
			opts: &ikuai.ListOptions{Offset: 0, Limit: 5},
			want: map[string]any{"TYPE": VRRPTypeData, "limit": "0,5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("vrrp_config", "show", vrrpSample())
			page, err := New(ts.Client()).VRRP.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if len(page.Rows) != 1 || page.Rows[0].ID != 1 {
				t.Fatalf("行解码不符: %+v", page)
			}
			wantParam(t, paramOf(t, ts.CallsFor("vrrp_config")[0]), tt.want)
		})
	}
}

// TestVRRPSave 覆盖 vrrp_config.save（写动作，离线）。
func TestVRRPSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("vrrp_config", "save", map[string]any{})
	cfg := VRRPConfig{ID: 1, Enabled: EnabledYes, Prio: 150, Interval: 3, Type: 1, Method: 2}
	if err := New(ts.Client()).VRRP.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("vrrp_config")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"].(float64) != 1 || p["enabled"] != "yes" || p["prio"].(float64) != 150 {
		t.Fatalf("save 参数不符: %v", p)
	}
}

// TestVRRPError 覆盖固件错误路径（Result!=30000 → *ikuai.APIError）。
func TestVRRPError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("vrrp_config", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "vrrp_config", Action: action}
	})
	if _, err := New(ts.Client()).VRRP.Get(context.Background()); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := New(ts.Client()).VRRP.Save(context.Background(), VRRPConfig{ID: 1}); err == nil {
		t.Fatal("save 错误应上抛")
	}
}
