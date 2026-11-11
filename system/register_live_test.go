package system

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveRegister 真机只读验证服务注册（register.show 两种 TYPE）。
func TestLiveRegister(t *testing.T) {
	c := livetest.Client(t)
	ctx := context.Background()
	svc := New(c).Register

	info, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("register.show TYPE=data,gwid 失败: %v", err)
	}
	if info.GWID == "" {
		t.Error("gwid 不应为空")
	}
	if len(info.Data) == 0 {
		t.Error("注册节点列表不应为空（真机观测 1 个节点）")
	}
	for _, n := range info.Data {
		if n.Code == "" {
			t.Errorf("注册码为空: %+v", n)
		}
	}
	t.Logf("gwid=%s 节点数=%d 首节点: id=%d code=%s comment=%s",
		info.GWID, len(info.Data), info.Data[0].ID, info.Data[0].Code, info.Data[0].Comment)

	full, err := svc.GetWithServiceCode(ctx)
	if err != nil {
		t.Fatalf("register.show TYPE=data,gwid,scode 失败: %v", err)
	}
	if full.SCode == nil {
		t.Fatal("scode 区块应存在")
	}
	t.Logf("服务码状态=%d errcode=%q errmsg=%q", full.SCode.Status, full.SCode.ServeErrcode, full.SCode.ServeErrmsg)
}
