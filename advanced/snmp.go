package advanced

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// netsnmp 的读写权限常量（固件 rw 字段，真机观测 "ro"；取值集对照
// 控制台 i18n read_write: r=只读、rw=读写）。
const (
	SNMPRWReadOnly  = "ro"
	SNMPRWReadWrite = "rw"
)

// SNMPConfig 是 SNMP 服务的单行配置（netsnmp.show data 数组元素，观测
// 恒为单行 id=1，真机 3.7.24 实证全部字段）。
type SNMPConfig struct {
	// ID 行 ID（观测恒为 1）。
	ID int64 `json:"id"`
	// Enabled SNMP 服务开关（"yes"/"no"，真机观测 "no"）。
	Enabled string `json:"enabled"`
	// ListenPort 监听端口（真机观测 161）。
	ListenPort int64 `json:"listen_port"`
	// SysName 系统信息（SNMP sysName.0）。
	SysName string `json:"sysname"`
	// SysContact 联系信息（SNMP sysContact.0）。
	SysContact string `json:"syscontact"`
	// SysLocation 物理位置信息（SNMP sysLocation.0）。
	SysLocation string `json:"syslocation"`
	// Version 协议版本（真机观测 2 = v2c；3 = v3，待真机验证）。
	Version int64 `json:"version"`
	// Community v2c 团体字（真机观测 "public"）。
	Community string `json:"community"`
	// Source 可访问设备 IP（允许查询本机 SNMP 的来源，IP/掩码格式
	// "1.1.1.1/24"；真机观测为空串）。
	Source string `json:"source"`
	// RW 读写权限（ro=只读 / rw=读写，真机观测 "ro"）。
	RW string `json:"rw"`
	// Username v3 用户名（真机观测为空串）。
	Username string `json:"username"`
	// Security v3 安全级别（无认证/认证/认证且加密；取值集待真机验证）。
	Security string `json:"security"`
	// AuthProto v3 认证模式（md5/sha，待真机验证）。
	AuthProto string `json:"auth_proto"`
	// AuthPass v3 认证密码。
	AuthPass string `json:"auth_pass"`
	// PrivProto v3 加密模式（des/aes，待真机验证）。
	PrivProto string `json:"priv_proto"`
	// PrivPass v3 加密密码。
	PrivPass string `json:"priv_pass"`
}

// IsEnabled SNMP 服务开关为 "yes"。
func (c SNMPConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// IsV3 是否为 SNMPv3 配置（v3 才有用户名/认证/加密字段）。
func (c SNMPConfig) IsV3() bool { return c.Version == 3 }

// SNMPService 是 SNMP 服务（netsnmp）的子服务。
type SNMPService struct {
	c *ikuai.Client
}

// Get 获取 SNMP 服务配置（开关、监听端口、团体字、v3 用户与认证加密）。
// 对应 /Action/call: netsnmp.show（TYPE=total,data，真机探测实证返回
// {"total":1,"data":[单行]}），即 Web 控制台「高级应用 → SNMP 服务」页面。
// 固件恒返回单行；列表为空时返回 nil。
func (s *SNMPService) Get(ctx context.Context) (*SNMPConfig, error) {
	page, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(page.Rows) == 0 {
		return nil, nil
	}
	return &page.Rows[0], nil
}

// List 以分页形态获取 SNMP 配置行（固件恒返回单行）。
// 对应 /Action/call: netsnmp.show（TYPE=total,data, limit="offset,rows"）。
func (s *SNMPService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SNMPConfig], error) {
	resp, err := s.c.Call(ctx, "netsnmp", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[SNMPConfig](resp.Data)
}

// Edit 保存 SNMP 服务配置（cfg 应为 Get 返回并修改后的完整行）。
// 对应 /Action/call: netsnmp.edit（cfg.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；action 名（edit）待真机验证。
func (s *SNMPService) Edit(ctx context.Context, cfg SNMPConfig) error {
	_, err := s.c.Call(ctx, "netsnmp", "edit", cfg)
	return err
}

// netsnmpc.show 的 TYPE 选择器常量（真机 3.7.24 探测验证）。
const (
	// SNMPClientTypeAll 采集配置区块：数据行 + 上报间隔 + IP 分组 +
	// MIB 文件元数据。
	SNMPClientTypeAll = "all"
)

// SNMPClient 是 SNMP 客户端的一行采集配置（netsnmpc.show data 数组
// 元素）。
//
// 空表推导，待真机验证：真机 3.7.24 未配置采集项（data 恒为空数组），
// 行字段无法实证（控制台 i18n snpm_client 提示行含 ip/port/oid/oid_type/
// datatype/name/value 等语义），故行对象以 Raw 原样保留，字段语义待
// 真机验证后再细化强类型。
type SNMPClient struct {
	// Raw 行的原始 JSON 对象。
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON 保留行对象的原始 JSON。
func (r *SNMPClient) UnmarshalJSON(b []byte) error {
	r.Raw = append(r.Raw[:0], b...)
	return nil
}

// MarshalJSON 把行对象原样序列化（Add/Edit 提交参数与原始行逐字段一致）。
func (r SNMPClient) MarshalJSON() ([]byte, error) {
	if len(r.Raw) == 0 {
		return []byte("{}"), nil
	}
	return r.Raw, nil
}

// Map 将行原始 JSON 解码为通用 map（字段语义待真机验证时的临时通道）。
func (r SNMPClient) Map() (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(r.Raw, &m); err != nil {
		return nil, fmt.Errorf("advanced: 解码 netsnmpc 行失败: %w", err)
	}
	return m, nil
}

// SNMPClientInfo 是 SNMP 客户端的采集配置区块（netsnmpc.show TYPE=all，
// 真机 3.7.24 实证顶层字段）。
type SNMPClientInfo struct {
	// Total 采集行数（真机观测 0）。
	Total int64 `json:"total"`
	// Rows 采集配置行列表（行结构待真机验证，见 SNMPClient）。
	Rows []SNMPClient `json:"data"`
	// SNMPInterval 上报间隔（真机观测 0；单位待真机验证，语义为向
	// 网管平台上报数据的周期）。
	SNMPInterval int64 `json:"snmp_interval"`
	// IPGroup 可选 IPv4 分组名列表（真机实证，控制台用于选择监控网段）。
	IPGroup []string `json:"ipgroup"`
	// MIBFile 已导入的自定义 MIB 文件列表（真机观测为空数组，行结构
	// 待真机验证）。
	MIBFile []json.RawMessage `json:"mibfile"`
	// MIBInfo MIB 文件详情列表（真机观测为空数组，行结构待真机验证）。
	MIBInfo []json.RawMessage `json:"mibinfo"`
}

// SNMPClientService 是 SNMP 客户端（netsnmpc）的子服务。
type SNMPClientService struct {
	c *ikuai.Client
}

// List 获取 SNMP 客户端采集配置行列表（分页）。
// 对应 /Action/call: netsnmpc.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total":0,"data":[]}），
// 即 Web 控制台「高级应用 → SNMP 客户端」页面。
func (s *SNMPClientService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[SNMPClient], error) {
	resp, err := s.c.Call(ctx, "netsnmpc", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[SNMPClient](resp.Data)
}

// Get 获取 SNMP 客户端采集配置区块（数据行、上报间隔、IP 分组、MIB
// 文件元数据）。
// 对应 /Action/call: netsnmpc.show（TYPE=all，真机探测实证）。
func (s *SNMPClientService) Get(ctx context.Context) (*SNMPClientInfo, error) {
	var info SNMPClientInfo
	if err := s.c.CallData(ctx, "netsnmpc", "show",
		map[string]any{"TYPE": SNMPClientTypeAll}, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Add 新增采集配置行，返回固件分配的行 ID。row 建议取控制台提交的
// 原始行字段（SDK 以 Raw 原样透传，不做字段加工）。
// 对应 /Action/call: netsnmpc.add。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *SNMPClientService) Add(ctx context.Context, row SNMPClient) (int64, error) {
	resp, err := s.c.Call(ctx, "netsnmpc", "add", row)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 修改采集配置行（row.Raw 应含待修改行的全部字段与 id）。
// 对应 /Action/call: netsnmpc.edit。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；行字段空表推导，待真机验证。
func (s *SNMPClientService) Edit(ctx context.Context, row SNMPClient) error {
	_, err := s.c.Call(ctx, "netsnmpc", "edit", row)
	return err
}

// Delete 删除采集配置行，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: netsnmpc.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *SNMPClientService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "netsnmpc", "del", ikuai.IDParam(ids))
	return err
}
