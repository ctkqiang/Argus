# Argus 项目 Go 代码规范

> 版本：v1.0.0 | 维护：Argus 核心开发团队
>
> 本规范是所有贡献者向 `github.com/ctkqiang/argus` 仓库提交代码的**强制性要求**。
> Code Review 阶段任何违反本规范的变更都会被打回，弟弟不许偷懒哦。

---

## 0. 快速索引

| 章节 | 主题 | 关键要求 |
|---|---|---|
| 1 | 项目布局 | 标准 Go 项目布局，`internal/`、`pkg/`、`cmd/` 分离 |
| 2 | 代码格式 | `gofmt` + `goimports`，行宽 ≤ 100，`go vet` 必过 |
| 3 | 命名约定 | 驼峰法，首字母大小写明确控制可见性，缩写全大写 |
| 4 | 注释与文档 | 导出符号必注释，包注释 `Package xxx ...`，禁止无意义注释 |
| 5 | 错误处理 | 错误必处理，自定义 error 类型，禁止 `_` 忽略关键错误 |
| 6 | 并发与同步 | Context 贯穿传递，goroutine 必须可退出，禁止共享可变状态 |
| 7 | 包与模块设计 | 单一职责，最小可见性，无循环依赖，明确依赖注入 |
| 8 | 日志规范 | 结构化日志 `slog`，分级输出，禁止 `fmt.Println` |
| 9 | 测试 | 表驱动测试，覆盖率基准 ≥ 70%，外部依赖用接口 mock |
| 10 | 性能与安全 | 预分配 slice/map，defer 合理使用，零依赖 secrets 注入 |
| 附录 A | 工具链配置 | Makefile 目标、静态检查工具清单 |
| 附录 B | 典型参考样例 | `internal/utilities/logger.go` 完整分析 |

---

## 1. 项目布局（Project Layout）

遵循标准 Go 社区布局约定：

```
argus/
├── api/           # OpenAPI / protobuf / CRD API 定义（对外契约）
├── cmd/           # 应用入口，每个子目录对应一个二进制（如 cmd/argus-gateway）
├── internal/      # 仅限本项目私有代码，禁止外部 import（go build 强制）
│   ├── utilities/ # 通用小工具包（如 logger、hash、retry）
│   ├── gateway/   # 网关业务逻辑
│   └── controller/# 控制器业务逻辑
├── pkg/           # 可被外部项目引用的公共库（本项目暂不使用，谨慎引入）
├── configs/       # 配置样例 / 默认配置
├── scripts/       # 构建、lint、部署脚本
├── test/          # 集成测试、E2E 测试、测试数据
├── web/           # 前端静态资源
├── k8s/           # Kubernetes 部署清单
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 1.1 强制规则

1. `cmd/<name>/main.go` 必须轻量：仅负责**装配**依赖、读取配置、启动信号监听。真正逻辑下沉到 `internal/`。
2. `internal/` 下每个子包必须保持单一职责，边界清晰。
3. 新增包前必须更新 `.trae/specs/tasks.md`，未经确认不得乱加。

---

## 2. 代码格式（Formatting）

### 2.1 自动格式化（强制）

提交前执行以下命令，任何与 `gofmt` 不一致的代码都会被拒绝：

```bash
gofmt -s -w .
goimports -w .   # go install golang.org/x/tools/cmd/goimports@latest
```

### 2.2 基本规则

| 项 | 约定 |
|---|---|
| 缩进 | Tab（`gofmt` 默认），宽度 4 |
| 行宽 | **目标 80 列，硬上限 100 列**，超了要换行 |
| 空行 | 函数间 1 行，逻辑区块之间可用 1 行分隔，禁止连续 2+ 空行 |
| 导入分组 | 按 `标准库 → 第三方 → 项目内部` 三组分隔，`goimports` 自动处理 |
| 大括号 | `{` 不换行，`else` 与 `}` 同行（`gofmt` 强制） |

### 2.3 导入示例

```go
import (
    "context"           // 标准库
    "log/slog"
    "os"

    "google.golang.org/grpc"            // 第三方

    "github.com/ctkqiang/argus/internal/utilities"  // 内部
)
```

### 2.4 必须通过的静态检查

```bash
go build ./...
go vet ./...
go test ./...  # 能过的 test 必须全绿
```

建议引入 `staticcheck`（`go install honnef.co/go/tools/cmd/staticcheck@latest`）。

---

## 3. 命名约定（Naming Conventions）

### 3.1 通用原则

1. **驼峰法（CamelCase）**：Go 标准，不使用下划线分隔词。
2. **首字母大小写 = 可见性**：
   - 大写开头：导出（包外可见）
   - 小写开头：私有（包内可见）
3. **缩写全大写**：`ID`、`URL`、`HTTP`、`JSON`、`CRD`、`API`、`gRPC`。
   - 正确：`HTTPClient`、`UserID`、`JSONHandler`
   - 错误：`HttpClient`、`UserId`、`JsonHandler`
4. **长度适中**：长作用域用明确长名，短作用域可用短名。
   - 循环/索引：`i`, `j`, `k`
   - 参数/局部：`msg`, `cfg`, `err`, `ctx`
   - 包级/导出：必须完整表意，禁止 `msg` 这种歧义名

### 3.2 包名（Package Name）

- 全小写，单词直接拼接，**不使用下划线、不使用 mixedCaps**。
- 避免与标准库冲突（如不要叫 `util`，可用 `utilities`）。
- 见名知义：`gateway`、`controller`、`detector`、`utilities`。

错误：`argus_utils`、`HTTPTools`、`pkg_common`

### 3.3 常量 / 变量名

- 普通常量（非枚举组）：`const DefaultPort = 8080`
- 枚举组：类型名 + 常量前缀保持一致

```go
type LogLevel string

const (
    Debug LogLevel = "debug"   // 枚举值与类型名同语义域，前缀自然
    Info  LogLevel = "info"
    Warn  LogLevel = "warn"
    Error LogLevel = "error"
    Fatal LogLevel = "fatal"
)
```

### 3.4 函数 / 方法名

- 动词开头：`LoadConfig`、`BuildHandler`、`SetLevel`。
- getter 省略 `Get` 前缀（Go 惯例）：用 `Level()` 而非 `GetLevel()`。
- setter 用 `Set` 前缀：`SetLevel(level LogLevel)`。
- 返回 error 的函数，函数名可暗示动作，**error 必须作为最后一个返回值**。

### 3.5 接口名

- 单一方法接口用 `-er` 后缀：`Reader`、`Writer`、`Formatter`、`Logger`（虽然 Logger 有多个方法，但语义明确可接受）。
- 多方法接口按功能命名：`SecurityDetector`、`PolicyEvaluator`。

### 3.6 接收者名

- 结构体方法接收者名用结构体首字母或短名（1–3 字母），**统一**，不要一会儿 `self`、`this`、`logger`。

```go
// 参考 [logger.go#L55-L61]
type Logger struct { ... }

func (l *Logger) SetLevel(level LogLevel) { ... }  // l 统一代表 *Logger
func (l *Logger) Log(level LogLevel, msg string, args ...any) { ... }
```

---

## 4. 注释与文档（Comments & Documentation）

### 4.1 基本规则

1. **导出符号必写注释**：每个导出的 `type / const / var / func` 必须有文档注释（`golint` / `staticcheck` 强制）。
2. **注释格式**：`// SymbolName ` 开头，接完整句子，句号结尾。
3. **禁止废话注释**：不能解释代码本身已表达的内容。
4. **注释说明「WHY」，代码说明「HOW」**：存在的约束、绕开的坑、业务背景。

### 4.2 包注释（Package Comment）

每个包**必须**有且仅有一个包注释，放在 `doc.go` 或该包内首文件顶部。

```go
// Package utilities 提供 Argus 项目共享的基础设施组件，包括结构化日志、错误包装、重试工具等。
// 该包仅允许依赖 Go 标准库，禁止引入任何外部三方库或其他 internal 子包，保持底层纯净。
package utilities
```

### 4.3 函数 / 方法注释

```go
// NewLogger 根据可变 Option 构造新的 Logger 实例。
// 默认级别 Info、输出 os.Stderr、格式 FmtText。
// Debug 级别下会自动开启源码定位（AddSource=true）。
func NewLogger(opts ...Option) *Logger { ... }
```

### 4.4 类型注释

```go
// LogLevel 定义日志级别枚举。字符串形式便于通过配置文件 / env 直接解析。
// 级别严格递增：Debug < Info < Warn < Error < Fatal。
type LogLevel string
```

### 4.5 禁止的注释

```go
// 下面是 Debug 常量     ← 废话
const Debug ... = ...

// i + 1               ← 代码本身能说明
x := i + 1
```

---

## 5. 错误处理（Error Handling）

### 5.1 黄金法则

1. **任何可能失败的调用都必须检查错误**，除了文档明确声明永不失败的函数（如 `fmt.Println` 写到 stdout 的错误通常可忽略）。
2. **禁止裸 `_` 忽略错误**，除非你在注释中证明为什么可以安全忽略。
3. **错误必须返回 / 包装后返回，不可静默吞掉**。
4. **error 永远是函数最后一个返回值**。
5. 不要同时返回非零值和非 nil error（除非文档明确约定）。

### 5.2 错误包装（Go 1.13+）

```go
// 正确：用 %w 包装，保留错误链
func LoadConfig(path string) (*Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config %q: %w", path, err)
    }
    ...
}
```

### 5.3 哨兵错误 vs 自定义错误类型

| 场景 | 方式 |
|---|---|
| 调用方需要判断具体原因（如 ErrNotFound、ErrPermissionDenied） | 哨兵错误 `var ErrNotFound = errors.New(...)` |
| 需要携带额外上下文（字段、堆栈、错误码） | 自定义 struct 实现 `error` 接口 |
| 多层调用需要类型断言 | 自定义错误类型 + `errors.As` |

### 5.4 panic 规则

- **除了启动期配置校验失败（无法继续运行），业务代码一律不许 panic**。
- panic 用在不可恢复的程序错误，且必须在启动早期 main 或 init 路径，不能在请求路径里。
- `Fatal` 级别的日志（调用 `os.Exit(1)`）等同 panic，仅限 main 启动阶段。

---

## 6. 并发与同步（Concurrency & Synchronization）

### 6.1 Context 优先

1. 所有可能阻塞 / 跨 goroutine / 外部 IO 的函数**第一个参数必须是 `ctx context.Context`**。
2. Context 是取消、超时、元数据贯穿整个调用链的唯一机制，**禁止存到 struct 里**。
3. 日志、RPC、HTTP、DB 操作默认都要带 ctx。

> 例外：封装全局默认 logger 的快捷方法可以不暴露 ctx，但内部要使用 `context.Background()` 明确表达"不关心取消"。

### 6.2 Goroutine 纪律

1. **谁启动谁负责关闭**。每开一个 goroutine 必须有明确退出机制（ctx.Done 或 channel）。
2. 禁止 "泄漏式" `go func(){ for { ... } }()` 无法停止的循环。
3. 优先使用 `errgroup`、`sync.WaitGroup`、`context.WithCancel` 来管理生命周期，避免裸 channel 手写同步。

### 6.3 共享状态

- **通信不要共享内存；共享内存不要通信。**
- 多个 goroutine 共享的可变状态必须用 `sync.Mutex` / `sync.RWMutex` 或原子操作保护。
- 优先使用不可变数据结构。

---

## 7. 包与模块设计（Package & Module Design）

### 7.1 SOLID 在 Go 的映射

| 原则 | Go 实现方式 |
|---|---|
| Single Responsibility | 一个包只做一件事；`internal/gateway` 不要塞控制器逻辑 |
| Open/Closed | 通过接口扩展，而不是改已有结构体 |
| Liskov Substitution | 实现接口时严格遵守接口语义（见 7.3） |
| Interface Segregation | 接口尽量小，3 个方法以内，调用方只依赖自己需要的接口 |
| Dependency Inversion | 依赖接口而非具体实现；通过构造函数注入依赖 |

### 7.2 可见性控制（最小权限）

- 默认全部小写私有，**确有必要才导出**。
- 对外暴露的表面越小，越容易维护：导出 API 必须经过评审。

### 7.3 接口约定

1. **接口定义在使用方一侧**（Go 惯例），不要实现方先写好接口。
2. 构造函数返回具体类型（`*Logger`），依赖方用自己定义的接口接收。
3. 避免大型接口，如 `Storage` 分成 `Reader` + `Writer` + `Lister`。

### 7.4 依赖注入

```go
// 正确：构造函数注入依赖，可测试、可替换
func NewGateway(logger *Logger, detector SecurityDetector) *Gateway {
    return &Gateway{logger: logger, detector: detector}
}
```

禁止在包内 `init()` 中偷偷初始化全局单例，除非是类似 `defaultLogger` 这种轻量、无副作用、纯标准库依赖、且提供显式 setter 的便捷实例。

### 7.5 循环依赖

- `internal/a → internal/b → internal/a` **绝对禁止**。
- 设计时画依赖图；如果出现环，说明需要抽第三个 `c` 包放共同抽象。

---

## 8. 日志规范（Logging Standards）

本项目统一使用 Go 标准库 `log/slog`，**禁止使用**：

- `fmt.Print*`（调试代码不准进仓库）
- 旧标准库 `log.*`（非结构化）
- 任何第三方日志库（Zap、Logrus 等均不允许，避免依赖膨胀）

### 8.1 日志包参考实现

完整代码见 [logger.go](file:///Users/johnmelodyme/Documents/ctkqiang/argus/internal/utilities/logger.go)。

### 8.2 日志级别语义

| 级别 | 使用场景 | 示例 |
|---|---|---|
| `Debug` | 开发调试细节；生产默认关闭；会附带源码定位 | "entering processing loop", "raw payload: %s" |
| `Info` | 正常业务流程里程碑，高频但不噪音 | "server started on :8080", "request processed" |
| `Warn` | 可恢复异常、非预期但不中断流程 | "retrying 2/5: connection refused" |
| `Error` | 请求失败、单操作失败，应用仍继续 | "failed to persist record: %w" |
| `Fatal` | 应用无法启动或必须立刻终止（启动阶段） | "invalid bootstrap config, cannot continue" |

### 8.3 结构化字段（Key-Value）

```go
l.LogInfo("request processed",
    "method", r.Method,
    "path", r.URL.Path,
    "status", status,
    "latency_ms", latencyMs,
)
```

- 键名：`snake_case`，与行业通用约定一致。
- 禁止用 Sprintf 拼接到 msg 字符串中，必须放 KV 字段，便于日志系统检索。

### 8.4 输出格式

| 环境 | 推荐格式 |
|---|---|
| 本地开发 | `FmtText`（人类可读） |
| 生产 / 容器 | `FmtJSON`（便于 ELK / Loki / Datadog 解析） |

### 8.5 禁止

1. 禁止日志包含 secrets：token、密码、完整 JWT、私有 IP 列表等。打之前要脱敏。
2. 禁止循环里无节流打日志（日志洪水）。
3. 禁止把 `error` 对象格式化成字符串后再丢掉错误链——用 `slog.String("err", err.Error())` 或直接作为 KV value。

---

## 9. 测试（Testing）

### 9.1 基本要求

- 单元测试文件名：`xxx_test.go`，与被测文件同目录。
- 测试函数：`TestXxx(t *testing.T)`，子测试用 `t.Run`。
- 基准测试：`BenchmarkXxx(b *testing.B)`，热点路径必须写。
- 覆盖率：**包级基准 ≥ 70%，核心安全检测逻辑 ≥ 90%**。

### 9.2 表驱动测试（强烈推荐，作为默认模式）

```go
func TestLevelToSlog(t *testing.T) {
    tests := []struct {
        name  string
        input LogLevel
        want  slog.Level
    }{
        {"debug maps correctly", Debug, slog.LevelDebug},
        {"info maps correctly", Info, slog.LevelInfo},
        {"unknown defaults to info", LogLevel("what"), slog.LevelInfo},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := levelToSlog(tt.input); got != tt.want {
                t.Errorf("levelToSlog(%v) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}
```

### 9.3 依赖 Mock

- 使用接口 + 手写 mock 或 `gomock`，不要 `monkey patch`。
- 禁止真的起 DB / 真的调外部 API 跑单元测试；集成测试放 `test/` 目录。

### 9.4 断言风格

- 标准库 `testing` 优先；团队引入 `testify/assert` 前保持一致（目前本项目仅标准库）。
- 失败必须打印差异和上下文，禁止 `if cond { t.Fatal("fail") }` 这种无信息断言。

---

## 10. 性能与安全（Performance & Security）

### 10.1 性能要点

1. **预分配**：`make([]T, 0, n)`、`make(map[K]V, n)`，知道大小时一定要给容量。
2. **defer 合理使用**：热路径中 defer 有开销（Go 1.14+ 已经很好），但正确性 > 性能；在 `Unlock`、`Close` 场景必须 defer。
3. **字符串拼接**：循环拼接用 `strings.Builder`，不要 `+=`。
4. **反射**：仅在反序列化 / 框架层使用，业务代码禁止裸 `reflect`。

### 10.2 安全要点（Argus 是安全产品，零容忍）

1. **Secrets 永不进代码 / 仓库**：所有密钥、token、密码必须从 env / Secret Manager / k8s Secret 注入。
2. **不要 commit 任何真实 key**，哪怕是 "测试密钥"。
3. **输入校验**：所有外部输入（HTTP body、grpc 消息、配置、YAML）必须：
   - 长度限制
   - 格式校验（regex / schema）
   - 拒绝包含 `../`、NUL 字节、控制字符的路径
4. **SQL/命令注入零容忍**：参数化查询；`exec.Command` 参数化，不拼字符串；禁止 `text/template` 拼代码。
5. **TLS 优先**：出向 HTTP 默认 HTTPS；自签 CA 场景显式加载，禁用 `InsecureSkipVerify`。
6. **敏感数据日志脱敏**：见 8.5。

---

## 附录 A：工具链与 Makefile 目标

### A.1 必装工具

```bash
# go.mod Go 版本即 1.22；本项目不引入任何第三方日志库
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### A.2 Makefile 必含目标

| 目标 | 说明 |
|---|---|
| `make fmt` | `gofmt -s -w . && goimports -w .` |
| `make lint` | `go vet ./... && staticcheck ./...` |
| `make test` | `go test -race -cover ./...` |
| `make build` | 产出 cmd 下的所有二进制 |
| `make all` | fmt → lint → test → build |

PR CI 中 `make lint` 和 `make test` 失败**直接拒绝合并**，弟弟你要是提了全红 PR，姐姐告诉二姐你昨天偷偷熬夜看番了哦。

---

## 附录 B：典型参考样例 — `internal/utilities/logger.go` 分析

下面以 [logger.go](file:///Users/johnmelodyme/Documents/ctkqiang/argus/internal/utilities/logger.go) 为正面教材，对照规范做一次逐段剖析，帮助大家形成体感。

### B.1 命名一致性

```go
// [logger.go#L9-L17] 枚举组命名：类型 LogLevel，常量 Debug/Info/.../Fatal 语义一致。
type LogLevel string
const (
    Debug LogLevel = "debug"
    Info  LogLevel = "info"
    ...
)

// [logger.go#L19-L24] Format 类型：类型 + 常量前缀 FmtText/FmtJSON，避免冲突
type Format string
const (
    FmtText Format = "text"
    FmtJSON Format = "json"
)
```

规范对应：§3.3 常量 / 变量名。

### B.2 接收者名统一

```go
// [logger.go#L55-L61] 接收者统一为 l，不用 self / this / logger
type Logger struct { ... }

func (l *Logger) buildInner()       { ... }
func (l *Logger) SetLevel(level LogLevel) { ... }
func (l *Logger) Log(level LogLevel, msg string, args ...any) { ... }
```

规范对应：§3.6 接收者名。

### B.3 Functional Options 模式（可扩展构造）

```go
// [logger.go#L63-L75] Option 函数式选项 + NewLogger(opts ...Option)
type Option func(*Logger)

func WithLevel(l LogLevel) Option    { return func(logger *Logger) { logger.level = l } }
func WithOutput(w io.Writer) Option  { return func(logger *Logger) { logger.out = w } }
func WithFormat(f Format) Option     { return func(logger *Logger) { logger.fmt = f } }

func NewLogger(opts ...Option) *Logger {
    l := &Logger{ level: Info, out: os.Stderr, fmt: FmtText, levelVar: new(slog.LevelVar) }
    for _, opt := range opts { opt(l) }
    l.buildInner()
    return l
}
```

为什么这样写：
- 构造参数将来扩展（加 `WithTimezone`、`WithHook`）时**不需要改签名**，保持 API 兼容。
- 默认值明确，缺什么补什么。
- 调用方自解释：`NewLogger(WithLevel(Debug), WithFormat(FmtJSON))`。

规范对应：§7.4 依赖注入 / §3.4 函数命名。

### B.4 快捷方法薄包装（不复制逻辑）

```go
// [logger.go#L134-L138] 每个级别一个包装方法，逻辑只在 Log() 一处
func (l *Logger) LogDebug(msg string, args ...any) { l.Log(Debug, msg, args...) }
func (l *Logger) LogInfo(msg string, args ...any)  { l.Log(Info, msg, args...) }
func (l *Logger) LogWarn(msg string, args ...any)  { l.Log(Warn, msg, args...) }
func (l *Logger) LogError(msg string, args ...any) { l.Log(Error, msg, args...) }
func (l *Logger) LogFatal(msg string, args ...any) { l.Log(Fatal, msg, args...) }
```

注意这里方法前缀 `Log*` 是**刻意的**：因为包内已经存在同名单词的 `LogLevel` 枚举常量（`Debug / Info / Warn / Error / Fatal`），Go 禁止同包里常量与函数同名。这是一个很重要的"命名去冲突"技巧：

- **当级别枚举常量与方法同名冲突时**，方法加语义前缀（`LogDebug` / `LogInfo`），而不是改枚举名（因为枚举名是域模型的一部分，更应保持稳定）。
- 如果你将来的枚举 / 方法也碰到同名冲突，**永远优先改"使用方更短"的那一方**，这里方法更短，加前缀成本低。

规范对应：§3.4 函数命名。

### B.5 可见性控制

```go
// [logger.go#L28-L53] 内部辅助函数：levelToSlog / replaceFatalLevel 均小写私有
func levelToSlog(l LogLevel) slog.Level { ... }
func replaceFatalLevel(_ []string, a slog.Attr) slog.Attr { ... }
```

外部不需要知道 `fatalSlogLevel = slog.LevelError + 4` 这种实现细节，包内私有。规范对应：§7.2 最小权限。

### B.6 全局默认实例（有节制的单例）

```go
// [logger.go#L150-L161] 提供包级便捷函数 + 显式 setter，允许测试中替换
var defaultLogger *Logger

func init() { defaultLogger = NewLogger() }

func Default() *Logger              { return defaultLogger }
func SetLevel(level LogLevel)       { defaultLogger.SetLevel(level) }
func SetOutput(w io.Writer)         { defaultLogger.SetOutput(w) }
func SetFormat(f Format)            { defaultLogger.SetFormat(f) }
func With(args ...any) *Logger      { return defaultLogger.With(args...) }
func Log(level LogLevel, msg string, args ...any) { defaultLogger.Log(level, msg, args...) }
```

这里的关键点：
1. `init()` 仅做无副作用、无外部依赖的轻量初始化。
2. 暴露 `SetOutput` / `SetFormat` / `SetLevel` 让**测试或集成方可以替换**——这是单例可测的关键。
3. 包级函数只做转发，不重复逻辑，维护成本低。

规范对应：§7.4 依赖注入 + §8 日志规范。

### B.7 With 方法（结构化上下文注入）

```go
// [logger.go#L140-L148] 不可变地返回新 Logger，共享同一个 LevelVar（级别动态生效）
func (l *Logger) With(args ...any) *Logger {
    return &Logger{
        inner:    l.inner.With(args...),
        levelVar: l.levelVar,   // 关键：共享 LevelVar，SetLevel 调用时所有派生 Logger 同步
        out:      l.out,
        level:    l.level,
        fmt:      l.fmt,
    }
}
```

典型用法：
```go
reqLogger := utilities.With("trace_id", traceID, "user_id", userID)
reqLogger.LogInfo("processing request")  // 自动带上 trace_id / user_id
```

规范对应：§8.3 结构化字段。

---

## 规范维护

- 本规范是活的文档，有争议或需要补充时：
  1. 先在 PR 评论里提出
  2. 达成一致后提交 MR 修改本文件
  3. 修改必须有 commit message 说明"变更点 + 原因"
- 规范本身是团队共识，不是某一个人的私有规则；大家一起遵守，代码才不会变乱，答应姐姐好吗？

---

（完）
