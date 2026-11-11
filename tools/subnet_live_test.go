package tools

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSubnetNoServerFunc 真机只读验证子网换算无服务端接口的行为
// 记录（subnet.show 空参，真机 3.7.24 实测返回 30002
// Not found funcname(subnet)，控制台在浏览器本地换算）。
func TestLiveSubnetNoServerFunc(t *testing.T) {
	c := livetest.Client(t)
	_, err := New(c).Raw(context.Background(), "subnet", "show", map[string]any{})
	if err == nil {
		t.Fatal("subnet.show 应返回错误（固件无服务端 subnet func）")
	}
	ae, ok := ikuai.AsAPIError(err)
	if !ok {
		t.Fatalf("期望 *ikuai.APIError, got %v", err)
	}
	if ae.Code != 30002 {
		t.Fatalf("期望 Result=30002 (Not found funcname), got %d: %s", ae.Code, ae.ErrMsg)
	}
	t.Logf("subnet.show 按预期返回 Result=%d ErrMsg=%q（本地计算实现成立）", ae.Code, ae.ErrMsg)
}
