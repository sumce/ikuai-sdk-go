package tools

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// speedtest 全局任务状态枚举值（global_status 字段；0 为真机实测，
// 其余取值语义待真机验证）。
const (
	// SpeedTestGlobalStatusStopped 测速任务已停止（真机实测 0，附
	// global_status_msg "测速任务已停止"）。
	SpeedTestGlobalStatusStopped = 0
)

// speedtest 单线路状态枚举值（SpeedTestLine.Status 字段；真机 3.7.24
// 实测 "0"=未测速、"4"=正在测速下载，其余取值待真机验证）。
const (
	// SpeedTestLineStatusIdle 未测速（真机实测 "0"，速率显示 "--"）。
	SpeedTestLineStatusIdle = "0"
	// SpeedTestLineStatusTestingDownload 正在测速下载（真机实测 "4"）。
	SpeedTestLineStatusTestingDownload = "4"
)

// SpeedTestLine 是一条外网线路的测速状态（speedtest.show 的 data 对象
// 按接口名为键的元素，真机 3.7.24 实测字段；速率类字段固件以字符串
// 返回，未测速时为 "--"）。
type SpeedTestLine struct {
	// Interface 线路接口名（与 data 的键一致）。
	Interface string `json:"interface"`
	// Status 线路测速状态（枚举 SpeedTestLineStatus*；其余取值待真机验证）。
	Status string `json:"status"`
	// StatusMsg 状态文案（实测 "未测速"/"正在测速下载" 等）。
	StatusMsg string `json:"status_msg"`
	// Download 下载速率（字符串形态，未测速为 "--"）。
	Download string `json:"download"`
	// Upload 上传速率（字符串形态，未测速为 "--"）。
	Upload string `json:"upload"`
	// DownAvg 下载平均速率（字符串形态）。
	DownAvg string `json:"down_avg"`
	// UpAvg 上传平均速率（字符串形态）。
	UpAvg string `json:"up_avg"`
	// DownMin 下载最小速率（i18n speed_test.down_min，字符串形态）。
	DownMin string `json:"down_min"`
	// UpMin 上传最小速率（i18n speed_test.up_min，字符串形态）。
	UpMin string `json:"up_min"`
	// DownMax 下载最大速率（i18n speed_test.down_max，字符串形态）。
	DownMax string `json:"down_max"`
	// UpMax 上传最大速率（i18n speed_test.up_max，字符串形态）。
	UpMax string `json:"up_max"`
}

// IsIdle 该线路是否未测速（status == "0"）。
func (l SpeedTestLine) IsIdle() bool { return l.Status == SpeedTestLineStatusIdle }

// SpeedTestState 是外网测速任务的当前状态（speedtest.show 的 Data，
// 真机 3.7.24 实测：data 为按接口名为键的对象而非数组）。
type SpeedTestState struct {
	// GlobalStatus 全局任务状态（枚举 SpeedTestGlobalStatus*；其余
	// 取值语义待真机验证）。
	GlobalStatus int64 `json:"global_status"`
	// GlobalStatusMsg 全局状态文案（实测 "测速任务已停止"）。
	GlobalStatusMsg string `json:"global_status_msg"`
	// Task 任务标识（真机观测空串，语义待真机验证）。
	Task string `json:"task"`
	// Lines 各外网线路的测速状态，键为接口名（如 "adsl01"）。
	Lines map[string]SpeedTestLine `json:"data"`
}

// SpeedTestService 是外网测速（speedtest）的子服务
// （tools.speed-test「线路测速」页面）。
type SpeedTestService struct {
	c *ikuai.Client
}

// Status 查询外网测速任务的当前状态与各线路结果（只读）。
// 对应 /Action/call: speedtest.show（空参数 {}，真机 3.7.24 实测返回
// global_status + 按接口名为键的 data 对象），
// 即 Web 控制台「应用工具 → 线路测速」页面。
func (s *SpeedTestService) Status(ctx context.Context) (*SpeedTestState, error) {
	var out SpeedTestState
	if err := s.c.CallData(ctx, "speedtest", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Start 对指定外网线路发起一次测速（诊断动作）。
// 对应 /Action/call: speedtest.start（param {"interface":<线路>}——
// action 名与参数形态按 i18n speed_test.start_speed「开始测速」及
// show 响应的接口名键推断，未对真机调用，待真机验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用：测速占用设备性能资源且消耗线路
// 流量（i18n speed_test.tips2），仅离线测试覆盖。
func (s *SpeedTestService) Start(ctx context.Context, iface string) error {
	_, err := s.c.Call(ctx, "speedtest", "start", map[string]any{"interface": iface})
	return err
}

// Stop 停止进行中的外网测速（诊断动作）。
// 对应 /Action/call: speedtest.stop（action 名按 i18n
// speed_test.stop_speed「停止测速」按钮推断，未对真机调用，待真机
// 验证）。
//
// ⚠️ 诊断动作，禁止对真机随意调用，仅离线测试覆盖。
func (s *SpeedTestService) Stop(ctx context.Context) error {
	_, err := s.c.Call(ctx, "speedtest", "stop", map[string]any{})
	return err
}
