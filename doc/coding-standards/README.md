# 代码规范文档

本目录包含 Aether-Defense-System 项目的代码规范和设计规则文档。

## 文档结构

### 核心规范文档

1. **[Go-Zero 框架约定](./go-zero-conventions.md)**
   - API 定义规范（.api 文件）
   - RPC 服务规范（.proto 文件）
   - 错误处理模式
   - 中间件使用规范

2. **[微服务设计规则](./service-design-rules.md)**
   - 服务边界划分
   - 服务间通信规范
   - 数据管理规范
   - 服务治理（限流、熔断、监控）

3. **[性能优化指南](./performance-guidelines.md)**
   - Go 语言性能优化
   - Redis 性能优化
   - 数据库性能优化
   - 消息队列优化

4. **[开发流程规范](./development-workflow.md)**
   - TDD（测试驱动开发）流程
   - 测试运行和验证
   - 文档更新规范
   - 代码提交检查清单

5. **[代码审查规范](./code-review-guidelines.md)**
   - Code Review 流程和标准
   - 审查检查清单
   - 审查最佳实践
   - PR/MR 模板使用

6. **[Git 工作流规范](./git-workflow.md)**
   - 分支管理规范
   - 提交信息规范（Conventional Commits）
   - PR 创建流程
   - Git 最佳实践

### 架构设计文档

4. **[架构设计标准](../design/architecture-standards.md)**
   - 架构原则
   - 分层架构设计
   - 服务设计规范
   - 数据架构规范
   - 安全架构

## 快速开始

### 1. 配置开发环境

#### 安装 golangci-lint

```bash
# 安装 golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# 验证安装
golangci-lint --version
```

#### 安装 pre-commit

```bash
# 安装 pre-commit
pip install pre-commit

# 安装 git hooks
pre-commit install

# 手动运行所有检查
pre-commit run --all-files
```

### 2. 使用 AI 辅助开发

项目 `.cursor/rules/coding-standards.mdc` 文件包含了 AI 辅助开发的规范。Cursor 会自动读取 `.cursor/rules/` 目录下的规则文件，在代码生成和审查时遵循项目约定。

### 3. 代码检查

#### 运行 golangci-lint

```bash
# 检查整个项目
golangci-lint run

# 检查特定目录
golangci-lint run ./service/trade/...

# 自动修复部分问题
golangci-lint run --fix
```

#### 运行 pre-commit hooks

```bash
# 在提交前自动运行（已通过 git hooks 配置）
git commit -m "your message"

# 手动运行所有检查
pre-commit run --all-files

# 运行特定 hook
pre-commit run golangci-lint
```

## 规范实施路径

### 阶段一：基础规范（已完成）

- ✅ 创建 `.cursorrules` 文件
- ✅ 配置 `golangci-lint`
- ✅ 配置 `pre-commit` hooks
- ✅ 编写核心架构规范文档

### 阶段二：工具集成（进行中）

1. 在 CI/CD 中集成 lint 检查
2. 配置代码质量门禁
3. 团队培训和文档完善

### 阶段三：持续优化（长期）

1. 根据项目演进调整规范
2. 收集团队反馈，优化规则
3. 定期审查和更新文档

## 规范文件说明

### `.cursor/rules/coding-standards.mdc`

AI 辅助开发的规范文件（Cursor 专用），位于 `.cursor/rules/` 目录下，包含：
- 项目结构约定
- Go-Zero 框架使用规范
- 代码风格和最佳实践
- 架构规则
- 开发流程（TDD、Git 工作流）
- 代码审查检查清单

### `.golangci.yml`

golangci-lint 配置文件，包含：
- 启用的 linter 列表
- Linter 特定配置
- 排除规则
- 问题报告配置

### `.pre-commit-config.yaml`

pre-commit hooks 配置，包含：
- 文件格式检查（YAML, JSON, TOML）
- Go 代码格式化（gofmt, goimports）
- Go 代码检查（go-vet, go-lint）
- golangci-lint 集成
- 提交信息格式检查

## 常见问题

### Q: 如何禁用某个 lint 规则？

A: 在 `.golangci.yml` 中的 `issues.exclude-rules` 部分添加排除规则：

```yaml
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

### Q: 如何添加新的代码规范？

A: 
1. 更新 `.cursor/rules/coding-standards.mdc` 文件（AI 辅助规范）
2. 更新相应的文档（`doc/coding-standards/` 目录）
3. 如需要，更新 `.golangci.yml` 配置

### Q: pre-commit hooks 运行太慢怎么办？

A: 
1. 只运行必要的 hooks
2. 使用 `SKIP` 环境变量跳过特定 hook：
   ```bash
   SKIP=golangci-lint git commit -m "message"
   ```

### Q: 如何为现有代码应用规范？

A: 
1. 运行 `golangci-lint run --fix` 自动修复部分问题
2. 逐步重构代码以符合规范
3. 在新代码中严格遵循规范

## 参考资源

- [Go-Zero 官方文档](https://go-zero.dev/)
- [golangci-lint 文档](https://golangci-lint.run/)
- [pre-commit 文档](https://pre-commit.com/)
- [项目架构设计文档](../design/architecture.md)

## 贡献指南

如需更新或改进代码规范：

1. 在相应的文档文件中进行修改
2. 更新 `.cursorrules` 文件（如适用）
3. 更新 `.golangci.yml` 配置（如适用）
4. 提交 PR 并说明变更原因

