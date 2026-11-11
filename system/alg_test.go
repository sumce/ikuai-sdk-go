package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// algSample 构造真机探测形态的 alg.show 响应 Data（TYPE=data）。
func algSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"sip_ports": "", "tftp_ports": "", "id": 1,
			"support_ftp": 1, "support_tftp": 1, "support_sip": 1,
			"support_h323": 1, "ftp_ports": "",
		}},
	}
}

// TestALGGet 覆盖 alg.show 的成功路径与参数断言。
func TestALGGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("alg", "show", algSample())
	cfg, err := New(ts.Client()).ALG.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.SupportFTP != 1 || cfg.SupportTFTP != 1 ||
		cfg.SupportSIP != 1 || cfg.SupportH323 != 1 || cfg.FTPPorts != "" {
		t.Fatalf("配置行解码不符: %+v", cfg)
	}
	wantParam(t, paramOf(t, ts.CallsFor("alg")[0]),
		map[string]any{"TYPE": ALGTypeData})
}

// TestALGList 覆盖 List 的 TYPE 归一参数。
func TestALGList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("alg", "show", algSample())
	page, err := New(ts.Client()).ALG.List(context.Background(), &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Fatalf("行解码不符: %+v", page)
	}
	wantParam(t, paramOf(t, ts.CallsFor("alg")[0]),
		map[string]any{"TYPE": ALGTypeData, "limit": "0,5"})
}

// TestALGSave 覆盖 alg.save（写动作，burp 抓包实证参数形态，离线）。
func TestALGSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("alg", "save", map[string]any{})
	cfg := ALGConfig{
		ID: 1, SupportFTP: 1, FTPPorts: "", SupportTFTP: 1, TFTPPorts: "",
		SupportSIP: 1, SIPPorts: "", SupportH323: 1,
	}
	if err := New(ts.Client()).ALG.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("alg")[0]
	if rec.Action != "save" {
		t.Fatalf("action = %q, 期望 save（burp 抓包实证）", rec.Action)
	}
	p := paramOf(t, rec)
	// burp 抓包实证：save 按整行全量提交（含 id 与空串端口）。
	for _, k := range []string{"id", "support_ftp", "support_h323", "ftp_ports", "sip_ports", "tftp_ports"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("save 参数缺少键 %q: %v", k, p)
		}
	}
	if len(p) != 8 {
		t.Fatalf("save 参数应为完整 8 字段: %v", p)
	}
}

// TestALGError 覆盖固件错误路径。
func TestALGError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("alg", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95007, ErrMsg: "失败", FuncName: "alg", Action: action}
	})
	if _, err := New(ts.Client()).ALG.Get(context.Background()); err == nil {
		t.Fatal("show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95007 {
		t.Fatalf("期望 APIError(95007), got %v", err)
	}
	if err := New(ts.Client()).ALG.Save(context.Background(), ALGConfig{ID: 1}); err == nil {
		t.Fatal("save 错误应上抛")
	}
}
