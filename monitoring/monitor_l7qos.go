package monitoring

import (
	"context"
	"encoding/json"

	"github.com/sumce/ikuai-sdk-go"
)

// monitor_l7qos.show 的 TYPE 选择器常量（抓包实证，可单独或组合使用）。
const (
	// L7QoSTypeData 流控策略状态区块（data）。
	L7QoSTypeData = "data"
	// L7QoSTypeInterface 可选接口嵌套列表区块（interface）。
	L7QoSTypeInterface = "interface"
	// L7QoSTypeDefault 控制台「策略监控」页面默认 TYPE 组合。
	L7QoSTypeDefault = L7QoSTypeData + "," + L7QoSTypeInterface
)

// L7QoSStatus 是流控策略监控数据（monitor_l7qos.show）。
type L7QoSStatus struct {
	// Data 流控策略状态行。抓包与真机（3.7.24 未配置流控策略）均返回
	// 空数组，行字段无法实证，故以 RawMessage 原样保留，字段语义待真机
	// 验证后再细化强类型。
	Data []json.RawMessage `json:"data"`
	// Interface 可选接口嵌套列表（[接口名] 或 [接口名, 备注]）。
	Interface [][]string `json:"interface"`
}

// L7QoSService 是策略监控（monitor_l7qos）的子服务。
type L7QoSService struct {
	c *ikuai.Client
}

// Status 获取流控策略状态与可选接口列表。
// 对应 /Action/call: monitor_l7qos.show（TYPE=data,interface, interface=""），
// 即 Web 控制台「状态监控 → 策略监控」页面。
// opts 为 nil 时使用抓包默认值；仅取部分区块可用 opts.Type =
// L7QoSTypeData / L7QoSTypeInterface，接口过滤可用 opts.Extra["interface"]。
func (s *L7QoSService) Status(ctx context.Context, opts *ikuai.ListOptions) (*L7QoSStatus, error) {
	param := selectorShowParam(opts, map[string]any{
		"TYPE":      L7QoSTypeDefault,
		"interface": "", // 控制台抓包原样（空串=未过滤接口）。
	})
	var out L7QoSStatus
	if err := s.c.CallData(ctx, "monitor_l7qos", "show", param, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
