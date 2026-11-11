package system

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// backup.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
const (
	// BackupTypeDataDisk 控制台「配置备份」页的组合（data,disk）：
	// 备份文件列表 + 备份分区磁盘用量。
	BackupTypeDataDisk = "data,disk"
	// BackupTypeDataTotal 标准分页形态（真机验证固件同样接受）。
	BackupTypeDataTotal = "total,data"
)

// BackupFile 是一个配置备份文件的行（backup.show data 数组元素）。
//
// 真机（3.7.24，无自动备份文件）与抓包中的 data 均为空数组，行字段无法
// 实证，故行对象以 Raw 原样保留，字段语义待真机验证后再细化强类型。
type BackupFile struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (f *BackupFile) UnmarshalJSON(b []byte) error {
	f.Raw = append(f.Raw[:0], b...)
	return nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (f BackupFile) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(f.Raw, &m); err != nil {
		return nil, fmt.Errorf("system: 解码 backup 行失败: %w", err)
	}
	return m, nil
}

// BackupStatus 是配置备份页的聚合结果（backup.show TYPE=data,disk）。
type BackupStatus struct {
	// Data 备份文件列表。
	Data []BackupFile `json:"data"`
	// DiskTotal 备份分区总容量（观测值 44454，推测单位 MB，待真机验证）。
	DiskTotal int64 `json:"disk_total"`
	// DiskUsed 备份分区已用容量（观测值 1417，单位待真机验证）。
	DiskUsed int64 `json:"disk_used"`
	// DiskAvail 备份分区可用容量（观测值 39453，单位待真机验证）。
	DiskAvail int64 `json:"disk_avail"`
}

// BackupService 是备份（backup）的子服务。
type BackupService struct {
	c *ikuai.Client
}

// Status 获取配置备份状态（备份文件列表 + 备份分区磁盘用量）。
// 对应 /Action/call: backup.show（TYPE=data,disk，burp 抓包实证），
// 即 Web 控制台「系统设置 → 配置备份」页面。
func (s *BackupService) Status(ctx context.Context) (*BackupStatus, error) {
	var out BackupStatus
	if err := s.c.CallData(ctx, "backup", "show",
		map[string]any{"TYPE": BackupTypeDataDisk}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List 以标准分页形态获取备份文件列表。
// 对应 /Action/call: backup.show（TYPE=total,data, limit="offset,rows"；
// 真机验证）。opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *BackupService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[BackupFile], error) {
	param := pageShowParam(opts, nil)
	resp, err := s.c.Call(ctx, "backup", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[BackupFile](resp.Data)
}
