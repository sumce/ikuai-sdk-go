package logcenter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestMessagesList 覆盖 ikmessages.show TYPE=total,data 分页（真机实测）与
// 按已读状态过滤（FILTER1="read,==,0" 抓包实证）。
func TestMessagesList(t *testing.T) {
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
			name: "未读过滤（抓包实证字段 read）",
			opts: &ikuai.ListOptions{Limit: 5,
				Filters: []ikuai.Filter{{Field: "read", Op: "==", Value: MessageReadNo}}},
			want: map[string]any{"TYPE": "total,data", "limit": "0,5", "FILTER1": "read,==,0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("ikmessages", "show", map[string]any{
				"total": 2,
				"data": []any{
					map[string]any{"id": 1, "timestamp": 1790279675,
						"title": "固件升级通知", "content": "3.7.25 已发布", "read": 0},
				},
			})
			page, err := New(ts.Client()).Messages.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 2 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || r.Timestamp != 1790279675 || r.Title != "固件升级通知" ||
				r.Content != "3.7.25 已发布" || r.Read != MessageReadNo {
				t.Fatalf("消息行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("ikmessages")[0]), tt.want)
		})
	}
}

// TestMessagesUnreadCount 覆盖 ikmessages.show TYPE=total + FILTER1=
// "read,==,0"（抓包实证 analysis/calls/ikmessages__show.json，响应 {"total":0}）。
func TestMessagesUnreadCount(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ikmessages", "show", map[string]any{"total": 0})
	n, err := New(ts.Client()).Messages.UnreadCount(context.Background())
	if err != nil {
		t.Fatalf("UnreadCount 失败: %v", err)
	}
	if n != 0 {
		t.Fatalf("未读数 = %d, 期望 0", n)
	}
	rec := ts.CallsFor("ikmessages")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	// 参数形态必须与抓包完全一致：{"TYPE":"total","FILTER1":"read,==,0"}。
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "total", "FILTER1": "read,==,0"})

	// 未读数非 0 场景。
	ts.Reset()
	ts.SetResponse("ikmessages", "show", map[string]any{"total": 7})
	if n, err = New(ts.Client()).Messages.UnreadCount(context.Background()); err != nil || n != 7 {
		t.Fatalf("UnreadCount = %d, %v, 期望 7", n, err)
	}
}

// TestMessagesStart 覆盖 ikmessages.start（抓包实证 param {} → Success）。
// 离线测试；真机禁调（写动作）。
func TestMessagesStart(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ikmessages", "start", nil)
	if err := New(ts.Client()).Messages.Start(context.Background()); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	rec := ts.CallsFor("ikmessages")[0]
	if rec.Action != "start" {
		t.Fatalf("action = %q, 期望 start", rec.Action)
	}
	// 抓包实证 param 为空对象 {}。
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestMessagesMarkAllRead 覆盖 ikmessages.set {"read":1}（推测形态，
// 无抓包实证，待真机验证；仅离线测试）。
func TestMessagesMarkAllRead(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ikmessages", "set", nil)
	if err := New(ts.Client()).Messages.MarkAllRead(context.Background()); err != nil {
		t.Fatalf("MarkAllRead 失败: %v", err)
	}
	rec := ts.CallsFor("ikmessages")[0]
	if rec.Action != "set" {
		t.Fatalf("action = %q, 期望 set", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"read": float64(MessageReadYes)})
}

// TestMessagesError 覆盖 ikmessages 的固件错误路径。
func TestMessagesError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ikmessages", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30002, ErrMsg: "Not found funcname(ikmessages)",
			FuncName: "ikmessages", Action: action}
	})
	svc := New(ts.Client()).Messages
	ctx := context.Background()
	if _, err := svc.List(ctx, nil); err == nil {
		t.Fatal("ikmessages.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30002 {
		t.Fatalf("期望 APIError(30002), got %v", err)
	}
	if _, err := svc.UnreadCount(ctx); err == nil {
		t.Fatal("ikmessages.show TYPE=total 错误应上抛")
	}
	if err := svc.Start(ctx); err == nil {
		t.Fatal("ikmessages.start 错误应上抛")
	}
	if err := svc.MarkAllRead(ctx); err == nil {
		t.Fatal("ikmessages.set 错误应上抛")
	}
}

// TestMessageReadConstants 已读标记常量与固件值一致（未读=0 抓包实证）。
func TestMessageReadConstants(t *testing.T) {
	if MessageReadNo != 0 {
		t.Fatalf("MessageReadNo = %d, 期望 0", MessageReadNo)
	}
	if MessageReadYes != 1 {
		t.Fatalf("MessageReadYes = %d, 期望 1", MessageReadYes)
	}
}
