# Git 工作流规范

本文档定义了 Aether-Defense-System 项目的 Git 工作流规范，包括分支管理、提交规范和 PR 流程。

## Git 工作流概述

所有代码变更必须遵循以下流程：

```
1. 创建新分支（从 main/master）
   ↓
2. 切换到新分支
   ↓
3. 进行开发（TDD → 代码 → 测试 → 文档）
   ↓
4. 提交代码（遵循提交信息规范）
   ↓
5. 推送分支到远程
   ↓
6. 创建 Pull Request
   ↓
7. Code Review
   ↓
8. 合并到主分支
```

## 分支管理规范

### 分支命名规范

分支名称格式：`<type>/<description>`

**类型（type）**：
- `feat/` - 新功能
- `fix/` - Bug 修复
- `test/` - 测试相关
- `refactor/` - 重构
- `perf/` - 性能优化
- `docs/` - 文档更新
- `chore/` - 构建/工具相关
- `hotfix/` - 紧急修复

**描述（description）**：
- 使用小写字母和连字符
- 简洁描述功能或修复
- 长度控制在 50 个字符以内

**示例**：
```bash
feat/order-coupon-support      # 订单支持优惠券功能
fix/redis-connection-pool       # 修复 Redis 连接池问题
test/trade-service-integration  # 交易服务集成测试
refactor/promotion-service      # 重构营销服务
perf/redis-lua-optimization     # Redis Lua 脚本优化
docs/api-documentation          # API 文档更新
chore/update-dependencies       # 更新依赖
hotfix/order-payment-bug        # 紧急修复订单支付 bug
```

### 分支创建流程

#### 1. 确保主分支是最新的

```bash
# 切换到主分支
git checkout main  # 或 git checkout master

# 拉取最新代码
git pull origin main
```

#### 2. 创建新分支

```bash
# 从主分支创建新分支
git checkout -b feat/order-coupon-support

# 或者分两步
git branch feat/order-coupon-support
git checkout feat/order-coupon-support
```

#### 3. 验证分支

```bash
# 查看当前分支
git branch

# 查看分支状态
git status
```

### 分支管理规则

1. **主分支（main/master）**
   - 受保护分支，不能直接推送
   - 只能通过 PR 合并
   - 始终保持可部署状态

2. **功能分支（feat/）**
   - 从主分支创建
   - 完成功能后通过 PR 合并
   - 合并后删除分支

3. **修复分支（fix/）**
   - 从主分支创建
   - 修复 bug 后通过 PR 合并
   - 合并后删除分支

4. **紧急修复分支（hotfix/）**
   - 从主分支创建
   - 修复后立即合并
   - 可能需要同时合并到开发分支

5. **测试分支（test/）**
   - 用于添加或修改测试
   - 遵循相同的 PR 流程

## 提交信息规范

### Conventional Commits 格式

提交信息格式：
```
<type>(<scope>): <subject>

<body>

<footer>
```

### 类型（Type）

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `test`: 测试相关（添加或修改测试）
- `refactor`: 重构（既不是新功能也不是 bug 修复）
- `perf`: 性能优化
- `chore`: 构建/工具相关
- `style`: 代码格式（不影响代码运行）
- `ci`: CI/CD 相关

### 范围（Scope）

可选，表示影响的范围，如：
- `trade`: 交易服务
- `promotion`: 营销服务
- `user`: 用户服务
- `api`: API 层
- `rpc`: RPC 层
- `common`: 公共组件

### 主题（Subject）

- 使用祈使句，现在时态
- 首字母小写
- 不以句号结尾
- 长度控制在 50 个字符以内

### 正文（Body）

可选，详细描述：
- 变更的原因
- 与之前行为的对比
- 关闭的 Issue

### 页脚（Footer）

可选，用于：
- 关闭 Issue: `Closes #123`
- 关联 Issue: `Refs #456`
- 破坏性变更: `BREAKING CHANGE: <description>`

### 提交信息示例

#### 简单提交（只有主题）

```bash
git commit -m "feat(trade): add coupon support to order placement"
```

#### 详细提交（包含正文）

```bash
git commit -m "feat(trade): add coupon support to order placement

- Add coupon validation in PlaceOrder method
- Calculate discount amount based on coupon rules
- Update order total amount after applying coupon
- Add unit tests for coupon validation

Closes #123"
```

#### 多行提交（使用编辑器）

```bash
git commit
```

在编辑器中输入：
```
feat(trade): add coupon support to order placement

This change adds support for applying coupons during order placement.
The system validates coupon eligibility and calculates discount amount.

Changes:
- Add coupon validation logic
- Update order calculation to include discount
- Add integration tests

Closes #123
Refs #456
```

### 提交规则

1. **每次提交只做一件事**
   - 不要在一个提交中包含多个不相关的变更
   - 如果发现提交了不相关的内容，使用 `git reset` 重新提交

2. **提交前检查**
   ```bash
   # 查看变更
   git status
   git diff
   
   # 运行测试
   go test ./...
   
   # 运行 lint
   golangci-lint run
   ```

3. **提交频率**
   - 完成一个小功能就提交
   - 不要积累太多变更再提交
   - 保持提交历史清晰

4. **提交信息质量**
   - 清晰描述做了什么
   - 说明为什么这样做（如需要）
   - 关联相关的 Issue

## 完整工作流示例

### 示例 1：添加新功能

```bash
# 1. 切换到主分支并更新
git checkout main
git pull origin main

# 2. 创建功能分支
git checkout -b feat/order-coupon-support

# 3. 进行开发（TDD → 代码 → 测试 → 文档）
# ... 编写代码 ...

# 4. 查看变更
git status
git diff

# 5. 添加文件到暂存区
git add .

# 6. 提交（遵循提交信息规范）
git commit -m "feat(trade): add coupon support to order placement

- Add coupon validation in PlaceOrder method
- Calculate discount amount based on coupon rules
- Update order total amount after applying coupon
- Add unit tests for coupon validation

Closes #123"

# 7. 推送分支到远程
git push origin feat/order-coupon-support

# 8. 在 GitHub/GitLab 创建 Pull Request
```

### 示例 2：修复 Bug

```bash
# 1. 切换到主分支并更新
git checkout main
git pull origin main

# 2. 创建修复分支
git checkout -b fix/redis-connection-pool

# 3. 修复 bug
# ... 修复代码 ...

# 4. 添加测试
# ... 添加回归测试 ...

# 5. 提交
git commit -m "fix(common): fix Redis connection pool leak

- Close connections properly after use
- Add connection pool monitoring
- Add unit tests for connection cleanup

Fixes #456"

# 6. 推送并创建 PR
git push origin fix/redis-connection-pool
```

### 示例 3：添加测试

```bash
# 1. 切换到主分支并更新
git checkout main
git pull origin main

# 2. 创建测试分支
git checkout -b test/trade-service-integration

# 3. 添加测试
# ... 编写集成测试 ...

# 4. 提交
git commit -m "test(trade): add integration tests for order placement

- Add integration tests for PlaceOrder method
- Test order creation with multiple courses
- Test order creation with coupons
- Test error handling scenarios

Refs #789"

# 5. 推送并创建 PR
git push origin test/trade-service-integration
```

## 推送和 PR 流程

### 推送分支

```bash
# 首次推送，设置上游分支
git push -u origin feat/order-coupon-support

# 后续推送
git push
```

### 创建 Pull Request

1. **在 GitHub/GitLab 上创建 PR**
   - 使用 PR 模板（`.github/pull_request_template.md`）
   - 填写所有相关信息
   - 关联相关的 Issue

2. **PR 标题规范**
   - 使用与提交信息相同的格式
   - 示例：`feat(trade): add coupon support to order placement`

3. **PR 描述**
   - 使用 PR 模板
   - 详细描述变更内容
   - 说明测试情况
   - 关联相关 Issue

### PR 合并后

```bash
# 1. 切换到主分支
git checkout main

# 2. 拉取最新代码（包含合并的 PR）
git pull origin main

# 3. 删除本地分支（可选）
git branch -d feat/order-coupon-support

# 4. 删除远程分支（可选）
git push origin --delete feat/order-coupon-support
```

## 常见场景处理

### 场景 1：需要更新主分支代码

```bash
# 在功能分支上
git checkout feat/order-coupon-support

# 拉取主分支最新代码
git fetch origin main

# 合并主分支到当前分支
git merge origin/main

# 或者使用 rebase（保持提交历史线性）
git rebase origin/main
```

### 场景 2：提交后发现错误

```bash
# 修改文件后，添加到上次提交
git add .
git commit --amend

# 如果已经推送，需要强制推送（谨慎使用）
git push --force-with-lease origin feat/order-coupon-support
```

### 场景 3：需要修改提交信息

```bash
# 修改最后一次提交信息
git commit --amend -m "feat(trade): add coupon support"

# 如果已经推送，需要强制推送
git push --force-with-lease origin feat/order-coupon-support
```

### 场景 4：需要拆分提交

```bash
# 交互式 rebase
git rebase -i HEAD~3

# 在编辑器中，将需要拆分的提交标记为 'edit'
# 然后使用 git reset 和 git add 重新提交
```

## 禁止的操作

1. **禁止直接推送到主分支**
   ```bash
   # ❌ 错误
   git push origin main
   
   # ✅ 正确：通过 PR 合并
   ```

2. **禁止使用 `--force` 推送共享分支**
   ```bash
   # ❌ 错误
   git push --force origin feat/branch-name
   
   # ✅ 正确：使用 --force-with-lease
   git push --force-with-lease origin feat/branch-name
   ```

3. **禁止在提交中包含敏感信息**
   - 密码、密钥、Token 等
   - 使用 `.gitignore` 排除敏感文件

4. **禁止提交未测试的代码**
   - 所有代码必须通过测试
   - 所有 lint 检查必须通过

## Git Hooks 集成

项目已配置 pre-commit hooks，会自动检查：

- 代码格式（gofmt, goimports）
- 代码质量（golangci-lint）
- 测试（go test）
- 提交信息格式（gitlint）

## 检查清单

在创建 PR 前，确保：

- [ ] 从主分支创建了新分支
- [ ] 分支命名符合规范（`type/description`）
- [ ] 所有提交信息符合规范
- [ ] 所有测试通过
- [ ] Linter 检查通过
- [ ] Pre-commit hooks 通过
- [ ] 文档已更新
- [ ] 代码已推送到远程
- [ ] PR 已创建并填写模板

## 参考资源

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [项目开发流程规范](./development-workflow.md)
- [项目代码审查规范](./code-review-guidelines.md)

