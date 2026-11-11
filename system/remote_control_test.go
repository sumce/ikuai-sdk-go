package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// remoteControlSample 构造真机探测形态的 remote_control.show 响应 Data。
func remoteControlSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"ftp_port": 21, "sshd_port": 22, "sshd_passwd": "",
			"http_port": 182, "https_port": 443, "force_https": 0,
			"ftp_access": 1, "id": 1, "open_telnetd": 0, "open_wanweb": 2,
			"open_sshd": 0, "open_ftp": 0,
		}},
	}
}

// TestRemoteControlGet 覆盖 remote_control.show 的成功路径与参数断言。
func TestRemoteControlGet(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts", nil, map[string]any{"TYPE": RemoteControlTypeData}},
		{"带分页但归一 TYPE", &ikuai.ListOptions{Limit: 5},
			map[string]any{"TYPE": RemoteControlTypeData, "limit": "0,5"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("remote_control", "show", remoteControlSample())
			cfg, err := New(ts.Client()).RemoteControl.Get(context.Background())
			if err != nil {
				t.Fatalf("Get 失败: %v", err)
			}
			if cfg == nil || cfg.ID != 1 || cfg.HTTPPort != 182 || cfg.HTTPSPort != 443 ||
				cfg.SSHDPort != 22 || cfg.FTPPort != 21 || cfg.OpenWANWeb != 2 ||
				cfg.ForceHTTPS != 0 || cfg.FTPAccess != 1 {
				t.Fatalf("配置行解码不符: %+v", cfg)
			}
			// Get 内部走 List。
			if len(ts.CallsFor("remote_control")) != 1 {
				t.Fatalf("应恰好一次调用: %d", len(ts.CallsFor("remote_control")))
			}
			// Get 不带分页 → 仅 TYPE。
			if tt.name == "nil-opts" {
				wantParam(t, paramOf(t, ts.CallsFor("remote_control")[0]), tt.want)
			}
		})
	}
}

// TestRemoteControlList 覆盖 List 参数断言。
func TestRemoteControlList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("remote_control", "show", remoteControlSample())
	page, err := New(ts.Client()).RemoteControl.List(context.Background(), &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(page.Rows) != 1 || page.Rows[0].HTTPPort != 182 {
		t.Fatalf("行解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("remote_control")[0]),
		map[string]any{"TYPE": RemoteControlTypeData, "limit": "0,5"})
}

// TestRemoteControlSave 覆盖 remote_control.save（写动作，离线）。
func TestRemoteControlSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("remote_control", "save", map[string]any{})
	cfg := RemoteControlConfig{ID: 1, HTTPPort: 182, HTTPSPort: 443, OpenWANWeb: 2}
	if err := New(ts.Client()).RemoteControl.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("remote_control")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"].(float64) != 1 || p["http_port"].(float64) != 182 || p["open_wanweb"].(float64) != 2 {
		t.Fatalf("save 参数不符: %v", p)
	}
}

// TestRemoteControlError 覆盖固件错误路径。
func TestRemoteControlError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("remote_control", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "remote_control", Action: action}
	})
	if _, err := New(ts.Client()).RemoteControl.Get(context.Background()); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := New(ts.Client()).RemoteControl.Save(context.Background(), RemoteControlConfig{ID: 1}); err == nil {
		t.Fatal("save 错误应上抛")
	}
}
