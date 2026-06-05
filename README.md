# Novel-to-Screenplay AI Toolkit

AI 小说转剧本工具，用于将小说内容整理、拆解并转换为剧本创作素材。

## Repository Name

Recommended repository name: `novel-to-screenplay-ai-toolkit`

## Tech Stack

- Frontend: Vue 3, TypeScript, Vite, Pinia, Vue Router
- Backend: Go, Gin
- Deployment: Docker Compose, Nginx, PostgreSQL, Redis

## Current Foundation

The first development phase provides the project foundation:

- Vue frontend shell with routes, state management, and API client.
- Gin backend shell with `/api/healthz`.
- Domain models for screenplay YAML drafts and conversion tasks.
- Docker Compose stack for frontend, backend, PostgreSQL, Redis, and Nginx.

See `docs/phase-1-foundation.md` for setup details.
