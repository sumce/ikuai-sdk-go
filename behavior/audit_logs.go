package behavior

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// 审计日志类 func 的 show 形态说明（真机 3.7.24 实测）：
//   - audit_url_log：仅接受无 TYPE 的 param（{} 或 {"limit":...}）；
//     TYPE=total,data 回 30006 unknown TYPE (total)；TYPE=all 额外返回
//     stream_filelist/stream_filelist_total（日志磁盘文件列表）。
//   - audit_im_log / audit_terminal_log / audit_terminal_stat：接受
//     TYPE=total,data 标准分页；audit_terminal_log 的 TYPE=all 额外返回
//     data6/total6（IPv6 双栈行）。
//   - 探测真机上 audit_terminal_stat 回 30002 Not found funcname（本固件
//     未注册该 func），照常建模（Raw/错误形态为准）。

// URLLogRow 是一条 URL 审计日志行（audit_url_log.show 的 data 元素）。
//
// ⚠️ 探测真机（3.7.24）上日志列表为空（total=-1），行字段无法从真机数据
// 核对；字段按固件 i18n web_log 列名（ip_addr/mac/host/timestamp）推导，
// 空表推导，待真机验证。
type URLLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// Timestamp 时间（unix 秒；i18n 列名 timestamp）。空表推导，待真机验证。
	Timestamp int64 `json:"timestamp"`
	// IPAddr 终端 IP 地址。空表推导，待真机验证。
	IPAddr string `json:"ip_addr"`
	// MAC 终端 MAC。空表推导，待真机验证。
	MAC string `json:"mac"`
	// Host 网址浏览记录（访问的 URL/host）。空表推导，待真机验证。
	Host string `json:"host"`
}

// URLLogFile 是 URL 审计日志的磁盘导出文件项（audit_url_log.show TYPE=all
// 的 stream_filelist 元素）。
//
// ⚠️ 探测真机上该列表为空（stream_filelist=[]），元素结构无法核对；按
// 固件 i18n web_log 的 fileName/fileSize/fileTime 推导，空表推导，待真机
// 验证。
type URLLogFile struct {
	// Name 文件名。空表推导，待真机验证。
	Name string `json:"name"`
	// Size 文件大小（字节）。空表推导，待真机验证。
	Size int64 `json:"size"`
	// Time 修改时间（unix 秒）。空表推导，待真机验证。
	Time int64 `json:"time"`
}

// URLLogFiles 是 URL 审计日志页的聚合结果（audit_url_log.show TYPE=all，
// 真机 3.7.24 实测区块 total/data/stream_filelist/stream_filelist_total）。
type URLLogFiles struct {
	// Total 日志总条数（真机观测 -1，表示后端未统计/按文件存储的形态；
	// 语义待真机验证）。
	Total int64 `json:"total"`
	// Rows 日志行。
	Rows []URLLogRow `json:"data"`
	// Files 磁盘上的日志导出文件列表。
	Files []URLLogFile `json:"stream_filelist"`
	// FilesTotal 文件总数。
	FilesTotal int64 `json:"stream_filelist_total"`
}

// IMLogRow 是一条 IM 审计日志行（audit_im_log.show 的 data 元素）。
//
// ⚠️ 探测真机（3.7.24）上日志列表为空，行字段无法从真机数据核对；字段按
// 固件 i18n im_log 列名（im_type/account/event/ip/mac/date_time）推导，
// 空表推导，待真机验证。
type IMLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// IMType IM 类型（值域见 IMType* 常量）。空表推导，待真机验证。
	IMType string `json:"im_type"`
	// Account IM 账号。空表推导，待真机验证。
	Account string `json:"account"`
	// Event IM 操作（上线/下线等描述文本）。空表推导，待真机验证。
	Event string `json:"event"`
	// IP 终端 IP。空表推导，待真机验证。
	IP string `json:"ip"`
	// MAC 终端 MAC。空表推导，待真机验证。
	MAC string `json:"mac"`
	// DateTime 时间字符串。空表推导，待真机验证。
	DateTime string `json:"date_time"`
}

// IM 类型值（固件 i18n im_log 的 all/类型下拉项实证，与固件完全一致）。
const (
	IMTypeAlitalkPC     = "Alitalk_pc"     // 阿里旺旺-PC
	IMTypeAlitalkMobile = "Alitalk_mobile" // 阿里旺旺-移动端
	IMTypeYYPC          = "yy_pc"          // YY语音-PC
	IMTypeYYMobile      = "yy_mobile"      // YY语音-移动端
	IMTypeQQPC          = "qq_pc"          // QQ-PC
	IMTypeQQMobile      = "qq_mobile"      // QQ-移动端
)

// TerminalLogRow 是一条终端上下线审计日志行（audit_terminal_log.show 的
// data 元素；字段为真机 3.7.24 实测行结构）。
type TerminalLogRow struct {
	// ID 行 ID。
	ID int64 `json:"id"`
	// Timestamp 上线时间（unix 秒）。
	Timestamp int64 `json:"timestamp"`
	// LogoutTime 下线时间（unix 秒）。
	LogoutTime int64 `json:"logout_time"`
	// OnlineTime 在线时长（秒）。
	OnlineTime int64 `json:"online_time"`
	// DateTime 上线时间字符串（如 "2026-09-25 03:54:36"）。
	DateTime string `json:"date_time"`
	// IPAddr 终端 IP。
	IPAddr string `json:"ip_addr"`
	// MAC 终端 MAC。
	MAC string `json:"mac"`
	// Comment MAC 备注（联动 mac_comment）。
	Comment string `json:"comment"`
	// Auth 认证状态（观测 0；枚举语义待真机验证）。
	Auth int64 `json:"auth"`
	// SysType 系统类型（观测 "Windows10"/"Unknown"）。
	SysType string `json:"systype"`
	// DevType 终端类型（观测 "Ruijie"/"VMware" 等）。
	DevType string `json:"devtype"`
	// ClientModel 终端型号。
	ClientModel string `json:"client_model"`
	// ClientTypeID 终端类型 ID（观测 0/102999999）。
	ClientTypeID int64 `json:"client_typeid"`
	// TotalUp 累计上传（字节）。
	TotalUp int64 `json:"total_up"`
	// TotalDown 累计下载（字节）。
	TotalDown int64 `json:"total_down"`
}

// TerminalLogAll 是终端上下线日志页的聚合结果（audit_terminal_log.show
// TYPE=all，真机 3.7.24 实测区块 data/data6/total/total6，data6 为 IPv6
// 双栈行）。
type TerminalLogAll struct {
	// Total IPv4 行总数（真机实测 229127）。
	Total int64 `json:"total"`
	// Total6 IPv6 行总数。
	Total6 int64 `json:"total6"`
	// Rows IPv4 日志行。
	Rows []TerminalLogRow `json:"data"`
	// Rows6 IPv6 日志行。
	Rows6 []TerminalLogRow `json:"data6"`
}

// TerminalStatRow 是一条终端流量统计行（audit_terminal_stat.show 的 data
// 元素）。
//
// ⚠️ 探测真机（3.7.24）对该 func 全部 show 形态回 30002 Not found
// funcname(audit_terminal_stat)——本固件未注册该 func（菜单仍挂「流量统计」
// 页）；字段按固件 i18n flow-statistics 列名（mac/uptrend/downward/comment/
// account）推导，空表推导，待真机验证。
type TerminalStatRow struct {
	// ID 行 ID。
	ID int64 `json:"id,omitempty"`
	// MAC 终端 MAC。空表推导，待真机验证。
	MAC string `json:"mac"`
	// Account 认证账号。空表推导，待真机验证。
	Account string `json:"account"`
	// Uptrend 累计上行（字节）。空表推导，待真机验证。
	Uptrend int64 `json:"uptrend"`
	// Downward 累计下行（字节）。空表推导，待真机验证。
	Downward int64 `json:"downward"`
	// Comment 备注。空表推导，待真机验证。
	Comment string `json:"comment"`
}

// AuditLogsService 是审计日志（audit_url_log/audit_im_log/audit_terminal_log/
// audit_terminal_stat，全部只读）的子服务。
type AuditLogsService struct {
	c *ikuai.Client
}

// URLLogs 获取 URL 审计日志列表（只读分页）。
// 对应 /Action/call: audit_url_log.show（无 TYPE，仅 limit/过滤键，真机
// 实测 TYPE=total,data 回 30006 unknown TYPE (total)），
// 即 Web 控制台「行为管控 → 行为记录查看 → 网址浏览记录」页面。
func (s *AuditLogsService) URLLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[URLLogRow], error) {
	resp, err := s.c.Call(ctx, "audit_url_log", "show", noTypeShowParam(opts))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[URLLogRow](resp.Data)
}

// URLLogFiles 获取 URL 审计日志页聚合信息（日志行 + 磁盘导出文件列表，
// 只读）。
// 对应 /Action/call: audit_url_log.show（TYPE=all，真机实测区块）。
func (s *AuditLogsService) URLLogFiles(ctx context.Context) (*URLLogFiles, error) {
	var out URLLogFiles
	if err := s.c.CallData(ctx, "audit_url_log", "show",
		map[string]any{"TYPE": SelectorsTypeAll}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IMLogs 获取 IM 审计日志列表（只读分页）。
// 对应 /Action/call: audit_im_log.show（TYPE=total,data），
// 即 Web 控制台「行为管控 → 行为记录查看 → IM 上线记录」页面。
func (s *AuditLogsService) IMLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[IMLogRow], error) {
	resp, err := s.c.Call(ctx, "audit_im_log", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[IMLogRow](resp.Data)
}

// TerminalLogs 获取终端上下线审计日志列表（只读分页，IPv4）。
// 对应 /Action/call: audit_terminal_log.show（TYPE=total,data，真机实测
// total=229127），
// 即 Web 控制台「行为管控 → 行为记录查看 → 终端上下线记录」页面。
func (s *AuditLogsService) TerminalLogs(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[TerminalLogRow], error) {
	resp, err := s.c.Call(ctx, "audit_terminal_log", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[TerminalLogRow](resp.Data)
}

// TerminalLogsAll 获取终端上下线审计日志聚合信息（IPv4/IPv6 双栈，只读）。
// 对应 /Action/call: audit_terminal_log.show（TYPE=all，真机实测区块
// data/data6/total/total6）。
func (s *AuditLogsService) TerminalLogsAll(ctx context.Context, opts *ikuai.ListOptions) (*TerminalLogAll, error) {
	p := typedShowParam(opts, SelectorsTypeAll, nil)
	var out TerminalLogAll
	if err := s.c.CallData(ctx, "audit_terminal_log", "show", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TerminalStats 获取终端流量统计列表（只读分页）。
// 对应 /Action/call: audit_terminal_stat.show（TYPE=total,data），
// 即 Web 控制台「行为管控 → 行为记录查看 → 流量统计」页面。
// 注意：探测真机（3.7.24）回 30002 Not found funcname，字段为空表推导，
// 待真机验证。
func (s *AuditLogsService) TerminalStats(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[TerminalStatRow], error) {
	resp, err := s.c.Call(ctx, "audit_terminal_stat", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[TerminalStatRow](resp.Data)
}
