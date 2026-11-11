# iKuai API 目录（依据真机 burp.xml 抓包 + 固件 menu.json，151 个 func）

> 数据来源：真机 `http://192.168.1.1`（iKuaiOS 3.7.24 x64 Build202608141210）
> 的 Burp Suite 导出（88 条完整请求/响应）与固件菜单树 `json/menu.json`。
> 标注 ✅ 的条目有真机抓包实证；其余条目的 action 集合按固件菜单语义推导，
> 字段以实机验证为准，不确定的字段在类型注释中标注待验证（见 RULES.md §10）。

## 通用协议

- **登录** `POST /Action/login`
  `{"username":"u","passwd":"<md5(password)>","pass":"...","remember_password":false}`
  → `Result=10000` + `Set-Cookie: sess_key=<32hex>`；凭据错误 → `Result=10001`。
  （实测 `pass` 字段可省略。）
- **调用** `POST /Action/call`
  `{"func_name":"...","action":"...","param":{...}}`
  Cookie: `sess_key=..; username=..; login=1`
  成功 → `{"Result":30000,"ErrMsg":"Success","Data":{...}}`；add 另有 `"RowId":N`。
- **会话失效** → `{"ErrMsg":"no login authentication","Result":10014}`（HTTP 仍 200）。
- **show 分页参数**：`TYPE`（默认 `total,data`）、`limit:"offset,rows"`、
  `ORDER_BY`、`ORDER`、`FINDS`、`KEYWORDS`、`FILTER1..n:"字段,操作符,值"`。
- **结果码**：10000 登录成功 / 10001 账号密码错误 / 10014 未登录 / 30000 调用成功。
- **怪癖**：部分响应 Data 为空且 JSON 非法（如 `auth_code:show` 回
  `"Data":}`），核心包已做容错修复。

## 各大类 func 目录

### 1. 系统概况 overview（无子菜单）
| func | 说明 | 实证 |
|---|---|---|
| homepage | 控制台首页仪表盘 | ✅ `show TYPE=sysstat,ac_status` |

### 2. 状态监控 monitoring
| func | 说明 | 实证 |
|---|---|---|
| monitor_iface | 接口监控 | |
| monitor_lanip | 内网监控 IPv4 | ✅ `show TYPE=data,total, ORDER_BY=ip_addr_int, FINDS, KEYWORDS` |
| monitor_lanipv6 | 内网监控 IPv6 | ✅ `show TYPE=interface` |
| monitor_app_flow | 协议/应用流量 | ✅ `show TYPE=app_history, interface, datetype=day, time_range=all` |
| monitor_l7qos | 流控策略监控 | ✅ `show TYPE=data,interface, interface=""` |
| monitor_system | 系统监控曲线 | ✅ `show TYPE=cpu,memory,disk_space_used,on_terminal,conn_num,cputemp1,cputemp2, datetype=hour, math=avg` |
| cflow | 流量监控汇总 | ✅ `show {}` |
| camera | 摄像头监控 | |
| cloud_switch | 交换机监控 | |
| dev_control | 智能设备管控 | |

### 3. 系统设置 system
| func | 说明 | 实证 |
|---|---|---|
| basic | 基础设置 | ✅ `set_language {"language":1}` |
| vrrp_config | 主备热备 | |
| disk_mgmt | 磁盘分区 | |
| file_mgmt | 文件管理 | |
| raid_mgmt | RAID | |
| dingtalk | 应用绑定（钉钉） | ✅ `show TYPE=bind_status` |
| register | 服务注册 | ✅ `show TYPE=data,gwid` |
| alg | ALG 设置 | |
| irqbalance | 负载监控 | |
| ik_sysctl | 内核参数 | |
| webuser | Web 账号管理 | ✅ `show TYPE=mod_passwd, username=...`；add 返回 RowId ✅ |
| remote_control | 远程控制 | |
| upgrade | 升级 | ✅ `show TYPE=data,fileinfo`（含 system_ver/new_system_ver 等） |
| backup | 备份 | |
| hardwareinfo | 硬件信息 | |
| reboots | 重启 | |

### 4. 网络设置 network（35 个）
| func | 说明 | 实证 |
|---|---|---|
| lan | 内外网设置 | |
| dhcp_server | DHCP 服务端 | |
| dhcp_static | DHCP 静态分配 | |
| dhcp_lease | DHCP 客户端列表 | |
| dhcp_acl_mac | DHCP 黑白名单 | |
| dns | DNS 设置 | |
| dns_replace | 多线 DNS 分流 | |
| vlan | VLAN 设置 | |
| ipgroup | IPv4 分组 | |
| ipv6group | IPv6 分组 | |
| macgroup | MAC 分组 | |
| dtgroup | 时间分组 | |
| static_rt | 静态路由 | |
| static_rt_table | 当前路由表 | |
| pptp_client | PPTP 客户端 | |
| l2tp_client | L2TP 客户端 | |
| openvpn-client | OpenVPN 客户端 | |
| ipsec-vpn | IPSec 客户端 | |
| wireguard | WireGuard 客户端 | |
| ike_client | IKEv2 客户端 | |
| ik_web_sdwan | SD-WAN | |
| wuying | 无影云 | |
| upnpd | UPnP 状态 | |
| upnpd_leases | UPnP 设置/租约 | |
| nat_rule | NAT 规则 | |
| dnat | 端口映射 | ✅ `show/add`（字段见下） |
| netmap | DMZ 主机 | |
| ipv6 | IPv6 设置 | |
| ipv6_stream | IPv6 分流 | |
| dhcp6_lease | DHCPv6 终端 | |
| ipv6_static | DHCPv6 静态分配 | |
| dhcp6_acl_mac | DHCPv6 黑白名单 | |
| ipv6_neighbor | IPv6 邻居表 | |
| igmp_proxy | IGMP 代理 | |
| iptv | IPTV | |

### 5. 流控分流 flowcontrol
| func | 说明 | 实证 |
|---|---|---|
| lb_pcc | 多线负载 | |
| stream_layer7 | 协议分流 | |
| stream_ipport | 端口分流 | |
| stream_domain | 域名分流 | |
| stream_updown | 流量上下线分流 | |
| stream_control | 智能流控 | |
| simple_qos | IP 限速 | |
| mac_qos | MAC 限速 | |
| dprotos | 自定义协议 | |
| dprotos_l7 | 高级自定义协议 | |

### 6. AC管理 ac
| func | 说明 | 实证 |
|---|---|---|
| ac_status | AP 统计 | |
| ac_server | AP 配置 | |
| ac_group | AP 分组 | |
| ac_upgrade | AP 升级 | |
| wls_black | 无线黑名单 | |
| wls_mvlan | 无线终端 VLAN | |
| ac_online_clt | 在线用户 | |
| ac_net_optimize | 无线优化 | |

### 7. 认证计费 accounting
| func | 说明 | 实证 |
|---|---|---|
| ppp_online | 在线用户 | |
| webauth | Web 认证服务 | |
| pppoe_server | PPPoE 服务端 | |
| pptp_server | PPTP 服务端 | |
| l2tp_server | L2TP 服务端 | |
| openvpn-server | OpenVPN 服务端 | |
| ike_server | IKEv2 服务端 | |
| ppp_package | 套餐管理 | |
| pppuser | 账号管理 | |
| ppp_passwd | 自助密码 | |
| ppp_paylog | 报表管理 | |
| coupon | 优惠券管理 | |
| pppoe_proxy | 代理账号/在线账号 | |
| notice_temp | 通知模板 | |
| notice_cycle | 周期通知 | |
| notice_remind | 到期提醒 | |
| notice_expires | 到期通知 | |

### 8. 行为管控 behavior
| func | 说明 | 实证 |
|---|---|---|
| audit | 行为审计设置 | |
| record_free_setting | 审计豁免 | |
| audit_url_log | URL 审计日志 | |
| audit_im_log | IM 审计日志 | |
| audit_terminal_log | 终端审计日志 | |
| audit_terminal_stat | 终端审计统计 | |
| mac_comment | MAC 备注 | |
| acl_mac | MAC 访问控制 | |
| url_black | 网页黑名单 | |
| domain_blacklist | 域名黑名单 | |
| domain_group | 域名分组 | |
| url_redirect | 网页重定向 | |
| url_keywords | 网页关键字 | |
| url_replace | 网页替换 | |
| acl_l7 | 七层协议控制 | |
| acl_l2route | 二层访问控制 | |
| acl_qq | QQ 准入控制 | |

### 9. 安全设置 security
| func | 说明 | 实证 |
|---|---|---|
| acl | 访问控制 | |
| arp | ARP 表 | |
| conn_limit | 连接数限制 | |
| advanced | 安全高级设置 | |
| firewall | 防火墙 | |

### 10. 高级应用 advanced
| func | 说明 | 实证 |
|---|---|---|
| ik_web_cache | 网页缓存 | |
| ddns | 动态域名 | |
| nat_ddns | NAT DDNS | |
| netsnmp | SNMP 服务 | |
| netsnmpc | SNMP 客户端 | |
| port_mirror | 端口镜像 | |
| qemu | 虚拟机 | |
| plugins | 插件 | |
| wakeup | 网络唤醒 | |
| ftp_server | FTP 服务 | |
| smbd | SMB 服务 | |
| http_server | HTTP 服务 | |
| udp_proxy | UDP 代理 | |

### 11. 应用工具 tools
| func | 说明 | 实证 |
|---|---|---|
| Ping | 在线 Ping | |
| Traceroute | 路由追踪 | |
| tcpdump | 抓包工具 | |
| iperf | 内网测速 | |
| subnet | 子网计算 | |
| speedtest | 外网测速 | |
| iksyscheck | 系统诊断 | |
| watchdog | 看门狗 | |

### 12. 日志中心 logcenter
| func | 说明 | 实证 |
|---|---|---|
| syslog-pppauth | PPP 拨号日志 | |
| syslog-arp | ARP 日志 | |
| syslog-apaction | AP 操作日志 | |
| syslog-dhcpd | DHCP 日志 | |
| syslog-ddns | DDNS 日志 | |
| syslog-wanpppoe | WAN PPPoE 日志 | |
| syslog-notice | 通知日志 | |
| syslog-sysevent | 系统事件 | |
| syslog-webadmin | 管理日志 | |
| ikmessages | 系统消息 | ✅ `show TYPE=total, FILTER1="read,==,0"`；`start {}` |
| rt_collect | 实时采集 | ✅ `show TYPE=warning_notify` |

## 已验证抓包样本（写类型时的字段基准）

### homepage:show TYPE=sysstat,ac_status（系统概况）
```json
{"sysstat":{"cpu":["6.98%","12.00%"],"cputemp":[39],"freq":["1281","799"],
 "gwid":"d072...","hostname":"iKuai","link_status":0,
 "memory":{"total":47914500,"available":44857584,"free":44940664,
           "cached":127156,"buffers":267852,"used":"6%"},
 "online_user":{"count":170,"count_2g":0,"count_5g":0,
                "count_wired":170,"count_wireless":0},
 "stream":{"connect_num":10229,"upload":22910823,"download":25402421,
           "total_up":37472227088584,"total_down":40669280008537},
 "uptime":1629218,
 "verinfo":{"modelname":"","verstring":"3.7.24 x64 Build202608141210",
            "version":"3.7.24","build_date":202608141210,"arch":"x86",
            "sysbit":"x64","verflags":"","is_enterprise":0,
            "support_i18n":0,"support_dingtalk":1,"support_lcd":0,
            "bootguide":"hd"}},
 "ac_status":{"ap_count":0,"ap_online":0}}
```

### dnat:show / dnat:add（端口映射）
```json
// show param: {"TYPE":"total,data","limit":"0,20","ORDER_BY":"","ORDER":""}
{"Result":30000,"ErrMsg":"Success","Data":{"total":23,"data":[
 {"wan_port":"7422,30501-45000","lan_port":"22,30501-45000",
  "lan_addr_int":2887123202,"id":1,"enabled":"yes","comment":"出租300",
  "interface":"adsl07","src_addr":"","lan_addr":"172.22.1.2",
  "protocol":"tcp+udp"}]}}
// add param:
{"enabled":"yes","comment":"test","interface":"adsl01","lan_addr":"172.22.2.1",
 "protocol":"tcp","wan_port":"22112","lan_port":"22","src_addr":""}
// add 响应: {"Result":30000,"ErrMsg":"Success","RowId":29}
```

### monitor_lanip:show（内网终端）
```json
{"TYPE":"data,total","ORDER_BY":"ip_addr_int","orderType":"IP","limit":"0,20",
 "ORDER":"","FINDS":"ip_addr,mac,comment,username,hostname","KEYWORDS":"172.22"}
// 行字段: signal,hostname,apmac,dtalk_name,client_model,client_typeid,
// comment,uptime,ip_addr_int,connect_num,auth_type,client_type,client_device,
// apname,id,webid,total_up,ip_addr,download,upload,mac,link_addr,bssid,upra...
```

### monitor_system:show（系统曲线）
```json
{"TYPE":"cpu,memory,disk_space_used,on_terminal,conn_num,cputemp1,cputemp2",
 "datetype":"hour","math":"avg"}
// Data.cpu: [{"cpu":6,"timestamp":1794319920}, ...]
```

### 其余已验证动作
| 调用 | param |
|---|---|
| monitor_lanipv6:show | `{"TYPE":"interface"}` → `{"interface":[["all"],["lan1","备注"],...]}` |
| monitor_app_flow:show | `{"TYPE":"app_history","interface":"all","datetype":"day","time_range":"all"}` |
| monitor_l7qos:show | `{"TYPE":"data,interface","interface":""}` |
| cflow:show | `{}` → `{"data":{"today":{...},"week":{...},"yesterday":{...}},"clean_time":0}` |
| ikmessages:show | `{"TYPE":"total","FILTER1":"read,==,0"}` |
| ikmessages:start | `{}` |
| rt_collect:show | `{"TYPE":"warning_notify"}` |
| webuser:show | `{"TYPE":"mod_passwd","username":"..."}` |
| upgrade:show | `{"TYPE":"data,fileinfo"}` |
| register:show | `{"TYPE":"data,gwid"}` |
| dingtalk:show | `{"TYPE":"bind_status"}` |
| basic:set_language | `{"language":1}` |
| auth_code:show | `{}`（注意非法 JSON 怪癖） |
