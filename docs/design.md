# CLI 无痛搬家（CLI Migration）设计文档

> 版本：v0.1.0-alpha  
> 日期：2026-07-08  
> 状态：设计草案 / 待评审

---

## 1. 项目概述

### 1.1 目标

CLI Migration（暂定名 `cli-mig`）是一个帮助用户在不同机器、不同账号甚至不同操作系统之间「无痛迁移命令行工具环境」的开源 CLI 工具。

**核心能力：**

1. **发现（Discover）**：扫描当前系统，识别用户已安装的全部 CLI 工具及其来源。
2. **记录（Record）**：生成一份可序列化、可版本控制、可人工审阅的「环境清单」。
3. **还原（Restore）**：在目标机器上根据清单自动安装对应工具，尽量还原等价环境。
4. **交换（Exchange）**：支持 Windows↔Windows、Linux↔Linux 的同系统一键搬家。

### 1.2 非目标（本期不做，但架构预留）

- 跨操作系统自动转换（如 Windows 上 `scoop` 安装的包 → Linux 上 `apt` 的等价包）。
- GUI 应用、桌面环境、字体、Shell 主题（如 PowerShell 配置文件只记录，不保证完全还原）。
- 商业软件许可证迁移。

### 1.3 使用场景

| 场景 | 示例 |
|------|------|
| 换新电脑 | 用户买了新笔记本，希望把旧电脑上的开发环境一键搬过去。 |
| 重装系统 | 格式化后快速恢复常用 CLI。 |
| 多设备同步 | 个人台式机与笔记本保持一致的命令行环境。 |
| 团队 onboarding | 新员工拿到一份 `env.snapshot.yaml` 即可还原团队推荐环境。 |

---

## 2. 术语表

| 术语 | 含义 |
|------|------|
| **CLI 条目（Entry）** | 一个可执行命令的元数据，如 `node`、`kubectl`、`git`。 |
| **来源（Source）** | CLI 的安装渠道，如 `scoop`、`apt`、`cargo`、`npm`、`manual`（手动下载二进制）。 |
| **清单（Manifest）** | 一份 JSON/YAML 文件，记录某台机器上所有 CLI 条目及其来源、版本、配置。 |
| **Provider** | 包管理器或安装渠道的抽象，负责解析、安装、升级 CLI。 |
| **Detector** | 负责发现当前系统中已安装 CLI 的模块。 |
| **Bundle** | 一次搬迁任务的打包输出，可能包含清单、配置文件、脚本等。 |

---

## 3. 架构设计

### 3.1 设计原则

1. **Provider 抽象**：把「发现」和「安装」统一抽象为 Provider，新增包管理器只需实现接口。
2. **平台解耦**：核心逻辑完全不知道平台细节，平台差异下沉到 `PlatformLayer`。
3. **清单与执行分离**：先生成、审阅清单，再执行还原；清单是人类可读的中间产物。
4. **可回滚**：还原操作尽量原子化，失败时给出清晰的撤销建议。
5. **渐进增强**：先保证同系统可靠，再逐步加入跨系统映射。

### 3.2 总体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI 入口层                            │
│   cli-mig discover  |  cli-mig plan  |  cli-mig apply        │
│   cli-mig export    |  cli-mig import  |  cli-mig diff       │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                      编排引擎（Orchestrator）                 │
│  • 命令解析、配置加载、任务调度、并发控制、回滚协调            │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────▼─────┐ ┌────▼─────┐ ┌───▼────────┐
│   发现引擎   │ │  计划引擎 │ │  执行引擎   │
│  Discover   │ │   Plan   │ │   Apply    │
└───────┬─────┘ └────┬─────┘ └───┬────────┘
        │            │           │
        └────────────┼───────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    Provider 层                               │
│  ScoopProvider | AptProvider | HomebrewProvider | ...        │
│  每个 Provider 实现：detect / resolve / install / uninstall   │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                  平台适配层（Platform Layer）                 │
│   WindowsAdapter  |  LinuxAdapter  |  （预留 macOS）          │
│  负责 PATH、Shell 配置、权限、包管理器可用性检测               │
└─────────────────────────────────────────────────────────────┘
```

### 3.3 核心模块职责

#### 3.3.1 CLI 入口层（`cmd/`）

只负责参数解析、日志输出、错误展示，业务逻辑委托给内部包。

| 命令 | 作用 |
|------|------|
| `discover` | 扫描当前系统，输出清单。 |
| `plan` | 对比当前环境与目标清单，生成执行计划（只读）。 |
| `apply` | 执行计划，安装/升级/删除 CLI。 |
| `export` | 将清单与相关配置文件打包为 `bundle.tar.gz`。 |
| `import` | 解包并校验 Bundle。 |
| `diff` | 对比两台机器或两个清单的差异。 |

#### 3.3.2 发现引擎（`internal/discover/`）

- 枚举 `PATH` 中的可执行文件。
- 调用各 Provider 的 `detect()`，识别每个 CLI 的来源。
- 对未知来源的 CLI 标记为 `manual`。
- 收集版本号（优先调用 `--version`，失败则读取文件元数据）。

**关键策略：**

- 不把「可执行文件存在」直接等同于「由包管理器安装」。优先让各 Provider 根据自己的元数据（如注册表、dpkg 数据库、安装目录）认领。
- 冲突处理：同一个命令被多个 Provider 认领会提示用户，默认选择「更有可能」的那个（例如 `scoop` 优先于 `manual`）。

#### 3.3.3 Provider 层（`internal/providers/`）

每个 Provider 实现统一接口：

```go
type Provider interface {
    Name() string                       // 如 "scoop"
    Available() bool                    // 当前系统是否可用
    Detect(ctx context.Context) ([]Entry, error)   // 发现已安装条目
    Resolve(entry Entry) (InstallPlan, error)      // 将条目解析为安装计划
    Install(ctx context.Context, plan InstallPlan) error
    Uninstall(ctx context.Context, plan InstallPlan) error
}
```

**首期支持的 Provider：**

| Provider | 平台 | 说明 |
|----------|------|------|
| `scoop` | Windows | 主流 Windows CLI 包管理器。 |
| `winget` | Windows | Windows 自带，作为 scoop 的补充。 |
| `apt` | Linux | Debian/Ubuntu 系。 |
| `dnf` / `yum` | Linux | RHEL/CentOS/Fedora 系（首期可只实现检测）。 |
| `pacman` | Linux | Arch 系（首期可只实现检测）。 |
| `homebrew` | Linux | Linuxbrew 用户。 |
| `cargo` | 跨平台 | Rust 工具链。 |
| `npm` / `pnpm` / `yarn` | 跨平台 | 全局安装的 JS CLI。 |
| `pipx` | 跨平台 | Python CLI 工具。 |
| `manual` | 跨平台 | 兜底，记录可执行文件路径与元数据，但不一定能自动还原。 |

#### 3.3.4 平台适配层（`internal/platform/`）

为不同 OS 提供统一接口：

```go
type Platform interface {
    OS() OSType                                    // windows / linux / darwin(预留)
    ExecutableExtensions() []string               // Windows: [.exe, .cmd, .bat, .ps1]
    HomeDir() string
    ListPathEntries() []string
    ShellConfigFiles() []string                    // 如 ~/.bashrc, ~/.zshrc
    IsElevated() bool                              // 是否管理员/root
    QuoteCommand(args []string) string             // 平台安全的命令拼接
}
```

该层是后续跨系统迁移的基石：当需要 Windows→Linux 时，计划引擎会询问 Linux 的 Platform 实现，并把 Provider 的解析结果转换为 Linux 可执行的方案。

### 3.4 清单数据模型（Manifest）

清单使用 YAML 作为默认格式，便于版本控制和人工审阅。

```yaml
version: "1.0"
generated_at: "2026-07-08T18:40:00+08:00"
source:
  os: windows
  arch: amd64
  shell: powershell
  platform_version: "10.0.26200"
entries:
  - name: node
    command: node
    version: "20.12.0"
    provider: scoop
    provider_args:
      bucket: main
      package: nodejs
    aliases: []
    config_refs:
      - type: env
        key: NODE_OPTIONS
        value: "--max-old-space-size=4096"
  - name: kubectl
    command: kubectl
    version: "1.30.0"
    provider: winget
    provider_args:
      package_id: Kubernetes.kubectl
  - name: my-custom-tool
    command: mytool
    version: "unknown"
    provider: manual
    provider_args:
      path: "C:\\tools\\mytool.exe"
      sha256: "..."
```

**说明：**

- `provider_args` 是 Provider 相关的原始数据，Provider 自己定义 schema。
- `config_refs` 记录环境变量、别名、Shell 函数等附加信息，不强制在首期实现自动还原。
- 跨系统迁移时，目标端根据 `provider` 和 `provider_args` 寻找等价 Provider 或生成警告。

---

## 4. 同系统交换流程（Windows↔Windows / Linux↔Linux）

### 4.1 源端：导出

```
1. 用户运行：cli-mig discover --output env.yaml
2. 发现引擎扫描 PATH 与各 Provider 数据库。
3. 生成 Manifest（env.yaml）。
4. （可选）cli-mig export --manifest env.yaml --bundle env.bundle.tar.gz
   打包清单 + 选定配置文件（如 ~/.gitconfig、~/.ssh/config 等，需用户确认）。
```

### 4.2 目标端：导入与还原

```
1. 用户复制 bundle 到目标机器。
2. 运行：cli-mig import env.bundle.tar.gz
3. 运行：cli-mig plan --manifest env.yaml
   输出执行计划：需要安装哪些、哪些已存在版本不同、哪些无法还原。
4. 用户审阅后运行：cli-mig apply --manifest env.yaml
   按 Provider 逐个安装/升级。
5. 输出还原报告与异常项。
```

### 4.3 计划引擎示例输出

```text
Plan for applying env.yaml on windows/amd64:

  [install] node@v20.12.0 via scoop (nodejs)
  [skip]    git@2.45.0 already installed via scoop
  [upgrade] kubectl@v1.30.0 via winget (current: v1.29.0)
  [warn]    my-custom-tool (manual) cannot be auto-installed
            path: C:\tools\mytool.exe
  [install] ripgrep@14.1.0 via scoop

Total: 4 actions, 1 warning. Use --dry-run to preview commands.
```

---

## 5. 跨系统迁移的架构预留

虽然首期只做同系统交换，但架构必须为 Windows↔Linux 预留扩展点。

### 5.1 关键抽象

1. **Provider 的跨平台等价映射**  
   引入 `CrossPlatformResolver` 接口：

   ```go
   type CrossPlatformResolver interface {
       // 判断能否将源 Provider 的条目转换到目标平台
       CanResolve(source Entry, targetOS OSType) bool
       // 返回目标平台推荐的 Provider + 参数
       Resolve(source Entry, targetOS OSType) (Entry, error)
   }
   ```

   例如：`scoop` 上的 `nodejs` 可以映射到 Linux 的 `apt` 包 `nodejs`，或更推荐的 `nvm`/`fnm` Provider。

2. **平台无关的 Entry 描述**  
   Manifest 中的 `provider_args` 应保持描述性而非平台绑定。例如记录包名而不是具体的安装路径。

3. **可选的 fallback 策略**  
   当目标平台没有等价 Provider 时，给出选项：
   - 跳过该 CLI；
   - 尝试从 GitHub Release 下载对应平台的二进制（manual provider 的增强）；
   - 提示用户手动安装。

4. **配置路径转换**  
   `config_refs` 中的文件路径在跨系统时应经过 `PathTranslator` 转换（如 `C:\Users\lzy\.gitconfig` → `/home/lzy/.gitconfig`）。

### 5.2 跨系统流程（未来）

```
Source (Windows)              Target (Linux)
    │                              │
    ▼                              ▼
 discover ──► Manifest ──► CrossPlatformResolver
                               │
                               ▼
                    生成 target-oriented Manifest
                               │
                               ▼
                          plan / apply on Linux
```

### 5.3 本期预留动作

- `internal/platform` 包必须从一开始就存在，不要写成一堆 `if runtime.GOOS == "windows"` 散在业务代码里。
- Provider 接口预留 `Resolve` 方法，同系统时直接透传，跨系统时启用映射。
- Manifest schema 预留 `target_overrides` 字段，允许为特定目标 OS 指定覆盖参数。

---

## 6. 技术栈

| 层级 | 选型 | 理由 |
|------|------|------|
| 语言 | Go 1.22+ | 单二进制、跨平台编译、静态类型、生态成熟。 |
| CLI 框架 | cobra | 子命令、Flag、自动补全、文档生成。 |
| 配置解析 | viper + gopkg.in/yaml.v3 | 支持多格式、环境变量覆盖。 |
| 日志 | slog | Go 标准库，结构化日志。 |
| 测试 | testify + 表格驱动 | 保持简洁。 |
| 打包 | 单二进制 + tar.gz bundle | 无需安装器，便于传播。 |
| CI | GitHub Actions | 自动在 Windows、Ubuntu 上跑测试与构建。 |

---

## 7. 编码计划（里程碑）

### Phase 0：项目骨架（约 1 周）

**目标：** 可编译、可运行、有基本测试框架。

- [ ] 初始化 Go module：`go mod init github.com/yourname/cli-mig`。
- [ ] 创建目录结构：
  ```
  cmd/
  internal/
    ├── cli/
    ├── config/
    ├── discover/
    ├── plan/
    ├── apply/
    ├── platform/
    ├── providers/
    │   ├── registry.go
    │   ├── scoop.go
    │   ├── apt.go
    │   └── manual.go
    └── manifest/
  pkg/
    └── api/          # 对外稳定类型（未来可被 UI/Server 复用）
  docs/
  examples/
  ```
- [ ] 实现 `cobra` 根命令与 `discover`、`plan`、`apply`、`version` 子命令骨架。
- [ ] 定义 `Provider` 接口、`Entry` 类型、`Manifest` 结构体。
- [ ] 实现 `Platform` 接口的 Windows 与 Linux 桩代码。
- [ ] 接入 `slog` 日志与基础配置加载。
- [ ] 编写第一个单元测试（`manifest` 序列化/反序列化）。

**验收标准：**

- `go build ./...` 通过。
- `cli-mig version` 能输出版本号。
- `go test ./...` 全绿。

### Phase 1：发现引擎 + 核心 Provider（约 2-3 周）

**目标：** 在 Windows 与 Linux 上都能生成可信的 Manifest。

- [ ] 实现 PATH 扫描：枚举所有可执行文件，去重，提取名称。
- [ ] 实现 `manual` Provider：标记未知来源的 CLI。
- [ ] 实现 `scoop` Provider：读取 `~/scoop/apps/` 与 `scoop export` 输出。
- [ ] 实现 `apt` Provider：调用 `dpkg -l` / `apt list --installed` 解析。
- [ ] 实现版本探测：优先 `--version` / `-v` / `-V`，支持常见输出格式解析。
- [ ] 实现 Provider 注册表（`registry.go`），支持按平台启用。
- [ ] 实现 `discover` 命令：输出 YAML/JSON Manifest。
- [ ] 实现 `Manifest` 校验（schema 版本、必填字段）。
- [ ] 为 Provider 编写 mock 与单元测试。

**验收标准：**

- 在测试用的 Windows/Linux 虚拟机上，`cli-mig discover` 能列出常见 CLI（git、node、python、kubectl 等）。
- 对同一个系统多次运行，输出稳定（排序一致、无随机项）。
- 未知来源 CLI 至少被记录为 manual，不丢失。

### Phase 2：计划与还原（同系统）（约 2-3 周）

**目标：** 支持同系统环境的导出与导入。

- [ ] 实现 `plan` 引擎：对比当前环境与 Manifest，输出 `ActionPlan`。
- [ ] 实现 `apply` 引擎：按 ActionPlan 调用 Provider 安装/升级。
- [ ] 实现 `export` / `import` 命令：打包/解包清单与可选配置文件。
- [ ] 实现冲突处理：版本不同是升级还是跳过？提供 `--conflict` 策略（prefer-source / prefer-target / fail）。
- [ ] 实现 `--dry-run`：预览将要执行的命令，不实际改动系统。
- [ ] 实现基础回滚：记录已安装/升级的条目，失败时给出撤销脚本。
- [ ] 实现 `diff` 命令：比较两个 Manifest 或当前环境与 Manifest。

**验收标准：**

- 在干净 Windows 上执行 `cli-mig apply --manifest sample.yaml --dry-run` 输出正确命令。
- 实际 apply 一组测试包（如 5-10 个 CLI）成功。
- 能生成并导入 bundle，解包后清单与原始一致。

### Phase 3：扩展 Provider 与稳定性（约 2 周）

**目标：** 覆盖更多安装渠道，提升识别准确率。

- [ ] 实现 `winget` Provider（检测 + 安装）。
- [ ] 实现 `cargo`、`npm`/`pnpm`/`yarn`（全局包）、`pipx` Provider 的检测能力。
- [ ] 实现 `dnf` / `pacman` / `homebrew` 的至少检测能力。
- [ ] 引入启发式规则：根据安装目录判断 Provider（如 `/usr/bin/` 多为 apt，`*\\scoop\\apps\\*` 为 scoop）。
- [ ] 实现进度条与更友好的终端输出（`mpb` 或自定义）。
- [ ] 增加集成测试：在 CI 中用 GitHub Actions 的 Windows/Linux runner 跑端到端流程。

**验收标准：**

- 对常见开发环境（Node、Rust、Python、Go、K8s 工具链）的识别覆盖率达到 80% 以上。
- CI 在 Windows 与 Ubuntu 上均通过。

### Phase 4：跨系统映射（预留，未来）

**目标：** 实现 Windows↔Linux 的初步自动迁移。

- [ ] 实现 `CrossPlatformResolver` 框架。
- [ ] 建立常见 CLI 的跨平台映射表（scoop bucket/package → apt package / GitHub release asset）。
- [ ] 实现 `PathTranslator` 转换配置文件路径。
- [ ] 在 `plan` 中加入跨系统风险提示。
- [ ] 支持 `--target-os` 参数生成 target-oriented Manifest。

---

## 8. 风险与对策

| 风险 | 影响 | 对策 |
|------|------|------|
| 某些 CLI 无法识别来源 | 中 | 兜底 `manual` Provider；鼓励社区贡献映射表。 |
| 版本号格式千奇百怪 | 中 | 提供可插拔的 `VersionParser`，主流格式内置，其余记录原字符串。 |
| 安装需要管理员/root | 中 | apply 前检查权限，提示用户用管理员/ sudo 重试；支持 `--sudo` 模式。 |
| 还原后版本仍不一致 | 低 | 默认按源版本安装，若包管理器无该版本则提示，不强行降级/升级。 |
| 跨系统配置路径差异 | 高（跨系统阶段） | 一开始就抽象 `PathTranslator`，避免后期大量重构。 |
| 用户手动安装的脚本/别名遗漏 | 中 | 首期聚焦 CLI 二进制，Shell 配置只记录不自动还原。 |

---

## 9. 开源与社区

- **License**：MIT，便于企业与个人使用。
- **贡献指南**：Provider 是最易扩展的点，欢迎社区提交新 Provider。
- **Issue 模板**：
  - Provider 支持请求（请附带 `cli-mig discover --output` 的相关片段）。
  - 跨平台映射请求（请说明源平台、目标平台、期望行为）。
- **命名**：项目仓库建议 `cli-migration`，二进制名 `cli-mig`，配置文件目录 `~/.config/cli-mig/`。

---

## 10. 下一步行动

1. 评审本设计文档，确认技术栈与里程碑是否满足预期。
2. 创建 GitHub 仓库并初始化 Phase 0 的骨架代码。
3. 定义并冻结 `Manifest` schema v1.0。
4. 编写第一个可工作的 `discover` 原型，验证 Provider 抽象是否合理。

---

*文档结束*
