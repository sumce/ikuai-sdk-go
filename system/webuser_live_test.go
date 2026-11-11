package system

import (
	"context"
	"os"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveWebUser 真机只读验证 Web 账号管理
// （webuser.show TYPE=mod_passwd）。
//
// ⚠️ 本 func 的 add/edit/del/mod_passwd 均为写动作（改管理员数据），
// 严禁对真机调用，本测试不涉及。
func TestLiveWebUser(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	// 用当前登录账号查询改密标记（只读）。
	username := os.Getenv("IKUAI_USERNAME")
	if username == "" {
		username = "admin"
	}
	flag, err := New(c).WebUser.PasswdFlag(ctx, username)
	if err != nil {
		t.Fatalf("webuser.show TYPE=mod_passwd 失败: %v", err)
	}
	t.Logf("账号 %s 的 mod_passwd 标记=%d", username, flag)
}
