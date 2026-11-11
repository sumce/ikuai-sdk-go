package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// basicSample 构造抓包形态的 basic.show 响应 Data。
func basicSample() map[string]any {
	return map[string]any{
		"timestamp":    int64(1794330079),
		"support_fast": 0,
		"data": []any{map[string]any{
			"id": 1, "hostname": "iKuai", "language": 1, "time_zone": 8,
			"time_zone_full": "0800", "switch_nat": 2, "switch_dpi": 1,
			"switch_ntp": 1, "switch_ntpd": 0, "switch_ntpserver": 0,
			"ntpserver_list": "", "ntp_sync_cycle": 60, "link_mode": 0,
			"fast_nat": 0, "lan_nat": 1, "listenport": "lan1", "backport": "wan1",
		}},
		"interface": []any{"lan1", "lan2", "wan1"},
	}
}

// TestBasicGet 覆盖 basic.show 的成功路径与参数断言。
func TestBasicGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("basic", "show", basicSample())
	info, err := New(ts.Client()).Basic.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if info.Timestamp != 1794330079 || info.SupportFast != 0 {
		t.Fatalf("区块解码不符: %+v", info)
	}
	if len(info.Interface) != 3 || info.Interface[0] != "lan1" {
		t.Fatalf("接口列表解码不符: %v", info.Interface)
	}
	row := info.Row()
	if row == nil || row.Hostname != "iKuai" || row.Language != 1 ||
		row.TimeZone != 8 || row.TimeZoneFull != "0800" || row.SwitchNat != 2 ||
		row.Listenport != "lan1" || row.Backport != "wan1" {
		t.Fatalf("设置行解码不符: %+v", row)
	}
	wantParam(t, paramOf(t, ts.CallsFor("basic")[0]),
		map[string]any{"TYPE": BasicTypeAll})
}

// TestBasicSetLanguage 覆盖 basic.set_language（写动作，离线）。
func TestBasicSetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		lang    int64
		wantErr bool
	}{
		{"中文(抓包值)", 1, false},
		{"非法零值", 0, true},
		{"非法负值", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("basic", "set_language", map[string]any{})
			err := New(ts.Client()).Basic.SetLanguage(context.Background(), tt.lang)
			if tt.wantErr {
				if err == nil {
					t.Fatal("非法 lang 应报错")
				}
				return
			}
			if err != nil {
				t.Fatalf("SetLanguage 失败: %v", err)
			}
			rec := ts.CallsFor("basic")[0]
			if rec.Action != "set_language" {
				t.Fatalf("action = %q, 期望 set_language", rec.Action)
			}
			wantParam(t, paramOf(t, rec), map[string]any{"language": tt.lang})
		})
	}
}

// TestBasicSyncTime 覆盖 basic.sync_time（写动作，离线）。
func TestBasicSyncTime(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("basic", "sync_time", map[string]any{})
	if err := New(ts.Client()).Basic.SyncTime(context.Background()); err != nil {
		t.Fatalf("SyncTime 失败: %v", err)
	}
	rec := ts.CallsFor("basic")[0]
	if rec.Action != "sync_time" {
		t.Fatalf("action = %q, 期望 sync_time", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestBasicError 覆盖固件错误路径。
func TestBasicError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("basic", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95001, ErrMsg: "失败", FuncName: "basic", Action: action}
	})
	if _, err := New(ts.Client()).Basic.Get(context.Background()); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95001 {
		t.Fatalf("期望 APIError(95001), got %v", err)
	}
	if err := New(ts.Client()).Basic.SetLanguage(context.Background(), 1); err == nil {
		t.Fatal("set_language 错误应上抛")
	}
}
