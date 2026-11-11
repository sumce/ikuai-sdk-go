package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// tcpdumpSample 构造 tcpdump.show 的响应 Data（空闲/进行中两态，行
// 字段形态为真机 3.7.24 实测：hit_num/size 为字符串）。
func tcpdumpSample(capturing bool) map[string]any {
	row := map[string]any{
		"interface": "", "mac": "", "hit_num": "100", "proto": "",
		"ip_addr": "", "port": "", "filename": "", "size": "0", "last_time": "",
	}
	if capturing {
		row["interface"] = "lan1"
		row["proto"] = "tcp"
		row["filename"] = "/Document/001"
		row["size"] = "1024"
		row["last_time"] = "2026-11-11 10:00:00"
	}
	return map[string]any{"data": []any{row}}
}

// TestTCPDumpStatus 覆盖 tcpdump.show 状态查询的成功路径与参数断言。
func TestTCPDumpStatus(t *testing.T) {
	tests := []struct {
		name      string
		capturing bool
	}{
		{name: "空闲态", capturing: false},
		{name: "抓包中", capturing: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("tcpdump", "show", tcpdumpSample(tt.capturing))
			state, err := New(ts.Client()).TCPDump.Status(context.Background())
			if err != nil {
				t.Fatalf("Status 失败: %v", err)
			}
			if state == nil {
				t.Fatal("状态行不应为 nil")
			}
			if state.HitNum != "100" {
				t.Fatalf("hit_num 解码不符: %+v", state)
			}
			if tt.capturing {
				if state.Filename == "" || state.Size != "1024" ||
					state.Interface != "lan1" || state.Proto != TCPDumpProtoTCP {
					t.Fatalf("抓包中状态解码不符: %+v", state)
				}
			} else if state.Filename != "" || state.Size != "0" {
				t.Fatalf("空闲态状态解码不符: %+v", state)
			}
			rec := ts.CallsFor("tcpdump")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %s, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), map[string]any{})
		})
	}
}

// TestTCPDumpStatusEmpty 覆盖 data 为空数组时返回 nil。
func TestTCPDumpStatusEmpty(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("tcpdump", "show", map[string]any{"data": []any{}})
	state, err := New(ts.Client()).TCPDump.Status(context.Background())
	if err != nil {
		t.Fatalf("Status 失败: %v", err)
	}
	if state != nil {
		t.Fatalf("空 data 应返回 nil, got %+v", state)
	}
}

// TestTCPDumpStart 覆盖 tcpdump.show 带参发起抓包的参数断言（仅离线）。
func TestTCPDumpStart(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("tcpdump", "show", tcpdumpSample(true))
	req := TCPDumpRequest{
		Interface: "lan1", IPAddr: "192.168.1.2", MAC: "00:0c:29:57:23:25",
		Port: "443", Proto: TCPDumpProtoTCP, HitNum: 200, Filename: "/Document/sdk-test",
	}
	state, err := New(ts.Client()).TCPDump.Start(context.Background(), req)
	if err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	if state == nil || state.Filename != "/Document/001" {
		t.Fatalf("返回状态解码不符: %+v", state)
	}
	rec := ts.CallsFor("tcpdump")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %s, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{
		"interface": "lan1", "ip_addr": "192.168.1.2", "mac": "00:0c:29:57:23:25",
		"port": "443", "proto": "tcp", "hit_num": float64(200), "filename": "/Document/sdk-test",
	})
}

// TestTCPDumpStop 覆盖 tcpdump.stop 停止抓包的参数断言（仅离线）。
func TestTCPDumpStop(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("tcpdump", "stop", map[string]any{})
	if err := New(ts.Client()).TCPDump.Stop(context.Background()); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	rec := ts.CallsFor("tcpdump")[0]
	if rec.Action != "stop" {
		t.Fatalf("action = %s, 期望 stop", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestTCPDumpError 覆盖固件错误路径。
func TestTCPDumpError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("tcpdump", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown action", FuncName: "tcpdump", Action: action}
	})
	svc := New(ts.Client()).TCPDump
	if _, err := svc.Status(context.Background()); err == nil {
		t.Fatal("Status 应失败")
	} else if ae, ok := ikuai.AsAPIError(err); !ok || ae.Code != 30006 {
		t.Fatalf("期望 APIError(30006), got %v", err)
	}
	if _, err := svc.Start(context.Background(), TCPDumpRequest{}); err == nil {
		t.Fatal("Start 应失败")
	}
	if err := svc.Stop(context.Background()); err == nil {
		t.Fatal("Stop 应失败")
	}
}
