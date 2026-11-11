package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveBackup 真机只读验证备份（backup.show 两种 TYPE）。
// ⚠️ 备份的导出/恢复等写动作严禁对真机调用。
func TestLiveBackup(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Backup

	st, err := svc.Status(ctx)
	if err != nil {
		t.Fatalf("backup.show TYPE=data,disk 失败: %v", err)
	}
	if st.DiskTotal <= 0 || st.DiskAvail <= 0 {
		t.Errorf("备份分区用量异常: %+v", st)
	}
	t.Logf("备份文件数=%d 磁盘 total=%d used=%d avail=%d",
		len(st.Data), st.DiskTotal, st.DiskUsed, st.DiskAvail)

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("backup.show TYPE=total,data 失败: %v", err)
	}
	if page.Total != int64(len(page.Rows)) {
		t.Errorf("total=%d 与行数=%d 不一致", page.Total, len(page.Rows))
	}
	for _, f := range page.Rows {
		m, err := f.Map()
		if err != nil {
			t.Errorf("行 map 解码失败: %v", err)
		}
		t.Logf("备份文件行: %v", m)
	}

	// 第二页形态。
	page2, err := svc.List(ctx, &ikuai.ListOptions{Offset: 0, Limit: 5})
	if err != nil {
		t.Fatalf("backup.show 分页形态失败: %v", err)
	}
	t.Logf("分页形态行数=%d", len(page2.Rows))
}
