package accounting

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// noticeContentB64 是一段 base64 编码的 HTML 通知正文（真机实证形态）。
func noticeContentB64() string {
	return base64.StdEncoding.EncodeToString([]byte("<!DOCTYPE html><html><body>到期提醒</body></html>"))
}

// TestNoticesGet 覆盖四个通知 func 的 show TYPE=data 成功路径与参数断言。
func TestNoticesGet(t *testing.T) {
	fixtures := []struct {
		funcName string
		data     map[string]any
		assert   func(t *testing.T, s *NoticesService)
	}{
		{
			funcName: "notice_temp",
			data: map[string]any{"data": []any{map[string]any{
				"id": 1, "content": "", "red_url": "http://www.baidu.com",
				"timeout": 60, "pppuser": 0, "listuser": 0, "ipgroup": "",
			}}},
			assert: func(t *testing.T, s *NoticesService) {
				cfg, err := s.Temp(context.Background())
				if err != nil {
					t.Fatalf("Temp 失败: %v", err)
				}
				if cfg == nil || cfg.ID != 1 || cfg.RedURL != "http://www.baidu.com" ||
					cfg.Timeout != 60 || cfg.PPPUser != 0 || cfg.ListUser != 0 {
					t.Fatalf("字段解码不符: %+v", cfg)
				}
				if c, err := cfg.DecodeContent(); err != nil || c != "" {
					t.Fatalf("DecodeContent 不符: %q %v", c, err)
				}
			},
		},
		{
			funcName: "notice_cycle",
			data: map[string]any{
				"total": 1,
				"data": []any{map[string]any{
					"id": 2, "content": noticeContentB64(), "week": "1234567",
					"time": "09:00", "red_url": "http://www.baidu.com",
					"timeout": 60, "pppuser": 1, "listuser": 0, "ipgroup": "",
				}},
			},
			assert: func(t *testing.T, s *NoticesService) {
				page, err := s.CycleList(context.Background(), nil)
				if err != nil {
					t.Fatalf("CycleList 失败: %v", err)
				}
				if page.Total != 1 || len(page.Rows) != 1 || page.Rows[0].Week != "1234567" {
					t.Fatalf("周期通知解码不符: %+v", page)
				}
				c, err := page.Rows[0].DecodeContent()
				if err != nil || c != "<!DOCTYPE html><html><body>到期提醒</body></html>" {
					t.Fatalf("DecodeContent 不符: %q %v", c, err)
				}
			},
		},
		{
			funcName: "notice_remind",
			data: map[string]any{"data": []any{map[string]any{
				"id": 1, "enabled": "yes", "content": noticeContentB64(),
				"red_url": "http://www.baidu.com", "days": 5,
				"time": "08:00,16:00", "timeout": 60,
			}}},
			assert: func(t *testing.T, s *NoticesService) {
				cfg, err := s.Remind(context.Background())
				if err != nil {
					t.Fatalf("Remind 失败: %v", err)
				}
				if cfg == nil || !cfg.IsEnabled() || cfg.Days != 5 ||
					cfg.Time != "08:00,16:00" || cfg.Timeout != 60 {
					t.Fatalf("字段解码不符: %+v", cfg)
				}
				if c, err := cfg.DecodeContent(); err != nil || c == "" {
					t.Fatalf("DecodeContent 不符: %q %v", c, err)
				}
			},
		},
		{
			funcName: "notice_expires",
			data: map[string]any{"data": []any{map[string]any{
				"id": 1, "content": noticeContentB64(), "whiteip": "", "whitedomain": "",
			}}},
			assert: func(t *testing.T, s *NoticesService) {
				cfg, err := s.Expires(context.Background())
				if err != nil {
					t.Fatalf("Expires 失败: %v", err)
				}
				if cfg == nil || cfg.ID != 1 || cfg.WhiteIP != "" || cfg.WhiteDomain != "" {
					t.Fatalf("字段解码不符: %+v", cfg)
				}
			},
		},
	}
	for _, fx := range fixtures {
		t.Run(fx.funcName, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse(fx.funcName, "show", fx.data)
			fx.assert(t, New(ts.Client()).Notices)
			recs := ts.CallsFor(fx.funcName)
			if fx.funcName == "notice_cycle" {
				// 周期通知是分页列表：TYPE=total,data + limit。
				wantParam(t, paramOf(t, recs[0]),
					map[string]any{"TYPE": "total,data", "limit": "0,20"})
				return
			}
			if len(recs) != 1 || recs[0].Action != "show" {
				t.Fatalf("调用记录不符: %+v", recs)
			}
			wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": cfgTypeData})
		})
	}
}

// TestNoticesSet 覆盖三个配置 func 的 save 写动作参数断言（仅离线）。
func TestNoticesSet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, fn := range []string{"notice_temp", "notice_remind", "notice_expires"} {
		ts.SetResponse(fn, "save", map[string]any{})
	}
	svc := New(ts.Client()).Notices
	ctx := context.Background()
	if err := svc.SetTemp(ctx, NoticeTempConfig{ID: 1, Timeout: 30, ListUser: 1}); err != nil {
		t.Fatalf("SetTemp 失败: %v", err)
	}
	if err := svc.SetRemind(ctx, NoticeRemindConfig{ID: 1, Enabled: EnabledYes, Days: 3, Time: "09:00"}); err != nil {
		t.Fatalf("SetRemind 失败: %v", err)
	}
	if err := svc.SetExpires(ctx, NoticeExpiresConfig{ID: 1, WhiteIP: "1.2.3.4"}); err != nil {
		t.Fatalf("SetExpires 失败: %v", err)
	}
	for _, fn := range []string{"notice_temp", "notice_remind", "notice_expires"} {
		recs := ts.CallsFor(fn)
		if len(recs) != 1 || recs[0].Action != "save" {
			t.Fatalf("%s 应有一次 save 调用, got %+v", fn, recs)
		}
		p := paramOf(t, recs[0])
		if p["id"] != float64(1) {
			t.Fatalf("%s save param 缺少 id: %v", fn, p)
		}
	}
}

// TestNoticesCycleCRUD 覆盖周期通知 add/edit/del 的参数断言（仅离线）。
func TestNoticesCycleCRUD(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("notice_cycle", "add", map[string]any{})
	ts.SetResponse("notice_cycle", "edit", map[string]any{})
	ts.SetResponse("notice_cycle", "del", map[string]any{})
	svc := New(ts.Client()).Notices
	ctx := context.Background()

	if _, err := svc.CycleAdd(ctx, NoticeCycle{Week: "1,3,5", Time: "10:00", PPPUser: 1}); err != nil {
		t.Fatalf("CycleAdd 失败: %v", err)
	}
	rec := ts.CallsFor("notice_cycle")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	if p["week"] != "1,3,5" || p["time"] != "10:00" || p["pppuser"] != float64(1) {
		t.Fatalf("add param 字段不符: %v", p)
	}

	if err := svc.CycleEdit(ctx, NoticeCycle{ID: 4, Week: "1234567"}); err != nil {
		t.Fatalf("CycleEdit 失败: %v", err)
	}
	rec = ts.CallsFor("notice_cycle")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	p = paramOf(t, rec)
	if p["id"] != float64(4) || p["week"] != "1234567" {
		t.Fatalf("edit param 字段不符: %v", p)
	}

	if err := svc.CycleDelete(ctx, 4); err != nil {
		t.Fatalf("CycleDelete 失败: %v", err)
	}
	rec = ts.CallsFor("notice_cycle")[2]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"id": float64(4)})
}

// TestNoticesError 覆盖固件错误路径。
func TestNoticesError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("notice_temp", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total,data)", FuncName: "notice_temp", Action: action}
	})
	ts.HandleFuncName("notice_cycle", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95000, ErrMsg: "失败", FuncName: "notice_cycle", Action: action}
	})
	svc := New(ts.Client()).Notices
	ctx := context.Background()
	if _, err := svc.Temp(ctx); err == nil {
		t.Fatal("Temp 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := svc.SetTemp(ctx, NoticeTempConfig{ID: 1}); err == nil {
		t.Fatal("SetTemp 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.CycleList(ctx, nil); err == nil {
		t.Fatal("CycleList 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95000 {
		t.Fatalf("期望 APIError(95000), got %v", err)
	}
	if _, err := svc.CycleAdd(ctx, NoticeCycle{}); err == nil {
		t.Fatal("CycleAdd 应失败")
	}
	if err := svc.CycleEdit(ctx, NoticeCycle{ID: 1}); err == nil {
		t.Fatal("CycleEdit 应失败")
	}
	if err := svc.CycleDelete(ctx, 1); err == nil {
		t.Fatal("CycleDelete 应失败")
	}
}
