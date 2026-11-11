package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// fileServicesSample 构造 ftp_server/smbd/http_server show 的综合响应
// Data（ftp/http 行字段与 smb 顶层字段真机实证；ftp/http 行字段空表推导）。
func fileServicesSample(funcName string) map[string]any {
	switch funcName {
	case "ftp_server":
		return map[string]any{
			"total": 1,
			"data": []any{map[string]any{
				"id": 4, "enabled": "yes", "username": "ftpuser", "passwd": "ftppass",
				"permission": "ro", "home_dir": "A:/share", "upload": 0,
				"download": 100, "tagname": "下载账号",
			}},
			"open_ftp": 1, "ftp_port": 21, "ftp_access": 1, "partname": []any{"A"},
		}
	case "smbd":
		return map[string]any{
			"data": []any{map[string]any{
				"workgroup": "WORKGROUP", "wsdd2": 1, "interface": "lan1",
				"access": 1, "id": 1, "enabled": "no",
			}},
			"dir_data":  []any{map[string]any{"name": "public"}},
			"dir_total": 1,
			"interface": []any{[]any{"lan1"}, []any{"lan2", "内网接入三层交换机"}},
		}
	case "http_server":
		return map[string]any{
			"total": 1,
			"data": []any{map[string]any{
				"id": 6, "enabled": "yes", "tagname": "软件站", "http_port": 8080,
				"server_name": "dl.example.com", "ssl_on": 0, "autoindex": 1,
				"download": 1, "home_dir": "A:/www", "access": 0,
			}},
			"partname": []any{"A"},
		}
	}
	return map[string]any{}
}

// TestFTPListGetConfig 覆盖 ftp_server.show 的两种形态成功路径（真机
// 实证 TYPE=total/data 与 TYPE=all 顶层字段）。
func TestFTPListGetConfig(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("ftp_server", "show", fileServicesSample("ftp_server"))
	svc := New(ts.Client()).FileServices
	ctx := context.Background()

	page, err := svc.FTPList(ctx, nil)
	if err != nil {
		t.Fatalf("FTPList 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	u := page.Rows[0]
	if u.ID != 4 || !u.IsEnabled() || u.Username != "ftpuser" || u.Passwd != "ftppass" ||
		u.Permission != "ro" || u.HomeDir != "A:/share" || u.Upload != 0 ||
		u.Download != 100 || u.Tagname != "下载账号" {
		t.Fatalf("FTP 用户行解码不符: %+v", u)
	}

	cfg, err := svc.FTPGetConfig(ctx)
	if err != nil {
		t.Fatalf("FTPGetConfig 失败: %v", err)
	}
	if cfg.OpenFTP != 1 || cfg.FTPPort != 21 || cfg.FTPAccess != 1 ||
		len(cfg.Partname) != 1 || cfg.Partname[0] != "A" {
		t.Fatalf("FTP 配置解码不符: %+v", cfg)
	}

	recs := ts.CallsFor("ftp_server")
	wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": "total,data", "limit": "0,20"})
	wantParam(t, paramOf(t, recs[1]), map[string]any{"TYPE": "all"})
}

// TestFTPSaveConfigAddEditDelete 覆盖 ftp_server 写方法（写动作，仅离线）。
func TestFTPSaveConfigAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	for _, act := range []string{"add", "edit", "del"} {
		ts.SetResponse("ftp_server", act, map[string]any{})
	}
	svc := New(ts.Client()).FileServices
	ctx := context.Background()

	if err := svc.FTPSaveConfig(ctx, FTPConfig{OpenFTP: 1, FTPPort: 2121, FTPAccess: 0}); err != nil {
		t.Fatalf("FTPSaveConfig 失败: %v", err)
	}
	user := FTPUser{Username: "u1", Passwd: "p1", Permission: "rw", HomeDir: "A:/pub"}
	rowID, err := svc.FTPAddUser(ctx, user)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	user.ID = 5
	user.Permission = "ro"
	if err := svc.FTPEditUser(ctx, user); err != nil {
		t.Fatalf("FTPEditUser 失败: %v", err)
	}
	if err := svc.FTPDeleteUser(ctx, 5, 6); err != nil {
		t.Fatalf("FTPDeleteUser 失败: %v", err)
	}

	recs := ts.CallsFor("ftp_server")
	if recs[0].Action != "edit" || recs[1].Action != "add" || recs[2].Action != "edit" || recs[3].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action)
	}
	p := paramOf(t, recs[0])
	if p["open_ftp"] != float64(1) || p["ftp_port"] != float64(2121) || p["ftp_access"] != float64(0) {
		t.Fatalf("配置 edit 参数不符: %v", p)
	}
	p = paramOf(t, recs[2])
	if p["id"] != float64(5) || p["permission"] != "ro" {
		t.Fatalf("用户 edit 参数不符: %v", p)
	}
	if p := paramOf(t, recs[3]); len(p["ids"].([]any)) != 2 {
		t.Fatalf("del 多 ID 参数不符: %v", p)
	}
}

// TestSMBGetEdit 覆盖 smbd.show TYPE=data 与 smbd.edit（真机实证配置行
// 与顶层区块；edit 为写动作，仅离线）。
func TestSMBGetEdit(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("smbd", "show", fileServicesSample("smbd"))
	ts.SetResponse("smbd", "edit", map[string]any{})
	svc := New(ts.Client()).FileServices
	ctx := context.Background()

	info, err := svc.SMBGet(ctx)
	if err != nil {
		t.Fatalf("SMBGet 失败: %v", err)
	}
	if info.DirTotal != 1 || len(info.DirData) != 1 || len(info.Interface) != 2 ||
		info.Interface[1][1] != "内网接入三层交换机" {
		t.Fatalf("SMB 区块解码不符: %+v", info)
	}
	cfg, err := svc.SMBGetConfig(ctx)
	if err != nil {
		t.Fatalf("SMBGetConfig 失败: %v", err)
	}
	if cfg.ID != 1 || cfg.IsEnabled() || cfg.Workgroup != "WORKGROUP" ||
		cfg.WSDD2 != 1 || cfg.Interface != "lan1" || cfg.Access != 1 {
		t.Fatalf("SMB 配置行解码不符: %+v", cfg)
	}

	cfg.Workgroup = "IKUAI"
	cfg.WSDD2 = 0
	if err := svc.SMBEdit(ctx, *cfg); err != nil {
		t.Fatalf("SMBEdit 失败: %v", err)
	}
	// recs: [0]=SMBGet.show, [1]=SMBGetConfig.show, [2]=SMBEdit.edit。
	recs := ts.CallsFor("smbd")
	rec := recs[len(recs)-1]
	if rec.Action != "edit" {
		t.Fatalf("action 不符: %s", rec.Action)
	}
	p := paramOf(t, rec)
	if p["id"] != float64(1) || p["workgroup"] != "IKUAI" || p["wsdd2"] != float64(0) {
		t.Fatalf("edit 参数不符: %v", p)
	}
	// show 参数断言：smbd 不支持 TYPE=total，必须发 TYPE=data。
	wantParam(t, paramOf(t, ts.CallsFor("smbd")[0]), map[string]any{"TYPE": "data"})
}

// TestSMBGetEmpty 覆盖空表时 SMBGetConfig 返回 nil。
func TestSMBGetEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("smbd", "show", map[string]any{
		"data": []any{}, "dir_data": []any{}, "dir_total": 0, "interface": []any{},
	})
	cfg, err := New(ts.Client()).FileServices.SMBGetConfig(context.Background())
	if err != nil || cfg != nil {
		t.Fatalf("空表 SMBGetConfig 应返回 nil,nil: %v,%+v", err, cfg)
	}
}

// TestHTTPListGetOptionsAddEditDelete 覆盖 http_server 全方法族（行字段
// 空表推导；写动作仅离线）。
func TestHTTPListGetOptionsAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("http_server", "show", fileServicesSample("http_server"))
	for _, act := range []string{"add", "edit", "del"} {
		ts.SetResponse("http_server", act, map[string]any{})
	}
	svc := New(ts.Client()).FileServices
	ctx := context.Background()

	page, err := svc.HTTPList(ctx, &ikuai.ListOptions{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("HTTPList 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	r := page.Rows[0]
	if r.ID != 6 || !r.IsEnabled() || r.Tagname != "软件站" || r.HTTPPort != 8080 ||
		r.ServerName != "dl.example.com" || r.SSL != 0 || r.AutoIndex != 1 ||
		r.Download != 1 || r.HomeDir != "A:/www" || r.Access != 0 {
		t.Fatalf("HTTP 行解码不符: %+v", r)
	}

	opts, err := svc.HTTPGetOptions(ctx)
	if err != nil {
		t.Fatalf("HTTPGetOptions 失败: %v", err)
	}
	if opts.Total != 1 || len(opts.Partname) != 1 || opts.Partname[0] != "A" {
		t.Fatalf("TYPE=all 区块解码不符: %+v", opts)
	}

	row := HTTPServerRow{Tagname: "新站", HTTPPort: 80, HomeDir: "A:/web", Access: 1}
	rowID, err := svc.HTTPAdd(ctx, row)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	row.ID = 8
	if err := svc.HTTPEdit(ctx, row); err != nil {
		t.Fatalf("HTTPEdit 失败: %v", err)
	}
	if err := svc.HTTPDelete(ctx, 8); err != nil {
		t.Fatalf("HTTPDelete 失败: %v", err)
	}

	recs := ts.CallsFor("http_server")
	if recs[0].Action != "show" || recs[1].Action != "show" ||
		recs[2].Action != "add" || recs[3].Action != "edit" || recs[4].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s,%s,%s",
			recs[0].Action, recs[1].Action, recs[2].Action, recs[3].Action, recs[4].Action)
	}
	p := paramOf(t, recs[2])
	if p["tagname"] != "新站" || p["http_port"] != float64(80) || p["access"] != float64(1) {
		t.Fatalf("add 参数不符: %v", p)
	}
	if _, ok := p["id"]; ok {
		t.Fatalf("add 参数不应携带 id: %v", p)
	}
	wantParam(t, paramOf(t, recs[4]), map[string]any{"id": float64(8)})
}

// TestFileServicesError 覆盖固件错误路径（三个 func 逐一）。
func TestFileServicesError(t *testing.T) {
	for _, fn := range []string{"ftp_server", "smbd", "http_server"} {
		t.Run(fn, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.HandleFuncName(fn, func(action string, param json.RawMessage) (any, error) {
				return nil, &ikuai.APIError{Code: 95005, ErrMsg: "失败", FuncName: fn, Action: action}
			})
			svc := New(ts.Client()).FileServices
			ctx := context.Background()
			var err error
			switch fn {
			case "ftp_server":
				_, err = svc.FTPList(ctx, nil)
			case "smbd":
				_, err = svc.SMBGet(ctx)
			case "http_server":
				_, err = svc.HTTPList(ctx, nil)
			}
			if err == nil {
				t.Fatal("错误路径应报错")
			}
			errIsAPICode(t, err, 95005)
		})
	}
}
