package accounting

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestProxyAccounts 覆盖 pppoe_proxy.show 两个子页方法的成功路径与参数断言
// （真机 func 未授权时为 30002，见 live 测试）。
func TestProxyAccounts(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, svc *ProxyService)
	}{
		{
			name: "代理账号列表",
			run: func(t *testing.T, svc *ProxyService) {
				page, err := svc.AgentAccounts(context.Background(), nil)
				if err != nil {
					t.Fatalf("AgentAccounts 失败: %v", err)
				}
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["agent_account"] != "adsl01" {
					t.Fatalf("行内容不符: %v", m)
				}
			},
		},
		{
			name: "在线账号列表",
			run: func(t *testing.T, svc *ProxyService) {
				page, err := svc.OnlineAccounts(context.Background(),
					&ikuai.ListOptions{Offset: 0, Limit: 5})
				if err != nil {
					t.Fatalf("OnlineAccounts 失败: %v", err)
				}
				if page.Total != 1 || len(page.Rows) != 1 {
					t.Fatalf("行数不符: %+v", page)
				}
				m, err := page.Rows[0].Map()
				if err != nil {
					t.Fatalf("Map() 失败: %v", err)
				}
				if m["agent_account"] != "adsl01" {
					t.Fatalf("行内容不符: %v", m)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("pppoe_proxy", "show", map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"agent_account": "adsl01", "agent_password": "pw",
					"account_to_be_dialed": "adsl00",
				}},
			})
			tt.run(t, New(ts.Client()).Proxy)
			recs := ts.CallsFor("pppoe_proxy")
			if len(recs) != 1 || recs[0].Action != "show" {
				t.Fatalf("调用记录不符: %+v", recs)
			}
		})
	}
}

// TestProxyError 覆盖固件错误路径（含真机未授权 30002 形态）。
func TestProxyError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("pppoe_proxy", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname(pppoe_proxy)", FuncName: "pppoe_proxy", Action: action}
	})
	svc := New(ts.Client()).Proxy
	ctx := context.Background()
	if _, err := svc.AgentAccounts(ctx, nil); err == nil {
		t.Fatal("AgentAccounts 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
	if _, err := svc.OnlineAccounts(ctx, nil); err == nil {
		t.Fatal("OnlineAccounts 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
}
