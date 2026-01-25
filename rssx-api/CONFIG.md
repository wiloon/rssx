# 配置管理说明

## 配置优先级

```
环境变量 > .env > config.toml > 代码默认值
```

## 使用场景

### 本地开发
1. 复制 `.env.example` 为 `.env`
2. 修改 `.env` 中的配置（如数据库路径、Redis 地址等）
3. `.env` 不会被提交到 Git

```bash
cp .env.example .env
# 编辑 .env，修改为本地开发配置
```

### 容器部署
直接通过容器环境变量注入配置，无需配置文件：

```yaml
# docker-compose.yml 示例
environment:
  - DATABASE_PATH=/data/rssx/rssx.db
  - REDIS_ADDRESS=redis:6379
  - RSSX_RSS_SYNC_AUTO=true
  - LOG_LEVEL=info
```

### Kubernetes 部署
使用 ConfigMap 或 Secret 注入环境变量：

```yaml
env:
  - name: DATABASE_PATH
    value: /data/rssx/rssx.db
  - name: REDIS_ADDRESS
    value: redis-service:6379
```

## 配置项说明

| 环境变量 | 说明 | 默认值 | 示例 |
|---------|------|--------|------|
| `DATABASE_PATH` | 数据库文件路径（环境变量优先） | `/data/rssx/rssx.db` | `./data/rssx.db` |
| `SQLITE_PATH` | SQLite 数据库文件路径（配置文件） | `/data/rssx/rssx.db` | `./data/rssx.db` |
| `REDIS_ADDRESS` | Redis 地址 | `127.0.0.1:6379` | `redis:6379` |
| `RSSX_RSS_SYNC_AUTO` | 是否自动同步 RSS | `true` | `true/false` |
| `RSSX_SECURITY_KEY` | 安全密钥 | 无 | 任意字符串 |
| `SYNC_DURATION` | 同步间隔（分钟） | `60` | `30` |
| `NEWS_EXPIRE_TIME` | 新闻过期时间 | `-720h` | `-1440h` |
| `NEWS_GC_DURATION` | GC 清理间隔 | `24h` | `12h` |
| `LOG_LEVEL` | 日志级别 | `info` | `debug/info/warn/error` |
| `LOG_PATH` | 日志目录 | `/var/log/rssx/` | `./logs/` |
| `LOG_FILE_NAME` | 日志文件名 | `rssx-api.log` | `api.log` |

## TOML 配置文件（逐步弃用）

`config.toml` 和 `config-local.toml` 仍然可用，但建议优先使用环境变量：

- **优点**：向后兼容，快速切换配置
- **缺点**：容器部署时不推荐，增加配置管理复杂度

## 最佳实践

1. **本地开发**：使用 `.env`，数据库路径使用相对路径 `./data/`
2. **容器部署**：只使用环境变量，不挂载配置文件
3. **敏感信息**：通过环境变量或 Secret 注入，不写入配置文件
4. **版本控制**：只提交 `.env.example`，不提交 `.env`
