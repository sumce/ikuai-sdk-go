package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// HardwareCPUInfo 是硬件信息的 CPU/主板区块（hardwareinfo.show 的
// hardwareinfo 键）。
type HardwareCPUInfo struct {
	// CPUModel CPU 型号（如 "Intel(R) Xeon(R) D-1747NTE CPU @ 2.50GHz"）。
	CPUModel string `json:"cpumodel"`
	// CPUBit CPU 位数（观测值 32；与 sysbit=x64 不一致为固件原始值，
	// 语义待真机验证）。
	CPUBit int64 `json:"cpubit"`
	// CPUMHz CPU 频率范围文本（如 "800 MHz - 3500 MHz"）。
	CPUMHz string `json:"cpumhz"`
	// CPUCache CPU 缓存文本（如 "15360 KB"）。
	CPUCache string `json:"cpucache"`
	// CPUCores 物理核数（观测值 10）。
	CPUCores int64 `json:"cpucores"`
	// CPUThread 线程数（观测值 20）。
	CPUThread int64 `json:"cputhread"`
	// Mainboard 主板型号文本。
	Mainboard string `json:"mainboard"`
	// Memory 内存容量（观测值 47730，推测单位 MB，待真机验证）。
	Memory int64 `json:"memory"`
	// HDD 硬盘描述文本（如 "ATA KINGSTON OM8P0S3 66.5 (119.24GB)"）。
	HDD string `json:"hdd"`
}

// HardwareInfo 是硬件信息页的聚合结果（hardwareinfo.show）。
type HardwareInfo struct {
	// HardwareInfo CPU/主板信息区块。
	HardwareInfo HardwareCPUInfo `json:"hardwareinfo"`
	// PCIInfo PCI 设备信息：键为设备类别（如 "Host bridge"、"VGA compatible
	// controller"），值为该类别下的设备描述列表。
	PCIInfo map[string][]string `json:"pciinfo"`
}

// HardwareInfoService 是硬件信息（hardwareinfo）的子服务。
type HardwareInfoService struct {
	c *ikuai.Client
}

// Get 获取硬件信息（CPU 型号/核数/内存/主板/硬盘与 PCI 设备列表）。
// 对应 /Action/call: hardwareinfo.show（param 为空对象；真机对
// TYPE=total,data 等分页 TYPE 回 30006 unknown TYPE），
// 即 Web 控制台「系统设置 → 硬件信息」页面。
func (s *HardwareInfoService) Get(ctx context.Context) (*HardwareInfo, error) {
	var out HardwareInfo
	if err := s.c.CallData(ctx, "hardwareinfo", "show",
		map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
