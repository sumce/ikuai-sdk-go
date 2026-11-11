package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestWeComStatus 覆盖 wecom.show TYPE=data,bindid 的抓包形态
// （未绑定：enable=no 且各字段为空/0）。
func TestWeComStatus(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want WeComStatus
	}{
		{
			name: "未绑定(抓包形态)",
			data: map[string]any{
				"enable": "no", "sn": "", "url": "",
				"status": 0, "bindid": 0,
			},
			want: WeComStatus{Enable: EnabledNo},
		},
		{
			name: "已绑定",
			data: map[string]any{
				"enable": "yes", "sn": "SN123456", "url": "https://open.work.weixin.qq.com/...",
				"status": 1, "bindid": 7,
			},
			want: WeComStatus{
				Enable: EnabledYes, SN: "SN123456",
				URL:    "https://open.work.weixin.qq.com/...",
				Status: 1, BindID: 7,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("wecom", "show", tt.data)
			got, err := New(ts.Client()).WeCom.Status(context.Background())
			if err != nil {
				t.Fatalf("Status 失败: %v", err)
			}
			if *got != tt.want {
				t.Fatalf("解码不符: got %+v want %+v", *got, tt.want)
			}
			if got.IsEnabled() != (tt.want.Enable == EnabledYes) {
				t.Fatalf("IsEnabled 不符: %+v", got)
			}
			rec := ts.CallsFor("wecom")[0]
			if rec.FuncName != "wecom" || rec.Action != "show" {
				t.Fatalf("func/action = %s/%s, 期望 wecom/show", rec.FuncName, rec.Action)
			}
			wantParam(t, paramOf(t, rec),
				map[string]any{"TYPE": WeComTypeDataBindID})
		})
	}
}

// TestWeComError 覆盖固件错误路径。
func TestWeComError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("wecom", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95005, ErrMsg: "失败", FuncName: "wecom", Action: action}
	})
	if _, err := New(ts.Client()).WeCom.Status(context.Background()); err == nil {
		t.Fatal("wecom.show 错误应上抛")
	}
}
