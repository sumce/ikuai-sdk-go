package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// auditSample 构造真机探测形态的 audit.show 响应 Data（无 TYPE，param={}；
// 真机 3.7.24 实测行结构与观测值）。
func auditSample() map[string]any {
	return map[string]any{
		"data": []any{
			map[string]any{
				"total_size": 60579692, "use_size": 82700,
				"open_url_record": 0, "open_im_record": 0,
				"open_terminal_record": 0, "open_appid_record": 0,
			},
		},
	}
}

// TestAuditGet 覆盖 audit.show（无 TYPE）配置行解码与参数断言。
func TestAuditGet(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantNil  bool
		wantFunc func(cfg *AuditConfig) error
	}{
		{
			name:    "真机实测单行",
			data:    auditSample(),
			wantNil: false,
			wantFunc: func(cfg *AuditConfig) error {
				if cfg.TotalSize != 60579692 || cfg.UseSize != 82700 {
					t.Errorf("容量字段解码不符: %+v", cfg)
				}
				if cfg.IsURLRecordEnabled() || cfg.IsIMRecordEnabled() ||
					cfg.IsTerminalRecordEnabled() || cfg.IsAppidRecordEnabled() {
					t.Errorf("观测全 0, 开关应全关: %+v", cfg)
				}
				return nil
			},
		},
		{
			name:    "空 data",
			data:    map[string]any{"data": []any{}},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("audit", "show", tt.data)
			cfg, err := New(ts.Client()).Audit.Get(context.Background())
			if err != nil {
				t.Fatalf("Get 失败: %v", err)
			}
			if tt.wantNil {
				if cfg != nil {
					t.Fatalf("空 data 应返回 nil, got %+v", cfg)
				}
			} else if cfg == nil {
				t.Fatal("配置行不应为 nil")
			} else if tt.wantFunc != nil {
				if err := tt.wantFunc(cfg); err != nil {
					t.Fatal(err)
				}
			}
			rec := ts.CallsFor("audit")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %q, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), map[string]any{})
		})
	}
}

// TestAuditInfo 覆盖 audit.show TYPE=all（clean_type + data 区块，真机实测）。
func TestAuditInfo(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit", "show", map[string]any{
		"clean_type": []any{"all", "im", "url", "terminal"},
		"data": []any{map[string]any{"total_size": 100, "use_size": 1,
			"open_url_record": 1, "open_im_record": 0, "open_terminal_record": 1, "open_appid_record": 0}},
	})
	info, err := New(ts.Client()).Audit.Info(context.Background())
	if err != nil {
		t.Fatalf("Info 失败: %v", err)
	}
	if len(info.CleanType) != 4 || info.CleanType[0] != AuditCleanTypeAll ||
		info.CleanType[1] != AuditCleanTypeIM || info.CleanType[2] != AuditCleanTypeURL ||
		info.CleanType[3] != AuditCleanTypeTerminal {
		t.Fatalf("clean_type 解码不符: %v", info.CleanType)
	}
	row := info.Row()
	if row == nil || !row.IsURLRecordEnabled() || row.IsIMRecordEnabled() ||
		!row.IsTerminalRecordEnabled() || row.IsAppidRecordEnabled() {
		t.Fatalf("配置行解码不符: %+v", row)
	}
	wantParam(t, paramOf(t, ts.CallsFor("audit")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestAuditSave 覆盖 audit.save 参数形态（写动作，离线）。
func TestAuditSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("audit", "save", map[string]any{})
	cfg := AuditConfig{TotalSize: 100, UseSize: 1, OpenURLRecord: 1}
	if err := New(ts.Client()).Audit.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("audit")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"total_size", "use_size", "open_url_record",
		"open_im_record", "open_terminal_record", "open_appid_record"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("save 参数缺少键 %q: %v", k, p)
		}
	}
	if len(p) != 6 {
		t.Fatalf("save 参数键数 = %d, 期望 6: %v", len(p), p)
	}
}

// TestRecordFreeGetSet 覆盖 record_free_setting 的 Get/Set 参数形态
// （空表推导，待真机验证；真机 30002）。
func TestRecordFreeGetSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("record_free_setting", "show", map[string]any{
		"ip_white_list": []any{
			map[string]any{"id": 1, "ip": "172.22.1.0/24", "explain": "测试段", "group": "办公"},
		},
		"mac_white_list": []any{
			map[string]any{"id": 2, "mac": "00:0c:29:3b:11:af", "explain": "虚拟机", "group": ""},
		},
	})
	ts.SetResponse("record_free_setting", "save", map[string]any{})
	svc := New(ts.Client()).Audit
	ctx := context.Background()

	cfg, err := svc.RecordFreeGet(ctx)
	if err != nil {
		t.Fatalf("RecordFreeGet 失败: %v", err)
	}
	if len(cfg.IPWhiteList) != 1 || cfg.IPWhiteList[0].IP != "172.22.1.0/24" ||
		cfg.IPWhiteList[0].Explain != "测试段" {
		t.Fatalf("IP 白名单解码不符: %+v", cfg.IPWhiteList)
	}
	if len(cfg.MACWhiteList) != 1 || cfg.MACWhiteList[0].MAC != "00:0c:29:3b:11:af" {
		t.Fatalf("MAC 白名单解码不符: %+v", cfg.MACWhiteList)
	}
	rec := ts.CallsFor("record_free_setting")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})

	if err := svc.RecordFreeSet(ctx, *cfg); err != nil {
		t.Fatalf("RecordFreeSet 失败: %v", err)
	}
	rec = ts.CallsFor("record_free_setting")[1]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	var p struct {
		IPWhiteList  []RecordFreeIP  `json:"ip_white_list"`
		MACWhiteList []RecordFreeMAC `json:"mac_white_list"`
	}
	if err := json.Unmarshal(rec.Param, &p); err != nil {
		t.Fatalf("解析 save 参数失败: %v", err)
	}
	if len(p.IPWhiteList) != 1 || len(p.MACWhiteList) != 1 {
		t.Fatalf("save 参数应携带两张白名单: %+v", p)
	}
}

// TestAuditError 覆盖固件错误路径（show/save → *ikuai.APIError）。
func TestAuditError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	// 真机形态：record_free_setting 在 3.7.24 回 30002 Not found funcname。
	ts.HandleFuncName("record_free_setting", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname(record_free_setting)",
			FuncName: "record_free_setting", Action: action}
	})
	ts.HandleFuncName("audit", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "audit", Action: action}
	})
	svc := New(ts.Client()).Audit
	ctx := context.Background()

	if _, err := svc.Get(ctx); err == nil {
		t.Fatal("audit.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
	if _, err := svc.Info(ctx); err == nil {
		t.Fatal("audit.show TYPE=all 错误应上抛")
	}
	if err := svc.Save(ctx, AuditConfig{}); err == nil {
		t.Fatal("audit.save 错误应上抛")
	}
	if _, err := svc.RecordFreeGet(ctx); err == nil {
		t.Fatal("record_free_setting.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
	if err := svc.RecordFreeSet(ctx, RecordFreeConfig{}); err == nil {
		t.Fatal("record_free_setting.save 错误应上抛")
	}
}
