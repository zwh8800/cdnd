# CDND 分层架构重构方案

## Context

当前 cdnd 项目的 internal 包采用扁平化功能域组织方式（character、combat、config、game、llm 等 12 个包平铺），存在依赖关系混乱、职责边界不清、循环导入风险高等问题。本方案将其重构为经典四层架构（Domain → Infrastructure → Application → Interface），建立清晰的依赖方向和职责边界。

---

## 目标架构

```
internal/
├── domain/                          # 领域层 - D&D 5e核心业务逻辑
│   ├── character/                   # 角色实体（9个文件）
│   ├── world/                       # 世界实体（scene、npc）
│   ├── monster/                     # 怪物实体
│   ├── combat/                      # 战斗状态
│   ├── quest/                       # 任务状态
│   ├── rules/                       # 规则引擎
│   ├── events/                      # 领域事件
│   └── llm/                         # LLM类型定义和Provider接口
│
├── infrastructure/                  # 基础设施层 - 外部依赖实现
│   ├── config/                      # 配置管理
│   ├── storage/                     # 存档系统（原save）
│   ├── llm/                         # LLM适配器（OpenAI/Anthropic/Ollama）
│   └── prompt/                      # 提示词构建（移除lipgloss依赖）
│
├── application/                     # 应用层 - 用例编排
│   ├── engine/                      # 游戏引擎（Agentic Loop）
│   ├── state/                       # 游戏状态管理
│   └── tools/                       # DM工具系统（25+工具）
│
└── interface/                       # 接口层 - 外部交互
    ├── cmd/                         # CLI命令（从cmd/移入）
    └── ui/                          # TUI界面（Bubble Tea）
```

## 依赖规则

```
interface     → application + infrastructure + domain
application   → domain + infrastructure(仅工厂)
infrastructure→ domain
domain        → 仅domain内 + pkg/*
```

---

## 迁移步骤

### Phase 1: Domain 层（基础）

1. **创建目录结构**
   ```bash
   mkdir -p internal/{domain/{character,world,monster,combat,quest,rules,events,llm},infrastructure/{config,storage,llm,prompt},application/{engine,state,tools},interface/{ui,cmd}}
   ```

2. **移动领域包**（无外部依赖，可并行）
   - `character` → `domain/character`
   - `world` → `domain/world`
   - `quest` → `domain/quest`
   - `rules` → `domain/rules`
   - `combat` → `domain/combat`
   - `monster` → `domain/monster`
   - `game/events.go` → `domain/events/`

3. **拆分 llm 包**
   - 提取纯类型到 `domain/llm/types.go`（Message、Response、Provider接口等）
   - 实现移至 `infrastructure/llm/`

4. **批量更新 import**
   - 全局替换 6 个领域包的 import 路径

5. **验证**：`go build ./internal/domain/...`

### Phase 2: Infrastructure 层

1. **移动基础设施包**
   - `config` → `infrastructure/config`
   - `save` → `infrastructure/storage`
   - `llm/{factory,registry,openai,anthropic,ollama}` → `infrastructure/llm/`
   - `llm/prompt` → `infrastructure/prompt`

2. **剥离 lipgloss 依赖**
   - 将 `ColorMarkerStyles` 和 `ParseColorMarkers` 移至 `interface/ui/colors.go`
   - `infrastructure/prompt` 仅保留文本模板逻辑

3. **批量更新 import**

4. **验证**：`go build ./internal/infrastructure/...`

### Phase 3: Application 层

1. **移动应用包**
   - `game/state` → `application/state`
   - `game/{engine,init}` → `application/engine/`
   - `tools` → `application/tools`

2. **批量更新 import**

3. **验证**：`go build ./internal/application/...`

### Phase 4: Interface 层

1. **移动接口包**
   - `ui` → `interface/ui`
   - `cmd/*` → `interface/cmd/`

2. **创建 `interface/ui/colors.go`**（从 prompt 分离的颜色处理）

3. **更新 cmd 的 import**

4. **验证**：`go build ./...`

### Phase 5: 清理验证

1. 删除空的原 internal 子目录
2. 运行测试：`go test ./...`
3. 运行 lint：`golangci-lint run`
4. 功能验证（角色创建、游戏启动、存档等）

---

## 关键文件

| 文件 | 重要性 | 处理要点 |
|------|--------|----------|
| `internal/llm/provider.go` | 高 | 拆分为 domain(接口+类型) + infrastructure(实现) |
| `internal/llm/prompt/builder.go` | 高 | 剥离 lipgloss，移至 interface/ui |
| `internal/game/state/state.go` | 高 | import 变更影响最多文件 |
| `internal/game/engine.go` | 高 | 更新所有子系统 import |
| `internal/tools/types.go` | 中 | StateAccessor 接口保持不变 |

---

## 验证方案

**编译验证**：
```bash
go build ./internal/domain/...
go build ./internal/infrastructure/...
go build ./internal/application/...
go build ./...
```

**依赖验证**：
```bash
# domain 不应依赖上层
go list ./internal/domain/... | xargs -I{} go list -f '{{.ImportPath}}: {{.Imports}}' {} | grep -E 'infrastructure|application|interface' && echo "FAIL" || echo "PASS"

# infrastructure 不应依赖 application/interface
go list ./internal/infrastructure/... | xargs -I{} go list -f '{{.ImportPath}}: {{.Imports}}' {} | grep -E 'application|interface' && echo "FAIL" || echo "PASS"
```

**功能验证**：
- `./cdnd character list`
- `./cdnd start --skip-creation`
- `./cdnd load 1`
