package logcenter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestRTCollectAlerts 覆盖 rt_collect.show TYPE=total,data 分页（真机实测
// total=33）与按已读状态过滤（FILTER1="readflag,==,0" 真机实测生效）。
func TestRTCollectAlerts(t *testing.T) {
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
			name: "未读过滤（真机实测形态）",
			opts: &ikuai.ListOptions{Limit: 3,
				Filters: []ikuai.Filter{{Field: "readflag", Op: "==", Value: AlertReadFlagUnread}}},
			want: map[string]any{"TYPE": "total,data", "limit": "0,3", "FILTER1": "readflag,==,0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("rt_collect", "show", map[string]any{
				"total": 33,
				"data": []any{
					map[string]any{"readflag": 0, "id": 1, "timestamp": 1790280007,
						"msgtype": "网络设置", "source": "本地设备(vwan1)", "priority": 2,
						"content": "网段冲突", "detail": "vwan1与lan3网段冲突,网段详情：192.168.1.23/24",
						"solution": "--"},
				},
			})
			page, err := New(ts.Client()).RTCollect.Alerts(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Alerts 失败: %v", err)
			}
			if page.Total != 33 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || r.Timestamp != 1790280007 || r.MsgType != "网络设置" ||
				r.Source != "本地设备(vwan1)" || r.Priority != AlertPriorityWarning ||
				r.Content != "网段冲突" ||
				r.Detail != "vwan1与lan3网段冲突,网段详情：192.168.1.23/24" ||
				r.Solution != "--" || r.ReadFlag != AlertReadFlagUnread {
				t.Fatalf("告警行解码不符: %+v", r)
			}
			wantParam(t, paramOf(t, ts.CallsFor("rt_collect")[0]), tt.want)
		})
	}
}

// TestRTCollectAlertsAll 覆盖 rt_collect.show TYPE=all 聚合区块（真机实测
// total/data/warning_notify）。
func TestRTCollectAlertsAll(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("rt_collect", "show", map[string]any{
		"total": 33,
		"data": []any{
			map[string]any{"readflag": 0, "id": 2, "timestamp": 1790351015,
				"msgtype": "状态监控", "source": "本地设备", "priority": 1,
				"content": "路由端口状态变更", "detail": "端口eth9状态变为down", "solution": "--"},
		},
		"warning_notify": 1,
	})
	out, err := New(ts.Client()).RTCollect.AlertsAll(context.Background())
	if err != nil {
		t.Fatalf("AlertsAll 失败: %v", err)
	}
	if out.Total != 33 || len(out.Rows) != 1 || out.WarningNotify != 1 {
		t.Fatalf("聚合区块解码不符: %+v", out)
	}
	r := out.Rows[0]
	if r.ID != 2 || r.MsgType != "状态监控" || r.Priority != AlertPriorityInfo ||
		r.Content != "路由端口状态变更" || r.Detail != "端口eth9状态变为down" {
		t.Fatalf("告警行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("rt_collect")[0]), map[string]any{"TYPE": "all"})
}

// TestRTCollectWarningNotify 覆盖 rt_collect.show TYPE=warning_notify
// （抓包实证 analysis/calls/rt_collect__show.json，响应 {"warning_notify":1}）。
func TestRTCollectWarningNotify(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("rt_collect", "show", map[string]any{"warning_notify": 1})
	n, err := New(ts.Client()).RTCollect.WarningNotify(context.Background())
	if err != nil {
		t.Fatalf("WarningNotify 失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("warning_notify = %d, 期望 1", n)
	}
	rec := ts.CallsFor("rt_collect")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "warning_notify"})

	// 标志为 0 场景。
	ts.Reset()
	ts.SetResponse("rt_collect", "show", map[string]any{"warning_notify": 0})
	if n, err = New(ts.Client()).RTCollect.WarningNotify(context.Background()); err != nil || n != 0 {
		t.Fatalf("WarningNotify = %d, %v, 期望 0", n, err)
	}
}

// TestRTCollectError 覆盖 rt_collect 的固件错误路径。
func TestRTCollectError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("rt_collect", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)",
			FuncName: "rt_collect", Action: action}
	})
	svc := New(ts.Client()).RTCollect
	ctx := context.Background()
	if _, err := svc.Alerts(ctx, nil); err == nil {
		t.Fatal("rt_collect.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.AlertsAll(ctx); err == nil {
		t.Fatal("rt_collect.show TYPE=all 错误应上抛")
	}
	if _, err := svc.WarningNotify(ctx); err == nil {
		t.Fatal("rt_collect.show TYPE=warning_notify 错误应上抛")
	}
}

// TestAlertConstants 告警级别/已读标记常量与固件值一致。
func TestAlertConstants(t *testing.T) {
	if AlertPriorityInfo != 1 || AlertPriorityWarning != 2 || AlertPriorityFault != 3 {
		t.Fatalf("AlertPriority* 常量不符: %d/%d/%d",
			AlertPriorityInfo, AlertPriorityWarning, AlertPriorityFault)
	}
	if AlertReadFlagUnread != 0 || AlertReadFlagRead != 1 {
		t.Fatalf("AlertReadFlag* 常量不符: %d/%d", AlertReadFlagUnread, AlertReadFlagRead)
	}
}
