package logcenter

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// ikmessages（站内信系统消息，Web 控制台「日志中心 → 系统日志 → 消息通知」
// 页面，官方最新动态信息推送）的 show 形态说明：
//   - show TYPE=total,data：标准分页 {"total":N,"data":[...]}（真机实测）。
//   - show TYPE=total + FILTER1="read,==,0"：未读消息数（抓包实证，
//     analysis/calls/ikmessages__show.json）。
//   - start：param {} 重启/拉起消息推送通道（抓包实证，响应无 Data；
//     属写动作，真机测试禁调）。
//   - 消息行字段：探测真机消息表为空（data=[]，未读数 0），行字段无法
//     从真机数据核对；字段名 read 由 FILTER1 实证，其余按固件 i18n
//     /notification 列名（title/detail）推导，待真机验证。

// MessageRow 是一条站内信系统消息行（ikmessages.show 的 data 元素）。
//
// ⚠️ 探测真机（3.7.24）上消息表为空（未读数 0），行字段无法从真机数据
// 核对；字段按固件 i18n /notification 列名（title/detail）推导。空表推导，
// 待真机验证。
type MessageRow struct {
	// ID 行 ID。空表推导，待真机验证。
	ID int64 `json:"id"`
	// Timestamp 时间（unix 秒）。空表推导，待真机验证。
	Timestamp int64 `json:"timestamp"`
	// Title 消息标题（i18n /notification title「标题」）。空表推导，待真机验证。
	Title string `json:"title"`
	// Content 消息详情内容（i18n /notification detail「通知详情」；
	// 固件字段名可能为 detail，待真机验证）。
	Content string `json:"content"`
	// Read 已读标记（字段名由抓包 FILTER1="read,==,0" 实证；0=未读，
	// 已读取值待真机验证）。
	Read int64 `json:"read"`
}

// 消息已读标记值（未读=0 由抓包 FILTER1="read,==,0" 实证；已读取值待真机验证）。
const (
	MessageReadNo  = 0 // 未读（抓包实证）
	MessageReadYes = 1 // 已读（推测，待真机验证）
)

// MessagesService 是站内信系统消息（ikmessages）的子服务。
type MessagesService struct {
	c *ikuai.Client
}

// List 获取站内信系统消息列表（只读分页）。
// 对应 /Action/call: ikmessages.show（TYPE=total,data，真机实测），
// 即 Web 控制台「日志中心 → 系统日志 → 消息通知」页面。
// 按已读状态过滤可用 opts.Filters 添加 {"read,==,0"}（字段名实证）。
func (s *MessagesService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[MessageRow], error) {
	resp, err := s.c.Call(ctx, "ikmessages", "show", pageShowParam(opts, nil))
	if err != nil {
		return nil, err
	}
	return ikuai.DecodePage[MessageRow](resp.Data)
}

// UnreadCount 获取未读站内信消息数（只读）。
// 对应 /Action/call: ikmessages.show（TYPE=total + FILTER1="read,==,0"，
// 抓包实证 analysis/calls/ikmessages__show.json）。
func (s *MessagesService) UnreadCount(ctx context.Context) (int64, error) {
	opts := &ikuai.ListOptions{
		Type: MessagesTypeTotal,
		Filters: []ikuai.Filter{
			{Field: "read", Op: "==", Value: MessageReadNo},
		},
	}
	var p struct {
		Total int64 `json:"total"`
	}
	if err := s.c.CallData(ctx, "ikmessages", "show", opts.BuildParam(), &p); err != nil {
		return 0, err
	}
	return p.Total, nil
}

// Start 重启/拉起站内信消息推送（前端登录后调用，配合 WebSocket 通道
// /Action/rtcollect/websocket 接收实时推送）。
// 对应 /Action/call: ikmessages.start（param {}，抓包实证，响应无 Data）。
//
// ⚠️ 本方法按抓包（analysis/calls/ikmessages__start.json）建模；属写动作，
// 真机测试禁调（见 docs/RULES.md §9.2）。
func (s *MessagesService) Start(ctx context.Context) error {
	_, err := s.c.Call(ctx, "ikmessages", "start", map[string]any{})
	return err
}

// MarkAllRead 全部站内信消息标记为已读。
// 对应 /Action/call: ikmessages.set（param {"read":1}）。
//
// ⚠️ action 与 param 为推测：固件 i18n /notification 页存在「全部已读/
// 全部清空」操作按钮，但抓包未包含该调用，action/param 形态待真机验证；
// 属写动作，真机测试禁调。
func (s *MessagesService) MarkAllRead(ctx context.Context) error {
	_, err := s.c.Call(ctx, "ikmessages", "set", map[string]any{"read": MessageReadYes})
	return err
}
