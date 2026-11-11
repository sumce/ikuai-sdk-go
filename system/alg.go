package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// alg.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
const (
	// ALGTypeData 配置行区块选择器。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）。
	ALGTypeData = "data"
)

// ALGConfig 是 ALG（应用层网关）设置的单行配置（alg.show data 数组
// 元素，观测恒为单行 id=1）。
type ALGConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// SupportFTP FTP ALG 开关（1=开启）。
	SupportFTP int64 `json:"support_ftp"`
	// FTPPorts FTP ALG 检测的端口串（如 "21"，逗号分隔；为空表示默认）。
	FTPPorts string `json:"ftp_ports"`
	// SupportTFTP TFTP ALG 开关（1=开启）。
	SupportTFTP int64 `json:"support_tftp"`
	// TFTPPorts TFTP ALG 检测的端口串（如 "69"；为空表示默认）。
	TFTPPorts string `json:"tftp_ports"`
	// SupportSIP SIP ALG 开关（1=开启）。
	SupportSIP int64 `json:"support_sip"`
	// SIPPorts SIP ALG 检测的端口串（如 "5060"；为空表示默认）。
	SIPPorts string `json:"sip_ports"`
	// SupportH323 H.323 ALG 开关（1=开启）。
	SupportH323 int64 `json:"support_h323"`
}

// ALGService 是 ALG 设置（alg）的子服务。
type ALGService struct {
	c *ikuai.Client
}

// Get 获取 ALG 设置（FTP/TFTP/SIP/H.323 开关与端口串）。
// 对应 /Action/call: alg.show（TYPE=data），
// 即 Web 控制台「高级设置 → ALG」页面（挂在系统设置大类下）。
func (s *ALGService) Get(ctx context.Context) (*ALGConfig, error) {
	page, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(page.Rows) == 0 {
		return nil, nil
	}
	return &page.Rows[0], nil
}

// List 以分页形态获取 ALG 配置行（固件恒返回单行）。
// 对应 /Action/call: alg.show（TYPE=data, limit="offset,rows"）。
// 本 func 不支持 TYPE=total（超出的 TYPE 值会被归一为 data）。
func (s *ALGService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[ALGConfig], error) {
	param := typedShowParam(opts, ALGTypeData, nil)
	resp, err := s.c.Call(ctx, "alg", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[ALGConfig](resp.Data)
}

// Save 保存 ALG 设置（写动作）。
// 对应 /Action/call: alg.save（burp 抓包实证：整行字段 + id 全量提交，
// 响应无 Data）。cfg 应为 Get 返回并修改后的完整行。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *ALGService) Save(ctx context.Context, cfg ALGConfig) error {
	_, err := s.c.Call(ctx, "alg", "save", cfg)
	return err
}
