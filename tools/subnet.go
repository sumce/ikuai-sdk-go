package tools

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"strings"

	"github.com/sumce/ikuai-sdk-go"
)

// SubnetInfo 是子网换算的结果（字段对应 i18n subnet 区块的输出项，
// 真机 3.7.24 实测固件无服务端 subnet func）。
type SubnetInfo struct {
	// IPAddr 输入的 IP 地址（i18n subnet.IPAddr）。
	IPAddr string
	// SubnetMask 点分十进制子网掩码（i18n subnet.SubnetMask）。
	SubnetMask string
	// NetworkBits 网络位（CIDR 前缀长度，i18n subnet.network_bit）。
	NetworkBits int64
	// NetworkAddr 网络地址（i18n subnet.NetworkAddr）。
	NetworkAddr string
	// UsableRange 可用主机范围（i18n subnet.UsableRange，形如
	// "a.b.c.d - e.f.g.h"；/31 按 RFC 3021 两地址均可用，/32 为自身）。
	UsableRange string
	// UsableNum 可用数量（i18n subnet.UsableNum；/24 为 254，/31 为 2，
	// /32 为 1，其余为 2^(32-前缀)-2）。
	UsableNum int64
	// RandomAddr 网段内随机可用地址（i18n subnet.RandomAddr；每次
	// 计算随机生成，仅在可用范围内）。
	RandomAddr string
}

// SubnetService 是子网换算（subnet）的子服务
// （tools.subnet「子网换算」页面）。
//
// 真机 3.7.24 实测：subnet.show 返回 30002 Not found funcname(subnet)，
// 即固件没有服务端子网计算接口，控制台在浏览器本地完成换算。本子
// 服务因此以本地纯计算实现（不发任何网络请求），与控制台行为一致；
// 若后续固件新增服务端接口，可经 Service.Raw 探测接入。
type SubnetService struct {
	c *ikuai.Client
}

// Calculate 在本地换算子网信息（纯计算，不发起网络请求）。
//
// ipAddr 为 IPv4 地址；mask 支持「点分掩码（255.255.255.0）」与
// 「前缀位数（24 或 /24）」两种形态；mask 为空时允许 ipAddr 携带
// CIDR 后缀（"192.168.1.5/24"，对应 i18n subnet.IPSegment IP网段
// 输入模式）。参数非法时返回包装 ikuai.ErrInvalidParam 的错误。
func (s *SubnetService) Calculate(ipAddr, mask string) (*SubnetInfo, error) {
	// mask 为空时接受 ipAddr 的 CIDR 后缀（IP网段输入模式）。
	ipAddr = strings.TrimSpace(ipAddr)
	mask = strings.TrimSpace(mask)
	if mask == "" && strings.Contains(ipAddr, "/") {
		idx := strings.Index(ipAddr, "/")
		mask = ipAddr[idx+1:]
		ipAddr = ipAddr[:idx]
	}
	ip := net.ParseIP(ipAddr)
	ip4 := ip.To4()
	if ip == nil || ip4 == nil {
		return nil, fmt.Errorf("tools: 子网计算: IP 地址 %q 非法或非 IPv4: %w", ipAddr, ikuai.ErrInvalidParam)
	}
	if mask == "" {
		return nil, fmt.Errorf("tools: 子网计算: 子网掩码不能为空: %w", ikuai.ErrInvalidParam)
	}
	var bits int
	var maskIP net.IPMask
	if strings.Contains(mask, ".") {
		m := net.ParseIP(mask)
		m4 := m.To4()
		if m == nil || m4 == nil {
			return nil, fmt.Errorf("tools: 子网计算: 子网掩码 %q 非法: %w", mask, ikuai.ErrInvalidParam)
		}
		ones, zeroBits := net.IPMask(m4).Size()
		if ones == 0 && zeroBits == 0 {
			return nil, fmt.Errorf("tools: 子网计算: 子网掩码 %q 非连续: %w", mask, ikuai.ErrInvalidParam)
		}
		bits, maskIP = ones, net.IPMask(m4)
	} else {
		n, err := parsePrefix(strings.TrimPrefix(mask, "/"))
		if err != nil {
			return nil, fmt.Errorf("tools: 子网计算: 子网掩码 %q 非法: %w", mask, ikuai.ErrInvalidParam)
		}
		bits, maskIP = int(n), net.CIDRMask(int(n), 32)
	}
	network := ip4.Mask(maskIP)
	broadcast := make([]byte, 4)
	for i := range network {
		broadcast[i] = network[i] | ^maskIP[i]
	}
	networkU := binary.BigEndian.Uint32(network)
	broadcastU := binary.BigEndian.Uint32(broadcast)
	var firstU, lastU, usable uint64
	switch bits {
	case 32:
		firstU, lastU, usable = uint64(networkU), uint64(networkU), 1
	case 31:
		firstU, lastU, usable = uint64(networkU), uint64(broadcastU), 2
	default:
		firstU, lastU = uint64(networkU)+1, uint64(broadcastU)-1
		usable = uint64(broadcastU) - uint64(networkU) - 1
	}
	randomU := firstU
	if usable > 1 {
		// 可用数量大于 1 时在 [first, last] 内随机取一个可用地址。
		nBig := new(big.Int).SetUint64(usable)
		r, err := rand.Int(rand.Reader, nBig)
		if err == nil {
			randomU = firstU + r.Uint64()
		}
	}
	return &SubnetInfo{
		IPAddr:      ip4.String(),
		SubnetMask:  net.IP(maskIP).String(),
		NetworkBits: int64(bits),
		NetworkAddr: u32ToString(uint64(networkU)),
		UsableRange: u32ToString(firstU) + " - " + u32ToString(lastU),
		UsableNum:   int64(usable),
		RandomAddr:  u32ToString(randomU),
	}, nil
}

// parsePrefix 解析前缀位数文本（0-32）。
func parsePrefix(s string) (int64, error) {
	var n int64
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("前缀位数不能为空")
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, fmt.Errorf("前缀位数 %q 含非数字", s)
		}
		n = n*10 + int64(s[i]-'0')
		if n > 32 {
			return 0, fmt.Errorf("前缀位数 %q 超出 0-32", s)
		}
	}
	return n, nil
}

// u32ToString 将 32 位整数渲染为点分 IPv4 字符串。
func u32ToString(v uint64) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}
