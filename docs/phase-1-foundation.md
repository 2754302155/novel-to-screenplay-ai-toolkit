# 第 1 阶段工程骨架说明

第 1 阶段对应项目开发计划中的 PR-001 至 PR-004，目标是完成前端 Vue、后端 Gin、Docker Compose 部署和基础领域模型，为后续导入、解析、转换、校验和导出功能提供稳定起点。

## 已交付内容

- 前端 Vue 3 + TypeScript + Vite 工程。
- 前端路由、Pinia 状态管理、API 客户端和健康检查页面。
- 后端 Gin 工程、配置加载和 `/api/healthz` 健康检查接口。
- 剧本 YAML 初稿核心领域模型。
- Docker Compose 开发部署配置，包含 frontend、backend、postgres、redis 和 nginx。

## 本地运行

后端：

```bash
cd backend
go test ./...
go run ./cmd/server
```

前端：

```bash
cd frontend
npm install
npm run dev
```

Docker Compose：

```bash
cd deploy
docker compose up --build
```

默认访问地址：

- 前端：`http://localhost:8088`
- 后端健康检查：`http://localhost:8088/api/healthz`
- Nginx 统一入口：`http://localhost:8088`
