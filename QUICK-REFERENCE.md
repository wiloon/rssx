# RSSX K8s 部署快速参考

## 🎯 核心修改

### 1. 数据库文件名
- **修改前:** `rssx-api.db`
- **修改后:** `rssx.db`

### 2. 数据库路径
- **默认路径:** `/data/rssx/rssx.db`
- **环境变量:** `DATABASE_PATH`（优先级最高）
- **配置文件:** `sqlite.path`（次优先级）

### 3. 默认数据
首次启动自动创建 2 个默认 RSS 源：
- Hacker News: https://hnrss.org/newest
- r/golang: https://www.reddit.com/r/golang/.rss

---

## 🚀 快速测试

### 本地测试
```bash
# 使用默认路径
cd ~/workspace/projects/rssx/rssx-api
go run .

# 使用自定义路径
export DATABASE_PATH="/tmp/test/rssx.db"
go run .

# 验证数据
sqlite3 /tmp/test/rssx.db "SELECT * FROM feeds;"
```

### 容器测试
```bash
# 构建镜像
cd ~/workspace/projects/rssx
nerdctl build -f Containerfile.tekton -t rssx-api:test .

# 运行容器
nerdctl run --rm -p 8080:8080 \
  -v /tmp/rssx-data:/data/rssx \
  -e DATABASE_PATH=/data/rssx/rssx.db \
  rssx-api:test

# 验证
curl http://localhost:8080/health
ls -lh /tmp/rssx-data/rssx.db
```

---

## 📦 K8s 部署配置

### 最小化 Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rssx-api
spec:
  replicas: 1
  template:
    spec:
      securityContext:
        fsGroup: 1000  # 重要：确保文件权限
      containers:
      - name: rssx-api
        image: your-registry/rssx-api:latest
        env:
        - name: DATABASE_PATH
          value: "/data/rssx/rssx.db"
        volumeMounts:
        - name: data
          mountPath: /data/rssx
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: rssx-data-pvc
```

### PVC 配置
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: rssx-data-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

---

## ⚙️ 环境变量配置

### 必需配置
```bash
DATABASE_PATH=/data/rssx/rssx.db
```

### 可选配置
```bash
RSSX_RSS_SYNC_AUTO=true         # 自动同步 RSS
RSSX_SECURITY_KEY=your-key      # JWT 密钥
REDIS_ADDRESS=redis:6379        # Redis 地址
LOG_LEVEL=info                  # 日志级别
```

---

## 📁 目录结构

```
/data/rssx/
├── rssx.db         # SQLite 数据库
└── logs/           # 日志目录
    └── rssx.log    # 应用日志
```

---

## 🔍 验证检查

### 1. 数据库文件
```bash
kubectl exec -it <pod-name> -n rssx -- ls -lh /data/rssx/rssx.db
```

### 2. 默认数据
```bash
kubectl exec -it <pod-name> -n rssx -- \
  sqlite3 /data/rssx/rssx.db "SELECT COUNT(*) FROM feeds;"
# 应返回: 2
```

### 3. 应用日志
```bash
kubectl logs -f <pod-name> -n rssx
# 应看到: "Seeded 2 feeds" 或 "Database already has data"
```

### 4. 健康检查
```bash
kubectl exec -it <pod-name> -n rssx -- \
  wget -qO- http://localhost:8080/health
```

---

## ⚠️ 常见问题

### Q: Pod 启动失败，报权限错误
**A:** 确保 Deployment 配置了 `securityContext.fsGroup: 1000`

### Q: 数据库文件找不到
**A:** 检查 PVC 是否正确挂载到 `/data/rssx`

### Q: 没有默认数据
**A:** 检查应用日志，确认 seed data 是否执行成功

### Q: 从旧版本迁移数据
**A:** 
```bash
# 在 Pod 中执行
mkdir -p /data/rssx
mv /var/lib/rssx-api/rssx-api.db /data/rssx/rssx.db
```

---

## 📝 相关文件

- **详细文档:** [DEPLOYMENT-SUMMARY.md](DEPLOYMENT-SUMMARY.md)
- **Containerfile:** [Containerfile.tekton](Containerfile.tekton)
- **数据库代码:** [rssx-api/common/sqlite.go](rssx-api/common/sqlite.go)
- **配置说明:** [rssx-api/CONFIG.md](rssx-api/CONFIG.md)

---

## ✅ 部署检查清单

- [ ] 创建 PVC (`rssx-data-pvc`)
- [ ] 配置 fsGroup (1000)
- [ ] 设置 DATABASE_PATH 环境变量
- [ ] 验证健康检查端点
- [ ] 检查数据库文件权限
- [ ] 确认默认数据已加载
- [ ] 测试 RSS 同步功能

---

**版本:** v1.0.0 | **日期:** 2026-01-25
