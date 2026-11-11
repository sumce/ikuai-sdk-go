package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// onlineSample 构造 ac_online_clt.show 的分页响应 Data（列名真机探测实证）。
func onlineSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 12, "mac": "11:22:33:44:55:66", "ip_addr": "172.22.1.88",
			"ip_addr_int": 2887123288, "apmac": "aa:bb:cc:dd:ee:ff",
			"apname": "会议室AP", "bssid": "aa:bb:cc:dd:ee:f1",
			"ssid": "Office-WiFi", "signal": "-52", "frequencies": "5G",
			"upload": 153600, "download": 1048576, "uprate": "866M",
			"downrate": "866M", "total_up": 37472227, "total_down": 40669280,
			"connect_time": 3600, "timestamp": 1794334466, "last_time": 1794334400,
			"manufacturer": "vivo", "comment": "测试终端",
		}},
	}
}

// TestOnlineClientsList 覆盖 ac_online_clt.show 的成功路径与参数断言
// （列名经真机 ORDER_BY/FINDS 探测实证）。
func TestOnlineClientsList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页（按 ip_addr_int 排序）",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20", "ORDER_BY": OnlineClientsOrderByIPAddr},
		},
		{
			name: "自定义分页",
			opts: &ikuai.ListOptions{Offset: 40, Limit: 20},
			want: map[string]any{"TYPE": "total,data", "limit": "40,20", "ORDER_BY": OnlineClientsOrderByIPAddr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ac_online_clt", "show", onlineSample())
			page, err := New(ts.Client()).OnlineClients.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 12 || r.MAC != "11:22:33:44:55:66" || r.IPAddr != "172.22.1.88" ||
				r.IPAddrInt != 2887123288 || r.APMAC != "aa:bb:cc:dd:ee:ff" ||
				r.APName != "会议室AP" || r.BSSID != "aa:bb:cc:dd:ee:f1" ||
				r.SSID != "Office-WiFi" || r.Signal != "-52" || r.Frequencies != "5G" ||
				r.Upload != 153600 || r.Download != 1048576 ||
				r.TotalUp != 37472227 || r.TotalDown != 40669280 ||
				r.ConnectTime != 3600 || r.Manufacturer != "vivo" {
				t.Fatalf("终端行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ac_online_clt")[0]), tt.want)
		})
	}
}

// TestOnlineClientsSearch 覆盖 KEYWORDS/FINDS 搜索参数（真机探测实证
// FINDS 字段集合法）。
func TestOnlineClientsSearch(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_online_clt", "show", onlineSample())
	page, err := New(ts.Client()).OnlineClients.Search(context.Background(), "aa:bb", nil)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}
	if page.Total != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ac_online_clt")[0]), map[string]any{
		"TYPE":     "total,data",
		"limit":    "0,20",
		"ORDER_BY": OnlineClientsOrderByIPAddr,
		"FINDS":    OnlineClientsFindsDefault,
		"KEYWORDS": "aa:bb",
	})
}

// TestOnlineClientsSearchValidation 覆盖 Search 的参数校验（keywords 必填）。
func TestOnlineClientsSearchValidation(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	if _, err := New(ts.Client()).OnlineClients.Search(context.Background(), "", nil); !errIsInvalidParam(err) {
		t.Fatalf("空 keywords 应返回 ErrInvalidParam, got %v", err)
	}
	if len(ts.CallsFor("ac_online_clt")) != 0 {
		t.Fatal("参数校验失败不应发起调用")
	}
}

// TestOnlineClientsError 覆盖固件错误路径（含真机怪癖：非法 ORDER_BY
// 返回损坏 JSON，核心包解码失败向上传错）。
func TestOnlineClientsError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ac_online_clt", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "ac_online_clt", Action: action}
	})
	_, err := New(ts.Client()).OnlineClients.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
