package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveSNMP 真机只读验证 SNMP 服务配置（netsnmp.show，真机探测实证
// 单行）。只读，不做任何写动作。
func TestLiveSNMP(t *testing.T) {
	c := livetest.Client(t)
	cfg, err := New(c).SNMP.Get(context.Background())
	if err != nil {
		t.Fatalf("netsnmp.show 失败: %v", err)
	}
	if cfg == nil {
		t.Skip("本机无 SNMP 配置行")
	}
	t.Logf("SNMP: enabled=%s port=%d version=%d community=%s rw=%s v3user=%s",
		cfg.Enabled, cfg.ListenPort, cfg.Version, cfg.Community, cfg.RW, cfg.Username)
	if cfg.ListenPort == 0 {
		t.Error("SNMP listen_port 为空")
	}
}

// TestLiveSNMPClient 真机只读验证 SNMP 客户端配置（netsnmpc.show
// TYPE=total,data / TYPE=all，真机探测实证）。只读，不做任何写动作。
func TestLiveSNMPClient(t *testing.T) {
	c := livetest.Client(t)
	svc := New(c).SNMPClient
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("netsnmpc.show total,data 失败: %v", err)
	}
	t.Logf("SNMP 客户端行 total=%d", page.Total)
	for _, r := range page.Rows {
		if m, merr := r.Map(); merr == nil {
			t.Logf("行: %v", m)
		}
	}

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("netsnmpc.show TYPE=all 失败: %v", err)
	}
	t.Logf("TYPE=all: snmp_interval=%d ipgroup=%v mibfile=%d mibinfo=%d total=%d",
		info.SNMPInterval, info.IPGroup, len(info.MIBFile), len(info.MIBInfo), info.Total)
}
