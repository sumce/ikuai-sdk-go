package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDingTalk 真机只读验证钉钉绑定（dingtalk.show 三种 TYPE）。
func TestLiveDingTalk(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).DingTalk

	st, err := svc.BindStatus(ctx)
	if err != nil {
		t.Fatalf("dingtalk.show TYPE=bind_status 失败: %v", err)
	}
	if st.ErrMsg == "" {
		t.Error("bind_status.errmsg 不应为空")
	}
	t.Logf("绑定状态=%d errmsg=%q url=%q", st.Status, st.ErrMsg, st.URL)

	id, err := svc.BindID(ctx)
	if err != nil {
		t.Fatalf("dingtalk.show TYPE=bindid 失败: %v", err)
	}
	t.Logf("绑定 ID=%d", id)

	sup, err := svc.AppBindSupport(ctx)
	if err != nil {
		t.Fatalf("dingtalk.show TYPE=appbind_support 失败: %v", err)
	}
	if sup.SupportDtalk == 0 && sup.SupportWecom == 0 {
		t.Errorf("支持能力不应全为 0: %+v", sup)
	}
	t.Logf("钉钉支持=%d 企业微信支持=%d", sup.SupportDtalk, sup.SupportWecom)
}
