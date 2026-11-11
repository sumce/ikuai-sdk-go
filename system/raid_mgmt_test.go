package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestRaidList 覆盖 raid_mgmt.show 的默认与自定义分页参数。
func TestRaidList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			// 控制台形态：TYPE=data（burp 抓包另带 limit，此处为最小参数）。
			name: "nil-opts 控制台形态",
			opts: nil,
			want: map[string]any{"TYPE": RaidTypeData},
		},
		{
			name: "显式 total,data（真机同样接受）",
			opts: &ikuai.ListOptions{Type: RaidTypeDataTotal, Limit: 5},
			want: map[string]any{"TYPE": RaidTypeDataTotal, "limit": "0,5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("raid_mgmt", "show", map[string]any{"total": 0, "data": []any{}})
			page, err := New(ts.Client()).Raid.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 0 || len(page.Rows) != 0 {
				t.Fatalf("空 RAID 应为 0 行: %+v", page)
			}
			wantParam(t, paramOf(t, ts.CallsFor("raid_mgmt")[0]), tt.want)
		})
	}
}

// TestRaidListRows 覆盖行解码（Raw 形态）。
func TestRaidListRows(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("raid_mgmt", "show", map[string]any{
		"total": 1,
		"data":  []any{map[string]any{"id": 1, "name": "raid0", "level": "raid0"}},
	})
	page, err := New(ts.Client()).Raid.List(context.Background(), nil)
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
	if m["name"] != "raid0" || m["level"] != "raid0" {
		t.Fatalf("行 map 解码不符: %v", m)
	}
}

// TestRaidError 覆盖固件错误路径。
func TestRaidError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("raid_mgmt", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95004, ErrMsg: "失败", FuncName: "raid_mgmt", Action: action}
	})
	_, err := New(ts.Client()).Raid.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95004 {
		t.Fatalf("期望 APIError(95004), got %v", err)
	}
}
