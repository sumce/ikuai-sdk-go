package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// lanipSample 构造抓包形态的 monitor_lanip.show 响应 Data（TYPE=data,total）。
func lanipSample() map[string]any {
	return map[string]any{
		"total": 164,
		"data": []any{map[string]any{
			"ac_gid": 0, "apmac": "", "apname": "", "auth_type": 0, "bssid": "",
			"client_device": "vivo", "client_model": "vivo%20T4%205G",
			"client_type": "Android", "client_typeid": 101004299,
			"comment": "fff", "connect_num": 45, "download": 8019,
			"downrate": "", "dtalk_name": "", "frequencies": "",
			"hostname": "", "id": 104, "ip_addr": "172.20.3.4",
			"ip_addr_int": 2886992644, "link_addr": "",
			"mac": "54:16:51:f6:31:03", "ppptype": "", "reject": 0,
			"signal": "", "ssid": "", "timestamp": 1793933460,
			"total_down": 18217806113, "total_up": 2334347929,
			"upload": 5908, "uprate": "", "uptime": "2026-11-06 10:51:00",
			"username": "", "webid": 0,
		}},
	}
}

// TestLanIPList 覆盖 monitor_lanip.show 分页列表的成功路径与参数断言。
func TestLanIPList(t *testing.T) {
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
			opts: &ikuai.ListOptions{Offset: 20, Limit: 20},
			want: map[string]any{
				"TYPE":      "total,data",
				"limit":     "20,20",
				"ORDER_BY":  LanIPOrderByIPAddr,
				"orderType": LanIPOrderTypeIP,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("monitor_lanip", "show", lanipSample())
			page, err := New(ts.Client()).LanIP.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 164 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.IPAddr != "172.20.3.4" || r.IPAddrInt != 2886992644 ||
				r.ConnectNum != 45 || r.TotalDown != 18217806113 {
				t.Fatalf("终端行解码不符: %+v", r)
			}
			if got := r.ClientModelDecoded(); got != "vivo T4 5G" {
				t.Fatalf("ClientModelDecoded = %q, 期望 vivo T4 5G", got)
			}
			wantParam(t, paramOf(t, ts.CallsFor("monitor_lanip")[0]), tt.want)
		})
	}
}

// TestLanIPSearch 覆盖关键字搜索路径（FINDS/KEYWORDS 断言）。
func TestLanIPSearch(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_lanip", "show", lanipSample())
	page, err := New(ts.Client()).LanIP.Search(context.Background(), "172.22", nil)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	if page.Total != 164 || len(page.Rows) != 1 {
		t.Fatalf("搜索结果解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_lanip")[0]), map[string]any{
		"TYPE":      "total,data",
		"limit":     "0,20",
		"ORDER_BY":  LanIPOrderByIPAddr,
		"orderType": LanIPOrderTypeIP,
		"FINDS":     LanIPFindsDefault,
		"KEYWORDS":  "172.22",
	})
}

// TestLanIPSearchEmptyKeywords 空关键字返回 ErrInvalidParam。
func TestLanIPSearchEmptyKeywords(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	_, err := New(ts.Client()).LanIP.Search(context.Background(), "", nil)
	if err == nil {
		t.Fatal("空关键字应报错")
	}
}

// TestLanIPInterfaces 覆盖 TYPE=interface 接口列表路径。
func TestLanIPInterfaces(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("monitor_lanip", "show", map[string]any{
		"interface": []any{
			[]any{"all"},
			[]any{"lan1"},
			[]any{"lan2", "内网接入三层交换机"},
		},
	})
	ifaces, err := New(ts.Client()).LanIP.Interfaces(context.Background())
	if err != nil {
		t.Fatalf("Interfaces 失败: %v", err)
	}
	if len(ifaces) != 3 || ifaces[0][0] != "all" || ifaces[2][1] != "内网接入三层交换机" {
		t.Fatalf("接口列表解码不符: %v", ifaces)
	}
	wantParam(t, paramOf(t, ts.CallsFor("monitor_lanip")[0]),
		map[string]any{"TYPE": LanIPTypeInterface})
}

// TestLanIPError 覆盖固件错误路径。
func TestLanIPError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("monitor_lanip", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "monitor_lanip", Action: action}
	})
	_, err := New(ts.Client()).LanIP.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
