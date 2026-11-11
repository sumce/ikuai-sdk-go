package security

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveAdvancedGet 真机只读验证安全高级设置（advanced.show）。
// Save 为写动作，严禁对真机调用。
func TestLiveAdvancedGet(t *testing.T) {
	c := livetest.Client(t)
	cfg, err := New(c).Advanced.Get(context.Background())
	if err != nil {
		t.Fatalf("advanced.show 失败: %v", err)
	}
	if cfg == nil {
		t.Fatal("advanced.show 应返回单行配置")
	}
	if cfg.ID != 1 {
		t.Errorf("配置行 id 应为 1, got %d", cfg.ID)
	}
	t.Logf("noping_lan=%d noping_wan=%d notracert=%d hijack_ping=%d invalid=%d",
		cfg.NoPingLAN, cfg.NoPingWAN, cfg.NoTracert, cfg.HijackPing, cfg.Invalid)
	t.Logf("dos_lan=%d dos_lan_num=%d tcp2p=(%d,%d) udp2p=(%d,%d) tcp_mss=(%d,%d)",
		cfg.DOSDefenseLAN, cfg.DOSDefenseNum,
		cfg.LimitTCP2P, cfg.LimitTCP2PNum, cfg.LimitUDP2P, cfg.LimitUDP2PNum,
		cfg.TCPMSS, cfg.TCPMSSNum)
	if cfg.DOSDefenseNum <= 0 || cfg.TCPMSSNum <= 0 {
		t.Error("dos_lan_num/tcp_mss_num 应为正数")
	}
}
