package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// syscheckSample 构造 iksyscheck.show 的响应 Data（未体检形态为真机
// 3.7.24 实测：res 为空数组、status 为 "-1"；res 元素结构未知，测试
// 以对象形态构造用于解码通道验证）。
func syscheckSample(checked bool) map[string]any {
	m := map[string]any{}
	sections := []string{
		"disk", "memory", "internet", "dhcp_discover",
		"pppoe_discover", "arp", "cloud_link",
	}
	for _, sec := range sections {
		status, res := "-1", []any{}
		if checked {
			status = "0"
			if sec == "internet" {
				res = []any{
					map[string]any{"interface": "adsl01", "result": "ok"},
					map[string]any{"interface": "adsl09", "result": "fail"},
				}
			}
		}
		m[sec+"_res"] = res
		m[sec+"_status"] = status
	}
	return m
}

// TestSysCheckResult 覆盖 iksyscheck.show 结果查询（未体检与已体检）。
func TestSysCheckResult(t *testing.T) {
	tests := []struct {
		name    string
		checked bool
	}{
		{name: "未体检", checked: false},
		{name: "已体检", checked: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("iksyscheck", "show", syscheckSample(tt.checked))
			state, err := New(ts.Client()).SysCheck.Result(context.Background())
			if err != nil {
				t.Fatalf("Result 失败: %v", err)
			}
			if !tt.checked {
				if state.DiskStatus != "-1" || state.MemoryStatus != "-1" ||
					state.InternetStatus != "-1" || state.DHCPDiscoverStatus != "-1" ||
					state.PPPoEDiscoverStatus != "-1" || state.ARPStatus != "-1" ||
					state.CloudLinkStatus != "-1" {
					t.Fatalf("未体检态 7 项 status 均应为 -1: %+v", state)
				}
				if len(state.DiskRes) != 0 || len(state.CloudLinkRes) != 0 {
					t.Fatalf("未体检态 res 应为空数组: %+v", state)
				}
			} else {
				if state.InternetStatus != "0" || len(state.InternetRes) != 2 {
					t.Fatalf("已体检态线路检测解码不符: %+v", state)
				}
				m, err := state.InternetRes[0].Map()
				if err != nil {
					t.Fatalf("体检项 Map() 失败: %v", err)
				}
				if m["interface"] != "adsl01" || m["result"] != "ok" {
					t.Fatalf("体检项字段解码不符: %v", m)
				}
			}
			rec := ts.CallsFor("iksyscheck")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %s, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), map[string]any{})
		})
	}
}

// TestSysCheckExecute 覆盖 iksyscheck.begin_test 发起体检的参数断言
// （仅离线）。
func TestSysCheckExecute(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("iksyscheck", "begin_test", map[string]any{})
	if err := New(ts.Client()).SysCheck.Execute(context.Background()); err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}
	rec := ts.CallsFor("iksyscheck")[0]
	if rec.Action != "begin_test" {
		t.Fatalf("action = %s, 期望 begin_test", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestSysCheckStop 覆盖 iksyscheck.stop_test 停止体检的参数断言
// （仅离线）。
func TestSysCheckStop(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("iksyscheck", "stop_test", map[string]any{})
	if err := New(ts.Client()).SysCheck.Stop(context.Background()); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	rec := ts.CallsFor("iksyscheck")[0]
	if rec.Action != "stop_test" {
		t.Fatalf("action = %s, 期望 stop_test", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestSysCheckError 覆盖固件错误路径。
func TestSysCheckError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("iksyscheck", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "iksyscheck", Action: action}
	})
	svc := New(ts.Client()).SysCheck
	if _, err := svc.Result(context.Background()); err == nil {
		t.Fatal("Result 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if err := svc.Execute(context.Background()); err == nil {
		t.Fatal("Execute 应失败")
	}
	if err := svc.Stop(context.Background()); err == nil {
		t.Fatal("Stop 应失败")
	}
}
