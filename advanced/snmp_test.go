package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// snmpSample 构造 netsnmp.show 的分页响应 Data（真机 3.7.24 探测实证
// 单行样例）。
func snmpSample() map[string]any {
	return map[string]any{
		"total": 1,
		"data": []any{map[string]any{
			"listen_port": 161, "syslocation": "机房", "syscontact": "admin@example.com",
			"sysname": "ikuai", "community": "public", "priv_pass": "privpass",
			"source": "192.168.1.0/24", "rw": "ro", "auth_pass": "authpass",
			"username": "snmpv3", "auth_proto": "sha", "security": "authPriv",
			"priv_proto": "aes", "id": 1, "enabled": "no", "version": 3,
		}},
	}
}

// TestSNMPGetList 覆盖 netsnmp.show 的成功路径与参数断言（真机实证
// TYPE=total,data 形态，恒单行）。
func TestSNMPGetList(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts 默认分页", nil, map[string]any{"TYPE": "total,data", "limit": "0,20"}},
		{"自定义分页", &ikuai.ListOptions{Offset: 0, Limit: 1}, map[string]any{"TYPE": "total,data", "limit": "0,1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("netsnmp", "show", snmpSample())
			svc := New(ts.Client()).SNMP
			cfg, err := svc.Get(context.Background())
			if err != nil {
				t.Fatalf("Get 失败: %v", err)
			}
			page, err := svc.List(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("List 失败: %v", err)
			}
			if page.Total != 1 || len(page.Rows) != 1 {
				t.Fatalf("分页解码不符: %+v", page)
			}
			c := page.Rows[0]
			if c.ID != 1 || c.IsEnabled() || c.ListenPort != 161 ||
				c.SysName != "ikuai" || c.SysContact != "admin@example.com" ||
				c.SysLocation != "机房" || !c.IsV3() || c.Community != "public" ||
				c.Source != "192.168.1.0/24" || c.RW != SNMPRWReadOnly ||
				c.Username != "snmpv3" || c.Security != "authPriv" ||
				c.AuthProto != "sha" || c.AuthPass != "authpass" ||
				c.PrivProto != "aes" || c.PrivPass != "privpass" {
				t.Fatalf("SNMP 配置解码不符: %+v", c)
			}
			if cfg == nil || cfg.ListenPort != 161 {
				t.Fatalf("Get 返回不符: %+v", cfg)
			}
			// recs[0] 为 Get（默认分页），recs[1] 为本次 List。
			wantParam(t, paramOf(t, ts.CallsFor("netsnmp")[1]), tt.want)
		})
	}
}

// TestSNMPEdit 覆盖 netsnmp.edit（写动作，仅离线）。
func TestSNMPEdit(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("netsnmp", "edit", map[string]any{})
	cfg := SNMPConfig{ID: 1, Enabled: EnabledYes, ListenPort: 1610,
		Version: 2, Community: "mycomm", Source: "10.0.0.0/8", RW: SNMPRWReadWrite}
	if err := New(ts.Client()).SNMP.Edit(context.Background(), cfg); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	rec := ts.CallsFor("netsnmp")[0]
	if rec.Action != "edit" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"] != float64(1) || p["enabled"] != "yes" || p["listen_port"] != float64(1610) ||
		p["community"] != "mycomm" || p["rw"] != "rw" {
		t.Fatalf("edit 参数不符: %v", p)
	}
}

// TestSNMPGetEmpty 覆盖空表时 Get 返回 nil。
func TestSNMPGetEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("netsnmp", "show", map[string]any{"total": 0, "data": []any{}})
	cfg, err := New(ts.Client()).SNMP.Get(context.Background())
	if err != nil || cfg != nil {
		t.Fatalf("空表 Get 应返回 nil,nil: %v,%+v", err, cfg)
	}
}

// TestSNMPError 覆盖固件错误路径。
func TestSNMPError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("netsnmp", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95001, ErrMsg: "失败", FuncName: "netsnmp", Action: action}
	})
	_, err := New(ts.Client()).SNMP.Get(context.Background())
	errIsAPICode(t, err, 95001)
}

// snmpClientAllSample 构造 netsnmpc.show TYPE=all 的响应 Data（真机
// 3.7.24 探测实证顶层字段，行数据空表）。
func snmpClientAllSample() map[string]any {
	return map[string]any{
		"data":          []any{map[string]any{"id": 3, "ip": "192.168.9.9", "port": 161}},
		"snmp_interval": 30,
		"ipgroup":       []any{"公共无线", "监控"},
		"mibfile":       []any{map[string]any{"filename": "MY-MIB.mib"}},
		"mibinfo":       []any{},
		"total":         1,
	}
}

// TestSNMPClientListGet 覆盖 netsnmpc.show 的两种 TYPE 成功路径。
func TestSNMPClientListGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("netsnmpc", "show", snmpClientAllSample())
	svc := New(ts.Client()).SNMPClient
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("List 解码不符: %+v", page)
	}
	m, err := page.Rows[0].Map()
	if err != nil {
		t.Fatalf("行 Map 失败: %v", err)
	}
	if m["ip"] != "192.168.9.9" {
		t.Fatalf("行原始 JSON 不符: %v", m)
	}

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if info.SNMPInterval != 30 || len(info.IPGroup) != 2 || info.IPGroup[0] != "公共无线" ||
		len(info.MIBFile) != 1 || info.Total != 1 || len(info.Rows) != 1 {
		t.Fatalf("TYPE=all 区块解码不符: %+v", info)
	}

	recs := ts.CallsFor("netsnmpc")
	wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": "total,data", "limit": "0,20"})
	wantParam(t, paramOf(t, recs[1]), map[string]any{"TYPE": "all"})
}

// TestSNMPClientAddEditDelete 覆盖 add/edit/del（写动作，仅离线，Raw 原样
// 透传断言）。
func TestSNMPClientAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("netsnmpc", "add", map[string]any{})
	ts.SetResponse("netsnmpc", "edit", map[string]any{})
	ts.SetResponse("netsnmpc", "del", map[string]any{})
	svc := New(ts.Client()).SNMPClient
	ctx := context.Background()

	raw := json.RawMessage(`{"id":3,"ip":"192.168.9.9","port":161,"oid":"1.3.6.1"}`)
	row := SNMPClient{}
	if err := json.Unmarshal(raw, &row); err != nil {
		t.Fatalf("构造行失败: %v", err)
	}
	rowID, err := svc.Add(ctx, row)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	if err := svc.Edit(ctx, row); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	recs := ts.CallsFor("netsnmpc")
	if recs[0].Action != "add" || recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action)
	}
	// Raw 行原样透传：add/edit 参数应与原始行逐字段一致。
	for i := 0; i < 2; i++ {
		var m map[string]any
		if err := json.Unmarshal(recs[i].Param, &m); err != nil {
			t.Fatalf("解析 param 失败: %v", err)
		}
		if m["ip"] != "192.168.9.9" || m["oid"] != "1.3.6.1" || m["id"] != float64(3) {
			t.Fatalf("Raw 透传不符: %v", m)
		}
	}
	wantParam(t, paramOf(t, recs[2]), map[string]any{"id": float64(3)})
}

// TestSNMPClientError 覆盖固件错误路径。
func TestSNMPClientError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("netsnmpc", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95002, ErrMsg: "失败", FuncName: "netsnmpc", Action: action}
	})
	if _, err := New(ts.Client()).SNMPClient.List(context.Background(), nil); err == nil {
		t.Fatal("错误路径应报错")
	} else {
		errIsAPICode(t, err, 95002)
	}
}
