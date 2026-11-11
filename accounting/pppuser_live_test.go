package accounting

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveUsers 真机只读验证账号管理（pppuser.show）。账号的增删改
// （写动作）禁止对真机调用。
func TestLiveUsers(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	page, err := svc.Users.List(ctx, nil)
	if err != nil {
		t.Fatalf("pppuser.show 失败: %v", err)
	}
	t.Logf("账号总数=%d 本页行数=%d", page.Total, len(page.Rows))
	for _, u := range page.Rows {
		// 不记录密码明文。
		t.Logf("账号: id=%d username=%s enabled=%s ppptype=%s share=%d packages=%d expires=%d",
			u.ID, u.Username, u.Enabled, u.PPPType, u.Share, u.Packages, u.Expires)
	}
}
