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
- [ ] Phase 4：跨操作系统映射（预留）

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

## 快速开始

### 1. 扫描当前环境

```bash
cli-mig discover --output my-env.yaml
```

使用 `--probe-versions` 会调用每个 CLI 的 `--version` 来探测版本（较慢）：

```bash
cli-mig discover --output my-env.yaml --probe-versions
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

## 项目结构

```
cmd/cli-mig/          # 程序入口
internal/
  cli/                # cobra 命令
  discover/           # 发现引擎
  plan/               # 计划引擎
  apply/              # 执行引擎
  bundle/             # 打包/解包
  manifest/           # 清单序列化
  platform/           # 平台抽象
  providers/          # Provider 抽象与实现
pkg/api/              # 对外稳定类型
```

## 跨系统迁移

当前版本优先保证同系统（Windows→Windows、Linux→Linux）的可靠迁移。
架构上已预留 `CrossPlatformResolver` 和 `target_overrides`，未来可实现 Windows↔Linux 的自动映射。

## License

MIT
