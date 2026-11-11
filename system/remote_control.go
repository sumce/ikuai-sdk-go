package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// remote_control.show 的 TYPE 选择器常量（真机 3.7.24 验证）。
const (
	// RemoteControlTypeData 配置行区块选择器。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）。
	RemoteControlTypeData = "data"
)

// RemoteControlConfig 是远程控制（远程管理）的单行配置
// （remote_control.show data 数组元素，观测恒为单行 id=1）。
type RemoteControlConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// HTTPPort Web 管理端口（观测值 182）。
	HTTPPort int64 `json:"http_port"`
	// HTTPSPort HTTPS 管理端口（观测值 443）。
	HTTPSPort int64 `json:"https_port"`
	// ForceHTTPS 强制 HTTPS 开关（0=关闭）。
	ForceHTTPS int64 `json:"force_https"`
	// OpenWANWeb 外网 Web 管理开关（观测值 2；枚举语义待真机验证，
	// 推测 0=关闭 1=开启 2=仅 HTTPS）。
	OpenWANWeb int64 `json:"open_wanweb"`
	// OpenSSHD SSH 远程开关（0=关闭）。
	OpenSSHD int64 `json:"open_sshd"`
	// SSHDPort SSH 端口（观测值 22）。
	SSHDPort int64 `json:"sshd_port"`
	// SSHDPasswd SSH 登录密码（为空表示使用 Web 账号密码）。
	SSHDPasswd string `json:"sshd_passwd"`
	// OpenFTP FTP 服务开关（0=关闭）。
	OpenFTP int64 `json:"open_ftp"`
	// FTPPort FTP 端口（观测值 21）。
	FTPPort int64 `json:"ftp_port"`
	// FTPAccess FTP 访问范围（观测值 1；枚举语义待真机验证）。
	FTPAccess int64 `json:"ftp_access"`
	// OpenTelnetd Telnet 开关（0=关闭）。
	OpenTelnetd int64 `json:"open_telnetd"`
}

// RemoteControlService 是远程控制（remote_control）的子服务。
type RemoteControlService struct {
	c *ikuai.Client
}

// Get 获取远程控制配置（Web/SSH/FTP/Telnet 端口与开关）。
// 对应 /Action/call: remote_control.show（TYPE=data；真机对空参数同样
// 返回 data 区块），
// 即 Web 控制台「系统设置 → 远程控制」页面。
func (s *RemoteControlService) Get(ctx context.Context) (*RemoteControlConfig, error) {
	page, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(page.Rows) == 0 {
		return nil, nil
	}
	return &page.Rows[0], nil
}

// List 以分页形态获取远程控制配置行（固件恒返回单行）。
// 对应 /Action/call: remote_control.show（TYPE=data, limit="offset,rows"）。
// 本 func 不支持 TYPE=total（超出的 TYPE 值会被归一为 data）。
func (s *RemoteControlService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[RemoteControlConfig], error) {
	param := typedShowParam(opts, RemoteControlTypeData, nil)
	resp, err := s.c.Call(ctx, "remote_control", "show", param)
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[RemoteControlConfig](resp.Data)
}

// Save 保存远程控制配置（写动作）。
// 对应 /Action/call: remote_control.save（action 名按同类单行配置 func
// 的 alg.save 推断）。cfg 应为 Get 返回并修改后的完整行。
// ⚠️ 写动作：action/字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *RemoteControlService) Save(ctx context.Context, cfg RemoteControlConfig) error {
	_, err := s.c.Call(ctx, "remote_control", "save", cfg)
	return err
}
