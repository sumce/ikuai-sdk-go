package advanced

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// qemuAllSample 构造 qemu.show TYPE=all 的响应 Data（真机 3.7.24 探测
// 实证顶层字段，行数据为空表推导样例）。
func qemuAllSample() map[string]any {
	return map[string]any{
		"MemAvailable": 44859692,
		"cpu_cores":    20,
		"data":         []any{map[string]any{"id": 1, "name": "win10", "status": "running"}},
		"ethlist":      []any{"eth0", "eth1", "usb0"},
		"img_info":     []any{map[string]any{"name": "debian.img"}},
		"interface":    []any{[]any{"lan1"}, []any{"lan3", "服务器虚拟交换机"}},
		"log":          "",
		"partlist":     map[string]any{},
		"snapshot":     []any{},
		"total":        1,
		"usb_info": []any{map[string]any{
			"name":   "1:1.4:046b:ff10",
			"desc":   " American Megatrends, Inc. Virtual Keyboard and Mouse",
			"status": 0,
		}},
		"vfio_ethlist": []any{[]any{"eth8", "44:a9:2c:21:c1:56"}},
	}
}

// TestQEMUListInfo 覆盖 qemu.show 的两种 TYPE 成功路径（真机实证
// TYPE=total,data 与 TYPE=all 均可用）。
func TestQEMUListInfo(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("qemu", "show", qemuAllSample())
	svc := New(ts.Client()).QEMU
	ctx := context.Background()

	page, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 {
		t.Fatalf("分页解码不符: %+v", page)
	}
	m, err := page.Rows[0].Map()
	if err != nil {
		t.Fatalf("行 Map 失败: %v", err)
	}
	if m["name"] != "win10" || m["status"] != "running" {
		t.Fatalf("行原始 JSON 不符: %v", m)
	}

	info, err := svc.Info(ctx)
	if err != nil {
		t.Fatalf("Info 失败: %v", err)
	}
	if info.MemAvailable != 44859692 || info.CPUCores != 20 || info.Total != 1 ||
		len(info.VMs) != 1 || len(info.EthList) != 3 || info.EthList[2] != "usb0" ||
		len(info.ImgInfo) != 1 || len(info.Interface) != 2 ||
		info.Interface[1][1] != "服务器虚拟交换机" || info.Log != "" ||
		len(info.PartList) == 0 || len(info.Snapshot) != 0 ||
		len(info.USBInfo) != 1 || info.USBInfo[0].Name != "1:1.4:046b:ff10" ||
		len(info.VFIOEthList) != 1 || info.VFIOEthList[0][0] != "eth8" {
		t.Fatalf("TYPE=all 区块解码不符: %+v", info)
	}

	recs := ts.CallsFor("qemu")
	wantParam(t, paramOf(t, recs[0]), map[string]any{"TYPE": "total,data", "limit": "0,20"})
	wantParam(t, paramOf(t, recs[1]), map[string]any{"TYPE": "all"})
}

// TestQEMUAddEditDelete 覆盖 add/edit/del（写动作，仅离线，Raw 原样透传
// 断言）。
func TestQEMUAddEditDelete(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("qemu", "add", map[string]any{})
	ts.SetResponse("qemu", "edit", map[string]any{})
	ts.SetResponse("qemu", "del", map[string]any{})
	svc := New(ts.Client()).QEMU
	ctx := context.Background()

	raw := json.RawMessage(`{"id":1,"name":"win10","cpu_cores":4,"memory":4096}`)
	vm := VirtualMachine{}
	if err := json.Unmarshal(raw, &vm); err != nil {
		t.Fatalf("构造行失败: %v", err)
	}
	rowID, err := svc.Add(ctx, vm)
	if err != nil || rowID != 0 {
		t.Fatalf("静态响应无 RowId, 应返回 0: %v (rowID=%d)", err, rowID)
	}
	if err := svc.Edit(ctx, vm); err != nil {
		t.Fatalf("Edit 失败: %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	recs := ts.CallsFor("qemu")
	if recs[0].Action != "add" || recs[1].Action != "edit" || recs[2].Action != "del" {
		t.Fatalf("action 序列不符: %s,%s,%s", recs[0].Action, recs[1].Action, recs[2].Action)
	}
	for i := 0; i < 2; i++ {
		var m map[string]any
		if err := json.Unmarshal(recs[i].Param, &m); err != nil {
			t.Fatalf("解析 param 失败: %v", err)
		}
		if m["name"] != "win10" || m["memory"] != float64(4096) || m["cpu_cores"] != float64(4) {
			t.Fatalf("Raw 透传不符: %v", m)
		}
	}
	wantParam(t, paramOf(t, recs[2]), map[string]any{"id": float64(1)})
}

// TestQEMUError 覆盖固件错误路径。
func TestQEMUError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("qemu", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95003, ErrMsg: "失败", FuncName: "qemu", Action: action}
	})
	if _, err := New(ts.Client()).QEMU.List(context.Background(), nil); err == nil {
		t.Fatal("错误路径应报错")
	} else {
		errIsAPICode(t, err, 95003)
	}
}
