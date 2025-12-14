# Cursor/VSCode 编辑器配置

本目录包含 Cursor 编辑器（基于 VSCode）的配置文件，用于自动代码检查和格式化。

## 配置文件说明

### `settings.json`
编辑器工作区设置，包含：
- **Go 语言配置**：启用 gopls 语言服务器，配置格式化工具
- **自动格式化**：保存时自动格式化代码（使用 goimports）
- **自动导入整理**：保存时自动整理导入语句
- **golangci-lint 集成**：保存时自动运行 linter
- **文件关联**：`.api` 和 `.proto` 文件的语法高亮

### `extensions.json`
推荐的扩展列表（Cursor 会自动提示安装）。

### `tasks.json`
预定义的任务，可以通过 `Ctrl+Shift+P` -> `Tasks: Run Task` 运行：
- `golangci-lint: Run` - 运行代码检查
- `golangci-lint: Fix` - 运行代码检查并自动修复
- `go: mod tidy` - 整理 Go 模块依赖
- `go: test` - 运行测试
- `pre-commit: Run all` - 运行所有 pre-commit hooks

### `launch.json`
调试配置，用于调试 Go 程序。

## 功能特性

### ✅ 自动格式化
- 保存文件时自动格式化 Go 代码
- 使用 `goimports` 自动整理导入语句
- 自动添加/删除未使用的导入

### ✅ 实时 Linting
- 保存文件时自动运行 `golangci-lint`
- 在编辑器中实时显示错误和警告
- 使用项目配置的 `.golangci.yml`

### ✅ 代码提示
- Go 语言服务器（gopls）提供智能代码补全
- 类型检查和错误提示
- 代码导航和重构支持

## 使用说明

### 首次设置

1. **安装 Go 工具**：
   - 打开任意 `.go` 文件
   - Cursor 会自动提示安装 Go 工具
   - 或手动运行：`Ctrl+Shift+P` -> `Go: Install/Update Tools`

2. **安装 golangci-lint（命令行工具）**：
   ```bash
   # 安装 golangci-lint v1.55.2
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
   
   # 验证安装
   golangci-lint --version
   ```

3. **验证配置**：
   - 打开任意 `.go` 文件
   - 修改代码并保存
   - 应该看到自动格式化和 lint 检查

### 运行任务

使用快捷键或命令面板运行预定义任务：

1. **快捷键**：`Ctrl+Shift+B`（运行默认任务）
2. **命令面板**：`Ctrl+Shift+P` -> `Tasks: Run Task` -> 选择任务

### 调试

1. 在代码中设置断点
2. 按 `F5` 开始调试
3. 或使用 `Ctrl+Shift+D` 打开调试面板

## 故障排除

### Linter 不工作

1. **检查 golangci-lint 是否安装**：
   ```bash
   golangci-lint --version
   ```
   如果未安装，请按照上面的步骤 2 安装。

2. **检查 golangci-lint 是否在 PATH 中**：
   ```bash
   which golangci-lint
   ```
   应该显示 golangci-lint 的路径（通常在 `$(go env GOPATH)/bin`）

3. **重新加载窗口**：
   - `Ctrl+Shift+P` -> `Developer: Reload Window`

### 格式化不工作

1. **检查 Go 工具是否安装**：
   - `Ctrl+Shift+P` -> `Go: Install/Update Tools`
   - 确保 `goimports` 已安装

2. **检查设置**：
   - 确保 `editor.formatOnSave` 为 `true`
   - 确保 `go.formatTool` 为 `goimports`

### 导入整理不工作

1. **检查 Go 语言服务器**：
   - `Ctrl+Shift+P` -> `Go: Restart Language Server`

2. **检查设置**：
   - 确保 `editor.codeActionsOnSave` 包含 `source.organizeImports`

## 参考

- [Go Extension for VSCode](https://github.com/golang/vscode-go)
- [golangci-lint 文档](https://golangci-lint.run/)
- [项目代码规范文档](../doc/coding-standards/README.md)
