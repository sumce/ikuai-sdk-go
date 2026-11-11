package system

import (
	"context"

	"github.com/sumce/ikuai-sdk-go"
)

// register.show 的 TYPE 选择器常量（抓包 + 真机 3.7.24 验证）。
const (
	// RegisterTypeData 注册节点列表区块。
	RegisterTypeData = "data"
	// RegisterTypeGWID 网关设备唯一 ID 区块。
	RegisterTypeGWID = "gwid"
	// RegisterTypeSCode 服务码（scode）状态区块。
	RegisterTypeSCode = "scode"
)

// register.show 控制台实际使用的 TYPE 组合（burp 抓包实证）。
const (
	// RegisterTypeDataGWID 是「服务注册」页的基础组合
	// （data,gwid），Register.Get 使用。
	RegisterTypeDataGWID = RegisterTypeData + "," + RegisterTypeGWID
	// RegisterTypeDataGWIDSCode 是含服务码状态的完整组合
	// （data,gwid,scode），Register.GetWithServiceCode 使用。
	RegisterTypeDataGWIDSCode = RegisterTypeDataGWID + "," + RegisterTypeSCode
)

// RegisterNode 是一个服务注册节点的行（register.show data 数组元素）。
type RegisterNode struct {
	// ID 注册节点行 ID。
	ID int64 `json:"id"`
	// Node 节点编号（观测值 0；语义待真机验证）。
	Node int64 `json:"node"`
	// Code 注册码（32 位十六进制字符串）。
	Code string `json:"code"`
	// Comment 注册备注（如 "核心路由"）。
	Comment string `json:"comment"`
}

// ServiceCodeState 是服务码（scode）状态块（register.show TYPE=scode）。
// 各字段语义按命名与控制台推断，待真机验证。
type ServiceCodeState struct {
	// Status 服务码状态（观测值 0；枚举语义待真机验证）。
	Status int64 `json:"status"`
	// ServeErrcode 服务端错误码（正常为空）。
	ServeErrcode string `json:"serve_errcode"`
	// ServeErrmsg 服务端错误描述（正常为空）。
	ServeErrmsg string `json:"serve_errmsg"`
	// ServeTimeout 服务超时描述（观测值为空）。
	ServeTimeout string `json:"serve_timeout"`
	// ServeNumber 服务数量描述（观测值为空）。
	ServeNumber string `json:"serve_number"`
}

// RegisterInfo 是服务注册页的聚合结果（register.show TYPE=data,gwid，
// 请求含 scode 时 SCode 非空）。
type RegisterInfo struct {
	// Data 注册节点列表。
	Data []RegisterNode `json:"data"`
	// GWID 网关设备唯一 ID（32 位十六进制，与 homepage.show sysstat.gwid 一致）。
	GWID string `json:"gwid"`
	// SCode 服务码状态块（仅在请求 TYPE 含 scode 时出现，否则 nil）。
	SCode *ServiceCodeState `json:"scode"`
}

// RegisterService 是服务注册（register）的子服务。
type RegisterService struct {
	c *ikuai.Client
}

// Get 获取服务注册信息（注册节点列表 + 网关 ID）。
// 对应 /Action/call: register.show（TYPE=data,gwid），
// 即 Web 控制台「系统设置 → 服务注册」页面。
func (s *RegisterService) Get(ctx context.Context) (*RegisterInfo, error) {
	return s.show(ctx, RegisterTypeDataGWID)
}

// GetWithServiceCode 获取服务注册信息（含服务码状态块）。
// 对应 /Action/call: register.show（TYPE=data,gwid,scode）。
func (s *RegisterService) GetWithServiceCode(ctx context.Context) (*RegisterInfo, error) {
	return s.show(ctx, RegisterTypeDataGWIDSCode)
}

// show 以指定 TYPE 组合调用 register.show 并解码。
func (s *RegisterService) show(ctx context.Context, types string) (*RegisterInfo, error) {
	var out RegisterInfo
	if err := s.c.CallData(ctx, "register", "show",
		map[string]any{"TYPE": types}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
