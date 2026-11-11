package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestBackupStatus 覆盖 backup.show TYPE=data,disk 的成功路径与参数断言。
func TestBackupStatus(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("backup", "show", map[string]any{
		"data":       []any{},
		"disk_total": 44454, "disk_used": 1417, "disk_avail": 39453,
	})
	st, err := New(ts.Client()).Backup.Status(context.Background())
	if err != nil {
		t.Fatalf("Status 失败: %v", err)
	}
	if len(st.Data) != 0 {
		t.Fatalf("备份文件应为空: %+v", st.Data)
	}
	if st.DiskTotal != 44454 || st.DiskUsed != 1417 || st.DiskAvail != 39453 {
		t.Fatalf("磁盘用量解码不符: %+v", st)
	}
	wantParam(t, paramOf(t, ts.CallsFor("backup")[0]),
		map[string]any{"TYPE": BackupTypeDataDisk})
}

// TestBackupList 覆盖标准分页形态（TYPE=total,data）与行 Raw 解码。
func TestBackupList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("backup", "show", map[string]any{
		"total": 1,
		"data":  []any{map[string]any{"id": 1, "name": "config-20261110.conf"}},
	})
	page, err := New(ts.Client()).Backup.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	m, err := page.Rows[0].Map()
	if err != nil {
		t.Fatalf("Map 失败: %v", err)
	}
	if m["name"] != "config-20261110.conf" {
		t.Fatalf("行 map 解码不符: %v", m)
	}
	wantParam(t, paramOf(t, ts.CallsFor("backup")[0]),
		map[string]any{"TYPE": BackupTypeDataTotal, "limit": "0,20"})
}

// TestBackupError 覆盖固件错误路径。
func TestBackupError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("backup", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95010, ErrMsg: "失败", FuncName: "backup", Action: action}
	})
	if _, err := New(ts.Client()).Backup.Status(context.Background()); err == nil {
		t.Fatal("status 错误应上抛")
	}
	if _, err := New(ts.Client()).Backup.List(context.Background(), nil); err == nil {
		t.Fatal("list 错误应上抛")
	}
}
