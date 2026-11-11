package accounting

import (
	"context"
	"time"

	"github.com/sumce/ikuai-sdk-go"
)

// PPPType 枚举常量：账号认证类型（pppuser.ppptype，真机实证 "any"；
// 取值集合按固件 i18n accountManagement.ppptype_data 推导）。
const (
	// PPPTypeAny 不限（真机实证取值）。
	PPPTypeAny = "any"
	// PPPTypeWeb WEB-帐号认证。
	PPPTypeWeb = "web"
	// PPPTypePPPoE PPPoE 认证。
	PPPTypePPPoE = "pppoe"
	// PPPTypePPTP PPTP 认证。
	PPPTypePPTP = "pptp"
	// PPPTypeL2TP L2TP 认证。
	PPPTypeL2TP = "l2tp"
	// PPPTypeOVPN OpenVPN 认证。
	PPPTypeOVPN = "ovpn"
	// PPPTypePPPoERelay PPPoE 透传。
	PPPTypePPPoERelay = "pppoe_relay"
	// PPPTypeIKE IKEv2 认证。
	PPPTypeIKE = "ike"
)

// IPType 枚举候选值：账号 IP 绑定类型（pppuser.ip_type，真机观测 0；
// 取值语义按固件 i18n accountManagement 推导，待真机验证）。
const (
	// IPTypeNone 不绑定固定 IP。
	IPTypeNone int64 = 0
	// IPTypeFixed 绑定固定 IP（仅限共享数 1）。
	IPTypeFixed int64 = 1
)

// PPPUser 是一条计费账号（pppuser.show 行，真机 3.7.24 实证 29 字段；
// 语义按固件 i18n accountManagement 页面标注）。
type PPPUser struct {
	// ID 行 ID（add 时无需填写；edit/del 必填）。omitempty 保证 add 参数
	// 不携带 id。
	ID int64 `json:"id,omitempty"`
	// Username 账号名（真机观测 "admin"）。
	Username string `json:"username"`
	// Passwd 账号密码（明文回显，真机实证）。
	Passwd string `json:"passwd"`
	// Enabled 账号状态（"yes"/"no"；真机观测 "yes"）。
	Enabled string `json:"enabled"`
	// PPPType 认证类型（见 PPPType* 常量；真机观测 "any"）。
	PPPType string `json:"ppptype"`
	// PPPName 拨号名称（PPPoE 服务名；真机观测为空）。
	PPPName string `json:"pppname"`
	// Name 用户姓名。
	Name string `json:"name"`
	// Cardid 证件号码。
	Cardid string `json:"cardid"`
	// Phone 联系电话。
	Phone string `json:"phone"`
	// Address 住址。
	Address string `json:"address"`
	// Comment 备注。
	Comment string `json:"comment"`
	// MAC 绑定 MAC（空=不绑定；多值形态待真机验证）。
	MAC string `json:"mac"`
	// AutoMAC 自动学习首次拨号 MAC（0/1）。
	AutoMAC int64 `json:"auto_mac"`
	// IPType IP 绑定类型（0=不绑定固定 IP，见 IPType* 常量）。
	IPType int64 `json:"ip_type"`
	// Packages 当前套餐 ID（0=未开通套餐/自定义）。
	Packages int64 `json:"packages"`
	// Share 共享数（同账号并发拨号数）。
	Share int64 `json:"share"`
	// Upload 上行速率（KB/s，999999=不限）。
	Upload int64 `json:"upload"`
	// Download 下行速率（KB/s，999999=不限）。
	Download int64 `json:"download"`
	// Duration 在线/离线时长（秒）。
	Duration int64 `json:"duration"`
	// Expires 账号到期时间（Unix 秒；真机观测 1892484059）。
	Expires int64 `json:"expires"`
	// StartTime 套餐开始时间（Unix 秒）。
	StartTime int64 `json:"start_time"`
	// CreateTime 创建时间（Unix 秒）。
	CreateTime int64 `json:"create_time"`
	// LastConntime 最近上线时间（Unix 秒）。
	LastConntime int64 `json:"last_conntime"`
	// LastOfftime 最近下线时间（Unix 秒）。
	LastOfftime int64 `json:"last_offtime"`
	// BindVlanID 绑定 VLAN（固件以字符串承载，"0"=不绑定；支持 QinQ 如 "2008.100"）。
	BindVlanID string `json:"bind_vlanid"`
	// AutoVlanID 自动绑定 VLAN（0/1；仅 PPPoE）。
	AutoVlanID int64 `json:"auto_vlanid"`
	// BindIfname 绑定网卡（真机观测 "any"）。
	BindIfname string `json:"bind_ifname"`
	// ProxyUsername 所属代理（代拨）账号（空=非代理账号）。
	ProxyUsername string `json:"proxy_username"`
	// PPPoEv6Wan PPPoEv6 使用的 WAN 线路（语义待真机验证）。
	PPPoEv6Wan string `json:"pppoev6_wan"`
}

// IsEnabled 报告账号是否启用。
func (u PPPUser) IsEnabled() bool { return u.Enabled == EnabledYes }

// ExpiresTime 到期时间。
func (u PPPUser) ExpiresTime() time.Time { return time.Unix(u.Expires, 0) }

// UnlimitedSpeed 报告上下行是否均不限速（固件不限速哨兵值 999999）。
func (u PPPUser) UnlimitedSpeed() bool { return u.Upload == 999999 && u.Download == 999999 }

// UsersService 是账号管理（pppuser）的子服务。
type UsersService struct {
	c *ikuai.Client
}

// List 获取计费账号列表（分页，支持关键字/过滤）。
// 对应 /Action/call: pppuser.show（TYPE=total,data, limit="offset,rows"，
// 真机实证），即 Web 控制台「认证计费 → 认证账号 → 账号管理」页面。
// opts 为 nil 时补 ikuai.DefaultListOptions()。
func (s *UsersService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[PPPUser], error) {
	resp, err := s.c.Call(ctx, "pppuser", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[PPPUser](resp.Data)
}

// Add 新增计费账号（开户），返回固件分配的行 ID。
// 对应 /Action/call: pppuser.add。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *UsersService) Add(ctx context.Context, user PPPUser) (int64, error) {
	resp, err := s.c.Call(ctx, "pppuser", "add", user)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改计费账号。
// 对应 /Action/call: pppuser.edit。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *UsersService) Edit(ctx context.Context, user PPPUser) error {
	_, err := s.c.Call(ctx, "pppuser", "edit", user)
	return err
}

// Delete 删除计费账号，支持单个或多个 ID。
// 对应 /Action/call: pppuser.del。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *UsersService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "pppuser", "del", ikuai.IDParam(ids))
	return err
}
