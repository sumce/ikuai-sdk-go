package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveFile 真机只读验证文件管理（file_mgmt.show path=/）。
func TestLiveFile(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	rows, err := New(c).File.ListDir(ctx, "/", nil)
	if err != nil {
		t.Fatalf("file_mgmt.show path=/ 失败: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("挂载根目录应有条目")
	}
	for _, e := range rows {
		if e.FName == "" {
			t.Errorf("条目名为空: %+v", e)
		}
	}
	t.Logf("挂载根目录条目 %d 个，首条: name=%s type=%d size=%d",
		len(rows), rows[0].FName, rows[0].StType, rows[0].StSize)

	// 固件怪癖：非根目录返回 {"data":{}}，应得到空列表而非错误。
	sub, err := New(c).File.ListDir(ctx, "/etc", nil)
	if err != nil {
		t.Fatalf("file_mgmt.show path=/etc 不应报错: %v", err)
	}
	t.Logf("path=/etc 条目数=%d（固件怪癖：子目录返回空对象）", len(sub))
}
