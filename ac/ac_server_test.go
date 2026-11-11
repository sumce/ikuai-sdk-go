package ac

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// acServerSample 构造 ac_server.show 的分页响应 Data（行字段空表推导）。
func acServerSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"id": 5, "mac": "aa:bb:cc:dd:ee:ff", "ip": "192.168.1.100",
			"name": "会议室AP", "model": "ikuai AP-300", "version": "1.3.0",
			"status": "normal", "uptime": 3600, "fre": "fre_5g",
			"channel": "36", "link_speed": "1000M", "maxload": 64, "load": 12,
			"upload": 1024000, "download": 2048000,
		}},
	}
}

// TestServerList 覆盖 ac_server.show 的成功路径与参数断言（真机探测实证
// TYPE=total,data 形态）。
func TestServerList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{
			name: "nil-opts 默认分页",
			opts: nil,
			want: map[string]any{"TYPE": "total,data", "limit": "0,20"},
		},
		{
			name: "自定义分页",
			opts: &ikuai.ListOptions{Offset: 10, Limit: 5},
			want: map[string]any{"TYPE": "total,data", "limit": "10,5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ac_server", "show", acServerSample())
			page, err := New(ts.Client()).Server.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			r := page.Rows[0]
			if r.ID != 5 || r.MAC != "aa:bb:cc:dd:ee:ff" || r.IP != "192.168.1.100" ||
				r.Name != "会议室AP" || r.Model != "ikuai AP-300" || r.Version != "1.3.0" ||
				r.Status != "normal" || r.Uptime != 3600 || r.Fre != "fre_5g" ||
				r.Channel != "36" || r.MaxLoad != 64 || r.Load != 12 ||
				r.Upload != 1024000 || r.Download != 2048000 {
				t.Fatalf("AP 行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ac_server")[0]), tt.want)
		})
	}
}

// TestServerEditDelete 覆盖 edit/del（写动作，仅离线）。
func TestServerEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ac_server", "edit", map[string]any{})
	ts.SetResponse("ac_server", "del", map[string]any{})
	svc := New(ts.Client()).Server
	ctx := context.Background()

	if err := svc.Edit(ctx, APDevice{ID: 5, Name: "新备注"}); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 5); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}
	recs := ts.CallsFor("ac_server")
	if recs[0].Action != "edit" || recs[1].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s", recs[0].Action, recs[1].Action)
	}
	p := paramOf(t, recs[0])
	if p["id"] != float64(5) || p["name"] != "新备注" {
		t.Fatalf("edit 参数不符: %v", p)
	}
	if p := paramOf(t, recs[1]); p["id"] != float64(5) {
		t.Fatalf("del 参数不符: %v", p)
	}
}

// TestServerError 覆盖固件错误路径。
func TestServerError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ac_server", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "ac_server", Action: action}
	})
	_, err := New(ts.Client()).Server.List(context.Background(), nil)
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95002 {
		t.Fatalf("期望 APIError(95002), got %v", err)
	}
}
