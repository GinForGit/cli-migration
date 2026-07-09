# cli-migration

CLI 无痛搬家工具：发现并迁移你的命令行环境。

## 愿景

在不同机器、不同账号之间一键迁移命令行工具环境。

## 当前阶段

- [x] 架构设计
- [x] Phase 0：项目骨架
- [x] Phase 1：发现引擎与核心 Provider
- [x] Phase 2：计划与还原
- [x] Phase 3：扩展 Provider 与 Bundle
- [x] Phase 4：跨操作系统映射（基础实现）

## 支持的 Provider

| Provider   | 检测 | 安装 | 平台 |
|------------|------|------|------|
| scoop      | ✓    | ✓    | Windows |
| winget     | ✓    | ✓    | Windows |
| choco      | ✓    | ✓    | Windows |
| apt        | ✓    | ✓    | Linux   |
| cargo      | ✓    | ✓    | 跨平台  |
| npm        | ✓    | ✓    | 跨平台  |
| pipx       | ✓    | ✓    | 跨平台  |
| manual     | ✓    | ✗    | 跨平台  |

## 安装

```bash
go install github.com/GinForGit/cli-migration/cmd/cli-mig@latest
```

或者从 [Releases](https://github.com/GinForGit/cli-migration/releases) 下载预编译二进制。

## 快速开始

### 1. 扫描当前环境

```bash
cli-mig discover --output my-env.yaml
```

使用 `--probe-versions` 会调用每个 CLI 的 `--version` 来探测版本（较慢）：

```bash
cli-mig discover --output my-env.yaml --probe-versions
```

如果要迁移到另一操作系统，可生成 `target_overrides`（当前支持 windows/linux 互转）：

```bash
# 在 Windows 上扫描，为 Linux 生成映射
cli-mig discover --output my-env.yaml --target-os linux

# 在 Linux 上扫描，为 Windows 生成映射
cli-mig discover --output my-env.yaml --target-os windows
```

### 2. 预览还原计划

```bash
cli-mig plan --manifest my-env.yaml
```

### 3. 执行还原

```bash
# 先预览
cli-mig apply --manifest my-env.yaml --dry-run

# 真正执行
cli-mig apply --manifest my-env.yaml
```

### 4. 打包搬家

```bash
# 仅打包清单
cli-mig export --manifest my-env.yaml --output my-env.bundle.tar.gz

# 同时带走 .gitconfig 等配置文件
cli-mig export --manifest my-env.yaml --include-configs --output my-env.bundle.tar.gz
```

### 5. 新机器解包并还原

```bash
cli-mig import my-env.bundle.tar.gz --output-dir my-env
cli-mig plan --manifest my-env/manifest.yaml
cli-mig apply --manifest my-env/manifest.yaml
```

### 6. 对比差异

```bash
cli-mig diff --manifest my-env.yaml
```

## 命令与参数

### `cli-mig version`

打印版本、commit 和构建时间。

```bash
cli-mig version
```

### `cli-mig discover`

扫描当前系统已安装的 CLI 工具，生成清单文件。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--output` | `-o` | `env.yaml` | 输出清单文件路径 |
| `--format` | `-f` | `yaml` | 输出格式：`yaml` 或 `json` |
| `--probe-versions` | - | `false` | 探测 `manual` 条目的版本（较慢） |
| `--target-os` | - | `""` | 为指定目标系统生成 `target_overrides`，可选 `windows`、`linux` |

```bash
cli-mig discover -o my-env.yaml -f yaml --target-os linux
```

### `cli-mig plan`

对比清单与当前环境，输出将要执行的操作（只读）。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--manifest` | `-m` | - | 清单文件路径（必填） |
| `--target-os` | - | 当前系统 | 目标操作系统：`windows`、`linux` |

```bash
cli-mig plan -m my-env.yaml --target-os windows
```

### `cli-mig apply`

根据清单在当前机器上安装、升级或跳过 CLI 工具。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--manifest` | `-m` | - | 清单文件路径（必填） |
| `--dry-run` | - | `false` | 只预览，不执行 |
| `--skip-manual` | - | `false` | 跳过无法自动安装的 `manual` 条目 |
| `--target-os` | - | 当前系统 | 目标操作系统：`windows`、`linux` |

```bash
cli-mig apply -m my-env.yaml --dry-run
cli-mig apply -m my-env.yaml --skip-manual
```

### `cli-mig export`

将清单文件与可选的配置文件打包成 `tar.gz` 归档。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--manifest` | `-m` | - | 清单文件路径（必填） |
| `--output` | `-o` | `env.bundle.tar.gz` | 输出 bundle 路径 |
| `--include-configs` | - | `false` | 包含常用配置文件 |

```bash
cli-mig export -m my-env.yaml -o my-env.bundle.tar.gz --include-configs
```

### `cli-mig import`

解压 bundle 归档并显示其中的清单路径。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--output-dir` | `-d` | 根据 bundle 名推断 | 解压目录 |
| `<bundle>` | - | - | bundle 文件路径（位置参数，必填） |

```bash
cli-mig import my-env.bundle.tar.gz -d my-env
```

### `cli-mig diff`

对比清单与当前环境，显示哪些条目不一致。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--manifest` | `-m` | - | 清单文件路径（必填） |

```bash
cli-mig diff -m my-env.yaml
```

## 项目结构

```
cmd/cli-mig/          # 程序入口
internal/
  cli/                # cobra 命令
  discover/           # 发现引擎
  plan/               # 计划引擎
  apply/               # 执行引擎
  bundle/             # 打包/解包
  manifest/           # 清单序列化
  platform/           # 平台抽象
  providers/          # Provider 抽象与实现
  crossplatform/      # 跨平台映射
pkg/api/              # 对外稳定类型
```

## 跨系统迁移

当前版本已实现 Windows ↔ Linux 的基础自动映射，通过 `CrossPlatformResolver` 和 `target_overrides` 完成。同系统迁移仍是优先保证场景，跨系统映射表会持续扩展。

## License

MIT
