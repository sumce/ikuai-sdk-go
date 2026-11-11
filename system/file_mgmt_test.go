package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// fileSample 构造真机探测形态的 file_mgmt.show 响应 Data（path="/"）。
func fileSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"st_size": int64(63233441792), "st_inode": 2,
			"f_name": "A", "st_mtime": int64(1792006588), "st_type": 0,
		}},
	}
}

// TestFileListDir 覆盖 file_mgmt.show 的成功路径与参数断言。
func TestFileListDir(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("file_mgmt", "show", fileSample())
	rows, err := New(ts.Client()).File.ListDir(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("ListDir 失败: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("应解码 1 个条目: %+v", rows)
	}
	e := rows[0]
	if e.FName != "A" || e.StSize != 63233441792 || e.StInode != 2 ||
		e.StMtime != 1792006588 || e.StType != 0 || !e.IsDir() {
		t.Fatalf("条目解码不符: %+v", e)
	}
	wantParam(t, paramOf(t, ts.CallsFor("file_mgmt")[0]), map[string]any{
		"path":     "/",
		"TYPE":     FileTypeData,
		"ORDER_BY": "",
		"ORDER":    "",
	})
}

// TestFileListDirEmptyObject 覆盖固件怪癖：非根目录返回 {"data":{}} 时
// 应返回空列表且不报错。
func TestFileListDirEmptyObject(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("file_mgmt", "show", map[string]any{"data": map[string]any{}})
	rows, err := New(ts.Client()).File.ListDir(context.Background(), "/etc", nil)
	if err != nil {
		t.Fatalf("空对象 data 不应报错: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("应返回空列表: %+v", rows)
	}
}

// TestFileListDirEmptyPath 空路径返回 ErrInvalidParam。
func TestFileListDirEmptyPath(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	_, err := New(ts.Client()).File.ListDir(context.Background(), "", nil)
	if err == nil {
		t.Fatal("空 path 应报错")
	}
}

// TestFileListDirError 覆盖固件错误路径。
func TestFileListDirError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("file_mgmt", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95003, ErrMsg: "目录不存在", FuncName: "file_mgmt", Action: action}
	})
	_, err := New(ts.Client()).File.ListDir(context.Background(), "/not-exist", nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95003 {
		t.Fatalf("期望 APIError(95003), got %v", err)
	}
}
