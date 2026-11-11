package ac

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveStatusOverview 真机只读验证无线概况（ac_status.show TYPE=all，
// 真机探测实证；只读，不做任何写动作）。
func TestLiveStatusOverview(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).Status

	ov, err := svc.Overview(context.Background())
	if err != nil {
		t.Fatalf("ac_status.show TYPE=all 失败: %v", err)
	}
	t.Logf("ac_status=%d ap_count=%d ap_online=%d clt_count=%d wphy=%+v",
		ov.ACStatus, ov.APStatus.APCount, ov.APStatus.APOnline,
		ov.CltStatus.CltCount, ov.WPHYUsage)
	if ov.APStatus.APCount < 0 || ov.CltStatus.CltCount < 0 {
		t.Error("统计值不应为负")
	}
}

// TestLiveStatusSelectors 真机只读验证各 TYPE 子选择器（真机探测实证）。
func TestLiveStatusSelectors(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Status

	if v, err := svc.ControlStatus(ctx); err != nil {
		t.Errorf("TYPE=ac_status 失败: %v", err)
	} else {
		t.Logf("ac_status=%d", v)
	}
	if v, err := svc.APStatus(ctx); err != nil {
		t.Errorf("TYPE=ap_status 失败: %v", err)
	} else {
		t.Logf("ap_status=%+v", v)
	}
	if v, err := svc.ClientStatus(ctx); err != nil {
		t.Errorf("TYPE=clt_status 失败: %v", err)
	} else {
		t.Logf("clt_status=%+v", v)
	}
	if v, err := svc.APBadStat(ctx); err != nil {
		t.Errorf("TYPE=ap_bad_stat 失败: %v", err)
	} else {
		t.Logf("ap_bad_stat=%+v up=%d down=%d", v.APBadStat, v.TotalUpload, v.TotalDownload)
	}
	if v, err := svc.APNetStatus(ctx); err != nil {
		t.Errorf("TYPE=ap_net_status 失败: %v", err)
	} else {
		t.Logf("ap_net_status=%+v", v)
	}
	if v, err := svc.WPHYUsage(ctx); err != nil {
		t.Errorf("TYPE=wphy_usage 失败: %v", err)
	} else {
		t.Logf("wphy_usage=%+v", v)
	}
	if v, err := svc.APChanInfo(ctx); err != nil {
		t.Errorf("TYPE=ap_chaninfo 失败: %v", err)
	} else {
		t.Logf("ap_chaninfo=%d 行", len(v))
	}
	if v, err := svc.APCountInfo(ctx); err != nil {
		t.Errorf("TYPE=ap_countinfo 失败: %v", err)
	} else {
		t.Logf("ap_countinfo=%d 行", len(v))
	}
	if v, err := svc.APEvents(ctx); err != nil {
		t.Errorf("TYPE=ap_event_data 失败: %v", err)
	} else {
		t.Logf("ap_event_data=%d 行", len(v))
	}
}
