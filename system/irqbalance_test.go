package system

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// TestIRQBalanceHardIRQ 覆盖 TYPE=hardirq 的成功路径与参数断言。
func TestIRQBalanceHardIRQ(t *testing.T) {
	tests := []struct {
		name string
		opts *ikuai.ListOptions
		want map[string]any
	}{
		{"nil-opts", nil, map[string]any{"TYPE": IRQBalanceTypeHardIRQ}},
		{"控制台分页", &ikuai.ListOptions{Limit: 20},
			map[string]any{"TYPE": IRQBalanceTypeHardIRQ, "limit": "0,20"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("irqbalance", "show", map[string]any{
				"hardirq_data": []any{
					map[string]any{"cpuid": "cpu0", "irq_num": "43", "ifname": "ice-eth0-TxRx-0", "set_cpuid": "auto"},
					map[string]any{"cpuid": "cpu1", "irq_num": "44", "ifname": "ice-eth0-TxRx-1", "set_cpuid": "auto"},
				},
			})
			rows, err := New(ts.Client()).IRQBalance.HardIRQ(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("HardIRQ 失败: %v", err)
			}
			if len(rows) != 2 || rows[0].Cpuid != "cpu0" || rows[0].IrqNum != "43" ||
				rows[0].Ifname != "ice-eth0-TxRx-0" || rows[0].SetCpuid != "auto" {
				t.Fatalf("硬中断行解码不符: %+v", rows)
			}
			wantParam(t, paramOf(t, ts.CallsFor("irqbalance")[0]), tt.want)
		})
	}
}

// TestIRQBalanceSoftIRQ 覆盖 TYPE=softirq 的成功路径与参数断言。
func TestIRQBalanceSoftIRQ(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("irqbalance", "show", map[string]any{
		"softirq_data": []any{map[string]any{
			"cpuid": "0", "freq": "2097", "phyid": "0", "coreid": "0",
			"used": "10.00%", "softirq": 1, "hardirq": 1,
		}},
	})
	rows, err := New(ts.Client()).IRQBalance.SoftIRQ(context.Background(), nil)
	if err != nil {
		t.Fatalf("SoftIRQ 失败: %v", err)
	}
	if len(rows) != 1 || rows[0].Cpuid != "0" || rows[0].Freq != "2097" ||
		rows[0].Used != "10.00%" || rows[0].Softirq != 1 || rows[0].Hardirq != 1 {
		t.Fatalf("软中断行解码不符: %+v", rows)
	}
	wantParam(t, paramOf(t, ts.CallsFor("irqbalance")[0]),
		map[string]any{"TYPE": IRQBalanceTypeSoftIRQ})
}

// TestIRQBalanceCPUList 覆盖 TYPE=cpulist,freqlist 的成功路径与参数断言。
func TestIRQBalanceCPUList(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("irqbalance", "show", map[string]any{
		"cpulist":         []any{"auto", "cpu0", "cpu1"},
		"cpufreq_support": 1,
		"current_cpufreq": "ondemand",
		"current_turbo":   0,
		"cpufreq_list":    []any{"conservative", "ondemand", "powersave", "performance", "schedutil"},
		"turbo_support":   1,
	})
	info, err := New(ts.Client()).IRQBalance.CPUList(context.Background())
	if err != nil {
		t.Fatalf("CPUList 失败: %v", err)
	}
	if len(info.Cpulist) != 3 || info.Cpulist[0] != "auto" ||
		info.CpufreqSupport != 1 || info.CurrentCpufreq != "ondemand" ||
		info.CurrentTurbo != 0 || len(info.CpufreqList) != 5 || info.TurboSupport != 1 {
		t.Fatalf("CPU 列表块解码不符: %+v", info)
	}
	wantParam(t, paramOf(t, ts.CallsFor("irqbalance")[0]),
		map[string]any{"TYPE": IRQBalanceTypeCPUList})
}

// TestIRQBalanceError 覆盖固件对空参数等回 30006 unknown TYPE 的错误路径。
func TestIRQBalanceError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("irqbalance", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 30006, ErrMsg: "unknown TYPE (data)", FuncName: "irqbalance", Action: action}
	})
	if _, err := New(ts.Client()).IRQBalance.HardIRQ(context.Background(), nil); err == nil {
		t.Fatal("hardirq 错误应上抛")
	}
	if _, err := New(ts.Client()).IRQBalance.SoftIRQ(context.Background(), nil); err == nil {
		t.Fatal("softirq 错误应上抛")
	}
	if _, err := New(ts.Client()).IRQBalance.CPUList(context.Background()); err == nil {
		t.Fatal("cpulist 错误应上抛")
	}
}
