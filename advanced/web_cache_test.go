package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// webCacheSample 构造 ik_web_cache.show TYPE=data 的响应 Data（真机
// 3.7.24 探测实证形态，行数据截取真机样例）。
func webCacheSample() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"line_count": 2,
			"line": []any{
				map[string]any{"id": 1, "name": "adsl03", "up": 0, "down": 0,
					"weight": 1, "comment": "", "enabled": "no", "status": "idle"},
				map[string]any{"id": 2, "name": "vwan1", "up": 100, "down": 200,
					"weight": 3, "comment": "备用", "enabled": "yes", "status": "invalid"},
			},
			"port":          650,
			"service_count": 2,
			"service": []any{
				map[string]any{"cache_min_uses": 1, "enabled": "no", "type": "video",
					"max_size": 10, "inactive_time": 10080},
				map[string]any{"cache_min_uses": 2, "enabled": "yes", "type": "data",
					"max_size": 20, "inactive_time": 43200},
			},
		},
	}
}

// webCacheStatusSample 构造 ik_web_cache.show TYPE=status 的响应 Data
// （真机 3.7.24 探测实证形态）。
func webCacheStatusSample() map[string]any {
	return map[string]any{
		"status": []any{
			map[string]any{"service_type": "video", "cache_status": "idle",
				"err": "没有挂载外部存储(Ext文件系统)"},
			map[string]any{"service_type": "data", "cache_status": "running", "err": ""},
		},
	}
}

// TestWebCacheGet 覆盖 ik_web_cache.show TYPE=data 的成功路径与参数断言
// （真机实证：本 func 不支持 TYPE=total，必须发 TYPE=data 且不带分页）。
func TestWebCacheGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_web_cache", "show", webCacheSample())
	cfg, err := New(ts.Client()).WebCache.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if cfg.LineCount != 2 || cfg.Port != 650 || cfg.ServiceCount != 2 {
		t.Fatalf("配置顶层字段解码不符: %+v", cfg)
	}
	if len(cfg.Lines) != 2 || len(cfg.Services) != 2 {
		t.Fatalf("子列表解码不符: %+v", cfg)
	}
	l := cfg.Lines[1]
	if l.ID != 2 || l.Name != "vwan1" || l.Up != 100 || l.Down != 200 ||
		l.Weight != 3 || l.Comment != "备用" || !l.IsEnabled() || l.Status != "invalid" {
		t.Fatalf("线路行解码不符: %+v", l)
	}
	s := cfg.Services[1]
	if s.Type != "data" || !s.IsEnabled() || s.MaxSize != 20 || s.InactiveTime != 43200 || s.CacheMinUses != 2 {
		t.Fatalf("服务配置解码不符: %+v", s)
	}
	rec := ts.CallsFor("ik_web_cache")[0]
	if rec.Action != "show" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{"TYPE": "data"})
}

// TestWebCacheStatus 覆盖 ik_web_cache.show TYPE=status 的成功路径。
func TestWebCacheStatus(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_web_cache", "show", webCacheStatusSample())
	rows, err := New(ts.Client()).WebCache.Status(context.Background())
	if err != nil {
		t.Fatalf("Status 失败: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("状态行数不符: %+v", rows)
	}
	if rows[0].ServiceType != "video" || rows[0].CacheStatus != "idle" ||
		rows[0].Err != "没有挂载外部存储(Ext文件系统)" {
		t.Fatalf("状态行解码不符: %+v", rows[0])
	}
	if rows[1].CacheStatus != "running" || rows[1].Err != "" {
		t.Fatalf("状态行解码不符: %+v", rows[1])
	}
	wantParam(t, paramOf(t, ts.CallsFor("ik_web_cache")[0]),
		map[string]any{"TYPE": "status"})
}

// TestWebCacheSave 覆盖 ik_web_cache.edit（写动作，仅离线）。
func TestWebCacheSave(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ik_web_cache", "edit", map[string]any{})
	cfg := WebCacheConfig{Port: 651, LineCount: 1,
		Lines:        []WebCacheLine{{ID: 1, Name: "adsl03", Enabled: EnabledYes, Weight: 2}},
		ServiceCount: 1,
		Services:     []WebCacheServiceCfg{{Type: "video", Enabled: EnabledYes, MaxSize: 30, InactiveTime: 60, CacheMinUses: 3}},
	}
	if err := New(ts.Client()).WebCache.Save(context.Background(), cfg); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}
	rec := ts.CallsFor("ik_web_cache")[0]
	if rec.Action != "edit" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	p := paramOf(t, rec)
	if p["port"] != float64(651) || p["line_count"] != float64(1) || p["service_count"] != float64(1) {
		t.Fatalf("edit 参数顶层不符: %v", p)
	}
	line := p["line"].([]any)[0].(map[string]any)
	if line["id"] != float64(1) || line["enabled"] != "yes" || line["weight"] != float64(2) {
		t.Fatalf("edit 参数线路不符: %v", line)
	}
	svc := p["service"].([]any)[0].(map[string]any)
	if svc["type"] != "video" || svc["max_size"] != float64(30) || svc["cache_min_uses"] != float64(3) {
		t.Fatalf("edit 参数服务不符: %v", svc)
	}
}

// TestWebCacheError 覆盖固件错误路径。
func TestWebCacheError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("ik_web_cache", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "ik_web_cache", Action: action}
	})
	if _, err := New(ts.Client()).WebCache.Get(context.Background()); err == nil {
		t.Fatal("错误路径应报错")
	} else {
		errIsAPICode(t, err, 30006)
	}
}
