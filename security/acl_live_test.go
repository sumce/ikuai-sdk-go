package security

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveACLList 真机只读验证 ACL 规则列表（acl.show）。
// Add/Edit/Delete 为写动作，严禁对真机调用。
func TestLiveACLList(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).ACL.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("acl.show 失败: %v", err)
	}
	t.Logf("ACL 规则 total=%d 本页行数=%d", page.Total, len(page.Rows))
	for i, r := range page.Rows {
		if i >= 5 {
			break
		}
		t.Logf("  行: id=%d enabled=%s src=%q dst=%q proto=%s action=%s dir=%s in=%s out=%s",
			r.ID, r.Enabled, r.SrcAddr, r.DstAddr, r.Protocol, r.Action,
			r.Dir, r.IInterface, r.OInterface)
		if r.Action != ACLActionAccept && r.Action != ACLActionDrop {
			t.Errorf("第 %d 行 action 非法: %q", i, r.Action)
		}
		if r.Dir != ACLDirForward && r.Dir != ACLDirInput {
			t.Errorf("第 %d 行 dir 非法: %q（行字段空表推导，待真机验证）", i, r.Dir)
		}
	}
}

// TestLiveACLSelectors 真机只读验证 ACL 表单选择器（acl.show TYPE=all）。
func TestLiveACLSelectors(t *testing.T) {
	c := livetest.Client(t)
	sel, err := New(c).ACL.Selectors(context.Background())
	if err != nil {
		t.Fatalf("acl.show TYPE=all 失败: %v", err)
	}
	if len(sel.Interface) == 0 {
		t.Error("接口选择器不应为空")
	}
	t.Logf("可选接口 %d 个, ipgroup=%d macgroup=%d dtgroup=%d iplib_status=%d",
		len(sel.Interface), len(sel.IPGroup), len(sel.MACGroup), len(sel.DTGroup), sel.IPLibStatus)
}
