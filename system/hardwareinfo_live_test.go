package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveHardwareInfo 真机只读验证硬件信息（hardwareinfo.show）。
func TestLiveHardwareInfo(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()

	info, err := New(c).HardwareInfo.Get(ctx)
	if err != nil {
		t.Fatalf("hardwareinfo.show 失败: %v", err)
	}
	hw := info.HardwareInfo
	if hw.CPUModel == "" {
		t.Error("cpumodel 不应为空")
	}
	if hw.CPUCores <= 0 || hw.CPUThread <= 0 {
		t.Errorf("核数/线程数异常: %+v", hw)
	}
	if hw.Memory <= 0 {
		t.Errorf("内存容量异常: %d", hw.Memory)
	}
	if len(info.PCIInfo) == 0 {
		t.Error("PCI 设备列表不应为空")
	}
	t.Logf("CPU=%s 核=%d 线程=%d 内存=%d 主板=%s PCI 类别数=%d",
		hw.CPUModel, hw.CPUCores, hw.CPUThread, hw.Memory, hw.Mainboard, len(info.PCIInfo))
}
