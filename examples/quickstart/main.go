// quickstart 展示 iKuai SDK 的最小可用流程：登录 → 系统概况 →
// 状态监控查询 → 端口映射列表 → 通用 CRUD 兜底。
//
// 运行（替换为你的路由器地址与账号）：
//
//	go run ./examples/quickstart -base http://192.168.1.1 -user admin -pass yourpass
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/monitoring"
	"github.com/sumce/ikuai-sdk-go/network"
	"github.com/sumce/ikuai-sdk-go/overview"
)

func main() {
	base := flag.String("base", "http://192.168.1.1", "路由器管理地址")
	user := flag.String("user", "admin", "管理账号")
	pass := flag.String("pass", "", "管理密码")
	flag.Parse()
	if *pass == "" {
		log.Fatal("请通过 -pass 提供密码")
	}

	c, err := ikuai.NewClient(*base, *user, *pass)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := c.Login(ctx); err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	defer c.Logout(context.WithoutCancel(ctx))

	// 1) 系统概况（homepage.show TYPE=sysstat,ac_status）。
	ov, err := overview.New(c).SystemOverview(ctx)
	if err != nil {
		log.Fatalf("系统概况: %v", err)
	}
	memPct, _ := ov.SysStat.Memory.UsedPercent()
	fmt.Printf("主机 %s | 固件 %s | 在线终端 %d | 内存 %.0f%% | 运行 %s\n",
		ov.SysStat.Hostname, ov.SysStat.VerInfo.Version,
		ov.SysStat.OnlineUser.Count, memPct,
		ov.UptimeDuration().Round(time.Hour))

	// 2) 状态监控：终端监控 IPv4，按关键字搜索并分页。
	mon := monitoring.New(c)
	page, err := mon.LanIP.List(ctx, &ikuai.ListOptions{
		Limit: 5, Finds: "ip_addr,mac,comment,hostname", Keywords: "172.22",
	})
	if err != nil {
		log.Fatalf("终端监控: %v", err)
	}
	fmt.Printf("\n内网终端（命中 %d，显示前 %d）：\n", page.Total, len(page.Rows))
	for _, t := range page.Rows {
		fmt.Printf("  %-15s %-17s %s\n", t.IPAddr, t.MAC, t.Hostname)
	}

	// 3) 网络设置：端口映射列表（抓包实证的标准分页）。
	dnats, err := network.New(c).DNAT.List(ctx, nil)
	if err != nil {
		log.Fatalf("端口映射: %v", err)
	}
	fmt.Printf("\n端口映射共 %d 条：\n", dnats.Total)
	for i, r := range dnats.Rows {
		if i >= 5 {
			fmt.Printf("  ... 其余 %d 条略\n", dnats.Total-5)
			break
		}
		fmt.Printf("  [%d] %s:%s -> %s:%s (%s) %s\n",
			r.ID, r.Interface, r.WANPort, r.LANAddr, r.LANPort, r.Protocol, r.Comment)
	}

	// 4) 兜底通道：任意 func_name/action/param 都可以通过 Call 直达。
	resp, err := c.Call(ctx, "cflow", "show", nil)
	if err != nil {
		log.Fatalf("通用调用: %v", err)
	}
	if total, ok := resp.Count(); ok {
		fmt.Printf("\n流量监控连接统计总数: %d（Result=%d）\n", total, resp.Result)
	}
}
