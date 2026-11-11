package overview

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSysStat 真机只读验证系统实时状态（sysstat.show）。
func TestLiveSysStat(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	r, err := svc.SysStat.Realtime(ctx, nil)
	if err != nil {
		t.Fatalf("sysstat.show 默认组合失败: %v", err)
	}
	if r.CPU == nil || r.Memory == nil || r.Stream == nil || r.VerInfo == nil {
		t.Fatal("真机应返回 cpu/cputemp/memory/stream/verinfo 区块")
	}
	t.Logf("cpu采样=%d cputemp=%v memory.total=%d stream.connect_num=%d 版本=%s",
		len(r.CPU), r.CPUTemp, r.Memory.Total, r.Stream.ConnectNum,
		r.VerInfo.Verstring)

	v, err := svc.SysStat.VerInfo(ctx)
	if err != nil {
		t.Fatalf("sysstat.show TYPE=verinfo 失败: %v", err)
	}
	t.Logf("version=%s arch=%s sysbit=%s build_date=%d",
		v.Version, v.Arch, v.Sysbit, v.BuildDate)
	if v.Version == "" || v.BuildDate == 0 {
		t.Error("真机应返回版本号与构建时间")
	}

	cf, err := svc.SysStat.CPUFreq(ctx)
	if err != nil {
		t.Fatalf("sysstat.show TYPE=cpu,freq 失败: %v", err)
	}
	t.Logf("cpu采样=%d freq采样=%d", len(cf.CPU), len(cf.Freq))
	if len(cf.CPU) == 0 || len(cf.Freq) == 0 {
		t.Error("真机应返回 cpu/freq 采样")
	}
}
