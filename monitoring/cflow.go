package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// CFlowValue 兼容固件的数字怪癖：cflow.show 对「今日/昨日」统计返回
// 字符串数字（如 "3622459"），对「本周」统计返回 JSON 整数。原始值
// 统一保留为字符串（遵守 docs/RULES.md §5 字符串数字保留 string），
// Int64 提供数值解析。注意重新编码时按字符串输出。
type CFlowValue string

// UnmarshalJSON 同时接受 JSON 数字与字符串两种形态。
func (v *CFlowValue) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*v = CFlowValue(s)
		return nil
	}
	*v = CFlowValue(b)
	return nil
}

// String 返回原始字符串形式。
func (v CFlowValue) String() string { return string(v) }

// Int64 解析数值；空值返回 0。非整数内容时报错。
func (v CFlowValue) Int64() (int64, error) {
	s := strings.TrimSpace(string(v))
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("monitoring: CFlowValue %q 不是整数: %w", string(v), err)
	}
	return n, nil
}

// CFlowCounters 是一个统计区间（今日/本周/昨日）的分流连接计数。
// 各计数字段语义按命名与控制台页面推断，待真机验证。
type CFlowCounters struct {
	// LlbCnt 多线负载（链路负载均衡）命中连接计数。
	LlbCnt CFlowValue `json:"llb_cnt"`
	// PortCnt 端口分流命中连接计数。
	PortCnt CFlowValue `json:"port_cnt"`
	// ProtoCnt 协议分流命中连接计数。
	ProtoCnt CFlowValue `json:"proto_cnt"`
	// DomainCnt 域名分流命中连接计数。
	DomainCnt CFlowValue `json:"domain_cnt"`
	// ConnCnt 连接总数。
	ConnCnt CFlowValue `json:"conn_cnt"`
}

// CFlowPeriods 是各统计区间的计数集合。
type CFlowPeriods struct {
	// Today 今日统计（固件返回字符串数字）。
	Today CFlowCounters `json:"today"`
	// Week 本周统计（固件返回 JSON 整数）。
	Week CFlowCounters `json:"week"`
	// Yesterday 昨日统计（固件返回字符串数字）。
	Yesterday CFlowCounters `json:"yesterday"`
}

// CFlowSummary 是分流监控连接统计汇总（cflow.show）。
type CFlowSummary struct {
	// Data 今日/本周/昨日统计集合。
	Data CFlowPeriods `json:"data"`
	// CleanTime 统计清零时间戳（0=未清零；字段语义待真机验证）。
	CleanTime int64 `json:"clean_time"`
}

// CFlowService 是分流监控（cflow）的子服务。
type CFlowService struct {
	c *ikuai.Client
}

// Summary 获取分流监控的连接统计汇总（今日/本周/昨日的多线负载/
// 端口分流/协议分流/域名分流/总连接计数）。
// 对应 /Action/call: cflow.show（param 为空对象），
// 即 Web 控制台「状态监控 → 流量监控」页面的连接统计汇总。
func (s *CFlowService) Summary(ctx context.Context) (*CFlowSummary, error) {
	var out CFlowSummary
	if err := s.c.CallData(ctx, "cflow", "show", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
