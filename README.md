# eino-stock — AI 股票选股分析平台

基于 **Kratos** + **Vue 3** 的智能股票行情与 AI 选股平台。后端 Go 语言构建，前端 Vue 3 SPA，集成多个数据源和 LLM 选股能力。

---

## 功能特性

- **AI 智能选股** — 自然语言输入选股条件，LLM 自动解析并调用选股工具返回结果
- **多维度并行选股** — 将用户条件自动拆分到估值、财务、市场三个维度并行筛选，取交集结果
- **AI 对话助手** — 流式 SSE 对话，支持多轮工具调用（选股/行情/公告/研报等）
- **实时行情** — 支持 A 股、港股实时行情，新浪/腾讯双数据源
- **K 线数据** — 东方财富主源 + 新浪备用，支持多周期（分钟/日/周/月/年）
- **股票关注** — SQLite 持久化关注列表
- **板块/ETF 搜索** — 东方财富选股器接口
- **热门策略** — 东方财富智能选股热门策略推荐
- **数据工具集** — 分时数据、公告、研报、全球指数、涨停热门板块、龙虎榜、行业估值/资金排名

---

## 技术栈

| 层次   | 技术                                                                 |
|--------|----------------------------------------------------------------------|
| 后端   | Go, [Kratos v2](https://github.com/go-kratos/kratos), GORM, Wire    |
| 数据库 | SQLite（[glebarez/sqlite](https://github.com/glebarez/sqlite)）      |
| 前端   | Vue 3 (Composition API + \<script setup\>), Vite, vue-router          |
| AI     | 字节跳动 [CloudWeGo Eino](https://github.com/cloudwego/eino) 框架     |
| LLM    | DeepSeek Chat（兼容 OpenAI API 格式）                               |
| 行情源 | 新浪财经、腾讯证券、东方财富、财联社                                  |

---

## 项目结构

```
eino-stock/
├── api/market/v1/                        # protobuf 定义的 API 层
│   ├── market.pb.go                      # protobuf 生成的消息结构
│   ├── market_grpc.pb.go                 # gRPC 服务定义
│   └── market_http.pb.go                 # HTTP 服务定义（注册到 Kratos）
│
├── cmd/eino-stock/                       # 应用入口
│   ├── main.go                           # Kratos App 启动、配置加载
│   ├── wire.go                           # Wire 依赖注入声明
│   └── wire_gen.go                       # Wire 生成代码
│
├── configs/
│   └── config.yaml                       # 主配置文件（服务端口、数据源、LLM 配置）
│
├── internal/
│   ├── biz/                              # 业务逻辑层（Usecase）
│   │   ├── biz.go                        # Wire ProviderSet
│   │   ├── ai/chat.go                    # AI 选股/并行选股 Usecase
│   │   ├── follow/follow.go              # 关注管理 Usecase
│   │   ├── market/market.go              # 行情查询 Usecase（Stock/Quote/KLine）
│   │   ├── market/kline.go               # KLine 模型 + KLineProvider 接口
│   │   ├── market/code.go                # 股票代码归一化工具
│   │   ├── market/code_test.go           # 代码归一化测试
│   │   └── screen/screen.go              # 板块/ETF/热门策略 Usecase
│   │
│   ├── conf/
│   │   └── conf.pb.go                    # 配置 protobuf 消息（Server/Data/DataSource）
│   │
│   ├── data/                             # 数据持久化层
│   │   ├── data.go                       # Data 初始化（SQLite + 自动迁移 + 种子数据）
│   │   ├── market.go                     # MarketRepo 实现（股票搜索）
│   │   ├── follow.go                     # FollowRepo 实现（关注 CRUD）
│   │   ├── seed.go                       # 默认 25 只股票种子数据
│   │   └── model/stock.go                # StockBasic GORM 模型
│   │
│   ├── infrastructure/                   # 基础设施层
│   │   ├── provider.go                   # Wire ProviderSet（行情/K线/选股客户端工厂）
│   │   ├── eastmoney/
│   │   │   ├── kline.go                  # 东方财富 K 线 API 客户端
│   │   │   └── screen.go                 # 东方财富选股器 API 客户端（板块/ETF/热门策略）
│   │   ├── sina/
│   │   │   └── kline.go                  # 新浪财经 K 线 API 客户端
│   │   ├── quote/
│   │   │   ├── client.go                 # 实时行情客户端（新浪+腾讯双源）
│   │   │   └── parse.go                  # 行情数据解析 + GB18030→UTF-8 转换
│   │   └── eino/                         # AI 基础设施层
│   │       ├── config.go                 # AI 配置（DeepSeek API Key/URL/Model）
│   │       ├── chat.go                   # ChatAgent — 对话模式（直接 API 调用 + 工具调用循环）
│   │       ├── parser.go                 # 用户选股条件 LLM 解析
│   │       ├── screener.go               # 单轮选股模式（Eino React Agent）
│   │       ├── parallel.go               # 多维度并行选股（估值/财务/市场三路并发）
│   │       └── tools/                    # AI 工具集（Eino Tool 实现）
│   │           ├── registry.go           # 工具注册表
│   │           ├── iwencai.go            # SelectAStock — 东方财富 iwencai 选股
│   │           ├── market_data.go        # GetMarketData — 财联社市场概况
│   │           ├── global_indexes.go     # GetGlobalStockIndexes — 全球指数
│   │           ├── minute_data.go        # GetStockMinuteData — 分时数据
│   │           ├── stock_detail.go       # GetStockDetail — 五档盘口
│   │           ├── stock_notice.go       # GetStockNotice — 公司公告
│   │           ├── research_report.go    # GetStockResearchReport — 研报
│   │           ├── uplimit_hot_plates.go  # GetUplimitHotPlates — 涨停热门板块
│   │           ├── long_tiger_list.go    # GetLongTigerList — 龙虎榜
│   │           ├── industry_valuation.go  # GetIndustryValuation — 行业估值
│   │           └── industry_money_rank.go # GetIndustryMoneyRank — 行业资金排名
│   │
│   ├── server/                           # 服务层
│   │   ├── server.go                     # Wire ProviderSet
│   │   ├── http.go                       # HTTP 路由注册（AI + Market + Screen + Tool + Follow + WebUI）
│   │   ├── grpc.go                       # gRPC 服务注册
│   │   └── web.go                        # 嵌入的静态 HTML 页面（embed）
│   │
│   └── service/                          # HTTP 处理器层
│       ├── service.go                    # Wire ProviderSet
│       ├── helper.go                     # writeJSON / writeError 工具函数
│       ├── market.go                     # 行情相关处理器（搜索/报价/K线）
│       ├── screen.go                     # 板块/ETF/热门策略处理器
│       ├── ai.go                         # AI 处理器（筛选/并行筛选/流式对话）
│       ├── follow.go                     # 关注处理器（列表/添加/删除）
│       └── tool.go                       # 工具直调处理器（绕过 AI，直接返回数据工具结果）
│
└── frontend/                             # Vue 3 前端 SPA
    ├── src/
    │   ├── main.ts                       # Vue 应用入口 + vue-router 路由
    │   ├── App.vue                       # 主布局（顶栏 + 内容区 + 底部 TabBar）
    │   ├── api/index.ts                  # API 客户端封装（fetch + 所有接口）
    │   ├── types/index.ts                # TypeScript 类型定义
    │   └── views/
    │       ├── Home.vue                  # 首页：热门策略 + 搜索框 + 股票卡片网格
    │       ├── StockDetail.vue           # 股票详情：行情卡片 + 分时/公告/研报 Tab
    │       └── Follow.vue                # 关注列表：卡片式布局 + 取消关注
    ├── env.d.ts                          # 类型声明
    └── vite.config.ts                    # Vite 配置 + API 代理到 localhost:8000
```

---

## 快速启动

### 前置条件

- Go 1.22+
- Node.js 18+（前端开发）
- LLM API Key（DeepSeek 或其他兼容 OpenAI 的 API）

### 后端启动

```bash
cd eino-stock
go mod tidy
go build -o bin/eino-stock.exe ./cmd/eino-stock

# 方式一：从环境变量读取 LLM API Key
set LLM_API_KEY=sk-your-key
./bin/eino-stock.exe -conf ./configs

# 方式二：直接配置在 configs/config.yaml 中
# 编辑 configs/config.yaml 设置 ai.llm_api_key
./bin/eino-stock.exe -conf ./configs
```

服务器默认监听 `:8000`（HTTP）和 `:9000`（gRPC）。

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器默认 `:5173`，API 请求代理到 `:8000`。

### 访问

| 方式          | 地址                         |
|---------------|------------------------------|
| Vue 前端 SPA  | http://localhost:5173        |
| 嵌入 HTML     | http://localhost:8000/       |
| AI 对话       | http://localhost:8000/ai.html |
| K 线图        | http://localhost:8000/kline.html |
| 选股器        | http://localhost:8000/screen.html |

---

## 配置说明

`configs/config.yaml`：

| 配置项                          | 说明                        | 默认值                        |
|---------------------------------|-----------------------------|-------------------------------|
| `server.http.addr`              | HTTP 监听地址               | `0.0.0.0:8000`               |
| `server.grpc.addr`              | gRPC 监听地址               | `0.0.0.0:9000`               |
| `data.database.source`          | SQLite 数据库路径           | `./data/eino-stock.db`       |
| `data_source.http_timeout`      | HTTP 请求超时               | `15s`                        |
| `data_source.qgqp_b_id`        | 东方财富选股器 fingerprint  | —                            |
| `ai.llm_api_key`               | LLM API Key                 | `LLM_API_KEY` 环境变量优先   |
| `ai.llm_base_url`              | LLM API 地址                | `https://api.deepseek.com`   |
| `ai.llm_model`                 | LLM 模型名                  | `deepseek-chat`              |

> `qgqp_b_id` 从东方财富选股器页面（`xuangu.eastmoney.com`）按 F12 → Application → Cookies 获取。

---

## API

### 行情与标的

| 方法 | 路径                              | 说明                           |
|------|-----------------------------------|--------------------------------|
| GET  | `/api/market/stocks?keyword=&limit=` | 搜索股票（支持代码/名称模糊） |
| GET  | `/api/market/quote/{code}`        | 单只实时行情                   |
| GET  | `/api/market/kline/{code}?ktype=101&limit=120` | K 线数据               |
| GET  | `/api/screen/bk/{keyword}?limit=` | 搜索板块                       |
| GET  | `/api/screen/etf/{keyword}?limit=` | 搜索 ETF                      |
| GET  | `/api/screen/hot-strategy`        | 热门选股策略                   |

### AI 选股与对话

| 方法 | 路径                              | 说明                           |
|------|-----------------------------------|--------------------------------|
| GET  | `/api/ai/screen?q=`              | AI 单轮选股                    |
| GET  | `/api/ai/parallel?q=`            | AI 多维并行选股                |
| GET  | `/api/ai/chat?q=`                | AI 流式对话（SSE）             |

### 数据工具（直调，绕过 AI）

| 方法 | 路径                              | 说明                           |
|------|-----------------------------------|--------------------------------|
| GET  | `/api/tool/screen?q=`            | 东方财富 iwencai 选股          |
| GET  | `/api/tool/screen-v2?q=`         | 选股 v2（结构化 JSON 返回）    |
| GET  | `/api/tool/minute?code=`         | 分时数据                       |
| GET  | `/api/tool/detail?code=`         | 五档盘口详情                   |
| GET  | `/api/tool/notice?code=`         | 公司公告                       |
| GET  | `/api/tool/report?code=&days=`   | 研报                           |
| GET  | `/api/tool/global-indexes`       | 全球指数                       |
| GET  | `/api/tool/hot-plates`           | 涨停热门板块                   |
| GET  | `/api/tool/long-tiger`           | 龙虎榜                         |
| GET  | `/api/tool/industry-valuation`   | 行业估值                       |
| GET  | `/api/tool/industry-money-rank`  | 行业资金排名                   |

### 关注管理

| 方法 | 路径                              | 说明                           |
|------|-----------------------------------|--------------------------------|
| GET  | `/api/follow/list`               | 关注列表                       |
| POST | `/api/follow/add?code=&name=`    | 添加关注                       |
| POST | `/api/follow/remove?code=`       | 取消关注                       |

---

## 数据源

| 数据源     | 用途                         | 接口地址                              |
|------------|------------------------------|---------------------------------------|
| 新浪财经   | 实时行情、股票详情（五档）   | `hq.sinajs.cn`                        |
| 腾讯证券   | 实时行情（港股）             | `qt.gtimg.cn`                         |
| 东方财富   | K 线、选股器、公告、研报     | `push2his.eastmoney.com`, `np-tjxg-g.eastmoney.com` |
| 财联社     | 市场概况（指数 + 涨跌分布）  | `x-quote.cls.cn`                      |
| 东方财富 iwencai | 选股引擎                  | `iwencai.com`                         |

---

## AI 选股架构

项目提供三种 AI 选股模式：

1. **单轮筛选** (`/api/ai/screen`)：用户输入中文条件 → LLM 解析为结构化的 iwencai 查询 → Eino React Agent 调用 `SelectAStock` 工具 → 返回文本结果
2. **多维并行筛选** (`/api/ai/parallel`)：LLM 解析后，将条件自动拆分到 **估值**、**财务**、**市场** 三个维度 → 三路并行查询 → 取交集（同时满足所有维度的股票）
3. **对话模式** (`/api/ai/chat`)：完整的多轮工具调用循环，支持选股、行情、公告、研报等 12 种工具的自动调用，SSE 流式输出

---

## 开发

```bash
# Wire 重新生成依赖注入
cd cmd/eino-stock && wire

# 运行测试
go test ./internal/...
```

---

*最后更新: 2026-06-09*
