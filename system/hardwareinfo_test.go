package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// hardwareInfoSample 构造真机探测形态的 hardwareinfo.show 响应 Data。
func hardwareInfoSample() map[string]any {
	return map[string]any{
		"hardwareinfo": map[string]any{
			"cpumodel": "Intel(R) Xeon(R) D-1747NTE CPU @ 2.50GHz",
			"cpubit":   32,
			"cpumhz":   "800 MHz - 3500 MHz",
			"cpucache": "15360 KB",
			"cpucores": 10, "cputhread": 20,
			"mainboard": "Joinus BBU B1208-1747NTE-D3/weber, ",
			"memory":    47730,
			"hdd":       "ATA KINGSTON OM8P0S3 66.5 (119.24GB)",
		},
		"pciinfo": map[string]any{
			"Host bridge": []any{"Intel Corporation Ice Lake IEH"},
			"VGA compatible controller": []any{
				"ASPEED Technology, Inc. ASPEED Graphics Family (rev 41)",
			},
		},
	}
}

// TestHardwareInfoGet 覆盖 hardwareinfo.show 的成功路径与参数断言。
func TestHardwareInfoGet(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("hardwareinfo", "show", hardwareInfoSample())
	info, err := New(ts.Client()).HardwareInfo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	hw := info.HardwareInfo
	if hw.CPUModel != "Intel(R) Xeon(R) D-1747NTE CPU @ 2.50GHz" ||
		hw.CPUBit != 32 || hw.CPUCores != 10 || hw.CPUThread != 20 ||
		hw.Memory != 47730 || hw.CPUCache != "15360 KB" {
		t.Fatalf("硬件信息解码不符: %+v", hw)
	}
	if len(info.PCIInfo) != 2 {
		t.Fatalf("PCI 信息键数不符: %v", info.PCIInfo)
	}
	if vga := info.PCIInfo["VGA compatible controller"]; len(vga) != 1 ||
		vga[0] != "ASPEED Technology, Inc. ASPEED Graphics Family (rev 41)" {
		t.Fatalf("VGA PCI 解码不符: %v", vga)
	}
	wantParam(t, paramOf(t, ts.CallsFor("hardwareinfo")[0]), map[string]any{})
}

// TestHardwareInfoError 覆盖固件错误路径。
func TestHardwareInfoError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("hardwareinfo", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (total)", FuncName: "hardwareinfo", Action: action}
	})
	_, err := New(ts.Client()).HardwareInfo.Get(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
}
