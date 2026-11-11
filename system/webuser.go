package system

import (
	"context"
	"fmt"

	"github.com/sumce/ikuai-sdk-go"
)

// webuser.show 的 TYPE 选择器常量（burp 抓包 + 真机 3.7.24 验证）。
const (
	// WebUserTypeModPasswd 改密标记查询区块（配合 username 参数）。
	WebUserTypeModPasswd = "mod_passwd"
)

// WebUserAccount 是 Web 管理账号的行（webuser.show data 数组元素）。
//
// 本包仅实证了 webuser.show TYPE=mod_passwd 的响应形态；账号列表的行
// 字段与 add/edit 的参数字段未在抓包中出现，以下字段按固件语义与控制台
// 表单建模：写动作，字段待真机验证。
type WebUserAccount struct {
	// ID 行 ID（add 时忽略，edit 时必填）。
	ID int64 `json:"id,omitempty"`
	// Username 登录用户名。
	Username string `json:"username"`
	// Passwd 登录密码（明文，固件接口负责处理；查询响应中不出现）。
	Passwd string `json:"passwd,omitempty"`
	// Comment 账号备注。
	Comment string `json:"comment,omitempty"`
}

// WebUserPasswdMod 是修改账号密码的参数（webuser.mod_passwd）。
// 动作名由 show 的 TYPE=mod_passwd 推断，字段待真机验证。
type WebUserPasswdMod struct {
	// Username 要改密的账号名。
	Username string `json:"username"`
	// Passwd 新密码（明文）。
	Passwd string `json:"passwd"`
	// OldPasswd 旧密码（是否必填待真机验证）。
	OldPasswd string `json:"old_passwd,omitempty"`
}

// WebUserService 是 Web 账号管理（webuser）的子服务。
type WebUserService struct {
	c *ikuai.Client
}

// PasswdFlag 查询账号的强制改密标记。
// 对应 /Action/call: webuser.show（TYPE=mod_passwd, username=...），
// burp 抓包实证。返回固件的 mod_passwd 原始数值（观测值 0；
// 推测 0=无需强制改密，非 0=下次登录需改密，待真机验证）。
// username 为空时返回 ikuai.ErrInvalidParam。
func (s *WebUserService) PasswdFlag(ctx context.Context, username string) (int64, error) {
	if username == "" {
		return 0, fmt.Errorf("system: username 不能为空: %w", ikuai.ErrInvalidParam)
	}
	var out struct {
		ModPasswd int64 `json:"mod_passwd"`
	}
	if err := s.c.CallData(ctx, "webuser", "show",
		map[string]any{"TYPE": WebUserTypeModPasswd, "username": username}, &out); err != nil {
		return 0, err
	}
	return out.ModPasswd, nil
}

// Add 新增 Web 管理账号，返回固件分配的行 ID（API_CATALOG 实证 add 返回
// RowId）。
// 对应 /Action/call: webuser.add。
// ⚠️ 写动作：字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *WebUserService) Add(ctx context.Context, account WebUserAccount) (int64, error) {
	resp, err := s.c.Call(ctx, "webuser", "add", account)
	if err != nil {
		return 0, err
	}
	return resp.RowId, nil
}

// Edit 按 ID 修改 Web 管理账号（account.ID 必须非 0）。
// 对应 /Action/call: webuser.edit。
// ⚠️ 写动作：字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *WebUserService) Edit(ctx context.Context, account WebUserAccount) error {
	if account.ID == 0 {
		return fmt.Errorf("system: account.ID 不能为 0: %w", ikuai.ErrInvalidParam)
	}
	_, err := s.c.Call(ctx, "webuser", "edit", account)
	return err
}

// Delete 删除 Web 管理账号，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: webuser.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *WebUserService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "webuser", "del", ikuai.IDParam(ids))
	return err
}

// ModPasswd 修改账号密码（写动作）。
// 对应 /Action/call: webuser.mod_passwd（动作名由 show 的
// TYPE=mod_passwd 推断）。
// ⚠️ 写动作：字段待真机验证，仅离线测试覆盖，禁止对真机调用。
func (s *WebUserService) ModPasswd(ctx context.Context, mod WebUserPasswdMod) error {
	_, err := s.c.Call(ctx, "webuser", "mod_passwd", mod)
	return err
}
