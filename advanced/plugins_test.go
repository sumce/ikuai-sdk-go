package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// pluginsSample 构造 plugins.show 的响应 Data（真机 3.7.24 未安装插件
// data 为空数组；此处以推导行样例验证解码通道）。
func pluginsSample() map[string]any {
	return map[string]any{
		"data": []any{map[string]any{
			"id": 1, "name": "adguard", "version": "0.1", "status": "running",
		}},
	}
}

// TestPluginsList 覆盖 plugins.show 的成功路径与参数断言（真机实证：
// 本 func 不支持 TYPE=total,data，必须发 TYPE=data 且不带分页）。
func TestPluginsList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts 强制 TYPE=data", nil, map[string]any{"TYPE": "data"}},
		{"opts.Type 显式", &ikuai.ListOptions{Type: "all"}, map[string]any{"TYPE": "all"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("plugins", "show", pluginsSample())
			page, err := New(ts.Client()).Plugins.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if len(page.Rows) != 1 {
				t.Fatalf("行数不符: %+v", page)
			}
			m, err := page.Rows[0].Map()
			if err != nil {
				t.Fatalf("行 Map 失败: %v", err)
			}
			if m["name"] != "adguard" || m["status"] != "running" {
				t.Fatalf("行原始 JSON 不符: %v", m)
			}
			rec := ts.CallsFor("plugins")[0]
			if rec.Action != "show" {
				t.Fatalf("action 不符: %s", rec.Action)
			}
			wantParam(t, paramOf(t, rec), tt.want)
		})
	}
}

// TestPluginsListEmpty 覆盖空插件列表（真机形态）。
func TestPluginsListEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("plugins", "show", map[string]any{"data": []any{}})
	page, err := New(ts.Client()).Plugins.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 0 || len(page.Rows) != 0 {
		t.Fatalf("空表解码不符: %+v", page)
	}
}

// TestPluginsError 覆盖固件错误路径。
func TestPluginsError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("plugins", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "plugins", Action: action}
	})
	_, err := New(ts.Client()).Plugins.List(context.Background(), nil)
	errIsAPICode(t, err, 30006)
}
