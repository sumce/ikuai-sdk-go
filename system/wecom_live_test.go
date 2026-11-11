package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWeCom 真机只读验证企业微信绑定（wecom.show TYPE=data,bindid）。
func TestLiveWeCom(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	st, err := New(c).WeCom.Status(ctx)
	if err != nil {
		t.Fatalf("wecom.show TYPE=data,bindid 失败: %v", err)
	}
	if st.Enable != EnabledYes && st.Enable != EnabledNo {
		t.Errorf("enable 应为 yes/no, got %q", st.Enable)
	}
	t.Logf("enable=%q sn=%q url=%q status=%d bindid=%d",
		st.Enable, st.SN, st.URL, st.Status, st.BindID)
}
