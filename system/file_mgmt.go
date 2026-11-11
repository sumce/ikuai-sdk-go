package system

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// file_mgmt.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
const (
	// FileTypeData 目录列表区块选择器。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）。
	FileTypeData = "data"
)

// FileEntry 是目录中的一个文件/目录条目（file_mgmt.show data 数组元素）。
type FileEntry struct {
	// StSize 文件大小（字节；目录条目观测为其内容总量，语义待真机验证）。
	StSize int64 `json:"st_size"`
	// StInode inode 号（目录观测值为 2）。
	StInode int64 `json:"st_inode"`
	// FName 文件/目录名（如 "A"）。
	FName string `json:"f_name"`
	// StMtime 最后修改时间（unix 秒）。
	StMtime int64 `json:"st_mtime"`
	// StType 条目类型（观测值 0=目录；其余枚举待真机验证，推测 1=文件）。
	StType int64 `json:"st_type"`
}

// IsDir 按 StType 推断是否为目录（观测值 0=目录，推断待真机验证）。
func (f FileEntry) IsDir() bool { return f.StType == 0 }

// FileService 是文件管理（file_mgmt）的子服务。
type FileService struct {
	c *ikuai.Client
}

// ListDir 列出指定目录下的文件条目。
// 对应 /Action/call: file_mgmt.show（param={"path":path,"TYPE":"data",
// "ORDER_BY":"","ORDER":""}，burp 抓包实证），即 Web 控制台
// 「系统设置 → 文件管理」页面。
//
// 固件怪癖（真机 3.7.24 验证）：path="/"（挂载分区根目录）返回
// {"data":[条目...]}；其余路径/缺省 path 返回 {"data":{}}（data 为空
// 对象而非数组），此时返回空列表而不报错。
// path 为空时返回 ikuai.ErrInvalidParam。
func (s *FileService) ListDir(ctx context.Context, path string, opts *ikuai.ListOptions) ([]FileEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("system: path 不能为空: %w", ikuai.ErrInvalidParam)
	}
	param := typedShowParam(opts, FileTypeData, map[string]any{
		"path":     path,
		"ORDER_BY": "",
		"ORDER":    "",
	})
	resp, err := s.c.Call(ctx, "file_mgmt", "show", param)
	if err != nil {
		return nil, err
	}
	// data 兼容数组与空对象两种固件形态。
	var block struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &block); err != nil {
		return nil, fmt.Errorf("system: decode file_mgmt data: %w", err)
	}
	raw := block.Data
	if len(raw) == 0 || raw[0] != '[' {
		return nil, nil
	}
	var rows []FileEntry
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("system: decode file_mgmt rows: %w", err)
	}
	return rows, nil
}
