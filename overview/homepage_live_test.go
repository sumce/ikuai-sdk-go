package overview

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveHomepage 真机只读验证首页仪表盘（homepage.show）。
func TestLiveHomepage(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c)

	o, err := svc.SystemOverview(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=sysstat,ac_status 失败: %v", err)
	}
	if o.SysStat == nil || o.ACStatus == nil {
		t.Fatal("真机应返回 sysstat 与 ac_status 区块")
	}
	avg, err := o.CPUAverage()
	if err != nil {
		t.Fatalf("CPUAverage 解析失败: %v", err)
	}
	used, err := o.SysStat.Memory.UsedPercent()
	if err != nil {
		t.Fatalf("UsedPercent 解析失败: %v", err)
	}
	t.Logf("hostname=%s cpu均值=%.2f%% 内存使用=%.0f%% 在线用户=%d 连接数=%d "+
		"上传速率=%d 下载速率=%d 累计上行=%d 累计下行=%d 运行时长=%v "+
		"版本=%s ap_count=%d ap_online=%d",
		o.SysStat.Hostname, avg, used, o.SysStat.OnlineUser.Count,
		o.SysStat.Stream.ConnectNum, o.SysStat.Stream.Upload,
		o.SysStat.Stream.Download, o.SysStat.Stream.TotalUp,
		o.SysStat.Stream.TotalDown, o.UptimeDuration(),
		o.SysStat.VerInfo.Verstring, o.ACStatus.APCount, o.ACStatus.APOnline)
	if len(o.SysStat.CPU) == 0 || o.SysStat.Memory.Total == 0 {
		t.Error("真机 sysstat 应含 CPU 采样与内存信息")
	}

	snap, err := svc.Homepage.InterfaceSnapshot(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=ether_info,snapshoot 失败: %v", err)
	}
	t.Logf("网口数=%d lan快照行=%d wan快照行=%d",
		len(snap.EtherInfo), len(snap.SnapshootLAN), len(snap.SnapshootWAN))
	if len(snap.EtherInfo) == 0 {
		t.Error("真机应返回物理网口信息")
	}

	ah, err := svc.Homepage.AppHistorical(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=app_historical 失败: %v", err)
	}
	t.Logf("app_historical 区间组数=%d", len(ah.Groups))
	if len(ah.Groups) == 0 {
		t.Error("真机应有应用流量历史数据")
	}

	af, err := svc.Homepage.AppFlow(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=app_flow,app_historical 失败: %v", err)
	}
	t.Logf("app_flow 快照行=%d app_historical 区间组数=%d",
		len(af.AppFlow.Rows), len(af.AppHistorical.Groups))
	if len(af.AppFlow.Rows) == 0 || len(af.AppHistorical.Groups) == 0 {
		t.Error("真机应有应用流量快照与历史数据")
	}

	ms, err := svc.Homepage.MonitorSystem(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=monitor_system 失败: %v", err)
	}
	t.Logf("bandwidth_total(up=%d,down=%d) every_day_stats=%d",
		ms.BandwidthTotal.Upload, ms.BandwidthTotal.Download, len(ms.EveryDayStats))

	n, err := svc.Homepage.DHCPAddrpoolNum(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=dhcp_addrpool_num 失败: %v", err)
	}
	t.Logf("dhcp available_num=%d", n)

	ap, err := svc.Homepage.Antipiracy(ctx)
	if err != nil {
		t.Fatalf("homepage.show TYPE=antipiracy 失败: %v", err)
	}
	t.Logf("antipiracy=%d", ap)

	// 任意 TYPE 组合逃生口。
	resp, err := svc.Homepage.Show(ctx, HomepageTypeSysStat, HomepageTypeDHCPAddrpoolNum)
	if err != nil {
		t.Fatalf("homepage.show 任意组合失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("任意组合应成功, got %+v", resp)
	}
}
