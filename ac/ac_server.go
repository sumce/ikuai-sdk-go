package ac

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// APDevice 是接入 AC 的 AP 设备一行（ac_server.show data 数组元素）。
//
// 空表推导，待真机验证：真机 3.7.24 无 AP 设备（ac_status ap_count=0），
// 列表恒为空数组，行字段无法实证。字段名按固件 i18n（ap-config 区块
// macorip/ap_model/name/channel/link_speed/maxload/upload 及 equipment
// 区块 mac/ip/ap_status/online_time）+ 控制台「AP列表」页面语义推导。
type APDevice struct {
	// ID 行 ID（edit/del 必填；add 时无需填写——AP 由固件自动注册）。
	ID int64 `json:"id,omitempty"`
	// MAC AP 的 MAC 地址。
	MAC string `json:"mac"`
	// IP AP 的 IP 地址。
	IP string `json:"ip"`
	// Name AP 备注（i18n ap-config.name「备注」）。
	Name string `json:"name"`
	// Model AP 型号（i18n ap_model「型号」）。
	Model string `json:"model"`
	// Version AP 固件版本（i18n ac.now_version「当前版本」语义）。
	Version string `json:"version"`
	// Status AP 状态（normal/exception/offline/upgrading 等，取值集待真机验证）。
	Status string `json:"status"`
	// Uptime 在线时长（i18n equipment.online_time；单位待真机验证）。
	Uptime int64 `json:"uptime"`
	// Fre 射频规格（fre_24g=2.4G 单频 / fre_5g=2.4G+5G 双频，取值集待真机验证）。
	Fre string `json:"fre"`
	// Channel 信道。
	Channel string `json:"channel"`
	// LinkSpeed 上行链路协商速率（i18n link_speed）。
	LinkSpeed string `json:"link_speed"`
	// MaxLoad 最大带机量（0 表示不限制，i18n max_amount_tip）。
	MaxLoad int64 `json:"maxload"`
	// Load 当前带机量（i18n maxload「当前/最大带机量」的当前侧）。
	Load int64 `json:"load"`
	// Upload 累计上行（字节，i18n upload「累计上下行」）。
	Upload int64 `json:"upload"`
	// Download 累计下行（字节）。
	Download int64 `json:"download"`
}

// ServerService 是 AP 列表（ac_server）的子服务。
type ServerService struct {
	c *ikuai.Client
}

// List 获取接入 AC 的 AP 设备列表（分页）。
// 对应 /Action/call: ac_server.show（TYPE=total,data, limit="offset,rows"，
// 真机探测实证返回 {"total","data"}），
// 即 Web 控制台「AC管理 → AP列表」页面。
func (s *ServerService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[APDevice], error) {
	resp, err := s.c.Call(ctx, "ac_server", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[APDevice](resp.Data)
}

// Edit 修改 AP 设备信息（如备注 name，写动作）。
// 对应 /Action/call: ac_server.edit（ap.ID 随 JSON 自动携带）。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用；字段集待真机验证。
func (s *ServerService) Edit(ctx context.Context, ap APDevice) error {
	_, err := s.c.Call(ctx, "ac_server", "edit", ap)
	return err
}

// Delete 卸载 AP 设备（控制台「卸载」，写动作），支持单个或多个 ID
// （内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: ac_server.del。
// ⚠️ 写动作：仅离线测试覆盖，禁止对真机调用。
func (s *ServerService) Delete(ctx context.Context, ids ...int64) error {
	_, err := s.c.Call(ctx, "ac_server", "del", ikuai.IDParam(ids))
	return err
}
