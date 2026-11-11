package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveIRQBalance 真机只读验证负载监控（irqbalance.show 三种 TYPE）。
func TestLiveIRQBalance(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).IRQBalance

	hard, err := svc.HardIRQ(ctx, nil)
	if err != nil {
		t.Fatalf("irqbalance.show TYPE=hardirq 失败: %v", err)
	}
	if len(hard) == 0 {
		t.Error("硬中断分布不应为空")
	}
	for _, r := range hard {
		if r.IrqNum == "" {
			t.Errorf("中断号为空: %+v", r)
			break
		}
	}
	t.Logf("硬中断行数=%d 首行: cpuid=%s irq=%s ifname=%s set=%s",
		len(hard), hard[0].Cpuid, hard[0].IrqNum, hard[0].Ifname, hard[0].SetCpuid)

	soft, err := svc.SoftIRQ(ctx, nil)
	if err != nil {
		t.Fatalf("irqbalance.show TYPE=softirq 失败: %v", err)
	}
	if len(soft) == 0 {
		t.Error("软中断统计不应为空")
	}
	t.Logf("软中断行数=%d 首行: cpuid=%s freq=%s used=%s",
		len(soft), soft[0].Cpuid, soft[0].Freq, soft[0].Used)

	// 控制台分页形态（固件接受 limit 参数但实测不生效，197 行全量返回，
	// 与 burp 抓包行为一致；此处仅验证调用成功）。
	hard2, err := svc.HardIRQ(ctx, &ikuai.ListOptions{Limit: 20})
	if err != nil {
		t.Fatalf("irqbalance.show 分页形态失败: %v", err)
	}
	if len(hard2) == 0 {
		t.Error("分页形态应有数据")
	}

	cpu, err := svc.CPUList(ctx)
	if err != nil {
		t.Fatalf("irqbalance.show TYPE=cpulist,freqlist 失败: %v", err)
	}
	if len(cpu.Cpulist) < 2 {
		t.Errorf("CPU 列表应含 auto 与至少一个 cpuN: %v", cpu.Cpulist)
	}
	t.Logf("CPU 数=%d 调频支持=%d 当前策略=%s 睿频=%d 策略集=%v",
		len(cpu.Cpulist), cpu.CpufreqSupport, cpu.CurrentCpufreq, cpu.CurrentTurbo, cpu.CpufreqList)
}
