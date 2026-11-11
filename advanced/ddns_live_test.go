package advanced

import (
	"context"
	"testing"

	"github.com/sumce/ikuai-sdk-go/internal/livetest"
)

// TestLiveDDNS 真机只读验证动态域名记录列表（ddns.show，真机探测实证
// 17 行）。只读，不做任何写动作（add/edit/del 绝对禁止对真机调用）。
func TestLiveDDNS(t *testing.T) {
	c := livetest.Client(t)
	page, err := New(c).DDNS.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("ddns.show 失败: %v", err)
	}
	t.Logf("DDNS 记录 total=%d", page.Total)
	if page.Total == 0 {
		t.Skip("本机无 DDNS 记录，行字段无法核验")
	}
	for _, r := range page.Rows {
		t.Logf("记录: id=%d server=%s interface=%s type=%s domain=%s ip=%s result=%s enabled=%s cf_token=%d location=%d",
			r.ID, r.Server, r.Interface, r.Type, r.Domain, r.IPAddress, r.Result, r.Enabled, r.CFToken, r.Location)
		if r.Server == "" || r.Domain == "" {
			t.Error("DDNS 行 server/domain 为空（字段待真机验证）")
		}
	}
}

// TestLiveNATDDNS 真机只读验证内网穿透状态（nat_ddns.show TYPE=data，
// 真机探测实证）。只读。
func TestLiveNATDDNS(t *testing.T) {
	c := livetest.Client(t)
	st, err := New(c).NATDDNS.Get(context.Background())
	if err != nil {
		t.Fatalf("nat_ddns.show 失败: %v", err)
	}
	t.Logf("NAT DDNS: status=%d errinfo=%q username=%q version=%q",
		st.Status, st.ErrInfo, st.Username, st.Version)
}
