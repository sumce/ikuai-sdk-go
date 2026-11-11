package advanced

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// qemu.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// QEMUTypeAll 虚拟机列表 + 宿主机资源区块（内存/CPU/网卡/分区/
	// USB/快照等）。
	QEMUTypeAll = "all"
)

// VirtualMachine 是一台虚拟机（qemu.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 未创建虚拟机（data 恒为空数组），
// 行字段无法实证（控制台 i18n service.virtual.header_title 提示行含
// name/cpu_cores/memory/disk/interface/mac/vnc_port/run_time/power_boot/
// start_priority/uefi/accel 等语义，且磁盘/网卡为嵌套结构），故行对象
// 以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type VirtualMachine struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (v *VirtualMachine) UnmarshalJSON(b []byte) error {
	v.Raw = append(v.Raw[:0], b...)
	return nil
}

// MarshalJSON 把行对象原样序列化（Add/Edit 提交参数与原始行逐字段一致）。
func (v VirtualMachine) MarshalJSON() ([]byte, error) {
	if len(v.Raw) == 0 {
		return []byte("{}"), nil
	}
	return v.Raw, nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (v VirtualMachine) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(v.Raw, &m); err != nil {
		return nil, fmt.Errorf("advanced: 解码 qemu 行失败: %w", err)
	}
	return m, nil
}

// QEMUUSBDevice 是一台可挂载给虚拟机的 USB 设备（qemu.show TYPE=all 的
// usb_info 数组元素，真机 3.7.24 实证）。
type QEMUUSBDevice struct {
	// Name USB 设备标识（真机观测 "1:1.4:046b:ff10" = 总线:端口:厂商ID:产品ID）。
	Name string `json:"name"`
	// Desc 设备描述（真机观测 " American Megatrends, Inc. Virtual Keyboard and Mouse"）。
	Desc string `json:"desc"`
	// Status 状态（真机观测 0；编码语义待真机验证，推测 0=可挂载）。
	Status int64 `json:"status"`
}

// QEMUInfo 是虚拟机宿主机的资源信息区块（qemu.show TYPE=all，真机
// 3.7.24 实证顶层字段）。
type QEMUInfo struct {
	// Total 虚拟机数（真机观测 0）。
	Total int64 `json:"total"`
	// VMs 虚拟机行列表（行结构待真机验证，见 VirtualMachine）。
	VMs []VirtualMachine `json:"data"`
	// MemAvailable 宿主机可用内存（真机观测 44859692；单位待真机验证，
	// 对照系统概况 memory.total 量级推测为 KB）。
	MemAvailable int64 `json:"MemAvailable"`
	// CPUCores 宿主机 CPU 核心数（真机观测 20）。
	CPUCores int64 `json:"cpu_cores"`
	// EthList 可选物理网卡名列表（真机实证 ["eth0",...,"usb0"]）。
	EthList []string `json:"ethlist"`
	// ImgInfo 可用镜像文件列表（真机观测为空数组，行结构待真机验证）。
	ImgInfo []json.RawMessage `json:"img_info"`
	// Interface 可桥接接口列表，每项 [接口名] 或 [接口名, 备注]
	// （真机实证，如 ["lan1"]、["lan3","服务器虚拟交换机"]）。
	Interface [][]string `json:"interface"`
	// Log 虚拟机相关日志（真机观测为空串）。
	Log string `json:"log"`
	// PartList 可用磁盘分区（真机观测为空对象 {}；结构待真机验证）。
	PartList json.RawMessage `json:"partlist"`
	// Snapshot 快照列表（真机观测为空数组，行结构待真机验证）。
	Snapshot []json.RawMessage `json:"snapshot"`
	// USBInfo 可挂载 USB 设备列表（真机实证）。
	USBInfo []QEMUUSBDevice `json:"usb_info"`
	// VFIOEthList 可直通（VFIO）网卡组，每项 [网卡名串, MAC 串]
	// （真机实证，如 ["eth8","44:a9:2c:21:c1:56"]）。
	VFIOEthList [][]string `json:"vfio_ethlist"`
}

// QEMUService 是虚拟机（qemu）的子服务。
type QEMUService struct {
	c *ikuai.Client
}

// List 获取虚拟机列表（分页）。
// 对应 /Action/call: qemu.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → 虚拟机」页面。
func (s *QEMUService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[VirtualMachine], error) {
	resp, err := s.c.Call(ctx, "qemu", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[VirtualMachine](resp.Data)
}

// Info 获取虚拟机列表与宿主机资源区块（内存/CPU/网卡/分区/USB/快照）。
// 对应 /Action/call: qemu.show（TYPE=all，真机探测实证）。
func (s *QEMUService) Info(ctx context.Context) (*QEMUInfo, error) {
	var info QEMUInfo
	if err := s.c.CallData(ctx, "qemu", "show",
		map[string]any{"TYPE": QEMUTypeAll}, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Add 新建虚拟机，返回固件分配的行 ID。vm 建议取控制台提交的原始行
// 字段（SDK 以 Raw 原样透传，不做字段加工）。
// 对应 /Action/call: qemu.add。
// ⚠️ 写动作（影响系统状态：创建磁盘/虚拟机）：仅离线测试覆盖，禁止对
// 真机调用；行字段空表推导，待真机验证。
func (s *QEMUService) Add(ctx context.Context, vm VirtualMachine) (int64, error) {
	resp, err := s.c.Call(ctx, "qemu", "add", vm)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 修改虚拟机配置（vm.Raw 应含待修改虚拟机的全部字段与 id）。
// 对应 /Action/call: qemu.edit。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *QEMUService) Edit(ctx context.Context, vm VirtualMachine) error {
	_, err := s.c.Call(ctx, "qemu", "edit", vm)
	return err
}

// Delete 删除虚拟机，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: qemu.del。
// ⚠️ 写动作（影响系统状态：销毁虚拟机/磁盘）：仅离线测试覆盖，禁止对
// 真机调用。
func (s *QEMUService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "qemu", "del", ikuai.IDParam(ids))
	return err
}
