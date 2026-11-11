# iKuai SDK for Go

[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

爱快（iKuaiOS）路由器 Web API 的全功能 Go SDK —— 基于真机（iKuaiOS 3.7.24）抓包逆向验证，
覆盖 Web 控制台 **12 大类 151 个功能**，全部强类型、中文注释、零第三方依赖。

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/overview"
)

func main() {
	c, err := ikuai.NewClient("http://192.168.1.1", "admin", "your-password")
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := c.Login(ctx); err != nil { // passwd=md5，会话自动管理
		log.Fatal(err)
	}
	defer c.Logout(context.WithoutCancel(ctx))

	ov, _ := overview.New(c).SystemOverview(ctx) // 系统概况仪表盘
	fmt.Printf("主机 %s，固件 %s，在线终端 %d，CPU %.1f%%\n",
		ov.SysStat.Hostname, ov.SysStat.VerInfo.Version,
		ov.SysStat.OnlineUser.Count, ov.SysStat.CPUAverage())
}
```

## 特性

- **全功能覆盖**：151 个 func_name 全部强类型实现（对照固件 `menu.json` 完整清单）
- **真机实证**：所有字段基于 Burp Suite 抓包 + 真机只读探测逐字段核对；无法实证的字段逐一标注「待真机验证」
- **会话全自动**：登录（`passwd=md5`）、`sess_key` Cookie、`Result=10014` 自动重登重试，业务层零感知
- **容错解码**：兼容固件怪癖（非法 JSON `"Data":}`、`"Data":{}` 空对象、字符串数字等）
- **双通道**：强类型 Service 之外，`Call` / `CRUD` 通用层保证任何 func/action/param 组合可达
- **零依赖**：纯标准库；函数式选项（超时/TLS 跳过验证/日志/重登策略）
- **双轨测试**：离线假路由器（`internal/testserver`）+ 真机联调（`internal/livetest`，环境变量门控）

## 安装

```
go get github.com/sumce/ikuai-sdk-go
```

## 使用方式（门面 + 按包构造）

**方式一：门面聚合**（`sdk` 包，一站式）

```go
router, _ := sdk.New(&sdk.Config{BaseURL: "http://192.168.1.1", Username: "admin", Password: "..."})
router.Login(ctx)
ov, _ := router.Overview.SystemOverview(ctx)   // 系统概况
dnats, _ := router.Network.DNAT.List(ctx, nil) // 端口映射列表
```

**方式二：按包构造**（用哪个引哪个）

```go
mon := monitoring.New(c)
terminals, _ := mon.LanIP.Search(ctx, "172.22", nil)   // 终端监控·关键字搜索
curves, _ := mon.System.Curves(ctx, nil)               // 负载监控·曲线
```

## 12 大类与功能清单

| 包 | 大类 | 功能数 | 亮点 |
|---|---|---|---|
| `overview` | 系统概况 | 1 | homepage 仪表盘 7 种 TYPE 组合 + 独立 `sysstat` func |
| `monitoring` | 状态监控 | 10 | 线路/终端(IPv4+IPv6)/协议/策略/负载/分流/摄像头/交换机/周边 |
| `system` | 系统设置 | 17 | 基础/热备/磁盘/文件/RAID/升级/备份/账号（含企业微信 wecom） |
| `network` | 网络设置 | 35 | LAN/DHCP/DNS/VLAN/分组/路由/VPN客户端×8/UPnP/NAT/IPv6×7/组播 |
| `flowcontrol` | 流控分流 | 10 | 多线负载/协议·端口·域名·上下线分流/智能流控/IP·MAC 限速/自定义协议 |
| `ac` | AC管理 | 8 | AP 统计 20 种 TYPE/配置/分组/升级/黑白名单/在线终端 21 列实证 |
| `accounting` | 认证计费 | 16 | Web认证 141 字段/PPPoE·PPTP·L2TP·OpenVPN·IKEv2 服务端/账号/套餐/优惠券 |
| `behavior` | 行为管控 | 17 | 审计设置与日志/MAC备注/ACL 族/网页过滤/域名过滤 |
| `security` | 安全设置 | 5 | ACL/ARP 表/连接数限制/安全高级/防火墙 |
| `advanced` | 高级应用 | 13 | 网页缓存/DDNS/SNMP/端口镜像/虚拟机/插件/唤醒/FTP·SMB·HTTP/UDP代理 |
| `tools` | 应用工具 | 8 | Ping/Traceroute/抓包/iperf/子网计算(本地)/测速/诊断/看门狗 |
| `logcenter` | 日志中心 | 11 | 9 类 syslog 日志/站内信 ikmessages/实时采集 rt_collect |

每个包均可独立构造，方法注释注明对应的 `func_name/action` 与控制台菜单位置。
完整 func 清单与抓包实证样本见 [docs/API_CATALOG.md](docs/API_CATALOG.md)。

## 架构

```
ikuai-sdk/
├── （根）ikuai 核心包    Client/Call/CallData/CallRaw · Login/自动重登 · envelope/容错JSON · APIError · ListOptions/Page[T] · CRUD 兜底
├── overview … logcenter  12 个业务包（每类一包，一 func 一文件，全部强类型）
├── sdk/                  门面聚合包
├── internal/testserver/  离线假路由器（所有测试共用）
├── internal/livetest/    真机测试辅助（环境变量门控）
├── examples/             可运行示例
└── docs/                 RULES.md（开发契约）· API_CATALOG.md（151 func 实证目录）
```

依赖方向严格单向：`业务包 → 核心包`，`sdk → 全部`。详见 [docs/RULES.md](docs/RULES.md)。

## 测试

**离线**（无需路由器，永远全绿）：

```
go test ./...
```

**真机**（只读；写动作一律不对真机执行）：

```
IKUAI_BASE_URL=http://192.168.1.1 IKUAI_USERNAME=admin IKUAI_PASSWORD=xxx go test ./...
```

- 环境变量未设置时真机测试自动跳过
- 写动作（add/edit/del/set）只做离线覆盖，代码中逐处标注「禁止对真机调用」
- 测试凭据只走环境变量，不入仓库

## 已知限制

- 未授权/未安装模块的功能（如 `wireguard`、`ike_client`、`pppoe_proxy`）固件返回
  `Result=30002 Not found funcname`，SDK 已如实建模并在 live 测试中断言该形态
- 空表功能的行字段由固件 i18n + 表单语义推导，已在注释标注「待真机验证」，
  有真实数据的设备上会自动验证
- 高危动作（插件装卸、虚拟机电源、AP 升级、系统重启等）不提供便捷方法，
  统一经各包 `Raw` 逃生口或核心 `Call` 显式调用

## License

[MIT](LICENSE)
