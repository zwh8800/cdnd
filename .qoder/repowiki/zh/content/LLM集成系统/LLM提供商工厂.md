# LLM提供商工厂

<cite>
**本文档引用的文件**
- [factory.go](file://internal/llm/factory.go)
- [provider.go](file://internal/llm/provider.go)
- [openai.go](file://internal/llm/openai.go)
- [anthropic.go](file://internal/llm/anthropic.go)
- [ollama.go](file://internal/llm/ollama.go)
- [config.go](file://internal/config/config.go)
- [registry.go](file://internal/llm/registry.go)
- [config.example.yaml](file://config.example.yaml)
- [load.go](file://cmd/load.go)
- [start.go](file://cmd/start.go)
- [provider.go](file://cmd/provider.go)
</cite>

## 目录
1. [简介](#简介)
2. [项目结构](#项目结构)
3. [核心组件](#核心组件)
4. [架构概览](#架构概览)
5. [详细组件分析](#详细组件分析)
6. [依赖关系分析](#依赖关系分析)
7. [性能考虑](#性能考虑)
8. [故障排除指南](#故障排除指南)
9. [结论](#结论)
10. [附录](#附录)

## 简介

CDND项目中的LLM提供商工厂是一个基于工厂模式设计的核心组件，负责根据配置动态创建和管理不同的大语言模型提供商。该工厂实现了统一的接口抽象，支持OpenAI、Anthropic和Ollama三种主流LLM服务提供商，并提供了灵活的配置管理和错误处理机制。

工厂模式在此处的应用体现了以下设计理念：
- **解耦合**：将具体的提供商实现与业务逻辑分离
- **可扩展性**：易于添加新的LLM提供商支持
- **配置驱动**：通过配置文件动态选择和配置提供商
- **统一接口**：为不同提供商提供一致的API接口

## 项目结构

LLM提供商工厂位于internal/llm目录下，主要包含以下关键文件：

```mermaid
graph TB
subgraph "LLM工厂模块"
F[factory.go<br/>工厂实现]
P[provider.go<br/>接口定义]
R[registry.go<br/>注册表管理]
end
subgraph "提供商实现"
O[openai.go<br/>OpenAI实现]
A[anthropic.go<br/>Anthropic实现]
L[ollama.go<br/>Ollama实现]
end
subgraph "配置管理"
C[config.go<br/>配置结构]
E[config.example.yaml<br/>示例配置]
end
subgraph "命令行集成"
S[start.go<br/>启动命令]
G[load.go<br/>加载命令]
T[provider.go<br/>提供商命令]
end
F --> P
F --> O
F --> A
F --> L
F --> C
R --> P
O --> P
A --> P
L --> P
S --> F
G --> F
T --> F
```

**图表来源**
- [factory.go:1-69](file://internal/llm/factory.go#L1-L69)
- [provider.go:1-114](file://internal/llm/provider.go#L1-L114)
- [openai.go:1-257](file://internal/llm/openai.go#L1-L257)
- [anthropic.go:1-269](file://internal/llm/anthropic.go#L1-L269)
- [ollama.go:1-261](file://internal/llm/ollama.go#L1-L261)

**章节来源**
- [factory.go:1-69](file://internal/llm/factory.go#L1-L69)
- [provider.go:1-114](file://internal/llm/provider.go#L1-L114)

## 核心组件

### 工厂接口和数据结构

工厂模式的核心是Provider接口，它定义了所有LLM提供商必须实现的标准方法：

```mermaid
classDiagram
class Provider {
<<interface>>
+Name() string
+Generate(ctx, req) Response
+GenerateStream(ctx, req) StreamChunk
+SetModel(model string)
+SetMaxTokens(maxTokens int)
+SetTemperature(temp float64)
}
class ProviderConfig {
+string APIKey
+string Model
+string BaseURL
+int MaxTokens
+float64 Temperature
}
class OpenAIProvider {
-Client client
-string model
-int maxTokens
-float64 temperature
-string baseURL
+Name() string
+Generate(ctx, req) Response
+GenerateStream(ctx, req) StreamChunk
+SetModel(model string)
+SetMaxTokens(maxTokens int)
+SetTemperature(temp float64)
}
class AnthropicProvider {
-Client client
-string model
-int maxTokens
-float64 temperature
+Name() string
+Generate(ctx, req) Response
+GenerateStream(ctx, req) StreamChunk
+SetModel(model string)
+SetMaxTokens(maxTokens int)
+SetTemperature(temp float64)
}
class OllamaProvider {
-Client client
-string model
-int maxTokens
-float64 temperature
-string baseURL
+Name() string
+Generate(ctx, req) Response
+GenerateStream(ctx, req) StreamChunk
+SetModel(model string)
+SetMaxTokens(maxTokens int)
+SetTemperature(temp float64)
}
Provider <|.. OpenAIProvider
Provider <|.. AnthropicProvider
Provider <|.. OllamaProvider
ProviderConfig --> OpenAIProvider
ProviderConfig --> AnthropicProvider
ProviderConfig --> OllamaProvider
```

**图表来源**
- [provider.go:64-83](file://internal/llm/provider.go#L64-L83)
- [provider.go:85-92](file://internal/llm/provider.go#L85-L92)
- [openai.go:11-18](file://internal/llm/openai.go#L11-L18)
- [anthropic.go:11-17](file://internal/llm/anthropic.go#L11-L17)
- [ollama.go:11-19](file://internal/llm/ollama.go#L11-L19)

### 工厂函数实现

工厂模式的两个核心函数分别处理不同的创建场景：

#### NewProvider函数
该函数使用配置中的默认提供商进行实例化：

```mermaid
sequenceDiagram
participant Client as "调用方"
participant Factory as "NewProvider"
participant Config as "配置系统"
participant Provider as "具体提供商"
Client->>Factory : 调用NewProvider(cfg)
Factory->>Config : 检查默认提供商
Config-->>Factory : 返回默认提供商名称
Factory->>Config : 获取提供商配置
Config-->>Factory : 返回ProviderConfig
Factory->>Factory : 转换配置格式
Factory->>Factory : 根据类型创建实例
alt OpenAI
Factory->>Provider : NewOpenAIProvider(llmCfg)
else Anthropic
Factory->>Provider : NewAnthropicProvider(llmCfg)
else Ollama
Factory->>Provider : NewOllamaProvider(llmCfg)
else 未知提供商
Factory-->>Client : 返回错误
end
Provider-->>Factory : 返回Provider实例
Factory-->>Client : 返回Provider实例
```

**图表来源**
- [factory.go:9-41](file://internal/llm/factory.go#L9-L41)

#### NewProviderByName函数
该函数允许按指定名称创建提供商实例：

```mermaid
flowchart TD
Start([函数调用]) --> GetConfig["获取指定名称的配置"]
GetConfig --> CheckExists{"配置是否存在?"}
CheckExists --> |否| ReturnError["返回'配置不存在'错误"]
CheckExists --> |是| ConvertConfig["转换配置格式"]
ConvertConfig --> SwitchType["根据名称选择提供商类型"]
SwitchType --> |openai| CreateOpenAI["创建OpenAI提供商"]
SwitchType --> |anthropic| CreateAnthropic["创建Anthropic提供商"]
SwitchType --> |ollama| CreateOllama["创建Ollama提供商"]
SwitchType --> |其他| ReturnUnknown["返回'未知提供商'错误"]
CreateOpenAI --> ReturnProvider["返回提供商实例"]
CreateAnthropic --> ReturnProvider
CreateOllama --> ReturnProvider
ReturnError --> End([结束])
ReturnUnknown --> End
ReturnProvider --> End
```

**图表来源**
- [factory.go:43-68](file://internal/llm/factory.go#L43-L68)

**章节来源**
- [factory.go:9-68](file://internal/llm/factory.go#L9-L68)

## 架构概览

LLM提供商工厂采用分层架构设计，确保了良好的可维护性和扩展性：

```mermaid
graph TB
subgraph "应用层"
CLI[命令行界面]
Game[游戏引擎]
end
subgraph "工厂层"
Factory[LLM工厂]
Registry[提供商注册表]
end
subgraph "接口层"
ProviderInterface[Provider接口]
ConfigInterface[配置接口]
end
subgraph "实现层"
OpenAI[OpenAI实现]
Anthropic[Anthropic实现]
Ollama[Ollama实现]
end
subgraph "外部服务"
OpenAIService[OpenAI API]
AnthropicService[Anthropic API]
OllamaService[Ollama服务]
end
CLI --> Factory
Game --> Factory
Factory --> ProviderInterface
Factory --> Registry
Registry --> ProviderInterface
ProviderInterface --> OpenAI
ProviderInterface --> Anthropic
ProviderInterface --> Ollama
OpenAI --> OpenAIService
Anthropic --> AnthropicService
Ollama --> OllamaService
```

**图表来源**
- [factory.go:1-69](file://internal/llm/factory.go#L1-L69)
- [registry.go:1-140](file://internal/llm/registry.go#L1-L140)
- [provider.go:64-83](file://internal/llm/provider.go#L64-L83)

### 配置管理系统

配置系统提供了灵活的配置管理机制：

```mermaid
erDiagram
CONFIG {
string default_provider
map providers
}
PROVIDER_CONFIG {
string api_key
string model
string base_url
int max_tokens
float temperature
}
LLM_CONFIG {
string default_provider
map providers
}
CONFIG ||--|| LLM_CONFIG : "包含"
LLM_CONFIG ||--o{ PROVIDER_CONFIG : "配置多个提供商"
```

**图表来源**
- [config.go:8-29](file://internal/config/config.go#L8-L29)

**章节来源**
- [config.go:8-29](file://internal/config/config.go#L8-L29)

## 详细组件分析

### OpenAI提供商实现

OpenAI提供商实现了完整的LLM功能，包括标准生成和流式生成：

#### 核心特性
- **API兼容性**：完全兼容OpenAI Chat Completions API
- **工具调用支持**：支持函数调用和工具集成
- **流式响应**：支持实时流式响应处理
- **配置灵活性**：支持自定义BaseURL进行代理或兼容

#### 实现特点
```mermaid
sequenceDiagram
participant App as "应用程序"
participant OpenAI as "OpenAI提供商"
participant Client as "OpenAI客户端"
participant API as "OpenAI API"
App->>OpenAI : Generate(ctx, request)
OpenAI->>OpenAI : 转换消息格式
OpenAI->>OpenAI : 处理工具调用
OpenAI->>Client : CreateChatCompletion()
Client->>API : HTTP请求
API-->>Client : 响应数据
Client-->>OpenAI : ChatCompletionResponse
OpenAI->>OpenAI : 解析响应
OpenAI->>OpenAI : 处理工具调用
OpenAI-->>App : Response对象
```

**图表来源**
- [openai.go:41-125](file://internal/llm/openai.go#L41-L125)

**章节来源**
- [openai.go:11-257](file://internal/llm/openai.go#L11-L257)

### Anthropic提供商实现

Anthropic提供商专注于Claude系列模型的支持：

#### 核心特性
- **系统提示支持**：原生支持system角色消息
- **工具调用集成**：支持Claude特定的工具调用格式
- **流式处理**：完整的流式响应支持
- **参数优化**：针对Claude模型的参数优化

#### 实现差异
```mermaid
flowchart TD
Message[输入消息] --> RoleCheck{检查消息角色}
RoleCheck --> |User| UserMsg[转换为User消息]
RoleCheck --> |Assistant| AssistMsg[转换为Assistant消息]
RoleCheck --> |System| SystemMsg[提取为系统提示]
RoleCheck --> |Tool| ToolMsg[转换为Tool结果消息]
UserMsg --> BuildParams[构建消息参数]
AssistMsg --> BuildParams
SystemMsg --> BuildParams
ToolMsg --> BuildParams
BuildParams --> SendRequest[发送请求]
SendRequest --> ParseResponse[解析响应]
ParseResponse --> ExtractContent[提取内容]
ExtractContent --> ExtractTools[提取工具调用]
ExtractTools --> ReturnResponse[返回响应]
```

**图表来源**
- [anthropic.go:41-139](file://internal/llm/anthropic.go#L41-L139)

**章节来源**
- [anthropic.go:11-269](file://internal/llm/anthropic.go#L11-L269)

### Ollama提供商实现

Ollama提供商通过OpenAI兼容模式实现本地模型支持：

#### 设计理念
- **兼容性优先**：利用OpenAI SDK进行本地模型访问
- **零配置启动**：默认指向本地11434端口
- **无缝集成**：对上层应用透明，无需修改代码

#### 实现机制
```mermaid
graph LR
subgraph "Ollama提供商"
OllamaClient[OpenAI兼容客户端]
LocalServer[本地Ollama服务器]
end
subgraph "配置处理"
BaseURL[BaseURL配置]
DefaultURL[默认localhost:11434]
end
BaseURL --> |存在| OllamaClient
BaseURL --> |不存在| DefaultURL
DefaultURL --> OllamaClient
OllamaClient --> LocalServer
```

**图表来源**
- [ollama.go:21-38](file://internal/llm/ollama.go#L21-L38)

**章节来源**
- [ollama.go:11-261](file://internal/llm/ollama.go#L11-L261)

### 注册表管理器

注册表提供了动态提供商管理能力：

#### 核心功能
- **并发安全**：使用读写锁保证线程安全
- **动态注册**：支持运行时注册和注销提供商
- **默认提供商管理**：自动管理默认提供商状态
- **全局访问**：提供全局静态方法便于使用

**章节来源**
- [registry.go:8-140](file://internal/llm/registry.go#L8-L140)

## 依赖关系分析

### 外部依赖管理

工厂模式有效隔离了外部依赖：

```mermaid
graph TB
subgraph "内部依赖"
Factory[LLM工厂]
Provider[Provider接口]
Config[配置结构]
end
subgraph "外部SDK"
OpenAISDK[OpenAI SDK]
AnthropicSDK[Anthropic SDK]
end
subgraph "Go标准库"
Context[context包]
IO[io包]
JSON[encoding/json包]
end
Factory --> Provider
Factory --> Config
OpenAIImpl[OpenAI实现] --> OpenAISDK
OpenAIImpl --> Context
OpenAIImpl --> IO
AnthropicImpl[Anthropic实现] --> AnthropicSDK
AnthropicImpl --> Context
AnthropicImpl --> JSON
```

**图表来源**
- [openai.go:3-9](file://internal/llm/openai.go#L3-L9)
- [anthropic.go:3-9](file://internal/llm/anthropic.go#L3-L9)

### 内部模块交互

```mermaid
sequenceDiagram
participant CMD as "命令行"
participant CFG as "配置系统"
participant FACT as "工厂"
participant PROV as "提供商"
participant GAME as "游戏引擎"
CMD->>CFG : 获取配置
CFG-->>CMD : 返回配置
CMD->>FACT : 创建提供商
FACT->>FACT : 验证配置
FACT->>PROV : 实例化提供商
PROV-->>FACT : 返回实例
FACT-->>CMD : 返回提供商
CMD->>GAME : 创建游戏引擎
GAME->>PROV : 使用提供商
```

**图表来源**
- [load.go:29-34](file://cmd/load.go#L29-L34)
- [start.go:32-37](file://cmd/start.go#L32-L37)

**章节来源**
- [load.go:1-120](file://cmd/load.go#L1-L120)
- [start.go:1-99](file://cmd/start.go#L1-L99)
- [provider.go:53-94](file://cmd/provider.go#L53-L94)

## 性能考虑

### 并发安全性

工厂和注册表都采用了读写锁机制来保证并发安全：

- **读操作优化**：多个读取操作可以并行执行
- **写操作保护**：注册和注销操作互斥执行
- **最小锁范围**：尽量缩短持有锁的时间

### 内存管理

- **连接池**：SDK客户端通常内置连接池管理
- **流式处理**：流式响应使用通道避免大量内存占用
- **及时释放**：流式连接在完成后及时关闭

### 配置缓存

- **配置预处理**：在工厂中完成配置转换，避免重复计算
- **实例复用**：提供商实例可以在应用生命周期内复用

## 故障排除指南

### 常见错误类型

#### 配置验证错误
```mermaid
flowchart TD
Start([开始验证]) --> CheckDefault["检查默认提供商"]
CheckDefault --> HasDefault{"是否有默认提供商?"}
HasDefault --> |否| NoDefault["返回'无默认提供商'错误"]
HasDefault --> |是| GetConfig["获取提供商配置"]
GetConfig --> ConfigExists{"配置是否存在?"}
ConfigExists --> |否| ConfigNotFound["返回'提供商未找到'错误"]
ConfigExists --> |是| ConvertSuccess["配置转换成功"]
ConvertSuccess --> CreateProvider["创建提供商实例"]
CreateProvider --> ProviderExists{"提供商类型有效?"}
ProviderExists --> |否| UnknownProvider["返回'未知提供商'错误"]
ProviderExists --> |是| Success["创建成功"]
NoDefault --> End([结束])
ConfigNotFound --> End
UnknownProvider --> End
Success --> End
```

**图表来源**
- [factory.go:11-40](file://internal/llm/factory.go#L11-L40)

#### 调试技巧

1. **配置验证**
   - 检查配置文件格式正确性
   - 验证API密钥的有效性
   - 确认网络连接正常

2. **日志记录**
   ```bash
   # 设置详细日志级别
   export LOG_LEVEL=debug
   
   # 测试特定提供商
   cdnd provider test openai
   ```

3. **环境变量**
   - OpenAI: `OPENAI_API_KEY`
   - Anthropic: `ANTHROPIC_API_KEY`
   - Ollama: 本地服务无需API密钥

**章节来源**
- [factory.go:11-40](file://internal/llm/factory.go#L11-L40)
- [provider.go:53-94](file://cmd/provider.go#L53-L94)

## 结论

CDND项目的LLM提供商工厂展现了优秀的软件工程实践：

### 设计优势
- **高度解耦**：通过接口抽象实现了良好的模块分离
- **灵活扩展**：新增提供商只需实现Provider接口
- **配置驱动**：支持动态配置和热切换
- **错误处理**：完善的错误传播和处理机制

### 最佳实践建议
1. **配置管理**：使用配置文件集中管理提供商设置
2. **错误处理**：在调用方妥善处理工厂返回的错误
3. **资源管理**：合理管理提供商实例的生命周期
4. **监控指标**：添加使用统计和性能监控

### 扩展指南
- 新增提供商时遵循现有接口规范
- 在config.example.yaml中添加示例配置
- 编写相应的单元测试和集成测试
- 更新命令行工具的文档和帮助信息

## 附录

### 支持的提供商列表

| 提供商 | 类型 | 特殊要求 | 默认端点 |
|--------|------|----------|----------|
| openai | 在线API | API密钥 | https://api.openai.com |
| anthropic | 在线API | API密钥 | https://api.anthropic.com |
| ollama | 本地服务 | 本地安装 | http://localhost:11434 |

### 配置示例

完整的配置示例可在config.example.yaml中找到，包含三个提供商的完整配置模板。