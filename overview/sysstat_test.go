package overview

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestSysStatRealtime 覆盖 sysstat.show 成功路径与参数断言。
func TestSysStatRealtime(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ikuai.ListOptions
		data   map[string]any
		want   map[string]any
		assert func(t *testing.T, r *SysStatRealtime)
	}{
		{
			name: "nil-opts 默认实时面板组合(sysstat.show 抓包默认)",
			opts: nil,
			data: map[string]any{
				"cpu":     []any{"6.00%", "4.00%", "7.92%"},
				"cputemp": []any{39},
				"memory": map[string]any{
					"total": 47914500, "available": 44857584, "free": 44940664,
					"cached": 127156, "buffers": 267852, "used": "6%",
				},
				"stream": map[string]any{
					"connect_num": 9149, "upload": 11377247, "download": 15008169,
					"total_up": 37471727016871, "total_down": 40668767125903,
				},
				"verinfo": map[string]any{
					"modelname": "", "verstring": "3.7.24 x64 Build202608141210",
					"version": "3.7.24", "build_date": 202608141210, "arch": "x86",
					"sysbit": "x64", "verflags": "", "is_enterprise": 0,
					"support_i18n": 0, "support_dingtalk": 1, "support_lcd": 0,
					"bootguide": "hd",
				},
			},
			want: map[string]any{"TYPE": SysStatTypeRealtime},
			assert: func(t *testing.T, r *SysStatRealtime) {
				if len(r.CPU) != 3 || r.CPU[2] != "7.92%" {
					t.Fatalf("cpu 解码不符: %v", r.CPU)
				}
				if len(r.CPUTemp) != 1 || r.CPUTemp[0] != 39 {
					t.Fatalf("cputemp 解码不符: %v", r.CPUTemp)
				}
				if r.Freq != nil {
					t.Fatalf("默认组合不含 freq: %v", r.Freq)
				}
				if r.Memory == nil || r.Memory.Total != 47914500 || r.Memory.Used != "6%" {
					t.Fatalf("memory 解码不符: %+v", r.Memory)
				}
				if r.Stream == nil || r.Stream.ConnectNum != 9149 ||
					r.Stream.TotalUp != 37471727016871 {
					t.Fatalf("stream 解码不符: %+v", r.Stream)
				}
				if r.VerInfo == nil || r.VerInfo.Version != "3.7.24" ||
					r.VerInfo.BuildDate != 202608141210 || r.VerInfo.Sysbit != "x64" {
					t.Fatalf("verinfo 解码不符: %+v", r.VerInfo)
				}
			},
		},
		{
			name: "opts.Type 单区块 cpu,freq",
			opts: &ikuai.ListOptions{Type: SysStatTypeCPU + "," + SysStatTypeFreq},
			data: map[string]any{
				"cpu":  []any{"10.00%", "3.00%"},
				"freq": []any{"1021", "1053"},
			},
			want: map[string]any{"TYPE": "cpu,freq"},
			assert: func(t *testing.T, r *SysStatRealtime) {
				if len(r.CPU) != 2 || r.CPU[0] != "10.00%" {
					t.Fatalf("cpu 解码不符: %v", r.CPU)
				}
				if len(r.Freq) != 2 || r.Freq[1] != "1053" {
					t.Fatalf("freq 解码不符: %v", r.Freq)
				}
				if r.Memory != nil || r.Stream != nil || r.VerInfo != nil || r.CPUTemp != nil {
					t.Fatalf("未请求区块应为 nil: %+v", r)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("sysstat", "show", tt.data)
			r, err := New(ts.Client()).SysStat.Realtime(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("Realtime 失败: %v", err)
			}
			tt.assert(t, r)
			rec := ts.CallsFor("sysstat")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %q, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec), tt.want)
		})
	}
}

// TestSysStatVerInfo 覆盖 TYPE=verinfo。
func TestSysStatVerInfo(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("sysstat", "show", map[string]any{
		"verinfo": map[string]any{
			"modelname": "", "verstring": "3.7.24 x64 Build202608141210",
			"version": "3.7.24", "build_date": 202608141210, "arch": "x86",
			"sysbit": "x64", "verflags": "", "is_enterprise": 0,
			"support_i18n": 0, "support_dingtalk": 1, "support_lcd": 0,
			"bootguide": "hd",
		},
	})
	v, err := New(ts.Client()).SysStat.VerInfo(context.Background())
	if err != nil {
		t.Fatalf("VerInfo 失败: %v", err)
	}
	if v.Version != "3.7.24" || v.Verstring != "3.7.24 x64 Build202608141210" ||
		v.BuildDate != 202608141210 || v.SupportDingTalk != 1 || v.BootGuide != "hd" {
		t.Fatalf("verinfo 解码不符: %+v", v)
	}
	wantParam(t, paramOf(t, ts.CallsFor("sysstat")[0]),
		map[string]any{"TYPE": SysStatTypeVerinfo})
}

// TestSysStatCPUFreq 覆盖 TYPE=cpu,freq。
func TestSysStatCPUFreq(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("sysstat", "show", map[string]any{
		"cpu":  []any{"10.00%", "3.03%", "3.96%"},
		"freq": []any{"1021", "1053", "1186"},
	})
	out, err := New(ts.Client()).SysStat.CPUFreq(context.Background())
	if err != nil {
		t.Fatalf("CPUFreq 失败: %v", err)
	}
	if len(out.CPU) != 3 || out.CPU[1] != "3.03%" {
		t.Fatalf("cpu 解码不符: %v", out.CPU)
	}
	if len(out.Freq) != 3 || out.Freq[2] != "1186" {
		t.Fatalf("freq 解码不符: %v", out.Freq)
	}
	wantParam(t, paramOf(t, ts.CallsFor("sysstat")[0]),
		map[string]any{"TYPE": "cpu,freq"})
}

// TestSysStatError 覆盖固件错误路径。
func TestSysStatError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("sysstat", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95006, ErrMsg: "失败", FuncName: "sysstat", Action: action}
	})
	svc := New(ts.Client())
	ctx := context.Background()
	for name, call := range map[string]func() error{
		"SysStat.Realtime": func() error {
			_, err := svc.SysStat.Realtime(ctx, nil)
			return err
		},
		"SysStat.VerInfo": func() error {
			_, err := svc.SysStat.VerInfo(ctx)
			return err
		},
		"SysStat.CPUFreq": func() error {
			_, err := svc.SysStat.CPUFreq(ctx)
			return err
		},
	} {
		err := call()
		ae, ok := ikuai.AsAPIError(err)
		if !ok || ae.Code != 95006 {
			t.Fatalf("%s 期望 APIError(95006), got %v", name, err)
		}
	}
}
