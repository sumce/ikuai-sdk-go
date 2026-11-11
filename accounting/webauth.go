package accounting

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// webauth.show 的 TYPE 选择器常量（真机 3.7.24 实证）。
const (
	// WebAuthTypeData 单行配置（真机实证，等价于空参数）。
	WebAuthTypeData = "data"
	// WebAuthTypeAll 全量（真机实证，返回形态与 data 相同）。
	WebAuthTypeAll = "all"
)

// WebAuthConfig 是 Web 认证服务的单行全局配置（webauth.show TYPE=data 行，
// 真机 3.7.24 实证 141 字段；语义按固件 i18n auth_expenses 与表单标注）。
//
// 认证方式开关字段（*_auth，0/1）与各方式的上下行限速（*_up/*_down，
// KB/s，0=不限）、倒计时（*_timeout，秒）、最长在线/闲置（*_max_time/
// *_idle_time，秒）按方式一一配对；0 表示未启用。
type WebAuthConfig struct {
	// ID 行 ID（恒为 1，save 时随行携带）。
	ID int64 `json:"id"`
	// Enabled Web 认证总开关（"yes"/"no"）。
	Enabled string `json:"enabled"`
	// MaxTime 默认最长在线时间（秒，0=不限）。
	MaxTime int64 `json:"max_time"`
	// IdleTime 默认闲置下线时间（秒，0=不限；真机观测 60）。
	IdleTime int64 `json:"idle_time"`
	// EncSSIDNoauth 未认证时广播 SSID 加密（0/1；语义待真机验证）。
	EncSSIDNoauth int64 `json:"enc_ssid_noauth"`
	// TimerRestart 定时重启 Web 认证服务（0/1）。
	TimerRestart int64 `json:"timer_restart"`
	// TimerEvent 定时重启事件描述（待真机验证）。
	TimerEvent string `json:"timer_event"`
	// TimerRestartWeek 定时重启的星期（"1234567"，周一~周日；真机观测 "1234567"）。
	TimerRestartWeek string `json:"timer_restart_week"`
	// TimerRestartTime 定时重启时刻（"07:00"）。
	TimerRestartTime string `json:"timer_restart_time"`

	// ---- 认证方式开关（0/1）----
	// UserAuth 用户名密码认证（WEB-用户认证）。
	UserAuth int64 `json:"user_auth"`
	// CouponAuth 上网码认证（WEB-上网码认证）。
	CouponAuth int64 `json:"coupon_auth"`
	// QQAuth QQ 认证。
	QQAuth int64 `json:"qq_auth"`
	// WeiboAuth 微博认证。
	WeiboAuth int64 `json:"weibo_auth"`
	// PhoneAuth 手机认证。
	PhoneAuth int64 `json:"phone_auth"`
	// StaticPwd 固定密码认证（WEB-固定密码）。
	StaticPwd int64 `json:"static_pwd"`
	// Nopasswd 一键认证（免密认证）。
	Nopasswd int64 `json:"nopasswd"`
	// Weixin 微信认证（真机观测 1）。
	Weixin int64 `json:"weixin"`
	// Weixin2 微信认证备用通道（语义待真机验证）。
	Weixin2 int64 `json:"weixin2"`
	// WeixinMini 微信小程序认证。
	WeixinMini int64 `json:"weixin_mini"`
	// FacebookAuth Facebook 认证。
	FacebookAuth int64 `json:"facebook_auth"`
	// TwitterAuth Twitter 认证。
	TwitterAuth int64 `json:"twitter_auth"`
	// GoogleAuth Google 认证。
	GoogleAuth int64 `json:"google_auth"`
	// RedpacketAuth 红包上网认证（WEB-红包上网）。
	RedpacketAuth int64 `json:"redpacket_auth"`
	// HotelAuth 酒店认证（语义待真机验证）。
	HotelAuth int64 `json:"hotel_auth"`
	// OpenapiAuth 开放 API 对接认证（语义待真机验证）。
	OpenapiAuth int64 `json:"openapi_auth"`
	// CustomAuth 自定义认证。
	CustomAuth int64 `json:"custom_auth"`
	// CustomAppkey 自定义认证应用 Key。
	CustomAppkey string `json:"custom_appkey"`
	// LDAPAuth LDAP 认证。
	LDAPAuth int64 `json:"ldap_auth"`
	// ApiRadius API-Radius 对接（0/1；语义待真机验证）。
	ApiRadius int64 `json:"api_radius"`
	// AutoAuth 免输密码自动认证（语义待真机验证）。
	AutoAuth int64 `json:"auto_auth"`

	// ---- 各认证方式的上下行限速（KB/s，0=不限）----
	PhoneUp      int64 `json:"phone_up"`
	PhoneDown    int64 `json:"phone_down"`
	NopasswdUp   int64 `json:"nopasswd_up"`
	NopasswdDown int64 `json:"nopasswd_down"`
	StaticPwdUp  int64 `json:"static_pwd_up"`
	StaticPwdDwn int64 `json:"static_pwd_down"`
	WeiboUp      int64 `json:"weibo_up"`
	WeiboDown    int64 `json:"weibo_down"`
	TryoutUp     int64 `json:"tryout_up"`
	TryoutDown   int64 `json:"tryout_down"`
	QQUp         int64 `json:"qq_up"`
	QQDown       int64 `json:"qq_down"`
	FacebookUp   int64 `json:"facebook_up"`
	FacebookDown int64 `json:"facebook_down"`
	GoogleUp     int64 `json:"google_up"`
	GoogleDown   int64 `json:"google_down"`
	TwitterUp    int64 `json:"twitter_up"`
	TwitterDown  int64 `json:"twitter_down"`
	RedpacketUp  int64 `json:"redpacket_up"`
	RedpacketDwn int64 `json:"redpacket_down"`
	HotelUp      int64 `json:"hotel_up"`
	HotelDown    int64 `json:"hotel_down"`
	CouponUp     int64 `json:"coupon_up"`
	CouponDown   int64 `json:"coupon_down"`
	WeixinUp     int64 `json:"weixin_up"`
	// WeixinDown 微信认证下行限速（KB/s）。
	WeixinDown int64 `json:"weixin_down"`

	// ---- 各认证方式的认证后倒计时（秒，0=不启用）----
	UserTimeout      int64 `json:"user_timeout"`
	CouponTimeout    int64 `json:"coupon_timeout"`
	QQTimeout        int64 `json:"qq_timeout"`
	WeiboTimeout     int64 `json:"weibo_timeout"`
	PhoneTimeout     int64 `json:"phone_timeout"`
	StaticTimeout    int64 `json:"static_timeout"`
	NopasswdTimeout  int64 `json:"nopasswd_timeout"`
	WeixinTimeout    int64 `json:"weixin_timeout"`
	RedpacketTimeout int64 `json:"redpacket_timeout"`
	FacebookTimeout  int64 `json:"facebook_timeout"`
	GoogleTimeout    int64 `json:"google_timeout"`
	TwitterTimeout   int64 `json:"twitter_timeout"`
	TryoutTimeout    int64 `json:"tryout_timeout"`
	HotelTimeout     int64 `json:"hotel_timeout"`

	// ---- 各认证方式的最长在线/闲置时间（秒，0=不限）----
	UserMaxTime      int64 `json:"user_max_time"`
	UserIdleTime     int64 `json:"user_idle_time"`
	CouponMaxTime    int64 `json:"coupon_max_time"`
	CouponIdleTime   int64 `json:"coupon_idle_time"`
	QQMaxTime        int64 `json:"qq_max_time"`
	QQIdleTime       int64 `json:"qq_idle_time"`
	WeiboMaxTime     int64 `json:"weibo_max_time"`
	WeiboIdleTime    int64 `json:"weibo_idle_time"`
	PhoneMaxTime     int64 `json:"phone_max_time"`
	PhoneIdleTime    int64 `json:"phone_idle_time"`
	StaticMaxTime    int64 `json:"static_max_time"`
	StaticIdleTime   int64 `json:"static_idle_time"`
	NopasswdMaxTime  int64 `json:"nopasswd_max_time"`
	NopasswdIdleTime int64 `json:"nopasswd_idle_time"`
	WeixinMaxTime    int64 `json:"weixin_max_time"`
	WeixinIdleTime   int64 `json:"weixin_idle_time"`
	FacebookMaxTime  int64 `json:"facebook_max_time"`
	FacebookIdleTime int64 `json:"facebook_idle_time"`
	GoogleMaxTime    int64 `json:"google_max_time"`
	GoogleIdleTime   int64 `json:"google_idle_time"`
	TwitterMaxTime   int64 `json:"twitter_max_time"`
	TwitterIdleTime  int64 `json:"twitter_idle_time"`
	RedpacketMaxTime int64 `json:"redpacket_max_time"`
	RedpacketIdle    int64 `json:"redpacket_idle_time"`
	TryoutMaxTime    int64 `json:"tryout_max_time"`
	TryoutIdleTime   int64 `json:"tryout_idle_time"`
	HotelMaxTime     int64 `json:"hotel_max_time"`
	HotelIdleTime    int64 `json:"hotel_idle_time"`

	// ---- 微信吸粉/HTTPS/弹窗等展示行为 ----
	// WeixinAbsorb 微信吸粉（关注公众号后上网；0/1）。
	WeixinAbsorb int64 `json:"weixin_absorb"`
	// HTTPSRedirect HTTPS 认证重定向（0/1；语义待真机验证）。
	HTTPSRedirect int64 `json:"https_redirect"`
	// Popups 认证成功弹窗（0/1；真机观测 1）。
	Popups int64 `json:"popups"`
	// PopupsIOS iOS 弹窗（0/1）。
	PopupsIOS int64 `json:"popups_ios"`
	// PopupsAndroid Android 弹窗（0/1）。
	PopupsAndroid int64 `json:"popups_android"`
	// AutoAuthTimeout 自动认证超时（真机观测 1；单位/语义待真机验证）。
	AutoAuthTimeout int64 `json:"auto_auth_timeout"`
	// Imperceptible 无感认证（0/1；语义待真机验证）。
	Imperceptible int64 `json:"imperceptible"`
	// URIAddAPInfo 认证 URI 附加 AP 信息（0/1；语义待真机验证）。
	URIAddAPInfo int64 `json:"uri_add_apinfo"`
	// SyncSwitch 与云平台同步（0/1；语义待真机验证）。
	SyncSwitch int64 `json:"sync_switch"`

	// ---- 接入范围与白名单 ----
	// Interface 认证生效的内网线路（如 "lan1"）。
	Interface string `json:"interface"`
	// IPGroup 生效的 IPv4 分组（逗号分隔的分组 ID；空=全部）。
	IPGroup string `json:"ipgroup"`
	// NoauthMac 免认证 MAC 列表（多行/逗号分隔）。
	NoauthMac string `json:"noauth_mac"`
	// Whitelist HTTP 白名单域名（多行）。
	Whitelist string `json:"whitelist"`
	// WhitelistHTTPS HTTPS 白名单域名（多行）。
	WhitelistHTTPS string `json:"whitelist_https"`
	// WhiteIP 白名单 IP（多行）。
	WhiteIP string `json:"whiteip"`
	// AuthIPMode 认证 IP 模式（0/1；语义待真机验证）。
	AuthIPMode int64 `json:"authip_mode"`
	// MacbindServer MAC 绑定服务地址（语义待真机验证）。
	MacbindServer string `json:"macbind_server"`

	// ---- 试用认证 ----
	// AllowTryout 允许试用认证（0/1）。
	AllowTryout int64 `json:"allow_tryout"`
	// TryoutTime 试用时长（真机观测 30；单位待真机验证，分钟/秒）。
	TryoutTime int64 `json:"tryout_time"`

	// ---- 对接/第三方 ----
	// GroupID 对接分组 ID（语义待真机验证）。
	GroupID int64 `json:"group_id"`
	// GroupKey 对接分组密钥（真机观测 "testing123"）。
	GroupKey string `json:"group_key"`
	// APISwitch API 对接开关（0/1）。
	APISwitch int64 `json:"api_switch"`
	// APIURL API 对接地址。
	APIURL string `json:"api_url"`
	// APIIPChangeURL API IP 变更通知地址（语义待真机验证）。
	APIIPChangeURL string `json:"api_ipchange_url"`
	// Passwd 认证页密码（如固定密码认证的口令；真机观测为空）。
	Passwd string `json:"passwd"`
	// ProxyService 认证代理服务（0/1；语义待真机验证）。
	ProxyService int64 `json:"proxy_service"`
	// ProxyIPaddr 认证代理地址（ProxyService=1 时使用）。
	ProxyIPaddr string `json:"proxy_ipaddr"`

	// ---- LDAP 认证 ----
	LDAPUserGroup  string `json:"ldap_usergroup"`    // LDAP 用户组
	LDAPURL        string `json:"ldap_url"`          // LDAP 服务器地址
	LDAPPort       int64  `json:"ldap_port"`         // LDAP 端口（真机观测 389）
	LDAPMode       int64  `json:"ldap_mode"`         // LDAP 模式（枚举语义待真机验证）
	LDAPBase       string `json:"ldap_base"`         // LDAP Base DN
	LDAPCName      string `json:"ldap_cname"`        // LDAP 用户名字段（真机观测 "cn"）
	LDAPAdminDN    string `json:"ldap_admin_dn"`     // LDAP 管理员 DN
	LDAPAdminPass  string `json:"ldap_admin_passwd"` // LDAP 管理员密码
	LDAPShareCount int64  `json:"ldap_share_count"`  // LDAP 共享数（语义待真机验证）

	// ---- Radius 认证 ----
	// NasName NAS 名称（真机观测 "iKuai"）。
	NasName string `json:"nasname"`
	// RadiusIP Radius 服务端 IP（真机观测 "127.0.0.1"）。
	RadiusIP string `json:"radius_ip"`
	// RadiusKey Radius 共享密钥（真机观测 "123"）。
	RadiusKey string `json:"radius_key"`
	// RadiusAuthport Radius 认证端口（真机观测 1812）。
	RadiusAuthport int64 `json:"radius_authport"`
	// RadiusAccountport Radius 计费端口（真机观测 1813）。
	RadiusAccountport int64 `json:"radius_accountport"`
	// RadiusInterval Radius 心跳间隔（秒，真机观测 60）。
	RadiusInterval int64 `json:"radius_interval"`
	// RadiusBasIP Radius BAS IP（语义待真机验证）。
	RadiusBasIP string `json:"radius_basip"`
}

// IsEnabled 报告 Web 认证是否开启。
func (c WebAuthConfig) IsEnabled() bool { return c.Enabled == EnabledYes }

// WebAuthService 是 Web 认证服务（webauth）的子服务。
type WebAuthService struct {
	c *ikuai.Client
}

// Get 获取 Web 认证服务全局配置（认证方式开关/上下行限速/Radius/LDAP/
// 白名单/弹窗等）。配置未落库（data 为空）时返回 (nil, nil)。
// 对应 /Action/call: webauth.show（TYPE=data，真机 3.7.24 实证；
// TYPE=total,data 回 30006 unknown TYPE），即 Web 控制台
// 「认证计费 → 认证服务 → Web 认证」页面。
func (s *WebAuthService) Get(ctx context.Context) (*WebAuthConfig, error) {
	var out struct {
		Data []WebAuthConfig `json:"data"`
	}
	if err := s.c.CallData(ctx, "webauth", "show",
		map[string]any{"TYPE": WebAuthTypeData}, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, nil
	}
	row := out.Data[0]
	return &row, nil
}

// Set 保存 Web 认证服务全局配置（整行保存，随行携带 ID）。
// 对应 /Action/call: webauth.save（save 动作参照同构配置页 alg.save 抓包；
// 行字段真机实证）。
// ⚠️ 写动作，字段待真机验证：仅离线测试覆盖，禁止对真机调用。
func (s *WebAuthService) Set(ctx context.Context, cfg WebAuthConfig) error {
	_, err := s.c.Call(ctx, "webauth", "save", cfg)
	return err
}
