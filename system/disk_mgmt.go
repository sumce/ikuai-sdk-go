package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// disk_mgmt.show 的 TYPE 选择器常量（真机 3.7.24 验证）。
const (
	// DiskTypeData 磁盘列表区块选择器。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）。
	DiskTypeData = "data"
)

// DiskFilesys 是分区的文件系统信息（disk_mgmt partition.filesys）。
type DiskFilesys struct {
	// FSType 文件系统类型（观测值 "ext4"）。
	FSType string `json:"fs_type"`
	// FSUUID 文件系统 UUID。
	FSUUID string `json:"fs_uuid"`
}

// DiskMounted 是分区的挂载信息（disk_mgmt partition.mounted）。
// 各容量字段观测口径为 KB（与 df 输出一致，如 mt_total=45520896 ≈ 43.4GiB），
// 单位待真机验证。
type DiskMounted struct {
	// MTAvail 可用空间。
	MTAvail string `json:"mt_avail"`
	// MTTotal 总空间。
	MTTotal string `json:"mt_total"`
	// MTUsed 已用空间。
	MTUsed string `json:"mt_used"`
	// MTUses 已用空间百分比文本（观测值 "3"/"0"，无百分号）。
	MTUses string `json:"mt_uses"`
	// MTPurpose 挂载用途（观测值 "0"=系统/未标记，"1"=数据分区；待真机验证）。
	MTPurpose string `json:"mt_purpose"`
	// MTName 挂载名（数据分区观测值为 "A"，其余为空）。
	MTName string `json:"mt_name"`
}

// DiskPartition 是磁盘上的一个分区行（disk_mgmt partition 数组元素）。
type DiskPartition struct {
	// Mounted 挂载信息。
	Mounted DiskMounted `json:"mounted"`
	// Name 分区设备名（如 sda3/sda5/sda6）。
	Name string `json:"name"`
	// Formating 正在格式化标志（1=格式化进行中）。
	Formating int64 `json:"formating"`
	// Type 分区类型（观测值 0=系统分区，2=数据分区；枚举语义待真机验证）。
	Type int64 `json:"type"`
	// Size 分区大小（字节；观测值 52428800 = 50MiB）。
	Size int64 `json:"size"`
	// Filesys 文件系统信息。
	Filesys DiskFilesys `json:"filesys"`
}

// DiskInfo 是一块磁盘的信息行（disk_mgmt.show data 数组元素）。
type DiskInfo struct {
	// Type 分区表类型（观测值 "dos"）。
	Type string `json:"type"`
	// Model 磁盘型号（如 "KINGSTON OM8P0S3"）。
	Model string `json:"model"`
	// Disk 磁盘设备名（如 "sda"）。
	Disk string `json:"disk"`
	// System 系统盘标志（1=装有系统的磁盘）。
	System int64 `json:"system"`
	// BlockSize 块大小（字节，观测值 512）。
	BlockSize int64 `json:"block_size"`
	// Creating 正在创建分区标志（1=分区操作进行中）。
	Creating int64 `json:"creating"`
	// Partition 分区列表。
	Partition []DiskPartition `json:"partition"`
	// Size 磁盘总容量（字节；观测值 128035676160 ≈ 119.3GiB）。
	Size int64 `json:"size"`
}

// DiskService 是磁盘分区（disk_mgmt）的子服务。
type DiskService struct {
	c *ikuai.Client
}

// List 获取磁盘与分区列表。
// 对应 /Action/call: disk_mgmt.show（TYPE=data），
// 即 Web 控制台「系统设置 → 磁盘分区」页面。
// opts 为 nil 时仅发送 {"TYPE":"data"}；本 func 不支持 TYPE=total
// （超出的 TYPE 值会被归一为 data），分页 limit 固件接受（真机验证）。
func (s *DiskService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DiskInfo], error) {
	param := typedShowParam(opts, DiskTypeData, nil)
	resp, err := s.c.Call(ctx, "disk_mgmt", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[DiskInfo](resp.Data)
}

// Get 获取磁盘列表的首块磁盘（控制台场景通常只有一块系统盘）；
// 列表为空时返回 nil。
func (s *DiskService) Get(ctx context.Context) (*DiskInfo, error) {
	page, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(page.Rows) == 0 {
		return nil, nil
	}
	return &page.Rows[0], nil
}
