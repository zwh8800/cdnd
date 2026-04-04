# 将 CLI 帮助信息翻译为中文

## Context

当前 `cdnd` 命令行工具的帮助信息是中英文混合的，部分命令（如 config、character 的子命令）已使用中文，但 root、start、version、load 等命令以及部分子命令仍使用英文。用户希望将所有帮助信息统一为中文显示。

## 需要修改的文件

### 1. `cmd/root.go`
- `Short`: "D&D CLI game powered by LLM" → 中文
- `Long`: 多行英文描述 → 中文
- `--config` flag: "config file (default is $HOME/.cdnd/config.yaml)" → 中文
- `--debug` flag: "enable debug mode" → 中文

### 2. `cmd/start.go`
- `start` 命令: Short、Long 描述 → 中文
- `-s, --save-slot` flag → 中文
- `-S, --scenario` flag → 中文
- `--skip-creation` flag → 中文
- `game` 子命令: Short → 中文
- `test-llm` 子命令: Short → 中文

### 3. `cmd/version.go`
- Short → 中文
- Long → 中文

### 4. `cmd/load.go`
- `load` 命令: Short、Long、Examples → 中文
- `-s, --slot` flag → 中文
- `saves` 子命令: Short → 中文

### 5. `cmd/character.go`
- `list` 子命令: Short → 中文
- `show` 子命令: Short → 中文
- (create、delete 已是中文，无需修改)

### 6. `cmd/config.go`
- 检查是否有残留英文文本需要翻译

## 验证方式

修改完成后运行以下命令验证帮助信息是否全部为中文：
```bash
go run main.go --help
go run main.go start --help
go run main.go load --help
go run main.go character --help
go run main.go config --help
go run main.go version --help
```
