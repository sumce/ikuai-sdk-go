package security

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// advancedSample 构造 advanced.show 的响应 Data（真机实测形态）。
func advancedSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"dos_lan_num": 300, "limit_tcp2p": 0, "limit_tcp2p_num": 50,
			"limit_udp2p": 0, "limit_udp2p_num": 300, "tcp_mss": 1,
			"tcp_mss_num": 1400, "id": 1, "noping_lan": 0, "noping_wan": 1,
			"notracert": 0, "hijack_ping": 0, "invalid": 0, "dos_lan": 1,
		}},
	}
}

// TestAdvancedGet 覆盖 advanced.show（Get）的成功路径与参数断言。
func TestAdvancedGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("advanced", "show", advancedSample())
	cfg, err := New(ts.Client()).Advanced.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.NoPingLAN != 0 || cfg.NoPingWAN != 1 ||
		cfg.NoTracert != 0 || cfg.HijackPing != 0 || cfg.Invalid != 0 ||
		cfg.DOSDefenseLAN != 1 || cfg.DOSDefenseNum != 300 ||
		cfg.LimitTCP2P != 0 || cfg.LimitTCP2PNum != 50 ||
		cfg.LimitUDP2P != 0 || cfg.LimitUDP2PNum != 300 ||
		cfg.TCPMSS != 1 || cfg.TCPMSSNum != 1400 {
		t.Fatalf("配置行解码不符: %+v", cfg)
	}
	// Get 内部走 List（opts=nil → 强制 TYPE=data，不带分页）。
	recs := ts.CallsFor("advanced")
	if len(recs) != 1 {
		t.Fatalf("应恰好一次调用: %d", len(recs))
	}
	wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": AdvancedTypeData})
}

// TestAdvancedList 覆盖 List 的 TYPE 归一与分页参数。
func TestAdvancedList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 强制 TYPE=data",
			opts: nil,
			want: map[string]any{"TYPE": AdvancedTypeData},
		},
		{
			name: "带分页但归一 TYPE",
			opts: &ikuai.ListOptions{Limit: 5},
			want: map[string]any{"TYPE": AdvancedTypeData, "limit": "0,5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("advanced", "show", advancedSample())
			page, err := New(ts.Client()).Advanced.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if len(page.Rows) != 1 || page.Rows[0].TCPMSSNum != 1400 {
				t.Fatalf("行解码不符: %+v", page)
			}
			wantParam(t, paramOf(t, ts.CallsFor("advanced")[0]), tt.want)
		})
	}
}

// TestAdvancedSave 覆盖 advanced.save（写动作，仅离线）。
func TestAdvancedSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("advanced", "save", map[string]any{})
	cfg := AdvancedConfig{
		ID: 1, NoPingWAN: 1, DOSDefenseLAN: 1, DOSDefenseNum: 300,
		TCPMSS: 1, TCPMSSNum: 1400,
	}
	if err := New(ts.Client()).Advanced.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("advanced")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"].(float64) != 1 || p["noping_wan"].(float64) != 1 ||
		p["dos_lan"].(float64) != 1 || p["dos_lan_num"].(float64) != 300 ||
		p["tcp_mss"].(float64) != 1 || p["tcp_mss_num"].(float64) != 1400 {
		t.Fatalf("save 参数不符: %v", p)
	}
}

// TestAdvancedError 覆盖固件错误路径（真机实测 TYPE=total 回 30006）。
func TestAdvancedError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("advanced", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "advanced", Action: action}
	})
	svc := New(ts.Client()).Advanced
	if _, err := svc.Get(context.Background()); err == nil {
		t.Fatal("Get 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.List(context.Background(), nil); err == nil {
		t.Fatal("List 应失败")
	}
	if err := svc.Save(context.Background(), AdvancedConfig{ID: 1}); err == nil {
		t.Fatal("Save 应失败")
	}
}
