package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestDingTalkBindStatus 覆盖 TYPE=bind_status 的两种抓包形态。
func TestDingTalkBindStatus(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want DingTalkBindStatus
	}{
		{
			name: "二维码就绪(抓包形态)",
			data: map[string]any{
				"status": 1,
				"url":    "https://page.dingtalk.com/wow/dingtalk/act/bind?token=x",
				"errmsg": "使用钉钉扫描二维码,绑定到钉钉",
			},
			want: DingTalkBindStatus{
				Status: DingTalkStatusQRReady,
				URL:    "https://page.dingtalk.com/wow/dingtalk/act/bind?token=x",
				ErrMsg: "使用钉钉扫描二维码,绑定到钉钉",
			},
		},
		{
			name: "正在获取二维码(抓包形态)",
			data: map[string]any{"status": 2, "url": "", "errmsg": "正在获取二维码"},
			want: DingTalkBindStatus{Status: DingTalkStatusGettingQR, ErrMsg: "正在获取二维码"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("dingtalk", "show", tt.data)
			got, err := New(ts.Client()).DingTalk.BindStatus(context.Background())
			if err != nil {
				t.Fatalf("BindStatus 失败: %v", err)
			}
			if *got != tt.want {
				t.Fatalf("解码不符: got %+v want %+v", *got, tt.want)
			}
			wantParam(t, paramOf(t, ts.CallsFor("dingtalk")[0]),
				map[string]any{"TYPE": DingTalkTypeBindStatus})
		})
	}
}

// TestDingTalkBindID 覆盖 TYPE=bindid 路径。
func TestDingTalkBindID(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dingtalk", "show", map[string]any{"bindid": 0})
	id, err := New(ts.Client()).DingTalk.BindID(context.Background())
	if err != nil {
		t.Fatalf("BindID 失败: %v", err)
	}
	if id != 0 {
		t.Fatalf("bindid 解码不符: %d", id)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dingtalk")[0]),
		map[string]any{"TYPE": DingTalkTypeBindID})
}

// TestDingTalkAppBindSupport 覆盖 TYPE=appbind_support 路径。
func TestDingTalkAppBindSupport(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("dingtalk", "show", map[string]any{"support_dtalk": 1, "support_wecom": 1})
	sup, err := New(ts.Client()).DingTalk.AppBindSupport(context.Background())
	if err != nil {
		t.Fatalf("AppBindSupport 失败: %v", err)
	}
	if sup.SupportDtalk != 1 || sup.SupportWecom != 1 {
		t.Fatalf("支持能力解码不符: %+v", sup)
	}
	wantParam(t, paramOf(t, ts.CallsFor("dingtalk")[0]),
		map[string]any{"TYPE": DingTalkTypeAppBindSupport})
}

// TestDingTalkError 覆盖固件错误路径。
func TestDingTalkError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("dingtalk", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95005, ErrMsg: "失败", FuncName: "dingtalk", Action: action}
	})
	if _, err := New(ts.Client()).DingTalk.BindStatus(context.Background()); err == nil {
		t.Fatal("bind_status 错误应上抛")
	}
	if _, err := New(ts.Client()).DingTalk.BindID(context.Background()); err == nil {
		t.Fatal("bindid 错误应上抛")
	}
	if _, err := New(ts.Client()).DingTalk.AppBindSupport(context.Background()); err == nil {
		t.Fatal("appbind_support 错误应上抛")
	}
}
