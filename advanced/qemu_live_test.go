package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveQEMU 真机只读验证虚拟机列表与宿主机资源（qemu.show
// TYPE=total,data / TYPE=all，真机探测实证）。只读；新建/删除/电源等
// 虚拟机操作绝对禁止对真机调用。
func TestLiveQEMU(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).QEMU
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("qemu.show 失败: %v", err)
	}
	t.Logf("虚拟机 total=%d", page.Total)
	for _, vm := range page.Rows {
		if m, merr := vm.Map(); merr == nil {
			t.Logf("虚拟机: %v", m)
		}
	}

	info, err := svc.Info(ctx)
	if err != nil {
		t.Fatalf("qemu.show TYPE=all 失败: %v", err)
	}
	t.Logf("宿主机: cpu_cores=%d mem_available=%d eth=%d iface=%d usb=%d vfio=%d img=%d",
		info.CPUCores, info.MemAvailable, len(info.EthList), len(info.Interface),
		len(info.USBInfo), len(info.VFIOEthList), len(info.ImgInfo))
	if info.CPUCores == 0 {
		t.Error("cpu_cores 为空")
	}
}
