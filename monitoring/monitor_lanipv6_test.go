package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// lanipv6Sample 构造抓包形态的 monitor_lanipv6.show 响应 Data。
func lanipv6Sample() map[string]any {
	return map[string]any{
		"total": 3,
		"data": []any{map[string]any{
			"ac_gid": 0, "apmac": "", "apname": "", "auth_type": 0, "bssid": "",
			"client_device": "", "client_model": "", "client_type": "Unknown",
			"client_typeid": 0, "comment": "", "connect_num": 12,
			"download": 1024, "downrate": "", "dtalk_name": "",
			"frequencies": "", "hostname": "host-v6", "id": 7,
			"ip_addr": "fe80::1", "ip_addr_int": 0, "link_addr": "",
			"mac": "00:0c:29:3b:11:af", "ppptype": "", "signal": "",
			"ssid": "", "timestamp": 1793933460, "total_down": 2048,
			"total_up": 4096, "upload": 2048, "uprate": "",
			"uptime": "2026-11-06 10:51:00", "username": "", "webid": 0,
		}},
	}
}

// TestLanIPv6List 覆盖 monitor_lanipv6.show 的成功路径与参数断言。
func TestLanIPv6List(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页",
			opts: nil,
			want: map[string]any{
				"TYPE":      "total,data",
				"limit":     "0,20",
				"ORDER_BY":  LanIPOrderByIPAddr,
				"orderType": LanIPOrderTypeIP,
			},
		},
		{
			name: "第二页",
			opts: &ikuai.ListOptions{Offset: 40, Limit: 40},
			want: map[string]any{
				"TYPE":      "total,data",
				"limit":     "40,40",
				"ORDER_BY":  LanIPOrderByIPAddr,
				"orderType": LanIPOrderTypeIP,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("monitor_lanipv6", "show", lanipv6Sample())
			page, err := New(ts.Client()).LanIPv6.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 3 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.IPAddr != "fe80::1" || r.Hostname != "host-v6" || r.ID != 7 {
				t.Fatalf("IPv6 终端行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("monitor_lanipv6")[0]), tt.want)
		})
	}
}

// TestLanIPv6Interfaces 覆盖 TYPE=interface 接口列表路径。
func TestLanIPv6Interfaces(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_lanipv6", "show", map[string]any{
		"interface": []any{
			[]any{"all"},
			[]any{"lan4", "李冰卧室"},
		},
	})
	ifaces, err := New(ts.Client()).LanIPv6.Interfaces(context.Background())
	if err != nil {
		t.Fatalf("Interfaces 失败: %v", err)
	}
	if len(ifaces) != 2 || ifaces[1][1] != "李冰卧室" {
		t.Fatalf("接口列表解码不符: %v", ifaces)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_lanipv6")[0]),
		map[string]any{"TYPE": LanIPv6TypeInterface})
}

// TestLanIPv6Error 覆盖固件错误路径。
func TestLanIPv6Error(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("monitor_lanipv6", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95003, ErrMsg: "失败", FuncName: "monitor_lanipv6", Action: action}
	})
	_, err := New(ts.Client()).LanIPv6.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95003 {
		t.Fatalf("期望 APIError(95003), got %v", err)
	}
}
