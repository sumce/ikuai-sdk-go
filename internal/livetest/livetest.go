// Package livetest 提供真机联调测试的公共辅助。
//
// 约定：所有业务包的真机测试（只读 smoke 测试 + 带清理的写测试）都
// 通过 Client(t) 获取真机客户端；环境变量未配置时测试自动跳过：
//
//	IKUAI_BASE_URL   例: http://192.168.1.1
//	IKUAI_USERNAME   管理账号
//	IKUAI_PASSWORD   管理密码
//
// 运行：
//
//	IKUAI_BASE_URL=http://192.168.1.1 \
//	IKUAI_USERNAME=xxx IKUAI_PASSWORD=xxx go test ./...
//
// 安全约定：凭据只允许来自环境变量，严禁写入仓库任何文件。
package livetest

import (
	"context"
	"os"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
)

// Available 报告真机测试环境是否就绪。
func Available() bool {
	return os.Getenv("IKUAI_BASE_URL") != "" &&
		os.Getenv("IKUAI_USERNAME") != "" &&
		os.Getenv("IKUAI_PASSWORD") != ""
}

// Client 返回已登录的真机客户端；环境未配置时 t.Skip。
func Client(t *testing.T) *ikuai.Client {
	t.Helper()
	if !Available() {
		t.Skip("跳过真机测试: 未设置 IKUAI_BASE_URL / IKUAI_USERNAME / IKUAI_PASSWORD")
	}
	c, err := ikuai.NewClient(
		os.Getenv("IKUAI_BASE_URL"),
		os.Getenv("IKUAI_USERNAME"),
		os.Getenv("IKUAI_PASSWORD"),
	)
	if err != nil {
		t.Fatalf("构造真机 Client 失败: %v", err)
	}
	if err := c.Login(context.Background()); err != nil {
		t.Fatalf("真机登录失败: %v", err)
	}
	t.Cleanup(func() {
		c.Logout(context.Background())
	})
	return c
}
