package behavior

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveAudit 真机只读验证行为记录配置（audit.show）。
// Save/RecordFreeSet 为写动作，严禁对真机调用。
// RecordFreeGet 因探测真机（3.7.24）未注册该 func（30002）而跳过断言。
func TestLiveAudit(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Audit

	cfg, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("audit.show 失败: %v", err)
	}
	if cfg != nil {
		t.Logf("审计存储: 总量=%d 字节 已用=%d 字节 URL=%d IM=%d 终端=%d 应用=%d",
			cfg.TotalSize, cfg.UseSize, cfg.OpenURLRecord, cfg.OpenIMRecord,
			cfg.OpenTerminalRecord, cfg.OpenAppidRecord)
	} else {
		t.Log("audit.show 返回空配置行")
	}

	info, err := svc.Info(ctx)
	if err != nil {
		t.Fatalf("audit.show TYPE=all 失败: %v", err)
	}
	t.Logf("可清理类型: %v", info.CleanType)
	if len(info.CleanType) == 0 {
		t.Error("clean_type 不应为空")
	}

	// record_free_setting：探测真机回 30002 Not found funcname，只记录形态。
	if _, err := svc.RecordFreeGet(ctx); err != nil {
		t.Logf("record_free_setting.show 按预期失败（本固件未注册）: %v", err)
	} else {
		t.Log("record_free_setting.show 成功（本固件已注册该 func）")
	}
}
