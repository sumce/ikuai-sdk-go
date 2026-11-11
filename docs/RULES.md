# iKuai SDK 开发契约（Agent 必读，违反即返工）

> **本文档是本项目所有开发工作（人类或 AI Agent）的唯一标准。**
> 开始写任何代码之前，必须完整读完本文档和 [API_CATALOG.md](API_CATALOG.md)。
> 与本文冲突的任何口头指示，以本文为准；确需变更标准时，先修改本文再写代码。

---

## 0. Agent 工作流程（必须按顺序执行）

1. **读规矩**：完整阅读本文档 + [API_CATALOG.md](API_CATALOG.md)。
2. **读架构**：阅读模块根目录 `doc.go` 与所负责包的 `doc.go`。
3. **领任务**：只实现分配给自己的**一个**包，不碰其他任何文件（核心包/其他业务包/文档一律只读）。
4. **开发**：按本文规范实现（见 §3~§8）。
5. **自检**：依次通过 `go build ./...`、`go vet ./...`、`go test ./...`（离线），再跑真机只读测试（见 §9）。
6. **汇报**：按 §11 格式汇报，不改就别交。

---

## 1. 模块布局与依赖规则

```
github.com/sumce/ikuai-sdk-go
├── （根）核心包 ikuai        client/auth/envelope/errors/option/list/crud/json
├── overview/      系统概况   1 功能
├── monitoring/    状态监控   10 功能
├── system/        系统设置   16 功能
├── network/       网络设置   35 功能
├── flowcontrol/   流控分流   10 功能
├── ac/            AC管理     8 功能
├── accounting/    认证计费   16 功能
├── behavior/      行为管控   17 功能
├── security/      安全设置   5 功能
├── advanced/      高级应用   13 功能
├── tools/         应用工具   8 功能
├── logcenter/     日志中心   11 功能
├── sdk/           门面聚合包（最后统一实现，Agent 不得触碰）
├── internal/testserver/   离线假路由器（只读使用，不得修改）
├── internal/livetest/     真机测试辅助（只读使用，不得修改）
├── examples/              可运行示例（仅集成阶段填写）
└── docs/                  本契约文档
```

**依赖方向（严格单向，禁止任何反向或横向 import）：**

```
业务包(overview...logcenter) ──> 核心包 ikuai
sdk/ ──> 核心包 + 全部业务包
internal/* ──> 核心包
```

- 禁止 import 其他业务包（`network` 不许 import `system`）。
- 禁止 import 任何第三方库（本项目零依赖，只用标准库）。
- 禁止修改核心包与 `internal/`、`docs/`、`go.mod`。

---

## 2. 核心包 API（唯一允许使用的核心面）

业务包只允许依赖以下核心 API；需要新能力时在汇报中提出，由集成方扩展核心。

```go
// 构造与会话
c, err := ikuai.NewClient(baseURL, username, password, opts...)
c.Login(ctx) / c.Logout(ctx) / c.KeepAlive(ctx) / c.LoggedIn() / c.BaseURL()

// 通用调用（一切功能的基础通道）
resp, err := c.Call(ctx, funcName, action string, param any)   // 返回 *ikuai.Response
err := c.CallData(ctx, funcName, action string, param, out any) // Data 解码进 out

// 响应封装
resp.OK()                      // Result==30000
resp.RowId                     // add 动作返回的行 ID
resp.Count()                   // ("total", true)
resp.Data                      // json.RawMessage

// 分页/过滤参数（show 的标准形态）
opts := ikuai.DefaultListOptions()          // TYPE=total,data, limit=0,20
opts.BuildParam()                           // → map[string]any（业务包统一用它构造 show 参数）

// 分页解码
page, err := ikuai.DecodePage[DNATRule](resp.Data)   // 强类型 {"total","data":[...]} 或裸数组
raw, err := ikuai.DecodeRawPage(resp.Data)           // map 形态

// 错误
var ae *ikuai.APIError; ikuai.AsAPIError(err)
ae.Code / ae.ErrMsg / ae.IsBadCredentials() / ae.IsAuthError()

// 兜底通用 CRUD（任何 func_name 可用）
crud := c.CRUD("dnat")
crud.Show / crud.List / crud.Count / crud.Add / crud.Edit / crud.Delete / crud.DeleteAll / crud.Set
```

**会话管理是核心包的职责**：业务包永远不调用 `Login`，不读 Cookie；核心层遇
`Result=10014` 自动重登并重试一次。

---

## 3. 包内文件组织规范

每个业务包统一为以下结构（**一个功能 = 一个实现文件**，禁止巨型混合文件）：

```
network/
├── doc.go              只读，不得修改
├── service.go          Service 结构 + New(*ikuai.Client) + 子服务聚合
├── dnat.go             dnat：类型 + 方法
├── dnat_test.go        dnat：离线测试（表驱动）
├── dnat_live_test.go   dnat：真机测试（env 门控）
├── types_dhcp.go       dhcp 族共用的类型定义（按族归并，避免碎文件）
├── dhcp.go             dhcp_server/dhcp_static/dhcp_lease/dhcp_acl_mac
└── ...
```

- 每包**必须**提供唯一入口：

```go
// Service 是网络设置大类的服务入口。
type Service struct {
    c *ikuai.Client

    DNAT    *DNATService    // dnat     端口映射
    DHCP    *DHCPService    // dhcp_*   DHCP 族
    ...
}

// New 构造网络设置服务。
func New(c *ikuai.Client) *Service { ... }
```

- 相邻的小功能（如 dhcp_server/dhcp_static）可合并进一个文件与一个子
  Service，但**类型必须各自定义**，方法粒度保持一功能一方法族。

---

## 4. 命名规范

| 对象 | 规则 | 示例 |
|---|---|---|
| 包名 | 固定，见 §1 | `network` |
| 包入口 | `Service` + `New(c *ikuai.Client) *Service` | `network.New(c)` |
| 子服务 | `<功能驼峰>Service` | `DNATService`、`DHCPService` |
| 方法名 | 动词在前、不加 func_name 前缀（挂在子服务上） | `DNAT.List/Add/Edit/Delete` |
| 无子服务的独立功能 | 直接挂在 `Service` 上，功能驼峰开头 | `svc.Reboot(ctx)` |
| 类型 | 固件语义 + 名词，不加 Type 后缀 | `DNATRule`、`DHCPLease` |
| 枚举常量 | 所属类型语义前缀，值与固件一致 | `EnabledYes = "yes"` |
| 文件名 | 小写下划线 = 功能名 | `dnat.go`、`dhcp.go` |

- Go 官方命名法（缩写全大写：`URL`、`ID`、`API`；不用匈牙利前缀）。
- 函数返回值命名统一：`(*Page[T], error)`、`(rowID int64, error)`、`error`。

---

## 5. 类型与 JSON 约定（最重要的一节）

1. **json tag 逐字段对齐固件**（以 [API_CATALOG.md](API_CATALOG.md) 的抓包样本为准）：
   固件字段 `lan_addr` → `json:"lan_addr"`，**不改名、不合并、不拆分**。
2. 整数一律 `int64`；浮点一律 `float64`。
3. **固件用字符串表示的值就保留 string**（如 `"6%"`、`"yes"/"no"`、
   端口段 `"1000-2000,3000"`），另提供解析辅助（见下），**禁止**擅自转
   数组/数值再编码。
4. 固件字段缺失/为 null 的场合：用指针（`*string`、`*int64`）或
   `json.RawMessage`，禁止用默认值猜。
5. 允许业务包给类型加**只读辅助方法**（不参与 JSON）：

```go
// WANPorts 解析固件的端口串 "7422,30501-45000" 为端口段列表。
func (r DNATRule) WANPorts() []PortRange { ... }

// IsEnabled 固件 enabled 字段为 "yes"/"no"。
func (r DNATRule) IsEnabled() bool { return r.Enabled == EnabledYes }
```

6. 布尔/枚举语义常量必须定义且值与固件完全一致：

```go
const (
    EnabledYes = "yes"
    EnabledNo  = "no"
)
```

---

## 6. 方法实现规范（标准形态）

show/add/edit/del 的**标准方法形态**（照此模板，禁止另起炉灶）：

```go
// List 获取端口映射列表。
// 对应 /Action/call: dnat.show（TYPE=total,data, limit="offset,rows"）。
func (s *DNATService) List(ctx context.Context, opts *ikuai.ListOptions) (*ikuai.Page[DNATRule], error) {
    if opts == nil {
        opts = ikuai.DefaultListOptions()
    }
    resp, err := s.c.Call(ctx, "dnat", "show", opts.BuildParam())
    if err != nil {
        return nil, err
    }
    return ikuai.DecodePage[DNATRule](resp.Data)
}

// Add 新增端口映射，返回固件分配的行 ID。
// 对应 /Action/call: dnat.add。
func (s *DNATService) Add(ctx context.Context, rule DNATRule) (int64, error) {
    resp, err := s.c.Call(ctx, "dnat", "add", rule)
    if err != nil {
        return 0, err
    }
    return resp.RowId, nil
}

// Edit 按 ID 修改端口映射。
// 对应 /Action/call: dnat.edit（param 自动合并 id）。
func (s *DNATService) Edit(ctx context.Context, rule DNATRule) error {
    _, err := s.c.Call(ctx, "dnat", "edit", rule)
    return err
}

// Delete 删除端口映射，支持单个或多个 ID（内部自动选 {"id"} / {"ids"}）。
// 对应 /Action/call: dnat.del。
func (s *DNATService) Delete(ctx context.Context, ids ...int64) error {
    _, err := s.c.Call(ctx, "dnat", "del", ikuai.IDParam(ids))
    return err
}
```

补充规则：

- 入参 `opts *ListOptions` 允许 nil（内部补默认值）。
- 带 TYPE 子选择器的功能（如 homepage/monitor_*），TYPE 定义为包级常量，
  方法签名显式化，不暴露裸字符串：

```go
const (
    HomepageTypeSysStat  = "sysstat"
    HomepageTypeACStatus = "ac_status"
)
func (s *Service) SystemOverview(ctx context.Context) (*SystemOverview, error)
```

- 固件特有 action（如 `basic:set_language`、`ikmessages:start`）实现为
  语义化方法，doc 注释注明原始 action。
- 需要"任意 TYPE 组合"逃生口时，每个大类 Service 提供
  `Raw(ctx, action, param)` 一个方法足矣，不许泛滥。

---

## 7. 注释规范（中文）

- 所有导出对象必须有 doc 注释：**一句话功能 + 对应 func/action**。
- 格式模板：

```go
// SystemOverview 获取系统概况仪表盘数据（CPU/内存/在线用户/流量/版本）。
// 对应 /Action/call: homepage.show（TYPE=sysstat,ac_status），
// 即 Web 控制台「系统概况」页面。
```

- 非显然的业务语义要写（如固件字段的单位：字节/秒/百分比）。
- 禁止把 TODO/调试注释留在交付代码里。

---

## 8. 错误处理

- 业务包**不吞错、不重试、不 log**：把错误原样向上传（重试/重登是核心包职责）。
- 包装错误用 `%w`，保持 `errors.Is/As` 与 `ikuai.AsAPIError` 可用。
- 参数校验错误返回 `fmt.Errorf("network: xxx: %w", ikuai.ErrInvalidParam)`
  形态；核心包提供 `ikuai.ErrInvalidParam` 哨兵。
- **禁止** `panic`、`os.Exit`、全局可变状态。

---

## 9. 测试规范（双轨制，强制）

### 9.1 离线单测（`<feature>_test.go`，默认 `go test ./...` 必须全绿）

- 用 `internal/testserver`（假路由器），**表驱动**。
- 每个方法至少覆盖：① 成功路径（含参数断言）② 固件错误路径
  （Result!=30000 → `*ikuai.APIError`）。
- 参数断言示例：

```go
ts := testserver.New()
defer ts.Close()
ts.SetResponse("dnat", "show", map[string]any{"total": 1, "data": []map[string]any{{...}}})
c := ts.Client()
_, _ = network.New(c).DNAT.List(ctx, nil) // opts=nil → 默认分页
rec := ts.CallsFor("dnat")[0]
if rec.Action != "show" { t.Fatalf(...) }
// 解析 rec.Param 断言 TYPE/limit 等
```

### 9.2 真机联调测试（`<feature>_live_test.go`，环境变量门控）

- 通过 `internal/livetest.Client(t)` 获取客户端；环境变量未配置时**自动
  t.Skip**（离线 CI 不受影响）：

```
IKUAI_BASE_URL=http://192.168.1.1
IKUAI_USERNAME=...  IKUAI_PASSWORD=...
```

- **只读测试**（show 类）直接跑。
- **写测试必须**：创建带唯一标记的临时记录（comment 前缀
  `ikuai-sdk-test-`）→ 断言 → `defer` 删除清理；失败也要清理。
- 真机测试**禁止**：改管理员密码、重启路由、删除既有业务数据、
  任何不可逆操作。

---

## 10. 禁止事项清单

- ❌ 修改核心包 / internal/ / docs/ / go.mod / 其他业务包
- ❌ 引入第三方依赖
- ❌ 跨业务包 import
- ❌ 硬编码凭据（凭据只能来自环境变量）
- ❌ `panic` / 全局可变状态 / `init()` 里做 IO
- ❌ 在业务包里实现会话管理、重试、日志
- ❌ 编造未在抓包/固件中见过的字段语义（不确定就在类型注释标注
  `// 字段语义待真机验证`，并在汇报中列出）
- ❌ 留 TODO、调试代码、被注释掉的代码

---

## 11. 交付自检清单与汇报格式

**自检清单（全部通过才算完成）：**

- [ ] `go build ./...` 通过
- [ ] `go vet ./...` 零告警
- [ ] `go test ./...`（离线）全绿
- [ ] 真机只读测试通过（配置环境变量后）
- [ ] API_CATALOG 中该包每个 func 都有：类型定义 + 标准方法族 + 测试
- [ ] 无禁止事项触碰
- [ ] doc 注释齐全（中文，含 func/action 对应说明）

**汇报格式：**

```
包: <包名>
实现 func: <列表>
文件: <新增文件列表>
自检: build/vet/test 结果
真机验证: 通过的 func 清单 / 未验证原因
待集成方决策: <不确定的字段语义、需要核心包新增的能力>
```
