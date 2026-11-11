package advanced

import (
	"context"
	"encoding/json"

	"github.com/sumce/ikuai-sdk-go"
)

// ftp_server.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// FTPTypeAll 用户行列表 + 服务配置区块（open_ftp/ftp_port/ftp_access/
	// partname）。
	FTPTypeAll = "all"
)

// smbd.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// SMBTypeData Samba 服务配置区块。本 func 不接受 TYPE=total
	// （真机回 30006 unknown TYPE (total)）；TYPE=all 与 data 等价。
	SMBTypeData = "data"
)

// FTPConfig 是 FTP 服务的全局配置（ftp_server.show TYPE=all 顶层字段，
// 真机 3.7.24 实证）。
type FTPConfig struct {
	// OpenFTP FTP 服务开关（真机观测 0；0=关闭，1=开启）。
	OpenFTP int64 `json:"open_ftp"`
	// FTPPort 服务端口（真机观测 21）。
	FTPPort int64 `json:"ftp_port"`
	// FTPAccess 允许外网访问（真机观测 1；0=拒绝，1=允许，语义对照
	// 控制台 i18n external_network「允许外网访问」）。
	FTPAccess int64 `json:"ftp_access"`
	// Partname 可选分区名列表（真机实证 ["A"]，控制台用于选择共享目录
	// 所在分区；需先在「磁盘管理-磁盘分区」新建普通存储分区）。
	Partname []string `json:"partname"`
}

// FTPUser 是一个 FTP 用户账号（ftp_server.show data 数组元素）。
//
// 行字段空表推导，待真机验证：真机 3.7.24 未配置 FTP 用户（data 恒为
// 空数组），字段名按控制台 i18n service.ftp.header_title（username/
// ftp_passwd/ftp_authority/File_Directory/external_network）+ 开源实现
// 核对（ikuai-cli ftpUserInputFields）推导。
type FTPUser struct {
	// ID 行 ID（edit/del 必填；add 时无需填写）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，空表推导，待真机验证）。
	Enabled string `json:"enabled"`
	// Username 用户名（匿名用户固定填 anonymous，密码任意，控制台
	// i18n ftp_passwd_tip）。
	Username string `json:"username"`
	// Passwd 密码（控制台 i18n ftp_passwd「密码」）。
	Passwd string `json:"passwd"`
	// Permission 权限（只读/读写，控制台 i18n ftp_read_only/ftp_read_write；
	// 取值编码待真机验证）。
	Permission string `json:"permission"`
	// HomeDir FTP 文件目录（分区/路径，控制台 i18n File_Directory
	// 「FTP文件目录」/ share_directory「共享目录」）。
	HomeDir string `json:"home_dir"`
	// Upload 上行限速（0 表示不限速；单位待真机验证）。
	Upload int64 `json:"upload"`
	// Download 下行限速（0 表示不限速；单位待真机验证）。
	Download int64 `json:"download"`
	// Tagname 名称/备注（空表推导，待真机验证）。
	Tagname string `json:"tagname"`
}

// IsEnabled 用户启用状态为 "yes"。
func (u FTPUser) IsEnabled() bool { return u.Enabled == EnabledYes }

// ftpAllWrapper 对应 ftp_server.show TYPE=all 的 Data 形态
// （{"data":[...],"open_ftp":..,...}，真机实证）。
type ftpAllWrapper struct {
	FTPConfig
	Total int64     `json:"total"`
	Rows  []FTPUser `json:"data"`
}

// SMBConfig 是 Samba 服务的配置行（smbd.show data 数组元素，观测恒为
// 单行 id=1，真机 3.7.24 实证全部字段）。
type SMBConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// Enabled Samba 服务开关（"yes"/"no"，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// Workgroup 工作组（真机观测 "WORKGROUP"，控制台 i18n workgroup）。
	Workgroup string `json:"workgroup"`
	// WSDD2 网络发现开关（wsdd2 服务，真机观测 1；0=关闭 1=开启，
	// 控制台 i18n net_discovery「网络发现」）。
	WSDD2 int64 `json:"wsdd2"`
	// Interface 监听接口（空串=全部接口，真机观测为空串）。
	Interface string `json:"interface"`
	// Access 允许外网访问（真机观测 1；0=拒绝 1=允许，控制台 i18n
	// network_access「允许外网访问」）。
	Access int64 `json:"access"`
}

// IsEnabled Samba 服务开关为 "yes"。
func (c SMBConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// SMBInfo 是 Samba 服务的完整区块（smbd.show TYPE=data，真机 3.7.24
// 实证顶层字段）。
type SMBInfo struct {
	// Configs 配置行列表（观测恒为单行）。
	Configs []SMBConfig `json:"data"`
	// DirData 共享目录列表（真机观测为空数组，行结构待真机验证；
	// 共享需先在「磁盘管理-磁盘分区」新建普通存储分区）。
	DirData []json.RawMessage `json:"dir_data"`
	// DirTotal 共享目录数（真机观测 0）。
	DirTotal int64 `json:"dir_total"`
	// Interface 可选接口列表，每项 [接口名] 或 [接口名, 备注]
	// （真机实证）。
	Interface [][]string `json:"interface"`
}

// HTTPServerRow 是一个 HTTP 文件服务实例（http_server.show data 数组
// 元素）。
//
// 行字段空表推导，待真机验证：真机 3.7.24 未配置 HTTP 服务（data 恒为
// 空数组），字段名按控制台 i18n service.http.header_title（File_Directory=
// 文件目录、Service_port=服务端口、Directory_permissions=目录浏览权限、
// domain=服务域名、network_access=外网访问）+ 开源实现核对（ikuai-cli
// httpUserInputFields）推导。
type HTTPServerRow struct {
	// ID 行 ID（edit/del 必填；add 时无需填写）。
	ID int64 `json:"id,omitempty"`
	// Enabled 是否启用（"yes"/"no"，空表推导，待真机验证）。
	Enabled string `json:"enabled"`
	// Tagname 名称/备注（空表推导，待真机验证）。
	Tagname string `json:"tagname"`
	// HTTPPort 服务端口（控制台 i18n Service_port「服务端口」）。
	HTTPPort int64 `json:"http_port"`
	// ServerName 服务域名（控制台 i18n domain「服务域名」）。
	ServerName string `json:"server_name"`
	// SSL SSL 开关（0=关闭 1=开启，空表推导，待真机验证）。
	SSL int64 `json:"ssl_on"`
	// AutoIndex 目录浏览权限（控制台 i18n Directory_permissions
	// 「目录浏览权限」；0=关闭 1=开启，待真机验证）。
	AutoIndex int64 `json:"autoindex"`
	// Download 访问方式（控制台 i18n interview_method「访问方式」；
	// 编码语义待真机验证）。
	Download int64 `json:"download"`
	// HomeDir 文件目录（控制台 i18n File_Directory「文件目录」）。
	HomeDir string `json:"home_dir"`
	// Access 外网访问（控制台 i18n network_access「外网访问」；
	// 0=拒绝 1=允许，待真机验证）。
	Access int64 `json:"access"`
}

// IsEnabled 服务启用状态为 "yes"。
func (r HTTPServerRow) IsEnabled() bool { return r.Enabled == EnabledYes }

// HTTPOptions 是 HTTP 服务的可选值区块（http_server.show TYPE=all，
// 真机 3.7.24 实证顶层字段）。
type HTTPOptions struct {
	// Total 服务实例数（真机观测 0）。
	Total int64 `json:"total"`
	// Rows 服务实例行列表。
	Rows []HTTPServerRow `json:"data"`
	// Partname 可选分区名列表（真机实证 ["A"]；需先在「磁盘管理-磁盘
	// 分区」新建普通存储分区）。
	Partname []string `json:"partname"`
}

// FileServicesService 是文件服务（ftp_server/smbd/http_server）的子
// 服务，覆盖 Web 控制台「高级应用」下的 FTP 服务、Samba 服务、HTTP
// 服务三个页面（相邻小功能合并进一个子服务，类型各自定义）。
type FileServicesService struct {
	c *ikuai.Client
}

// ---- FTP（ftp_server）----

// FTPList 获取 FTP 用户列表（分页）。
// 对应 /Action/call: ftp_server.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → FTP 服务」页面。
func (s *FileServicesService) FTPList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[FTPUser], error) {
	resp, err := s.c.Call(ctx, "ftp_server", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[FTPUser](resp.Data)
}

// FTPGetConfig 获取 FTP 服务全局配置（开关/端口/外网访问/可选分区）。
// 对应 /Action/call: ftp_server.show（TYPE=all，真机探测实证返回
// {"data":[],"open_ftp":0,"ftp_port":21,"ftp_access":1,"partname":["A"],
// "total":0}）。
func (s *FileServicesService) FTPGetConfig(ctx context.Context) (*FTPConfig, error) {
	var w ftpAllWrapper
	if err := s.c.CallData(ctx, "ftp_server", "show",
		map[string]any{"TYPE": FTPTypeAll}, &w); err != nil {
		return nil, err
	}
	return &w.FTPConfig, nil
}

// FTPSaveConfig 保存 FTP 服务全局配置（开关/端口/外网访问）。
// 对应 /Action/call: ftp_server.edit。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；action 名（edit）与字段
// 集待真机验证。
func (s *FileServicesService) FTPSaveConfig(ctx context.Context, cfg FTPConfig) error {
	_, err := s.c.Call(ctx, "ftp_server", "edit", cfg)
	return err
}

// FTPAddUser 新增 FTP 用户，返回固件分配的行 ID。
// 对应 /Action/call: ftp_server.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *FileServicesService) FTPAddUser(ctx context.Context, user FTPUser) (int64, error) {
	resp, err := s.c.Call(ctx, "ftp_server", "add", user)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// FTPEditUser 修改 FTP 用户。
// 对应 /Action/call: ftp_server.edit（user.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *FileServicesService) FTPEditUser(ctx context.Context, user FTPUser) error {
	_, err := s.c.Call(ctx, "ftp_server", "edit", user)
	return err
}

// FTPDeleteUser 删除 FTP 用户，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: ftp_server.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *FileServicesService) FTPDeleteUser(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ftp_server", "del", ikuai.IDParam(ids))
	return err
}

// ---- SMB（smbd）----

// SMBGet 获取 Samba 服务配置区块（配置行、共享目录、可选接口）。
// 对应 /Action/call: smbd.show（TYPE=data，真机探测实证），
// 即 Web 控制台「高级应用 → Samba 服务」页面。
func (s *FileServicesService) SMBGet(ctx context.Context) (*SMBInfo, error) {
	var info SMBInfo
	if err := s.c.CallData(ctx, "smbd", "show",
		typedShowParam(nil, SMBTypeData, nil), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// SMBGetConfig 获取 Samba 服务配置行（固件恒返回单行；列表为空时返回
// nil）。
// 对应 /Action/call: smbd.show（TYPE=data）。
func (s *FileServicesService) SMBGetConfig(ctx context.Context) (*SMBConfig, error) {
	info, err := s.SMBGet(ctx)
	if err != nil {
		return nil, err
	}
	if len(info.Configs) == 0 {
		return nil, nil
	}
	return &info.Configs[0], nil
}

// SMBEdit 保存 Samba 服务配置（cfg 应为 SMBGetConfig 返回并修改后的
// 完整行）。
// 对应 /Action/call: smbd.edit（cfg.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；action 名（edit）待真机验证。
func (s *FileServicesService) SMBEdit(ctx context.Context, cfg SMBConfig) error {
	_, err := s.c.Call(ctx, "smbd", "edit", cfg)
	return err
}

// ---- HTTP（http_server）----

// HTTPList 获取 HTTP 文件服务实例列表（分页）。
// 对应 /Action/call: http_server.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → HTTP 服务」页面。
func (s *FileServicesService) HTTPList(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[HTTPServerRow], error) {
	resp, err := s.c.Call(ctx, "http_server", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[HTTPServerRow](resp.Data)
}

// HTTPGetOptions 获取 HTTP 服务实例列表与可选分区区块。
// 对应 /Action/call: http_server.show（TYPE=all，真机探测实证返回
// {"data":[],"partname":["A"],"total":0}）。
func (s *FileServicesService) HTTPGetOptions(ctx context.Context) (*HTTPOptions, error) {
	var opts HTTPOptions
	if err := s.c.CallData(ctx, "http_server", "show",
		map[string]any{"TYPE": "all"}, &opts); err != nil {
		return nil, err
	}
	return &opts, nil
}

// HTTPAdd 新增 HTTP 文件服务实例，返回固件分配的行 ID。
// 对应 /Action/call: http_server.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *FileServicesService) HTTPAdd(ctx context.Context, row HTTPServerRow) (int64, error) {
	resp, err := s.c.Call(ctx, "http_server", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// HTTPEdit 修改 HTTP 文件服务实例。
// 对应 /Action/call: http_server.edit（row.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *FileServicesService) HTTPEdit(ctx context.Context, row HTTPServerRow) error {
	_, err := s.c.Call(ctx, "http_server", "edit", row)
	return err
}

// HTTPDelete 删除 HTTP 文件服务实例，支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: http_server.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *FileServicesService) HTTPDelete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "http_server", "del", ikuai.IDParam(ids))
	return err
}
