# AI 小说转剧本工具

AI 小说转剧本工具用于将中文小说 TXT 文本解析为章节，并借助 OpenAI 兼容接口生成结构化 YAML 剧本初稿。系统面向小说作者、编剧和内容改编团队，重点解决长文本导入、章节选择、AI 分块生成、结果校验、YAML 下载和任务结果持久化。

## 功能概览

- TXT 小说导入：支持粘贴正文或上传 TXT，单文件上限 10MB。
- 编码自动识别：支持 UTF-8、UTF-16、GB18030、GBK、Big5 等常见中文文本编码。
- 章节解析：自动识别章节边界、标题、字数和中文比例。
- 章节确认：章节过多时分页展示，可选择连续章节生成剧本。
- 生成约束：单次至少选择 3 个连续章节，所选章节总字数不得超过 100000 字。
- AI 设置：前端可配置 Base URL、模型名和 API Key，并支持联通测试。
- OpenAI 兼容接口：支持 OpenAI、DashScope 兼容模式及其他兼容 Chat Completions 的服务。
- 长文本分块：后端按章节和段落拆分文本，分块调用 AI，再合并为完整 YAML。
- YAML 编辑与校验：任务完成后可查看、编辑、重新校验并下载 YAML 剧本初稿。
- YAML 帮助文档：任务详情页可下载 YAML 字段说明 PDF。
- 任务持久化：转换任务、进度、Draft 和 YAML 结果写入 PostgreSQL，服务重启后可恢复。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3, TypeScript, Vite, Pinia, Vue Router |
| 后端 | Go, Gin |
| AI 接入 | OpenAI Compatible Chat Completions |
| 数据存储 | PostgreSQL |
| 辅助服务 | Redis |
| 部署 | Docker Compose, Nginx |

## 系统架构

```text
Browser
  |
  v
Nginx :8088
  |-- Frontend Vue static assets
  |
  `-- /api/* -> Gin backend
                 |
                 |-- Chapter parser
                 |-- YAML validator and repair
                 |-- OpenAI compatible AI client
                 |-- Task service with chunked generation
                 |
                 `-- PostgreSQL conversion_tasks
```

## 快速启动

### 环境要求

- Docker Desktop 或 Docker Engine
- Docker Compose
- Node.js 24+，仅本地前端开发时需要
- Go 1.26+，仅本地后端开发时需要

### 使用 Docker Compose 启动

```bash
cd deploy
docker compose up -d --build
```

启动后访问：

```text
http://localhost:8088
```

常用命令：

```bash
cd deploy
docker compose ps
docker compose logs -f backend
docker compose logs -f frontend
docker compose stop
```

## 使用流程

1. 打开首页，粘贴小说正文或上传 TXT 文件。
2. 在“模型设置”中填写 Base URL、模型名和 API Key，可点击“测试联通”。
3. 点击“解析章节”，系统会识别章节标题、字数和中文比例。
4. 在章节确认页选择连续章节范围。
5. 点击“生成所选章节”，进入转换任务详情页。
6. 等待任务完成后，在 YAML 结果区编辑、校验或下载 YAML。
7. 如需理解 YAML 字段，点击“下载 YAML 帮助 PDF”。

## AI 配置说明

前端配置保存在浏览器 localStorage 中，创建任务时随请求发送给后端。后端使用 OpenAI 兼容 Chat Completions 接口。

| 配置项 | 说明 | 示例 |
| --- | --- | --- |
| Base URL | AI 服务地址，可填写域名、`/v1` 地址或完整接口地址 | `https://api.openai.com/v1` |
| 模型 | 模型名称 | `gpt-4.1-mini` |
| API Key | 服务商提供的密钥 | `sk-...` |

Base URL 会自动规范化：

- 未填写路径时：自动补为 `/v1/chat/completions`
- 以 `/v1` 结尾时：自动补为 `/chat/completions`
- 已包含 `/v1/xxx` 时：按用户填写的完整路径请求，不再追加 `/v1/chat/completions`

安全说明：

- API Key 不会出现在任务列表或任务详情接口响应中。
- 任务完成或失败后，后端会清空任务中的 API Key，避免长期持久化密钥。
- 不建议在公开环境中使用共享密钥直接暴露给普通用户。

## 数据持久化

Docker Compose 默认启用 PostgreSQL：

```yaml
DATABASE_URL=postgres://screenplay:screenplay@postgres:5432/screenplay?sslmode=disable
```

后端启动时会自动创建 `conversion_tasks` 表。任务创建、进度更新、完成后的 Draft 和 YAML 都会写入数据库。

如果不配置 `DATABASE_URL`，后端会退回内存仓储，适合单元测试或临时开发；此时服务重启后任务会丢失。

## YAML 剧本结构

生成结果使用 YAML，主要字段如下：

| 字段 | 作用 |
| --- | --- |
| `schema_version` | YAML 结构版本 |
| `project` | 项目信息，例如标题、作者和生成时间 |
| `source` | 本次转换的章节范围 |
| `adaptation` | 改编格式、故事钩子、梗概和主题 |
| `characters` | 人物表 |
| `scenes` | 场景列表，包含来源章节、地点、人物、戏剧目的和节拍 |
| `continuity` | 可选的连续性信息，例如时间线、伏笔和待解决问题 |
| `quality_report` | 覆盖率、警告和人工复核项 |

完整字段解释可在任务详情页下载：

```text
/docs/yaml-screenplay-guide.pdf
```

## API 概览

所有接口都挂载在 `/api` 下。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/healthz` | 健康检查 |
| POST | `/api/chapters/parse` | 解析小说章节 |
| POST | `/api/ai/test` | 测试 AI 配置联通 |
| GET | `/api/conversion-tasks` | 查询转换任务列表 |
| POST | `/api/conversion-tasks` | 创建转换任务 |
| GET | `/api/conversion-tasks/:id` | 查询任务详情 |
| GET | `/api/conversion-tasks/:id/export` | 下载任务生成的 YAML |
| POST | `/api/yaml/validate` | 校验 YAML 结构 |

创建任务请求示例：

```json
{
  "source_text": "第一章\n正文...\n\n第二章\n正文...\n\n第三章\n正文...",
  "ai_config": {
    "provider": "openai_compatible",
    "base_url": "https://api.openai.com/v1",
    "model": "gpt-4.1-mini",
    "api_key": "sk-..."
  },
  "chapters": [
    {
      "id": "CH001",
      "title": "第一章",
      "word_count": 3200,
      "body": "第一章正文..."
    }
  ]
}
```

## 本地开发

### 前端

```bash
cd frontend
npm install
npm run dev
```

前端默认通过 Vite 开发服务器运行。生产环境由 Docker 构建静态资源并交给 Nginx。

### 后端

```bash
cd backend
go test ./...
go run ./cmd/server
```

本地直接运行后端时，可按需设置环境变量：

```bash
export PORT=8080
export APP_ENV=development
export APP_VERSION=0.1.0
export DATABASE_URL='postgres://screenplay:screenplay@localhost:5432/screenplay?sslmode=disable'
go run ./cmd/server
```

## 测试与构建

前端：

```bash
cd frontend
npm test
npm run build
```

后端：

```bash
cd backend
go test ./...
```

Docker：

```bash
cd deploy
docker compose build frontend
docker compose build backend
docker compose up -d
```

## 项目结构

```text
.
├── backend
│   ├── cmd/server              # Gin 服务入口
│   └── internal
│       ├── ai                  # OpenAI 兼容客户端、本地适配器和提示词
│       ├── api                 # HTTP 路由与处理器
│       ├── config              # 环境配置
│       ├── domain              # 剧本和任务领域模型
│       ├── repository          # 任务仓储与 PostgreSQL 持久化
│       ├── schema              # YAML 校验、修复和质量报告
│       └── service             # 章节解析、任务生成和分块合并
├── deploy
│   ├── docker-compose.yml      # 本地部署编排
│   └── nginx.conf              # 前后端反向代理配置
├── docs                        # 阶段文档
├── frontend
│   ├── public/docs             # YAML 帮助 HTML/PDF
│   └── src
│       ├── services            # API 客户端
│       ├── stores              # Pinia 状态
│       ├── utils               # TXT 解码和章节选择工具
│       └── views               # 页面组件
└── AI长文本剧本生成优化设计.md
```

## 常见问题

### 任务完成后重启服务，记录还在吗？

使用 Docker Compose 默认配置时会保存到 PostgreSQL，重启 backend 后仍可恢复任务列表和详情。如果没有配置 `DATABASE_URL`，任务只保存在内存里，重启后会丢失。

### 为什么大文本生成会分成多个文本块？

长篇小说一次性发送给 AI 容易超时或超过上下文限制。后端会按章节正文和段落拆分，每块最多约 6000 字符，分别生成局部剧本，再合并为完整 YAML。

### 为什么至少要选择 3 个章节？

少于 3 章时故事信息不足，AI 很难稳定提炼人物关系、冲突和场景节奏。系统因此限制至少选择 3 个连续章节。

### 为什么所选章节不能超过 100000 字？

这是为了控制单次任务的生成时间、AI 成本和结果稳定性。超出时建议缩小连续章节范围，分批转换。

### TXT 上传后乱码怎么办？

系统会自动检测常见中文编码。如果仍然乱码，建议先用文本编辑器将文件另存为 UTF-8，再重新上传。

### AI 输出校验失败怎么办？

可以重试生成，或在任务详情页编辑 YAML 后点击“重新校验”。重点检查数组字段是否仍为数组、场景是否有 `source_refs`、人物引用是否对应 `characters[].id`。
