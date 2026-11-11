package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// diskSample 构造真机探测形态的 disk_mgmt.show 响应 Data（TYPE=data）。
func diskSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"type": "dos", "model": "KINGSTON OM8P0S3", "disk": "sda",
			"system": 1, "block_size": 512, "creating": 0,
			"size": int64(128035676160),
			"partition": []any{
				map[string]any{
					"mounted": map[string]any{
						"mt_avail": "40399872", "mt_purpose": "0", "mt_name": "",
						"mt_total": "45520896", "mt_used": "1451008", "mt_uses": "3",
					},
					"name": "sda3", "formating": 0, "type": 0,
					"size": int64(52428800),
					"filesys": map[string]any{
						"fs_type": "ext4", "fs_uuid": "ca05f54d-70d7-4422-81f8-743533208df4",
					},
				},
				map[string]any{
					"mounted": map[string]any{
						"mt_avail": "59988058112", "mt_purpose": "1", "mt_name": "A",
						"mt_total": "63233441792", "mt_used": "24576", "mt_uses": "0",
					},
					"name": "sda6", "formating": 0, "type": 2,
					"size": int64(64571661824),
					"filesys": map[string]any{
						"fs_type": "ext4", "fs_uuid": "9831899a-1c63-4627-953f-0de41ba1c8a3",
					},
				},
			},
		}},
	}
}

// TestDiskList 覆盖 disk_mgmt.show 的成功路径与参数断言。
func TestDiskList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts", nil, map[string]any{"TYPE": DiskTypeData}},
		{"带分页但归一 TYPE", &ikuai.ListOptions{Limit: 5},
			map[string]any{"TYPE": DiskTypeData, "limit": "0,5"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("disk_mgmt", "show", diskSample())
			page, err := New(ts.Client()).Disk.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if len(page.Rows) != 1 {
				t.Fatalf("应解码 1 块磁盘: %+v", page)
			}
			d := page.Rows[0]
			if d.Disk != "sda" || d.Type != "dos" || d.System != 1 ||
				d.BlockSize != 512 || d.Size != 128035676160 || len(d.Partition) != 2 {
				t.Fatalf("磁盘行解码不符: %+v", d)
			}
			p0 := d.Partition[0]
			if p0.Name != "sda3" || p0.Type != 0 || p0.Size != 52428800 ||
				p0.Filesys.FSType != "ext4" || p0.Mounted.MTUses != "3" {
				t.Fatalf("分区行解码不符: %+v", p0)
			}
			if d.Partition[1].Mounted.MTName != "A" || d.Partition[1].Mounted.MTPurpose != "1" {
				t.Fatalf("数据分区解码不符: %+v", d.Partition[1])
			}
			wantParam(t, paramOf(t, ts.CallsFor("disk_mgmt")[0]), tt.want)
		})
	}
}

// TestDiskGet 覆盖首磁盘便捷方法。
func TestDiskGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("disk_mgmt", "show", diskSample())
	d, err := New(ts.Client()).Disk.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if d == nil || d.Disk != "sda" {
		t.Fatalf("首磁盘解码不符: %+v", d)
	}
}

// TestDiskGetEmpty 空列表时 Get 返回 nil 且不报错。
func TestDiskGetEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("disk_mgmt", "show", map[string]any{"data": []any{}})
	d, err := New(ts.Client()).Disk.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if d != nil {
		t.Fatalf("空列表应返回 nil, got %+v", d)
	}
}

// TestDiskError 覆盖固件错误路径。
func TestDiskError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("disk_mgmt", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "disk_mgmt", Action: action}
	})
	_, err := New(ts.Client()).Disk.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
