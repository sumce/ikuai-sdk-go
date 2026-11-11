package monitoring

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestCFlowSummary 覆盖 cflow.show 汇总成功路径（含固件字符串数字/
// 整数混用怪癖的解码）与参数断言。
func TestCFlowSummary(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	// 逐字复刻抓包怪癖：today/yesterday 为字符串数字，week 为 JSON 整数。
	ts.SetResponse("cflow", "show", map[string]any{
		"data": map[string]any{
			"today":     map[string]any{"llb_cnt": "0", "port_cnt": "6819694", "proto_cnt": "0", "domain_cnt": "0", "conn_cnt": "6867388"},
			"week":      map[string]any{"llb_cnt": 0, "port_cnt": 45227912, "proto_cnt": 0, "domain_cnt": 0, "conn_cnt": 45533602},
			"yesterday": map[string]any{"llb_cnt": "0", "port_cnt": "3622459", "proto_cnt": "0", "domain_cnt": "0", "conn_cnt": "3670473"},
		},
		"clean_time": 0,
	})
	out, err := New(ts.Client()).CFlow.Summary(context.Background())
	if err != nil {
		t.Fatalf("Summary 失败: %v", err)
	}
	if out.CleanTime != 0 {
		t.Fatalf("clean_time 解码不符: %d", out.CleanTime)
	}
	// 今日（字符串数字形态）。
	if got, err := out.Data.Today.PortCnt.Int64(); err != nil || got != 6819694 {
		t.Fatalf("today.port_cnt = %q/%v, 期望 6819694", out.Data.Today.PortCnt, err)
	}
	if got, err := out.Data.Today.ConnCnt.Int64(); err != nil || got != 6867388 {
		t.Fatalf("today.conn_cnt = %q/%v, 期望 6867388", out.Data.Today.ConnCnt, err)
	}
	// 本周（整数形态）。
	if got, err := out.Data.Week.PortCnt.Int64(); err != nil || got != 45227912 {
		t.Fatalf("week.port_cnt = %q/%v, 期望 45227912", out.Data.Week.PortCnt, err)
	}
	if out.Data.Week.ConnCnt.String() != "45533602" {
		t.Fatalf("week.conn_cnt 字符串形式不符: %q", out.Data.Week.ConnCnt)
	}
	// 昨日。
	if got, err := out.Data.Yesterday.PortCnt.Int64(); err != nil || got != 3622459 {
		t.Fatalf("yesterday.port_cnt = %q/%v, 期望 3622459", out.Data.Yesterday.PortCnt, err)
	}
	// 参数断言：cflow.show 发送空对象。
	rec := ts.CallsFor("cflow")[0]
	if rec.Action != "show" {
		t.Fatalf("action = %q, 期望 show", rec.Action)
	}
	wantParam(t, paramOf(t, rec), map[string]any{})
}

// TestCFlowValueUnmarshalJSON 覆盖 CFlowValue 对数字/字符串/null 的解码。
func TestCFlowValueUnmarshalJSON(t *testing.T) {
	tests := []struct {
		raw     string
		want    string
		wantN   int64
		wantErr bool // Int64() 是否应报错
	}{
		{raw: `"3622459"`, want: "3622459", wantN: 3622459},
		{raw: `45227912`, want: "45227912", wantN: 45227912},
		{raw: `""`, want: "", wantN: 0},
		{raw: `null`, want: "", wantN: 0},
		{raw: `"abc"`, want: "abc", wantErr: true},
	}
	for _, tt := range tests {
		var v CFlowValue
		if err := json.Unmarshal([]byte(tt.raw), &v); err != nil {
			t.Fatalf("解码 %s 失败: %v", tt.raw, err)
		}
		if v.String() != tt.want {
			t.Fatalf("CFlowValue(%s) = %q, 期望 %q", tt.raw, v, tt.want)
		}
		n, err := v.Int64()
		if tt.wantErr {
			if err == nil {
				t.Fatalf("Int64(%s) 应报错", tt.raw)
			}
			continue
		}
		if err != nil || n != tt.wantN {
			t.Fatalf("Int64(%s) = %d/%v, 期望 %d", tt.raw, n, err, tt.wantN)
		}
	}
}

// TestCFlowError 覆盖固件错误路径。
func TestCFlowError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("cflow", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95007, ErrMsg: "失败", FuncName: "cflow", Action: action}
	})
	_, err := New(ts.Client()).CFlow.Summary(context.Background())
	ae, ok := ikuai.AsAPIError(err)
	if !ok || ae.Code != 95007 {
		t.Fatalf("期望 APIError(95007), got %v", err)
	}
}
