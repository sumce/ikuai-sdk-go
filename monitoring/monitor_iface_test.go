package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// ifaceSample 构造抓包形态的 monitor_iface.show 响应 Data。
func ifaceSample(full bool) map[string]any {
	data := map[string]any{
		"iface_check": []any{map[string]any{
			"id": 1, "interface": "adsl01", "parent_interface": "wan1",
			"ip_addr": "198.51.100.10", "gateway": "198.51.100.1",
			"internet": "PPPOE", "updatetime": "1793736570",
			"auto_switch": "自动", "result": "success",
			"errmsg": "线路检测成功", "comment": "移动私网",
		}},
		"iface_stream": []any{map[string]any{
			"interface": "lan1", "comment": "", "ip_addr": "192.168.9.1",
			"connect_num": "--", "upload": 0, "download": 0,
			"total_up": 0, "total_down": 0, "updropped": 0,
			"downdropped": 0, "uppacked": 0, "downpacked": 0,
		}},
	}
	if !full {
		return data
	}
	data["ether_info"] = map[string]any{
		"eth0": map[string]any{
			"driver": "ice", "type": "FIBRE", "mac": "44:a9:2c:21:c1:4e",
			"link": 1, "speed": 10000, "duplex": 1,
			"model": "Intel E823-L", "interface": "wan1", "lock": 0,
		},
	}
	data["snapshoot_lan"] = []any{map[string]any{
		"id": 1, "comment": "", "interface": "lan1", "bandmode": 0,
		"linkmode": 4, "mac": "44:a9:2c:21:c1:56", "member": []any{"eth8"},
		"ip_addr": "192.168.9.1", "netmask": "255.255.255.0",
	}}
	data["snapshoot_wan"] = []any{map[string]any{
		"id": 1, "comment": "", "interface": "wan1",
		"mac": "44:a9:2c:21:c1:4e", "member": []any{"eth0"},
		"bandmode": 0, "default_route": 0, "internet": 4,
		"ip_addr": "", "netmask": "", "gateway": "", "dns1": "", "dns2": "",
		"count_static": 0, "count_dhcp": 0, "count_pppoe": 13,
		"count_check_fail": 0, "updatetime": 0, "check_res": 0, "errmsg": "",
		"power": "", "isp": "", "imei": "", "qnw": "", "ccid": "",
		"isnr": "", "pcid": "",
	}}
	return data
}

// TestInterfaceMonitor 覆盖 monitor_iface.show 的成功路径与参数断言。
func TestInterfaceMonitor(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		full   bool
		want   map[string]any
		assert func(t *testing.T, m *IfaceMonitor)
	}{
		{
			name: "nil-opts 请求全量 TYPE 且不带 limit",
			opts: nil,
			full: true,
			want: map[string]any{"TYPE": IfaceTypeAll},
			assert: func(t *testing.T, m *IfaceMonitor) {
				if len(m.IfaceCheck) != 1 || m.IfaceCheck[0].Interface != "adsl01" ||
					m.IfaceCheck[0].Internet != "PPPOE" || m.IfaceCheck[0].Updatetime != "1793736570" {
					t.Fatalf("iface_check 解码不符: %+v", m.IfaceCheck)
				}
				if len(m.IfaceStream) != 1 || m.IfaceStream[0].ConnectNum != "--" {
					t.Fatalf("iface_stream 解码不符: %+v", m.IfaceStream)
				}
				ei, ok := m.EtherInfo["eth0"]
				if !ok || ei.Speed != 10000 || ei.Driver != "ice" {
					t.Fatalf("ether_info 解码不符: %+v", m.EtherInfo)
				}
				if len(m.SnapshootLAN) != 1 || len(m.SnapshootLAN[0].Member) != 1 ||
					m.SnapshootLAN[0].Member[0] != "eth8" {
					t.Fatalf("snapshoot_lan 解码不符: %+v", m.SnapshootLAN)
				}
				if len(m.SnapshootWAN) != 1 || m.SnapshootWAN[0].CountPPPoE != 13 ||
					m.SnapshootWAN[0].Internet != 4 {
					t.Fatalf("snapshoot_wan 解码不符: %+v", m.SnapshootWAN)
				}
			},
		},
		{
			name: "opts.Type 指定部分区块",
			opts: &ikuai.ListOptions{Type: IfaceTypeCheck + "," + IfaceTypeStream},
			full: false,
			want: map[string]any{"TYPE": "iface_check,iface_stream"},
			assert: func(t *testing.T, m *IfaceMonitor) {
				if len(m.IfaceCheck) != 1 || len(m.IfaceStream) != 1 {
					t.Fatalf("部分区块解码不符: %+v", m)
				}
				if m.EtherInfo != nil || m.SnapshootLAN != nil || m.SnapshootWAN != nil {
					t.Fatal("未请求的区块应为 nil")
				}
			},
		},
		{
			name: "单区块 TYPE",
			opts: &ikuai.ListOptions{Type: IfaceTypeSnapshoot},
			full: true,
			want: map[string]any{"TYPE": "snapshoot"},
			assert: func(t *testing.T, m *IfaceMonitor) {
				if len(m.SnapshootWAN) != 1 {
					t.Fatalf("snapshoot 解码不符: %+v", m.SnapshootWAN)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("monitor_iface", "show", ifaceSample(tt.full))
			m, err := New(ts.Client()).Interface.Monitor(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Monitor 失败: %v", err)
			}
			tt.assert(t, m)
			calls := ts.CallsFor("monitor_iface")
			if len(calls) != 1 || calls[0].Action != "show" {
				t.Fatalf("应恰好一次 monitor_iface.show, got %v", calls)
			}
			wantParam(t, paramOf(t, calls[0]), tt.want)
		})
	}
}

// TestInterfaceMonitorError 覆盖固件错误路径（Result!=30000）。
func TestInterfaceMonitorError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("monitor_iface", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95001, ErrMsg: "参数缺失", FuncName: "monitor_iface", Action: action}
	})
	_, err := New(ts.Client()).Interface.Monitor(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95001 || ae.ErrMsg != "参数缺失" {
		t.Fatalf("期望 APIError(95001), got %v", err)
	}
}
