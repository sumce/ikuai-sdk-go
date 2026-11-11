package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveProxy 真机只读验证代理账号子页（pppoe_proxy.show）。真机
// 3.7.24 实证：未授权时返回 30002 Not found funcname(pppoe_proxy)，
// 本测试对两种形态（成功/30002）均视为通过并记录实际行为。
func TestLiveProxy(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Proxy

	page, err := svc.AgentAccounts(ctx, nil)
	if err != nil {
		ae, ok := ikuai.AsAPIError(err)
		if !ok {
			t.Fatalf("pppoe_proxy.show 非预期错误: %v", err)
		}
		if ae.Code != 30002 {
			t.Fatalf("pppoe_proxy.show 非预期结果码: %v", err)
		}
		t.Logf("真机行为（已记录）: 代理账号不可用: %v", err)
		return
	}
	t.Logf("代理账号总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码代理账号行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}
}

// TestLiveProxyOnline 真机只读验证在线账号子页（与代理账号共用
// func_name pppoe_proxy，固件 menu.json 中挂两个子页）。
func TestLiveProxyOnline(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Proxy

	page, err := svc.OnlineAccounts(ctx, nil)
	if err != nil {
		ae, ok := ikuai.AsAPIError(err)
		if !ok {
			t.Fatalf("pppoe_proxy.show 非预期错误: %v", err)
		}
		if ae.Code != 30002 {
			t.Fatalf("pppoe_proxy.show 非预期结果码: %v", err)
		}
		t.Logf("真机行为（已记录）: 在线账号不可用: %v", err)
		return
	}
	t.Logf("在线账号总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, row := range page.Rows {
		m, err := row.Map()
		if err != nil {
			t.Fatalf("解码在线账号行失败: %v", err)
		}
		t.Logf("行: %v", m)
	}
}
