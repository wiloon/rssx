# RSSX API

RSSX API 是一个基于 Go 语言开发的 RSS 订阅管理后端服务。

## 功能特性

- 🔐 用户认证与授权（JWT）
- 📰 RSS 订阅源管理
- 📝 新闻文章获取与存储
- 🔄 RSS 订阅源自动同步
- 💾 支持 SQLite 和 Redis 存储
- 🗑️ 自动垃圾回收机制
- 📊 订阅源列表管理
- 🔍 新闻文章查询

## 技术栈

- Go
- SQLite
- Redis
- JWT 认证

## 项目结构

```
rssx-api/
├── common/          # 通用工具（SQLite）
├── feed/            # 订阅源管理
├── feeds/           # 订阅源列表
├── news/            # 新闻文章处理
├── rss/             # RSS 同步与垃圾回收
├── storage/         # 存储层（Redis）
├── user/            # 用户管理
└── utils/           # 工具函数
    ├── config/      # 配置管理
    ├── jwt/         # JWT 工具
    ├── logger/      # 日志工具
    └── response/    # 响应工具
```

## 安装

```bash
# 克隆项目
git clone <repository-url>
cd rssx-api

# 安装依赖
go mod download
```

## 配置

编辑配置文件：
- `config-local.toml` - 本地开发环境
- `config-k8s.toml` - Kubernetes 环境
- `config.toml` - 默认配置

## 运行

### 本地运行

```bash
go run rssx-api.go
```

### Docker 运行

```bash
docker build -t rssx-api .
docker run -p 8080:8080 rssx-api
```

### Kubernetes 部署

```bash
cd deploy/k8s
./deploy-k8s.sh
```

## 测试

```bash
# 运行所有测试
./test.sh

# 或者
go test ./...
```

## API 端点

### 用户管理
- `POST /user/register` - 用户注册
- `POST /user/login` - 用户登录

### 订阅源管理
- `GET /feeds` - 获取订阅源列表
- `POST /feed` - 添加订阅源
- `GET /feed/:id` - 获取订阅源详情

### 新闻管理
- `GET /news` - 获取新闻列表
- `GET /news/:id` - 获取新闻详情
- `GET /feed/:id/news` - 获取指定订阅源的新闻

## 开发

### 构建

```bash
go build -o rssx-api rssx-api.go
```

### 代码格式化

```bash
go fmt ./...
```

### 代码检查

```bash
go vet ./...
```

## 部署

### Podman 部署

```bash
cd deploy
./deploy-podman.sh
```

### Kubernetes 部署

```bash
kubectl apply -f deploy/k8s/rssx-api-deployment.yaml
```

## 许可证

请参阅项目根目录的 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！
