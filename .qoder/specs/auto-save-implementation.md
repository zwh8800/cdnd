# 自动保存功能实现计划

## Context

CDND 项目的配置系统中已经定义了 `Autosave` 和 `AutosaveInterval` 配置项，但实际功能并未实现。玩家在游戏过程中无法获得自动保存的保护，一旦意外退出可能丢失进度。本计划旨在实现完整的自动保存功能，支持回合级和时间间隔两种触发模式，使用专用槽位 slot 0，并采用异步执行不影响游戏性能。

## 实现步骤

### Step 1: 修改 save 包支持 slot 0

**文件: `internal/save/manager.go`**

- 新增常量 `AutosaveSlot = 0`
- 修改 `Save()`、`Load()`、`Delete()`、`Exists()` 方法的槽位验证逻辑：
  - 原来: `if slot < 1 || slot > MaxSlots`
  - 改为: `if slot != AutosaveSlot && (slot < 1 || slot > MaxSlots)`
- `ListSlots()` 保持不变（仅列出 1-10，不暴露 slot 0）

### Step 2: Engine 添加自动保存管理

**文件: `internal/game/engine.go`**

**新增字段:**
```go
autosaveCancel context.CancelFunc
autosaveWg     sync.WaitGroup
autosaveSaving atomic.Bool
```

**新增方法:**

1. **`StartAutosave()`** — 在 `Start()` 和 `LoadGame()` 末尾调用
   - 检查 `config.Game.Autosave`，未启用则直接返回
   - 创建 cancelable context，启动 ticker goroutine
   - 按 `AutosaveInterval` 定时触发自动保存

2. **`StopAutosave()`** — 在游戏退出时调用
   - 调用 cancel 停止 ticker
   - `wg.Wait()` 等待当前保存完成

3. **`triggerAutosave(ctx)`** — 内部异步保存逻辑
   - 使用 `atomic.Bool` CAS 防止并发保存
   - 在 goroutine 中执行 `SaveGame(save.AutosaveSlot)`
   - recover panic，log 记录错误，不中断游戏

4. **`triggerAutosaveByTurn()`** — 回合级触发器
   - 检查配置启用状态
   - 调用 `triggerAutosave(ctx)`

**修改现有方法:**

- `Start()` — 末尾添加 `e.StartAutosave()`
- `LoadGame()` — 末尾添加 `e.StartAutosave()`
- `Process()` — 正常返回前添加 `go e.triggerAutosaveByTurn()`

### Step 3: 命令行集成

**文件: `cmd/start.go`**

- 在 `gameP.Run()` 返回后调用 `engine.StopAutosave()`

**文件: `cmd/load.go`**

- 在 `p.Run()` 返回后调用 `engine.StopAutosave()`

### Step 4: 添加命令行控制选项（可选）

**文件: `cmd/start.go` 和 `cmd/load.go`**

- 新增 `--no-autosave` 布尔标志
- 设置时覆盖 `config.Game.Autosave = false`

## 关键文件

- `internal/save/manager.go` — 槽位验证修改
- `internal/game/engine.go` — 自动保存核心逻辑
- `cmd/start.go` — 启动命令清理
- `cmd/load.go` — 加载命令清理

## 验证方法

1. **编译检查**: `go build ./...` 确保无编译错误
2. **单元测试**: 运行现有测试 `go test ./...`
3. **功能测试**:
   - 启动游戏，执行几次交互，检查 `~/.cdnd/saves/slot_0.json` 是否创建
   - 修改配置 `autosave: false`，验证不创建 slot_0
   - 验证手动保存（slot 1-10）与自动保存共存
   - 验证游戏退出后自动保存正确停止
