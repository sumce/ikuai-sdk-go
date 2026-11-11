package behavior

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// Service 是行为管控大类的服务入口，聚合 Web 控制台「行为管控」菜单下的
// 全部 17 个功能（菜单树 func 与页面的对应关系见 docs/API_CATALOG.md §8）。
type Service struct {
	c *ikuai.Client

	// Audit 审计设置（audit 行为记录配置 + record_free_setting 审计豁免）。
	Audit *AuditService
	// AuditLogs 审计日志（audit_url_log/audit_im_log/audit_terminal_log/
	// audit_terminal_stat，全部只读分页）。
	AuditLogs *AuditLogsService
	// MACComment MAC 备注（mac_comment）：终端名称/备注管理。
	MACComment *MACCommentService
	// ACL 访问控制族（acl_mac/acl_l7/acl_l2route/acl_qq）。
	ACL *ACLService
	// WebFilter 网页过滤（url_black/url_redirect/url_keywords/url_replace）。
	WebFilter *WebFilterService
	// DomainFilter 域名过滤（domain_blacklist/domain_group）。
	DomainFilter *DomainFilterService
}

// New 构造行为管控服务。
func New(c *ikuai.Client) *Service {
	return &Service{
		c:            c,
		Audit:        &AuditService{c: c},
		AuditLogs:    &AuditLogsService{c: c},
		MACComment:   &MACCommentService{c: c},
		ACL:          &ACLService{c: c},
		WebFilter:    &WebFilterService{c: c},
		DomainFilter: &DomainFilterService{c: c},
	}
}

// 固件通用的启用/禁用字符串值（enabled 字段，与固件完全一致）。
const (
	EnabledYes = "yes"
	EnabledNo  = "no"
)

// SelectorsTypeAll 是本包各列表 func（acl_l7/acl_qq/acl_l2route/url_black/
// url_redirect/url_keywords/url_replace/domain_blacklist/domain_group 等）
// show 动作的表单选择器 TYPE 值：响应在 total/data 之外附带表单可选值区块
// （IP 分组、MAC 分组、时间分组、匹配模式等，真机 3.7.24 验证）。
const SelectorsTypeAll = "all"

// Raw 以任意 func_name/action/param 调用行为管控大类接口，作为未封装
// action/TYPE 组合的逃生口（每个大类 Service 仅此一个，见 docs/RULES.md §6）。
// 对应 /Action/call: <funcName>.<action>。
// 注意：本包全部写动作（add/edit/save/set/del 等）未对真机验证，只应通过
// Raw 在充分知晓后果的前提下调用未封装动作，且严禁对真机执行写测试
// （见 docs/RULES.md §9.2 与本包各写方法注释）。
func (s *Service) Raw(ctx context.Context, funcName, action string, param any) (*ikuai.Response, error) {
	return s.c.Call(ctx, funcName, action, param)
}

// selectorShowParam 构造带固定默认键的 show 参数：opts == nil 时仅由
// defaults 生成参数；opts 非空时以 opts.BuildParam() 为底，defaults 只补
// 缺失键，用户在 opts.Type / opts.Extra 中显式给出的值优先。
func selectorShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	var p map[string]any
	if opts != nil {
		p = opts.BuildParam()
	}
	if p == nil {
		p = map[string]any{}
	}
	for k, v := range defaults {
		if _, ok := p[k]; !ok {
			p[k] = v
		}
	}
	return p
}

// pageShowParam 构造标准分页类 func（本包接受 TYPE=total,data 的列表型
// func）的 show 参数：opts == nil 时补 ikuai.DefaultListOptions()
// （TYPE=total,data, limit="0,20"），defaults 只补缺失的本功能专有键。
func pageShowParam(opts *ikuai.ListOptions, defaults map[string]any) map[string]any {
	if opts == nil {
		opts = ikuai.DefaultListOptions()
	}
	return selectorShowParam(opts, defaults)
}

// typedShowParam 构造以固定 TYPE 区块为主的 func（audit_url_log 之外的单
// 行配置 func 等不接受 TYPE=total 的场合）的 show 参数：opts 为 nil 或未
// 显式指定 Type 时强制写入 defaultType（固件对这类 func 的 TYPE=total 回
// 30006 unknown TYPE (total)，真机 3.7.24 验证）；opts.Type 显式给出的值
// 被尊重。其余键仍尊重 opts（limit/ORDER_BY/ORDER 等）。
func typedShowParam(opts *ikuai.ListOptions, defaultType string, defaults map[string]any) map[string]any {
	p := selectorShowParam(opts, defaults)
	if opts == nil || opts.Type == "" {
		p["TYPE"] = defaultType
	}
	return p
}

// noTypeShowParam 构造拒绝任何 TYPE 的日志类 func（audit_url_log 真机实测：
// TYPE=total,data 回 30006 unknown TYPE (total)，仅接受无 TYPE 的
// {"limit":...} 形态）的 show 参数：opts == nil 时按默认分页（limit="0,20"）
// 但剥掉 TYPE；opts.Type 显式给出的值被尊重（用于用户自担风险的特殊 TYPE）。
func noTypeShowParam(opts *ikuai.ListOptions) map[string]any {
	p := pageShowParam(opts, nil)
	if opts == nil || opts.Type == "" {
		delete(p, "TYPE")
	}
	return p
}
