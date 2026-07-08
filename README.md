# cli-migration

CLI 无痛搬家工具：发现并迁移你的命令行环境。

## 愿景

在不同机器、不同账号之间一键迁移命令行工具环境。

## 当前阶段

- [x] 架构设计
- [ ] Phase 0：项目骨架
- [ ] Phase 1：发现引擎与核心 Provider
- [ ] Phase 2：计划与还原
- [ ] Phase 3：扩展 Provider 与 CI

## 核心命令

```bash
cli-mig discover --output my-env.yaml
cli-mig plan --manifest my-env.yaml
cli-mig apply --manifest my-env.yaml
cli-mig export --manifest my-env.yaml --output bundle.tar.gz
cli-mig import bundle.tar.gz
```

## License

MIT
