package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// upgradeSample 构造抓包形态的 upgrade.show 响应 Data（TYPE=data,fileinfo）。
func upgradeSample() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"auto_upgrade_lib_dpi": 1, "auto_upgrade_lib_im": 1,
			"auto_upgrade_lib_domain": 1, "auto_upgrade_sec": 1,
			"libproto_ver": "2.0.486", "libaudit_ver": "2.1.17", "libdomain_ver": "2.1.0",
			"system_ver": "3.7.24", "build_date": "202608141210",
			"new_libproto_ver": "2.0.486", "new_libaudit_ver": "2.1.17",
			"new_libdomain_ver": "2.1.0", "new_system_ver": "3.7.26",
			"new_build_date":     "202609111743",
			"update_content":     "1.部分环境下OpenVPN拨号获取的IP地址错误",
			"ignore_upgrade_ver": "", "bootguide": "hd", "firmware_channel": 0,
		},
		"fileinfo": map[string]any{},
	}
}

// TestUpgradeStatus 覆盖 upgrade.show TYPE=data,fileinfo 的成功路径与参数断言。
func TestUpgradeStatus(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("upgrade", "show", upgradeSample())
	st, err := New(ts.Client()).Upgrade.Status(context.Background())
	if err != nil {
		t.Fatalf("Status 失败: %v", err)
	}
	d := st.Data
	if d.SystemVer != "3.7.24" || d.BuildDate != "202608141210" ||
		d.NewSystemVer != "3.7.26" || d.NewBuildDate != "202609111743" ||
		d.LibprotoVer != "2.0.486" || d.FirmwareChannel != 0 ||
		d.AutoUpgradeLibDPI != 1 || d.UpdateContent == "" {
		t.Fatalf("升级信息解码不符: %+v", d)
	}
	if !d.HasNewVersion() {
		t.Fatal("3.7.24 → 3.7.26 应判定有新版本")
	}
	if _, err := d.BuildTime(); err != nil {
		t.Fatalf("BuildTime 解析失败: %v", err)
	}
	if _, err := d.NewBuildTime(); err != nil {
		t.Fatalf("NewBuildTime 解析失败: %v", err)
	}
	// fileinfo 为空对象怪癖：原始 JSON 原样保留。
	if st.Fileinfo == nil || len(st.Fileinfo) == 0 {
		t.Fatal("fileinfo 应原样保留")
	}
	wantParam(t, paramOf(t, ts.CallsFor("upgrade")[0]),
		map[string]any{"TYPE": UpgradeTypeDataFileinfo})
}

// TestUpgradeHasNewVersionFalse 覆盖无新版本的判定。
func TestUpgradeHasNewVersionFalse(t *testing.T) {
	d := UpgradeData{SystemVer: "3.7.26", NewSystemVer: "3.7.26"}
	if d.HasNewVersion() {
		t.Fatal("相同版本不应判定有新版本")
	}
	d2 := UpgradeData{SystemVer: "3.7.26", NewSystemVer: ""}
	if d2.HasNewVersion() {
		t.Fatal("空新版本不应判定有新版本")
	}
}

// TestUpgradeCloudSync 覆盖 TYPE=cloud_sync 路径。
func TestUpgradeCloudSync(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("upgrade", "show", map[string]any{"cloud_sync": 1})
	v, err := New(ts.Client()).Upgrade.CloudSync(context.Background())
	if err != nil {
		t.Fatalf("CloudSync 失败: %v", err)
	}
	if v != 1 {
		t.Fatalf("cloud_sync 解码不符: %d", v)
	}
	wantParam(t, paramOf(t, ts.CallsFor("upgrade")[0]),
		map[string]any{"TYPE": UpgradeTypeCloudSync})
}

// TestUpgradeError 覆盖固件错误路径。
func TestUpgradeError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("upgrade", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95009, ErrMsg: "失败", FuncName: "upgrade", Action: action}
	})
	if _, err := New(ts.Client()).Upgrade.Status(context.Background()); err == nil {
		t.Fatal("status 错误应上抛")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 95009 {
		t.Fatalf("期望 APIError(95009), got %v", err)
	}
	if _, err := New(ts.Client()).Upgrade.CloudSync(context.Background()); err == nil {
		t.Fatal("cloud_sync 错误应上抛")
	}
}
