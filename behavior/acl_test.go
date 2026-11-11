package behavior

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// aclMACSample 构造真机探测形态的 acl_mac.show 响应 Data（真机 3.7.24
// 实测行结构，含 "comment:1" 特殊键）。
func aclMACSample() map[string]any {
	return map[string]any{
		"total": 7,
		"data": []any{
			map[string]any{"week": "1234567", "comment": "滑子豪", "expires": 0,
				"id": 1, "enabled": "no", "time": "00:00-23:59", "comment:1": "",
				"mac": "6c:4b:90:b0:20:ca"},
		},
	}
}

// TestACLMACList 覆盖 acl_mac.show 分页（真机实测行结构）。
func TestACLMACList(t *testing.T) {
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
			name: "第二页",
			opts: &ikuai.ListOptions{Offset: 20, Limit: 20},
			want: map[string]any{"TYPE": "total,data", "limit": "20,20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("acl_mac", "show", aclMACSample())
			page, err := New(ts.Client()).ACL.MACList(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("MACList 失败: %v", err)
			}
			if page.Total != 7 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
			}
			r := page.Rows[0]
			if r.ID != 1 || r.MAC != "6c:4b:90:b0:20:ca" || r.Comment != "滑子豪" ||
				r.Enabled != EnabledNo || r.Time != "00:00-23:59" || r.Week != "1234567" ||
				r.Expires != 0 || r.Comment1 != "" {
				t.Fatalf("MAC 规则行解码不符: %+v", r)
			}
			if r.IsEnabled() {
				t.Fatal("enabled=no, IsEnabled 应为 false")
			}
			wantParam(t, paramOf(t, ts.CallsFor("acl_mac")[0]), tt.want)
		})
	}
}

// TestACLMACWrites 覆盖 acl_mac add/edit/del 的参数形态（写动作，离线）。
func TestACLMACWrites(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl_mac", "add", map[string]any{})
	ts.SetResponse("acl_mac", "edit", map[string]any{})
	ts.SetResponse("acl_mac", "del", map[string]any{})
	svc := New(ts.Client()).ACL
	ctx := context.Background()

	rowID, err := svc.MACAdd(ctx, MACRule{MAC: "aa:bb:cc:dd:ee:ff", Enabled: EnabledYes,
		Time: "00:00-23:59", Week: "1234567", Comment: "封禁"})
	if err != nil {
		t.Fatalf("MACAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("acl_mac")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"mac", "enabled", "time", "week", "comment"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应含 id: %v", p)
	}

	if err := svc.MACEdit(ctx, MACRule{ID: 1, MAC: "6c:4b:90:b0:20:ca", Enabled: EnabledYes}); err != nil {
		t.Fatalf("MACEdit 失败: %v", err)
	}
	rec = ts.CallsFor("acl_mac")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.MACDelete(ctx, 1); err != nil {
		t.Fatalf("MACDelete 失败: %v", err)
	}
	rec = ts.CallsFor("acl_mac")[2]
	wantParam(t, paramOf(t, rec), map[string]any{"id": 1})

	if err := svc.MACDelete(ctx, 1, 2); err != nil {
		t.Fatalf("MACDelete 多 ID 失败: %v", err)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl_mac")[3]), map[string]any{"ids": []any{1, 2}})
}

// aclSelectorsSample 构造真机探测形态的 TYPE=all 分组选择器区块
// （acl_l7/acl_qq/acl_l2route 实测：ipgroup 为分组名字符串数组）。
func aclSelectorsSample(extra map[string]any) map[string]any {
	data := map[string]any{
		"total":    0,
		"data":     []any{},
		"ipgroup":  []any{"公共无线", "前台", "旭阳手机", "旭阳电脑", "监控", "运营", "销售"},
		"macgroup": []any{},
		"dtgroup":  []any{},
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

// TestACLL7List 覆盖 acl_l7.show 分页（行字段空表推导，待真机验证）。
func TestACLL7List(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl_l7", "show", map[string]any{
		"total": 1,
		"data": []any{
			map[string]any{"id": 1, "enabled": "yes", "action": L7ActionDrop,
				"app_proto": "p2p,video", "src_addr": "172.22.1.0/24",
				"dst_addr": "", "prio": 32, "time": "00:00-23:59",
				"week": "1234567", "comment": "封 P2P"},
		},
	})
	page, err := New(ts.Client()).ACL.L7List(context.Background(), nil)
	if err != nil {
		t.Fatalf("L7List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.Action != L7ActionDrop ||
		r.AppProto != "p2p,video" || r.SrcAddr != "172.22.1.0/24" ||
		r.Prio != 32 || r.Comment != "封 P2P" {
		t.Fatalf("L7 规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl_l7")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})
}

// TestACLL7Selectors 覆盖 acl_l7.show TYPE=all（真机实测区块）。
func TestACLL7Selectors(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl_l7", "show", aclSelectorsSample(nil))
	sel, err := New(ts.Client()).ACL.L7Selectors(context.Background())
	if err != nil {
		t.Fatalf("L7Selectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 7 || sel.IPGroup[0] != "公共无线" || sel.IPGroup[6] != "销售" {
		t.Fatalf("IP 分组选择器解码不符: %v", sel.IPGroup)
	}
	if len(sel.MACGroup) != 0 || len(sel.DTGroup) != 0 {
		t.Fatalf("MAC/时间分组应为空数组: %v %v", sel.MACGroup, sel.DTGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl_l7")[0]), map[string]any{"TYPE": SelectorsTypeAll})
}

// TestACLL7Writes 覆盖 acl_l7 add/edit/del（写动作，离线）。
func TestACLL7Writes(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl_l7", "add", map[string]any{})
	ts.SetResponse("acl_l7", "edit", map[string]any{})
	ts.SetResponse("acl_l7", "del", map[string]any{})
	svc := New(ts.Client()).ACL
	ctx := context.Background()

	rowID, err := svc.L7Add(ctx, L7Rule{Enabled: EnabledYes, Action: L7ActionAccept,
		AppProto: "stream_media", SrcAddr: "公共无线", Prio: 31,
		Time: "00:00-23:59", Week: "1234567"})
	if err != nil {
		t.Fatalf("L7Add 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("acl_l7")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	p := paramOf(t, rec)
	for _, k := range []string{"enabled", "action", "app_proto", "src_addr", "prio", "time", "week"} {
		if _, ok := p[k]; !ok {
			t.Fatalf("add 参数缺少键 %q: %v", k, p)
		}
	}

	if err := svc.L7Edit(ctx, L7Rule{ID: 2, Action: L7ActionDrop, Prio: 32}); err != nil {
		t.Fatalf("L7Edit 失败: %v", err)
	}
	rec = ts.CallsFor("acl_l7")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(2) || p["action"] != "drop" {
		t.Fatalf("edit 参数不符: %v", p)
	}

	if err := svc.L7Delete(ctx, 2, 3); err != nil {
		t.Fatalf("L7Delete 失败: %v", err)
	}
	rec = ts.CallsFor("acl_l7")[2]
	if rec.Action != "del" {
		t.Fatalf("action = %q, 期望 del", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"ids": []any{2, 3}})
}

// TestACLL2Route 覆盖 acl_l2route 的 Get/Selectors/Set（Get 真机实测行）。
func TestACLL2Route(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl_l2route", "show", map[string]any{
		"data": []any{
			map[string]any{"nol2rt": 0, "nol2rt_ip": "", "times": "00:00-23:59",
				"ttl_num": 1, "id": 1},
		},
		"ipgroup": []any{"监控"}, "macgroup": []any{}, "dtgroup": []any{},
	})
	svc := New(ts.Client()).ACL
	ctx := context.Background()

	cfg, err := svc.L2RouteGet(ctx)
	if err != nil {
		t.Fatalf("L2RouteGet 失败: %v", err)
	}
	if cfg == nil || cfg.ID != 1 || cfg.Nol2rt != 0 || cfg.Nol2rtIP != "" ||
		cfg.Times != "00:00-23:59" || cfg.TTLNum != 1 {
		t.Fatalf("配置行解码不符: %+v", cfg)
	}
	if cfg.IsEnabled() {
		t.Fatal("nol2rt=0, IsEnabled 应为 false")
	}
	// Get 为无 TYPE 空参数（真机实测 TYPE=total,data 回 30006）。
	rec := ts.CallsFor("acl_l2route")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})

	sel, err := svc.L2RouteSelectors(ctx)
	if err != nil {
		t.Fatalf("L2RouteSelectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 1 || sel.IPGroup[0] != "监控" {
		t.Fatalf("分组选择器解码不符: %v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl_l2route")[1]), map[string]any{"TYPE": SelectorsTypeAll})

	ts.Reset()
	ts.SetResponse("acl_l2route", "edit", map[string]any{})
	cfg.Nol2rt = 1
	cfg.Nol2rtIP = "172.22.1.0/24"
	if err := svc.L2RouteSet(ctx, *cfg); err != nil {
		t.Fatalf("L2RouteSet 失败: %v", err)
	}
	rec = ts.CallsFor("acl_l2route")[0]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"] != float64(1) || p["nol2rt"] != float64(1) || p["nol2rt_ip"] != "172.22.1.0/24" {
		t.Fatalf("edit 参数不符: %v", p)
	}
}

// TestACLQQ 覆盖 acl_qq 的 List/Selectors/Add/Edit/Delete（行字段空表推导，
// 待真机验证）。
func TestACLQQ(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("acl_qq", "show", map[string]any{
		"total": 1,
		"data": []any{
			map[string]any{"id": 1, "enabled": "yes", "mode": QQModeWhite,
				"qq": "123456789,987654321", "ip_addr": "监控",
				"time": "00:00-23:59", "week": "1234567", "comment": "白名单"},
		},
		"ipgroup": []any{"监控"}, "macgroup": []any{}, "dtgroup": []any{},
	})
	svc := New(ts.Client()).ACL
	ctx := context.Background()

	page, err := svc.QQList(ctx, nil)
	if err != nil {
		t.Fatalf("QQList 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: total=%d rows=%d", page.Total, len(page.Rows))
	}
	r := page.Rows[0]
	if r.ID != 1 || !r.IsEnabled() || r.Mode != QQModeWhite ||
		r.QQ != "123456789,987654321" || r.IPAddr != "监控" {
		t.Fatalf("QQ 规则行解码不符: %+v", r)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl_qq")[0]),
		map[string]any{"TYPE": "total,data", "limit": "0,20"})

	sel, err := svc.QQSelectors(ctx)
	if err != nil {
		t.Fatalf("QQSelectors 失败: %v", err)
	}
	if len(sel.IPGroup) != 1 || sel.IPGroup[0] != "监控" {
		t.Fatalf("分组选择器解码不符: %v", sel.IPGroup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("acl_qq")[1]), map[string]any{"TYPE": SelectorsTypeAll})

	ts.Reset()
	ts.SetResponse("acl_qq", "add", map[string]any{})
	ts.SetResponse("acl_qq", "edit", map[string]any{})
	ts.SetResponse("acl_qq", "del", map[string]any{})

	rowID, err := svc.QQAdd(ctx, QQRule{Enabled: EnabledYes, Mode: QQModeBlack,
		QQ: "123456789", Time: "00:00-23:59", Week: "1234567"})
	if err != nil {
		t.Fatalf("QQAdd 失败: %v", err)
	}
	if rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %d", rowID)
	}
	rec := ts.CallsFor("acl_qq")[0]
	if rec.Action != "add" {
		t.Fatalf("action = %q, 期望 add", rec.Action)
	}
	if p := paramOf(t, rec); p["mode"] != float64(QQModeBlack) || p["qq"] != "123456789" {
		t.Fatalf("add 参数不符: %v", p)
	}

	if err := svc.QQEdit(ctx, QQRule{ID: 1, Mode: QQModeWhite}); err != nil {
		t.Fatalf("QQEdit 失败: %v", err)
	}
	rec = ts.CallsFor("acl_qq")[1]
	if rec.Action != "edit" {
		t.Fatalf("action = %q, 期望 edit", rec.Action)
	}
	if p := paramOf(t, rec); p["id"] != float64(1) {
		t.Fatalf("edit 参数应携带 id=1: %v", p)
	}

	if err := svc.QQDelete(ctx, 1); err != nil {
		t.Fatalf("QQDelete 失败: %v", err)
	}
	rec = ts.CallsFor("acl_qq")[2]
	wantParam(t, paramOf(t, rec), map[string]any{"id": 1})
}

// TestACLQQMode 常量与固件 i18n 值一致。
func TestACLQQMode(t *testing.T) {
	if QQModeBlack != 0 || QQModeWhite != 1 {
		t.Fatalf("QQ 模式常量不符: %d %d", QQModeBlack, QQModeWhite)
	}
	if L7ActionAccept != "accept" || L7ActionDrop != "drop" {
		t.Fatalf("L7 动作常量不符: %q %q", L7ActionAccept, L7ActionDrop)
	}
}

// TestACLError 覆盖固件错误路径。
func TestACLError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("acl_mac", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "acl_mac", Action: action}
	})
	ts.HandleFuncName("acl_l7", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "acl_l7", Action: action}
	})
	ts.HandleFuncName("acl_l2route", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)",
			FuncName: "acl_l2route", Action: action}
	})
	ts.HandleFuncName("acl_qq", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "acl_qq", Action: action}
	})
	svc := New(ts.Client()).ACL
	ctx := context.Background()

	if _, err := svc.MACList(ctx, nil); err == nil {
		t.Fatal("acl_mac.show 错误应上抛")
	}
	if _, err := svc.MACAdd(ctx, MACRule{}); err == nil {
		t.Fatal("acl_mac.add 错误应上抛")
	}
	if err := svc.MACDelete(ctx, 1); err == nil {
		t.Fatal("acl_mac.del 错误应上抛")
	}
	if _, err := svc.L7List(ctx, nil); err == nil {
		t.Fatal("acl_l7.show 错误应上抛")
	}
	if _, err := svc.L7Selectors(ctx); err == nil {
		t.Fatal("acl_l7.show TYPE=all 错误应上抛")
	}
	if err := svc.L7Delete(ctx, 1); err == nil {
		t.Fatal("acl_l7.del 错误应上抛")
	}
	if _, err := svc.L2RouteGet(ctx); err == nil {
		t.Fatal("acl_l2route.show 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := svc.L2RouteSet(ctx, L2RouteConfig{}); err == nil {
		t.Fatal("acl_l2route.edit 错误应上抛")
	}
	if _, err := svc.QQList(ctx, nil); err == nil {
		t.Fatal("acl_qq.show 错误应上抛")
	}
	if err := svc.QQDelete(ctx, 1); err == nil {
		t.Fatal("acl_qq.del 错误应上抛")
	}
}
