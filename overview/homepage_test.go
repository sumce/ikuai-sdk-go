package overview

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sumce/ikuai-sdk-go"
	"github.com/sumce/ikuai-sdk-go/internal/testserver"
)

// homepageSystemOverviewData 构造 homepage.show TYPE=sysstat,ac_status
// 的抓包样本数据（值取自 analysis/calls/homepage__show.json）。
func homepageSystemOverviewData() map[string]any {
	return map[string]any{
		"sysstat": map[string]any{
			"cpu":         []any{"6.98%", "12.00%"},
			"cputemp":     []any{39},
			"freq":        []any{"1281", "799"},
			"gwid":        "d072dadeaacc4feb98a22dd0878b867a",
			"hostname":    "iKuai",
			"link_status": 0,
			"memory": map[string]any{
				"total": 47914500, "available": 44857584, "free": 44940664,
				"cached": 127156, "buffers": 267852, "used": "6%",
			},
			"online_user": map[string]any{
				"count": 170, "count_2g": 0, "count_5g": 0,
				"count_wired": 170, "count_wireless": 0,
			},
			"stream": map[string]any{
				"connect_num": 10229, "upload": 22910823, "download": 25402421,
				"total_up": 37472227088584, "total_down": 40669280008537,
			},
			"uptime": 1629218,
			"verinfo": map[string]any{
				"modelname": "", "verstring": "3.7.24 x64 Build202608141210",
				"version": "3.7.24", "build_date": 202608141210, "arch": "x86",
				"sysbit": "x64", "verflags": "", "is_enterprise": 0,
				"support_i18n": 0, "support_dingtalk": 1, "support_lcd": 0,
				"bootguide": "hd",
			},
		},
		"ac_status": map[string]any{"ap_count": 0, "ap_online": 0},
	}
}

// TestServiceSystemOverview 覆盖 homepage.show TYPE=sysstat,ac_status
// 成功路径、参数断言与只读辅助方法。
func TestServiceSystemOverview(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		assert func(t *testing.T, o *SystemOverview)
	}{
		{
			name: "抓包样本全字段解码",
			data: homepageSystemOverviewData(),
			assert: func(t *testing.T, o *SystemOverview) {
				if o.SysStat == nil || o.ACStatus == nil {
					t.Fatalf("sysstat/ac_status 区块均应解码: %+v", o)
				}
				ss := o.SysStat
				if len(ss.CPU) != 2 || ss.CPU[0] != "6.98%" || ss.CPU[1] != "12.00%" {
					t.Fatalf("cpu 解码不符: %v", ss.CPU)
				}
				if len(ss.CPUTemp) != 1 || ss.CPUTemp[0] != 39 {
					t.Fatalf("cputemp 解码不符: %v", ss.CPUTemp)
				}
				if len(ss.Freq) != 2 || ss.Freq[0] != "1281" {
					t.Fatalf("freq 解码不符: %v", ss.Freq)
				}
				if ss.GWID != "d072dadeaacc4feb98a22dd0878b867a" || ss.Hostname != "iKuai" ||
					ss.LinkStatus != 0 || ss.Uptime != 1629218 {
					t.Fatalf("sysstat 标量字段解码不符: %+v", ss)
				}
				m := ss.Memory
				if m.Total != 47914500 || m.Available != 44857584 || m.Free != 44940664 ||
					m.Cached != 127156 || m.Buffers != 267852 || m.Used != "6%" {
					t.Fatalf("memory 解码不符: %+v", m)
				}
				u := ss.OnlineUser
				if u.Count != 170 || u.Count2G != 0 || u.Count5G != 0 ||
					u.CountWired != 170 || u.CountWireless != 0 {
					t.Fatalf("online_user 解码不符: %+v", u)
				}
				st := ss.Stream
				if st.ConnectNum != 10229 || st.Upload != 22910823 ||
					st.Download != 25402421 || st.TotalUp != 37472227088584 ||
					st.TotalDown != 40669280008537 {
					t.Fatalf("stream 解码不符: %+v", st)
				}
				v := ss.VerInfo
				if v.Modelname != "" || v.Verstring != "3.7.24 x64 Build202608141210" ||
					v.Version != "3.7.24" || v.BuildDate != 202608141210 ||
					v.Arch != "x86" || v.Sysbit != "x64" || v.Verflags != "" ||
					v.IsEnterprise != 0 || v.SupportI18n != 0 ||
					v.SupportDingTalk != 1 || v.SupportLCD != 0 || v.BootGuide != "hd" {
					t.Fatalf("verinfo 解码不符: %+v", v)
				}
				if o.ACStatus.APCount != 0 || o.ACStatus.APOnline != 0 {
					t.Fatalf("ac_status 解码不符: %+v", o.ACStatus)
				}
				// 只读辅助方法。
				if d := o.UptimeDuration(); d != 1629218*time.Second {
					t.Fatalf("UptimeDuration = %v, 期望 %v", d, 1629218*time.Second)
				}
				avg, err := o.CPUAverage()
				if err != nil {
					t.Fatalf("CPUAverage 失败: %v", err)
				}
				if want := (6.98 + 12.00) / 2; avg != want {
					t.Fatalf("CPUAverage = %v, 期望 %v", avg, want)
				}
				pct, err := ss.Memory.UsedPercent()
				if err != nil || pct != 6 {
					t.Fatalf("UsedPercent = %v, %v; 期望 6", pct, err)
				}
				bt, err := v.BuildTime()
				if err != nil {
					t.Fatalf("BuildTime 失败: %v", err)
				}
				want := time.Date(2026, 8, 14, 12, 10, 0, 0, time.Local)
				if !bt.Equal(want) {
					t.Fatalf("BuildTime = %v, 期望 %v", bt, want)
				}
			},
		},
		{
			name: "仅 sysstat 区块时 ac_status 为 nil 且辅助方法安全",
			data: map[string]any{
				"sysstat": map[string]any{
					"cpu": []any{"3.00%", "5.00%", "7.00%"},
					"memory": map[string]any{
						"total": 47914500, "available": 44857584, "free": 44940664,
						"cached": 127156, "buffers": 267852, "used": "6%",
					},
					"uptime": 100,
				},
			},
			assert: func(t *testing.T, o *SystemOverview) {
				if o.SysStat == nil {
					t.Fatal("sysstat 区块应解码")
				}
				if o.ACStatus != nil {
					t.Fatalf("缺失的 ac_status 区块应为 nil: %+v", o.ACStatus)
				}
				avg, err := o.CPUAverage()
				if err != nil || avg != 5.0 {
					t.Fatalf("CPUAverage = %v, %v; 期望 5.0", avg, err)
				}
				if d := o.UptimeDuration(); d != 100*time.Second {
					t.Fatalf("UptimeDuration = %v, 期望 100s", d)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := testserver.New()
			defer ts.Close()
			ts.SetResponse("homepage", "show", tt.data)
			o, err := New(ts.Client()).SystemOverview(context.Background())
			if err != nil {
				t.Fatalf("SystemOverview 失败: %v", err)
			}
			tt.assert(t, o)
			rec := ts.CallsFor("homepage")[0]
			if rec.Action != "show" {
				t.Fatalf("action = %q, 期望 show", rec.Action)
			}
			wantParam(t, paramOf(t, rec),
				map[string]any{"TYPE": HomepageTypeSystemOverview})
		})
	}
}

// TestHomepageDHCPAddrpoolNum 覆盖 TYPE=dhcp_addrpool_num。
func TestHomepageDHCPAddrpoolNum(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{"available_num": 17})
	n, err := New(ts.Client()).Homepage.DHCPAddrpoolNum(context.Background())
	if err != nil {
		t.Fatalf("DHCPAddrpoolNum 失败: %v", err)
	}
	if n != 17 {
		t.Fatalf("available_num = %d, 期望 17", n)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": HomepageTypeDHCPAddrpoolNum})
}

// TestHomepageInterfaceSnapshot 覆盖 TYPE=ether_info,snapshoot。
func TestHomepageInterfaceSnapshot(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{
		"ether_info": map[string]any{
			"eth0": map[string]any{
				"driver": "ice", "duplex": 1, "interface": "wan1", "link": 1,
				"lock": 0, "mac": "44:a9:2c:21:c1:4e",
				"model": "Intel Corporation Ethernet Connection E823-L",
				"speed": 10000, "type": "FIBRE",
			},
			"eth11": map[string]any{
				"driver": "r8125", "duplex": 0, "interface": "Null", "link": 0,
				"lock": 0, "mac": "80:ae:54:6c:c0:b8",
				"model": "Realtek Semiconductor Co., Ltd. RTL8125 2.5GbE",
				"speed": 0, "type": "TP",
			},
		},
		"snapshoot_lan": []any{map[string]any{
			"bandmode": 0, "comment": "李冰卧室", "id": 4, "interface": "lan4",
			"ip_addr": "192.168.2.1", "linkmode": 4, "mac": "44:a9:2c:21:c1:51",
			"member": []any{"eth3", "eth4"}, "netmask": "255.255.255.0",
		}},
		"snapshoot_wan": []any{map[string]any{
			"bandmode": 0, "ccid": "", "check_res": 0, "comment": "",
			"count_check_fail": 0, "count_dhcp": 0, "count_pppoe": 13,
			"count_static": 0, "default_route": 0, "dns1": "", "dns2": "",
			"errmsg": "", "gateway": "", "id": 1, "imei": "",
			"interface": "wan1", "internet": 4, "ip_addr": "", "isnr": "",
			"isp": "", "mac": "44:a9:2c:21:c1:4e", "member": []any{"eth0"},
			"netmask": "", "pcid": "", "power": "", "qnw": "", "updatetime": 0,
		}},
	})
	out, err := New(ts.Client()).Homepage.InterfaceSnapshot(context.Background())
	if err != nil {
		t.Fatalf("InterfaceSnapshot 失败: %v", err)
	}
	eth0, ok := out.EtherInfo["eth0"]
	if !ok || eth0.Driver != "ice" || eth0.Link != 1 || eth0.Speed != 10000 ||
		eth0.Interface != "wan1" || eth0.Type != "FIBRE" {
		t.Fatalf("ether_info[eth0] 解码不符: %+v (%v)", eth0, ok)
	}
	eth11 := out.EtherInfo["eth11"]
	if eth11.Interface != "Null" || eth11.Link != 0 || eth11.Type != "TP" {
		t.Fatalf("ether_info[eth11] 解码不符: %+v", eth11)
	}
	if len(out.SnapshootLAN) != 1 {
		t.Fatalf("snapshoot_lan 行数不符: %d", len(out.SnapshootLAN))
	}
	lan := out.SnapshootLAN[0]
	if lan.Interface != "lan4" || lan.IPAddr != "192.168.2.1" ||
		len(lan.Member) != 2 || lan.Member[0] != "eth3" || lan.Member[1] != "eth4" ||
		lan.Netmask != "255.255.255.0" || lan.Comment != "李冰卧室" {
		t.Fatalf("snapshoot_lan 解码不符: %+v", lan)
	}
	if len(out.SnapshootWAN) != 1 {
		t.Fatalf("snapshoot_wan 行数不符: %d", len(out.SnapshootWAN))
	}
	wan := out.SnapshootWAN[0]
	if wan.Interface != "wan1" || wan.CountPPPoE != 13 || wan.Internet != 4 ||
		wan.ID != 1 || len(wan.Member) != 1 || wan.Member[0] != "eth0" {
		t.Fatalf("snapshoot_wan 解码不符: %+v", wan)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": HomepageTypeEtherInfoSnapshoot})
}

// TestHomepageAppHistorical 覆盖 TYPE=app_historical（固件双重包裹形态）。
func TestHomepageAppHistorical(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{
		"app_historical": map[string]any{
			"app_historical": []any{
				[]any{
					map[string]any{"download": "19074932", "name": "Total", "upload": "18420757"},
					map[string]any{"download": "16022297", "name": "Unknown", "upload": "16540064"},
					map[string]any{"download": "1025", "name": "Others", "upload": "4708"},
				},
				[]any{
					map[string]any{"download": "0", "name": "Test", "upload": "0"},
					map[string]any{"download": "1032", "name": "Others", "upload": "1558"},
				},
			},
		},
	})
	out, err := New(ts.Client()).Homepage.AppHistorical(context.Background())
	if err != nil {
		t.Fatalf("AppHistorical 失败: %v", err)
	}
	if len(out.Groups) != 2 {
		t.Fatalf("app_historical 组数不符: %d", len(out.Groups))
	}
	g0 := out.Groups[0]
	if len(g0) != 3 || g0[0].Name != "Total" || g0[0].Download != "19074932" ||
		g0[0].Upload != "18420757" || g0[2].Name != "Others" {
		t.Fatalf("app_historical[0] 解码不符: %+v", g0)
	}
	g1 := out.Groups[1]
	if len(g1) != 2 || g1[0].Name != "Test" || g1[0].Download != "0" {
		t.Fatalf("app_historical[1] 解码不符: %+v", g1)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": HomepageTypeAppHistorical})
}

// TestHomepageAppFlow 覆盖 TYPE=app_flow,app_historical（固件对两个区块
// 均双重包裹）。
func TestHomepageAppFlow(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{
		"app_flow": map[string]any{
			"app_flow": []any{map[string]any{
				"Total": 95619009052, "HTTP": 13492960546, "Download": 1099198888,
				"Transport": 38558, "IM": 63425829, "Video": 160837579,
				"Common": 564333075, "Game": 136056172, "Others": 55025233,
				"Test": 0, "Unknown": 80047133172,
			}},
		},
		"app_historical": map[string]any{
			"app_historical": []any{
				[]any{map[string]any{"download": "0", "name": "Test", "upload": "0"}},
			},
		},
	})
	out, err := New(ts.Client()).Homepage.AppFlow(context.Background())
	if err != nil {
		t.Fatalf("AppFlow 失败: %v", err)
	}
	if len(out.AppFlow.Rows) != 1 || out.AppFlow.Rows[0].Total != 95619009052 ||
		out.AppFlow.Rows[0].Unknown != 80047133172 || out.AppFlow.Rows[0].Test != 0 {
		t.Fatalf("app_flow 解码不符: %+v", out.AppFlow.Rows)
	}
	if len(out.AppHistorical.Groups) != 1 || out.AppHistorical.Groups[0][0].Name != "Test" {
		t.Fatalf("app_historical 解码不符: %+v", out.AppHistorical.Groups)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": HomepageTypeAppFlowHistorical})
}

// TestHomepageMonitorSystem 覆盖 TYPE=monitor_system。
func TestHomepageMonitorSystem(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{
		"monitor_system": map[string]any{
			"bandwidth_total": map[string]any{"upload": 0, "download": 0},
			"every_day_stats": []any{map[string]any{
				"clean_time": 20261110, "id": 1, "wifi": 0,
			}},
		},
	})
	out, err := New(ts.Client()).Homepage.MonitorSystem(context.Background())
	if err != nil {
		t.Fatalf("MonitorSystem 失败: %v", err)
	}
	if out.BandwidthTotal.Upload != 0 || out.BandwidthTotal.Download != 0 {
		t.Fatalf("bandwidth_total 解码不符: %+v", out.BandwidthTotal)
	}
	if len(out.EveryDayStats) != 1 || out.EveryDayStats[0].CleanTime != 20261110 ||
		out.EveryDayStats[0].ID != 1 || out.EveryDayStats[0].WiFi != 0 {
		t.Fatalf("every_day_stats 解码不符: %+v", out.EveryDayStats)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": HomepageTypeMonitorSystem})
}

// TestHomepageAntipiracy 覆盖 TYPE=antipiracy。
func TestHomepageAntipiracy(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{"antipiracy": -1})
	n, err := New(ts.Client()).Homepage.Antipiracy(context.Background())
	if err != nil {
		t.Fatalf("Antipiracy 失败: %v", err)
	}
	if n != -1 {
		t.Fatalf("antipiracy = %d, 期望 -1", n)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": HomepageTypeAntipiracy})
}

// TestHomepageShow 覆盖任意 TYPE 组合的 Show 方法与参数断言。
func TestHomepageShow(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.SetResponse("homepage", "show", map[string]any{"sysstat": map[string]any{}})
	resp, err := New(ts.Client()).Homepage.Show(context.Background(),
		HomepageTypeSysStat, HomepageTypeMonitorSystem, HomepageTypeAntipiracy)
	if err != nil {
		t.Fatalf("Show 失败: %v", err)
	}
	if !resp.OK() {
		t.Fatalf("Show 应成功, got %+v", resp)
	}
	wantParam(t, paramOf(t, ts.CallsFor("homepage")[0]),
		map[string]any{"TYPE": "sysstat,monitor_system,antipiracy"})
}

// TestHomepageShowEmptyTypes 空 TYPE 列表应返回 ErrInvalidParam。
func TestHomepageShowEmptyTypes(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	if _, err := New(ts.Client()).Homepage.Show(context.Background()); !errors.Is(err, ikuai.ErrInvalidParam) {
		t.Fatalf("空 TYPE 应返回 ErrInvalidParam, got %v", err)
	}
}

// TestHomepageError 覆盖固件错误路径（Result!=30000 → *ikuai.APIError）。
func TestHomepageError(t *testing.T) {
	ts := testserver.New()
	defer ts.Close()
	ts.HandleFuncName("homepage", func(action string, param json.RawMessage) (any, error) {
		return nil, &ikuai.APIError{Code: 95006, ErrMsg: "失败", FuncName: "homepage", Action: action}
	})
	svc := New(ts.Client())
	ctx := context.Background()
	for name, call := range map[string]func() error{
		"SystemOverview": func() error {
			_, err := svc.SystemOverview(ctx)
			return err
		},
		"Homepage.DHCPAddrpoolNum": func() error {
			_, err := svc.Homepage.DHCPAddrpoolNum(ctx)
			return err
		},
		"Homepage.InterfaceSnapshot": func() error {
			_, err := svc.Homepage.InterfaceSnapshot(ctx)
			return err
		},
		"Homepage.AppHistorical": func() error {
			_, err := svc.Homepage.AppHistorical(ctx)
			return err
		},
		"Homepage.AppFlow": func() error {
			_, err := svc.Homepage.AppFlow(ctx)
			return err
		},
		"Homepage.MonitorSystem": func() error {
			_, err := svc.Homepage.MonitorSystem(ctx)
			return err
		},
		"Homepage.Antipiracy": func() error {
			_, err := svc.Homepage.Antipiracy(ctx)
			return err
		},
		"Homepage.Show": func() error {
			_, err := svc.Homepage.Show(ctx, HomepageTypeSysStat)
			return err
		},
	} {
		err := call()
		ae, ok := ikuai.AsAPIError(err)
		if !ok || ae.Code != 95006 {
			t.Fatalf("%s 期望 APIError(95006), got %v", name, err)
		}
	}
}
