package tools

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/sumce/ikuai-sdk-go"
)

// TestSubnetCalculate 覆盖子网换算的输入形态与边界（本地纯计算）。
func TestSubnetCalculate(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		mask    string
		wantIP  string
		wantMk  string
		wantNet string
		wantRng string
		wantNum int64
	}{
		{
			name: "CIDR 后缀形态", ip: "192.168.1.5/24", mask: "",
			wantIP: "192.168.1.5", wantMk: "255.255.255.0",
			wantNet: "192.168.1.0", wantRng: "192.168.1.1 - 192.168.1.254", wantNum: 254,
		},
		{
			name: "点分掩码形态", ip: "10.0.0.77", mask: "255.255.0.0",
			wantIP: "10.0.0.77", wantMk: "255.255.0.0",
			wantNet: "10.0.0.0", wantRng: "10.0.0.1 - 10.0.255.254", wantNum: 65534,
		},
		{
			name: "前缀位数形态", ip: "172.22.1.2", mask: "/16",
			wantIP: "172.22.1.2", wantMk: "255.255.0.0",
			wantNet: "172.22.0.0", wantRng: "172.22.0.1 - 172.22.255.254", wantNum: 65534,
		},
		{
			name: "全零掩码", ip: "1.2.3.4", mask: "0.0.0.0",
			wantIP: "1.2.3.4", wantMk: "0.0.0.0",
			wantNet: "0.0.0.0", wantRng: "0.0.0.1 - 255.255.255.254", wantNum: 4294967294,
		},
		{
			name: "31 位 RFC3021", ip: "10.0.0.2", mask: "255.255.255.254",
			wantIP: "10.0.0.2", wantMk: "255.255.255.254",
			wantNet: "10.0.0.2", wantRng: "10.0.0.2 - 10.0.0.3", wantNum: 2,
		},
		{
			name: "32 位单机", ip: "10.0.0.2", mask: "32",
			wantIP: "10.0.0.2", wantMk: "255.255.255.255",
			wantNet: "10.0.0.2", wantRng: "10.0.0.2 - 10.0.0.2", wantNum: 1,
		},
		{
			name: "30 位", ip: "192.168.1.5", mask: "30",
			wantIP: "192.168.1.5", wantMk: "255.255.255.252",
			wantNet: "192.168.1.4", wantRng: "192.168.1.5 - 192.168.1.6", wantNum: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := New(nil).Subnet.Calculate(tt.ip, tt.mask)
			if err != nil {
				t.Fatalf("Calculate 失败: %v", err)
			}
			if info.IPAddr != tt.wantIP || info.SubnetMask != tt.wantMk ||
				info.NetworkAddr != tt.wantNet || info.UsableRange != tt.wantRng ||
				info.UsableNum != tt.wantNum || info.NetworkBits < 0 {
				t.Fatalf("换算结果不符: %+v", info)
			}
			// RandomAddr 必须落在可用范围内。
			if ip := net.ParseIP(info.RandomAddr); ip == nil || ip.To4() == nil {
				t.Fatalf("随机地址非法: %q", info.RandomAddr)
			}
			if !strings.Contains(tt.wantRng, "-") {
				t.Fatalf("测试用例 usable range 形态异常: %q", tt.wantRng)
			}
			first, last := tt.wantRng[:strings.Index(tt.wantRng, " - ")],
				tt.wantRng[strings.LastIndex(tt.wantRng, " - ")+3:]
			if !ipv4InRange(info.RandomAddr, first, last) {
				t.Fatalf("随机地址 %q 不在可用范围 %q 内", info.RandomAddr, tt.wantRng)
			}
		})
	}
}

// ipv4InRange 判断 ip 是否落在 [first, last] 区间内。
func ipv4InRange(ip, first, last string) bool {
	v, f, l := ipv4ToUint(ip), ipv4ToUint(first), ipv4ToUint(last)
	return v >= f && v <= l
}

// ipv4ToUint 将点分 IPv4 转为无符号整数（非法输入返回 0）。
func ipv4ToUint(s string) uint64 {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return 0
	}
	var out uint64
	for _, p := range parts {
		n, err := strconv.ParseUint(p, 10, 8)
		if err != nil {
			return 0
		}
		out = out<<8 | n
	}
	return out
}

// TestSubnetCalculateRandomDistribution 抽样验证随机地址确实随机
// （多次计算应出现不同结果）。
func TestSubnetCalculateRandomDistribution(t *testing.T) {
	svc := New(nil).Subnet
	seen := map[string]bool{}
	for i := 0; i < 32; i++ {
		info, err := svc.Calculate("192.168.1.5", "24")
		if err != nil {
			t.Fatalf("Calculate 失败: %v", err)
		}
		seen[info.RandomAddr] = true
	}
	if len(seen) < 2 {
		t.Fatalf("多次计算随机地址应出现不同值, got %v", seen)
	}
}

// TestSubnetCalculateErrors 覆盖参数非法路径（错误链含
// ikuai.ErrInvalidParam）。
func TestSubnetCalculateErrors(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		mask string
	}{
		{name: "IP 非法", ip: "999.1.1.1", mask: "24"},
		{name: "IP 为空", ip: "", mask: "24"},
		{name: "IPv6 不支持", ip: "2001:db8::1", mask: "24"},
		{name: "掩码为空", ip: "192.168.1.5", mask: ""},
		{name: "掩码非数字", ip: "192.168.1.5", mask: "abc"},
		{name: "前缀越界", ip: "192.168.1.5", mask: "33"},
		{name: "掩码非连续", ip: "192.168.1.5", mask: "255.0.255.0"},
		{name: "掩码 IP 非法", ip: "192.168.1.5", mask: "255.255.x.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(nil).Subnet.Calculate(tt.ip, tt.mask)
			if err == nil {
				t.Fatalf("Calculate(%q,%q) 应失败", tt.ip, tt.mask)
			}
			if !errors.Is(err, ikuai.ErrInvalidParam) {
				t.Fatalf("错误应包装 ikuai.ErrInvalidParam: %v", err)
			}
		})
	}
}
