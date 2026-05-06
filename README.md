# GoMind

GoMind 是一款基于全栈技术构建的 AI 对话应用程序。它提供了现代化的用户界面（Vue 3 + Element Plus）以及高性能的后端服务（Go + Gin），支持与多种主流 LLM（大语言模型，如 Gemini、OpenAI、Ollama 等）进行交互，并内置 RAG（检索增强生成）能力，旨在为用户提供快速、智能且具有长期记忆的对话体验。

## 🌟 核心特性

- **多模型支持**：通过 `CloudWeGo/Eino` 框架，灵活接入 OpenAI、Gemini、Ollama 等多种主流大模型。
- **RAG 检索增强**：内置 Embedding 处理引擎和文档目录扫描，利用 Redis / Vector 检索能力实现知识库问答。
- **历史记录与会话管理**：支持长记忆会话，利用 MySQL 持久化消息数据并能在服务重启时无缝恢复上下文。
- **高性能架构**：基于 Go 1.24 和 Gin 框架构建，使用 Redis 作为高速缓存，引入 RabbitMQ 处理异步任务解耦。
- **JWT 安全认证**：提供基于 JWT 的用户授权与安全验证机制。
- **邮件服务集成**：支持通过 SMTP (QQ邮箱等) 发送邮件。
- **前端现代化**：使用 Vue 3 结合 Element Plus，提供响应式、优雅的交互界面。
- **开箱即用**：提供完整的 `docker-compose.yml` 环境，一键启动所有依赖中间件。

## 🛠️ 技术栈

### 后端 (Backend)

- **语言**: Go 1.24
- **Web 框架**: [Gin](https://gin-gonic.com/)
- **AI 框架**: [CloudWeGo Eino](https://github.com/cloudwego/eino)
- **ORM**: [GORM](https://gorm.io/)
- **中间件驱动**: `go-redis`、`amqp`
- **其他**: JWT (`golang-jwt`)、邮件 (`gomail.v2`)

### 前端 (Frontend)

- **核心框架**: Vue 3
- **UI 组件库**: Element Plus
- **路由 & 请求**: Vue Router 4, Axios

### 中间件 & 基础设施

- **数据库**: MySQL 8.4
- **缓存/向量**: Redis Stack
- **消息队列**: RabbitMQ 3.13
- **容器化**: Docker & Docker Compose

## 📁 目录结构

```text
.
├── common/        # 公共组件 (AI助手抽象、MySQL/Redis/RabbitMQ 初始化等)
├── config/        # 配置文件目录 (config.toml, 包含 docker/local 等环境配置)
├── controller/    # 路由控制器，处理 HTTP 请求
├── dao/           # 数据访问层，负责与数据库交互
├── middleware/    # Gin 中间件 (JWT 拦截器等)
├── model/         # 数据结构定义
├── router/        # 路由注册与初始化
├── service/       # 核心业务逻辑处理
├── utils/         # 工具函数库
├── vue-frontend/  # Vue3 前端源码目录
├── main.go        # 后端入口程序
├── docker-compose.yml # 容器化编排文件
└── Dockerfile     # 后端构建文件
```

## 🚀 快速开始

您可以选择使用 Docker 快速部署。

项目提供了完整的 docker-compose 环境，可以一键启动 MySQL、Redis、RabbitMQ 以及前后端服务。

1. **环境准备**: 确保本机已安装 Docker 和 Docker Compose。
2. **配置环境变量**:
   将项目根目录下的 `.env.example` 复制为 `.env` 并填写您的 AI API Key：
   ```bash
   cp .env.example .env
   # 在 .env 中填入你的 GEMINI_API_KEY / OPENAI_API_KEY 等信息
   ```
3. **启动服务**:
   ```bash
   docker-compose up -d --build
   ```
4. **访问应用**:
   前端运行在：`http://localhost:8080`
   后端接口在：`http://localhost:9090`

## ⚙️ 核心配置说明

后端的主要配置文件位于 `config/config.toml` 中：

- **\[mainConfig]**: 配置服务启动地址与端口 (默认 `9090`)。
- **\[mysqlConfig] / \[redisConfig] / \[rabbitmqConfig]**: 配置各中间件连接的凭证及端口。
- **\[jwtConfig]**: 配置鉴权过期时间及秘钥。
- **\[ragModelConfig]**: RAG 配置核心。定义 embedding 模型类型（如 ark, openai）、维度 (dimension) 和本地文档扫描目录 (docDir)。
- **\[voiceServiceConfig]**: （可选）语音服务配置（如百度语音 API 相关 Key）。

## 🤝 贡献指南

欢迎提交 Issue 或 Pull Request 来完善此项目！

## 📄 开源协议

MIT License.
