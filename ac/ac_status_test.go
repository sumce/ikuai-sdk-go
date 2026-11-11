package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// acStatusAllSample 构造真机探测形态的 ac_status.show TYPE=all 响应 Data。
func acStatusAllSample() map[string]any {
	return map[string]any{
		"ac_status": 0,
		"ap_bad_stat": map[string]any{
			"deny_total": 0, "delay_total": 1, "dropptk_total": 2,
			"low_downrate_total": 3, "coverage_total": 4, "channf_total": 5,
		},
		"total_upload": 100, "total_download": 200,
		"ap_chaninfo": []any{}, "ap_chanload": []any{}, "ap_clients": []any{},
		"ap_load": map[string]any{
			"ap_load_low": 0, "ap_load_medium": 1, "ap_load_high": 2, "ap_num": 3,
		},
		"ap_countinfo": []any{}, "ap_event_data": []any{}, "ap_event_total": 7,
		"ap_net_status": map[string]any{
			"score_channel_2g": 10, "score_channel_5g": 20, "score_dropptk_2g": 30,
			"score_dropptk_5g": 40, "score_conn_load": 50, "score_delay": 60,
			"coverage": 70, "delay": 80,
		},
		"ap_puniness": 9,
		"ap_status": map[string]any{
			"ap_count": 11, "ap_online": 12, "ap_offline": 13,
			"ap_perfer_5g": 14, "ap_roaming": 15,
		},
		"clt_status": map[string]any{
			"clt_count": 21, "clt_count_2g": 22, "clt_count_5g": 23,
			"clt_max_online": 24, "clt_active": 25, "clt_inactive": 26,
		},
		"total_count_channel": map[string]any{"apmac_top5": []any{}, "channel_top5": []any{}},
		"total_count_flow":    []any{}, "total_count_low_rssi": []any{},
		"total_count_net_status": map[string]any{
			"score_active_user": 31, "score_chutil_load": 32, "score_channf_load": 33,
			"score_conn_sta": 34, "delay": 35, "coverage": 36, "dropptk": 37,
		},
		"total_count_netquality": []any{}, "total_count_online": []any{},
		"total_online_active": []any{}, "total_online_manufacturer": []any{},
		"total_online_system_type": []any{}, "total_online_wifi_proto": []any{},
		"wphy_usage": map[string]any{
			"chutil_2g": -255, "channf_2g": -255, "chutil_5g": -255, "channf_5g": -255,
		},
	}
}

// TestStatusOverview 覆盖 TYPE=all 聚合的参数与解码（真机探测实证）。
func TestStatusOverview(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_status", "show", acStatusAllSample())
	svc := New(ts.Client()).Status

	ov, err := svc.Overview(context.Background())
	if err != nil {
		t.Fatalf("Overview 失败: %v", err)
	}
	if ov.ACStatus != 0 || ov.APLoad.APLoadMedium != 1 || ov.APEventTotal != 7 ||
		ov.APPuniness != 9 {
		t.Fatalf("标量区块解码不符: %+v", ov)
	}
	if ov.APStatus.APCount != 11 || ov.APStatus.APRoaming != 15 {
		t.Fatalf("ap_status 解码不符: %+v", ov.APStatus)
	}
	if ov.CltStatus.CltCount != 21 || ov.CltStatus.CltCount5G != 23 {
		t.Fatalf("clt_status 解码不符: %+v", ov.CltStatus)
	}
	if ov.APBadStat.DelayTotal != 1 || ov.TotalUpload != 100 || ov.TotalDownload != 200 {
		t.Fatalf("ap_bad_stat 解码不符: %+v", ov.APBadStat)
	}
	if ov.APNetStatus.ScoreChannel5G != 20 || ov.APNetStatus.Coverage != 70 {
		t.Fatalf("ap_net_status 解码不符: %+v", ov.APNetStatus)
	}
	if ov.TotalCountNetStatus.ScoreActiveUser != 31 || ov.TotalCountNetStatus.Dropptk != 37 {
		t.Fatalf("total_count_net_status 解码不符: %+v", ov.TotalCountNetStatus)
	}
	if ov.WPHYUsage.Chutil2G != -255 || ov.WPHYUsage.Channf5G != -255 {
		t.Fatalf("wphy_usage 解码不符: %+v", ov.WPHYUsage)
	}
	if len(ov.APChanInfo) != 0 || len(ov.TotalCountFlow) != 0 {
		t.Fatalf("空数组区块解码不符: %+v", ov)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ac_status")[0]),
		map[string]any{"TYPE": ACStatusTypeAll})
}

// TestStatusSelectors 覆盖各 TYPE 子选择器方法的参数与解码。
func TestStatusSelectors(t *testing.T) {
	tests := []struct {
		name     string
		call     func(s *StatusService) error
		wantType string
	}{
		{
			name: "ControlStatus",
			call: func(s *StatusService) error {
				v, err := s.ControlStatus(context.Background())
				if err != nil {
					return err
				}
				if v != 1 {
					t.Fatalf("ac_status 应为 1, got %d", v)
				}
				return nil
			},
			wantType: ACStatusTypeACStatus,
		},
		{
			name: "APStatus",
			call: func(s *StatusService) error {
				v, err := s.APStatus(context.Background())
				if err != nil {
					return err
				}
				if v.APCount != 3 || v.APOnline != 2 {
					t.Fatalf("ap_status 解码不符: %+v", v)
				}
				return nil
			},
			wantType: ACStatusTypeAPStatus,
		},
		{
			name: "ClientStatus",
			call: func(s *StatusService) error {
				v, err := s.ClientStatus(context.Background())
				if err != nil {
					return err
				}
				if v.CltCount != 5 || v.CltActive != 4 {
					t.Fatalf("clt_status 解码不符: %+v", v)
				}
				return nil
			},
			wantType: ACStatusTypeCltStatus,
		},
		{
			name: "APBadStat",
			call: func(s *StatusService) error {
				v, err := s.APBadStat(context.Background())
				if err != nil {
					return err
				}
				if v.APBadStat.DenyTotal != 6 || v.TotalDownload != 900 {
					t.Fatalf("ap_bad_stat 解码不符: %+v", v)
				}
				return nil
			},
			wantType: ACStatusTypeAPBadStat,
		},
		{
			name: "APNetStatus",
			call: func(s *StatusService) error {
				v, err := s.APNetStatus(context.Background())
				if err != nil {
					return err
				}
				if v.ScoreDelay != 88 {
					t.Fatalf("ap_net_status 解码不符: %+v", v)
				}
				return nil
			},
			wantType: ACStatusTypeAPNetStatus,
		},
		{
			name: "WPHYUsage",
			call: func(s *StatusService) error {
				v, err := s.WPHYUsage(context.Background())
				if err != nil {
					return err
				}
				if v.Chutil5G != 42 {
					t.Fatalf("wphy_usage 解码不符: %+v", v)
				}
				return nil
			},
			wantType: ACStatusTypeWPHYUsage,
		},
		{
			name: "APChanInfo",
			call: func(s *StatusService) error {
				v, err := s.APChanInfo(context.Background())
				if err != nil {
					return err
				}
				if len(v) != 0 {
					t.Fatalf("ap_chaninfo 应为空: %v", v)
				}
				return nil
			},
			wantType: ACStatusTypeAPChanInfo,
		},
		{
			name: "APCountInfo",
			call: func(s *StatusService) error {
				v, err := s.APCountInfo(context.Background())
				if err != nil {
					return err
				}
				if len(v) != 0 {
					t.Fatalf("ap_countinfo 应为空: %v", v)
				}
				return nil
			},
			wantType: ACStatusTypeAPCountInfo,
		},
		{
			name: "APEvents",
			call: func(s *StatusService) error {
				v, err := s.APEvents(context.Background())
				if err != nil {
					return err
				}
				if len(v) != 0 {
					t.Fatalf("ap_event_data 应为空: %v", v)
				}
				return nil
			},
			wantType: ACStatusTypeAPEventData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ac_status", "show", map[string]any{
				"ac_status":    1,
				"ap_status":    map[string]any{"ap_count": 3, "ap_online": 2},
				"clt_status":   map[string]any{"clt_count": 5, "clt_active": 4},
				"ap_bad_stat":  map[string]any{"deny_total": 6},
				"total_upload": 800, "total_download": 900,
				"ap_net_status": map[string]any{"score_delay": 88},
				"wphy_usage":    map[string]any{"chutil_5g": 42},
				"ap_chaninfo":   []any{},
				"ap_countinfo":  []any{},
				"ap_event_data": []any{},
			})
			if err := tt.call(New(ts.Client()).Status); err != nil {
				t.Fatalf("%s 失败: %v", tt.name, err)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ac_status")[0]),
				map[string]any{"TYPE": tt.wantType})
		})
	}
}

// TestStatusOverviewBy 覆盖 OverviewBy 的参数透传（opts 显式 TYPE 优先）。
func TestStatusOverviewBy(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_status", "show", acStatusAllSample())
	opts := &ikuai.ListOptions{Type: ACStatusTypeAPStatus + "," + ACStatusTypeCltStatus}
	if _, err := New(ts.Client()).Status.OverviewBy(context.Background(), opts); err != nil {
		t.Fatalf("OverviewBy 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("ac_status")[0]),
		map[string]any{"TYPE": ACStatusTypeAPStatus + "," + ACStatusTypeCltStatus})
}

// TestStatusError 覆盖固件错误路径（未知 TYPE 真机回 30006）。
func TestStatusError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ac_status", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (bogus)", FuncName: "ac_status", Action: action}
	})
	_, err := New(ts.Client()).Status.Overview(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
