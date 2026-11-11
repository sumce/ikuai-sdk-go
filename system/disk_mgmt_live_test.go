package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDisk 真机只读验证磁盘分区（disk_mgmt.show TYPE=data）。
func TestLiveDisk(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Disk

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("disk_mgmt.show 失败: %v", err)
	}
	if len(page.Rows) == 0 {
		t.Fatal("真机应至少有一块磁盘")
	}
	d := page.Rows[0]
	if d.Disk == "" || d.Size <= 0 || len(d.Partition) == 0 {
		t.Errorf("磁盘行字段异常: %+v", d)
	}
	for _, p := range d.Partition {
		if p.Name == "" {
			t.Errorf("分区名为空: %+v", p)
		}
	}
	t.Logf("磁盘=%s 型号=%s 容量=%dB 分区数=%d", d.Disk, d.Model, d.Size, len(d.Partition))

	// 带分页参数的形态。
	page2, err := svc.List(ctx, &ikuai.ListOptions{Limit: 5})
	if err != nil {
		t.Fatalf("disk_mgmt.show 分页形态失败: %v", err)
	}
	if len(page2.Rows) == 0 {
		t.Error("分页形态应有数据")
	}

	first, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("disk_mgmt.Get 失败: %v", err)
	}
	if first == nil || first.Disk != d.Disk {
		t.Errorf("Get 与 List 结果不一致: %+v vs %+v", first, d)
	}
}
