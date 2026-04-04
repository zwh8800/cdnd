# TUI 加载动画实现计划

## Context

当前 `internal/ui/game.go` 中的"处理中..."状态是静态文本，用户等待 LLM 响应时缺乏视觉反馈。需要实现类似 Claude Code 或 npm install 的动态加载动画，提升用户体验。

## 实现方案

### 修改文件
- `/Users/wastecat/code/go/cdnd/internal/ui/game.go`（唯一需要修改的文件）

### 1. 添加动画字段到 GameModel

在 `loading` 字段附近添加：
```go
loadingFrame  int  // Braille 动画帧索引 (0-5)
loadingDots   int  // 省略号计数 (1-3)
loadingTimer  int  // 进度点动画帧 (0-3)
```

### 2. 定义 TickMsg 和动画常量

```go
type LoadingTickMsg time.Time

var brailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠦"}
```

### 3. 动画启动函数

```go
func startLoadingAnimation() tea.Cmd {
    return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
        return LoadingTickMsg(t)
    })
}
```

### 4. Update 方法修改

**添加 LoadingTickMsg 处理**（在 switch 语句中）：
- 仅当 `m.loading == true` 时更新帧计数器并继续调度
- `loadingFrame = (loadingFrame + 1) % 6`
- `loadingDots = (loadingDots % 3) + 1`
- `loadingTimer = (loadingTimer + 1) % 4`
- 返回 `startLoadingAnimation()` 命令形成循环

**修改 Enter 键处理**（设置 loading=true 时）：
- 初始化动画帧：`loadingFrame=0, loadingDots=1, loadingTimer=0`
- 在 `tea.Batch` 中同时启动 `processInput(input)` 和 `startLoadingAnimation()`

**清除 loading 的位置**（DMResponseMsg、StreamChunkMsg 错误/完成）：
- 只需设置 `m.loading = false`，动画会自动停止（因为 LoadingTickMsg 检查 loading 状态）

### 5. renderInput 方法修改

组合动画效果：
```
⠋ 处理中. ·
⠙ 处理中.. ··
⠹ 处理中... ···
⠸ 处理中.    
⠼ 处理中.. ·
⠦ 处理中... ··
```

- Braille spinner: `brailleFrames[m.loadingFrame]`
- 省略号: `strings.Repeat(".", m.loadingDots)`
- 进度点: `strings.Repeat("·", m.loadingTimer) + strings.Repeat(" ", 3-m.loadingTimer)`

### 6. 动画生命周期

```
空闲 → 按 Enter → loading=true → 启动动画 + processInput
     → 每 80ms 更新帧 → 收到响应 → loading=false → 动画自然停止
```

关键机制：`LoadingTickMsg` 处理中检查 `m.loading`，为 false 时不继续调度，动画自动停止。

## 验证方式

1. 编译检查：`go build ./...`
2. 运行程序：`go run .`（或相应的启动命令）
3. 输入文本按 Enter，观察输入框是否显示动态加载动画
4. 等待 LLM 响应完成后，动画应平滑停止并恢复输入框
5. 测试流式输出场景，动画应持续到流式完成
