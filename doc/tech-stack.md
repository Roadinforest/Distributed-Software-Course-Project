# 技术栈选型说明

## 1. 选型总览

| 类别 | 技术选型 | 版本 | 用途 |
|------|---------|------|------|
| 编程语言 | Go | 1.21+ | 服务端开发 |
| Web 框架 | Gin | v1.9+ | HTTP 路由与中间件 |
| ORM 框架 | GORM | v2.0+ | 数据库操作 |
| 关系数据库 | MySQL | 8.0+ | 持久化存储 |
| 缓存 | Redis | 7.0+ | 缓存、会话存储 |
| 认证 | golang-jwt | v5 | JWT Token 签发与验证 |
| 配置管理 | Viper | v1.18+ | 多格式配置文件读取 |
| 日志 | Zap | v1.27+ | 结构化高性能日志 |
| API 文档 | Swaggo | v1.16+ | Swagger 自动文档生成 |
| 密码加密 | bcrypt | (标准库) | 用户密码哈希 |
| 参数校验 | validator | v10 | 请求参数校验 |
| 容器化 | Docker | 24+ | 服务容器化部署 |
| 容器编排 | Docker Compose | v2 | 本地多服务编排 |

## 2. 编程语言 — Go

### 选型理由

| 优势 | 说明 |
|------|------|
| **高性能** | 编译型语言，运行效率接近 C/C++，远高于 Java/Python |
| **天然并发** | goroutine 轻量级线程 + channel 通信，非常适合高并发微服务 |
| **编译快速** | 编译速度极快，支持交叉编译，方便 CI/CD |
| **部署简单** | 编译为单一二进制文件，无需运行时环境依赖 |
| **内存占用低** | 单个服务内存占用小，适合容器化部署 |
| **生态成熟** | 丰富的微服务工具链和社区生态 |

### 对比

| 语言 | 性能 | 并发模型 | 部署复杂度 | 学习曲线 | 微服务生态 |
|------|------|---------|-----------|---------|-----------|
| **Go** | ★★★★★ | goroutine | 低（单二进制） | 中 | ★★★★★ |
| Java | ★★★★ | 线程池 | 中（需JVM） | 高 | ★★★★★ |
| Python | ★★★ | asyncio | 中 | 低 | ★★★★ |
| Node.js | ★★★★ | 事件循环 | 中 | 低 | ★★★★ |

## 3. Web 框架 — Gin

### 选型理由

| 优势 | 说明 |
|------|------|
| **高性能** | 基于 httprouter，Go 生态中性能最好的 HTTP 框架之一 |
| **轻量级** | API 简洁，核心库很小，启动快 |
| **中间件机制** | 支持链式中间件，方便添加认证、日志、限流等功能 |
| **参数绑定** | 内置 JSON/Form/Query 参数自动绑定和校验 |
| **社区活跃** | GitHub 70k+ Stars，文档完善，生态丰富 |

### 对比

| 框架 | 性能 | 易用性 | 功能丰富度 | 社区 |
|------|------|--------|-----------|------|
| **Gin** | ★★★★★ | ★★★★★ | ★★★★ | ★★★★★ |
| Echo | ★★★★★ | ★★★★ | ★★★★ | ★★★★ |
| Fiber | ★★★★★ | ★★★★ | ★★★ | ★★★ |
| Beego | ★★★ | ★★★ | ★★★★★ | ★★★ |

## 4. ORM 框架 — GORM

### 选型理由

| 优势 | 说明 |
|------|------|
| **功能全面** | 支持 CRUD、关联、事务、迁移、Hook 等完整功能 |
| **链式 API** | 流畅的链式调用风格，代码简洁 |
| **自动迁移** | 支持 AutoMigrate，根据结构体自动创建/修改表结构 |
| **多数据库** | 支持 MySQL、PostgreSQL、SQLite、SQL Server |
| **插件生态** | 丰富的插件（分页、软删除、乐观锁等） |

### 示例代码

```go
// 定义模型
type User struct {
    ID        uint   `gorm:"primaryKey"`
    Username  string `gorm:"uniqueIndex;size:32;not null"`
    Password  string `gorm:"size:128;not null"`
    Email     string `gorm:"size:128"`
    Phone     string `gorm:"size:20;index"`
    Status    int8   `gorm:"default:1"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 查询
var user User
db.Where("username = ?", username).First(&user)
```

## 5. 数据库 — MySQL 8.0

### 选型理由

| 优势 | 说明 |
|------|------|
| **成熟稳定** | 全球使用最广泛的开源关系数据库 |
| **ACID 事务** | InnoDB 引擎完整支持事务 |
| **性能优秀** | 针对读多写少场景优化，适合电商系统 |
| **运维生态** | 监控、备份、主从复制等工具链成熟 |
| **GORM 支持** | GORM 对 MySQL 支持最完善 |

### 关键配置

```yaml
# 字符集使用 utf8mb4 支持 emoji
character-set-server: utf8mb4
collation-server: utf8mb4_unicode_ci

# InnoDB 引擎
default-storage-engine: InnoDB
```

## 6. 缓存 — Redis

### 选型理由

| 优势 | 说明 |
|------|------|
| **高性能** | 内存存储，读写延迟微秒级 |
| **数据结构丰富** | String、Hash、List、Set、Sorted Set |
| **原子操作** | 支持原子性递增/递减，适合库存扣减场景 |
| **过期策略** | 支持 TTL，适合 Token 和缓存管理 |

### 使用场景

| 场景 | 数据结构 | 说明 |
|------|---------|------|
| JWT Token 黑名单 | String + TTL | 注销时将 Token 加入黑名单 |
| 商品信息缓存 | Hash | 缓存热点商品数据，减少数据库压力 |
| 库存预扣减 | String + DECRBY | 利用原子操作防止超卖 |

### Go 客户端选择

使用 `go-redis/redis` v9，Go 生态中最流行的 Redis 客户端。

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})
```

## 7. 认证 — JWT (golang-jwt)

### 选型理由

| 优势 | 说明 |
|------|------|
| **无状态** | Token 自包含用户信息，服务端无需存储会话 |
| **分布式友好** | 适合微服务架构，各服务可独立验证 Token |
| **标准化** | 基于 RFC 7519 标准 |

### Token 设计

```json
{
    "header": {
        "alg": "HS256",
        "typ": "JWT"
    },
    "payload": {
        "user_id": 1,
        "username": "zhangsan",
        "exp": 1710331200,
        "iat": 1710244800
    }
}
```

| 字段 | 说明 |
|------|------|
| user_id | 用户ID |
| username | 用户名 |
| exp | 过期时间（24小时） |
| iat | 签发时间 |

## 8. 配置管理 — Viper

### 选型理由

| 优势 | 说明 |
|------|------|
| **多格式支持** | YAML、JSON、TOML、环境变量 |
| **环境变量** | 自动绑定环境变量，适合容器化部署 |
| **热重载** | 支持配置文件变更监听 |

### 配置文件示例

```yaml
# config/config.yaml
server:
  port: 8081

database:
  host: localhost
  port: 3306
  user: root
  password: root123
  dbname: user_db

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key"
  expire: 24  # 小时
```

## 9. 日志 — Zap

### 选型理由

| 优势 | 说明 |
|------|------|
| **高性能** | Uber 出品，零内存分配设计 |
| **结构化日志** | JSON 格式输出，方便日志采集和分析 |
| **多级别** | Debug、Info、Warn、Error、Fatal |

### 使用方式

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
logger.Info("user login",
    zap.String("username", "zhangsan"),
    zap.Int64("user_id", 1),
)
```

## 10. 容器化 — Docker + Docker Compose

### 选型理由

| 优势 | 说明 |
|------|------|
| **环境一致** | 开发、测试、生产环境一致 |
| **一键部署** | Docker Compose 统一编排所有服务和中间件 |
| **资源隔离** | 各服务容器隔离运行 |
| **弹性伸缩** | 方便进行水平扩展 |

### Dockerfile 示例

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
EXPOSE 8081
CMD ["./server"]
```

### Docker Compose 示例

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root123
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7.0-alpine
    ports:
      - "6379:6379"

  user-service:
    build: ./user-service
    ports:
      - "8081:8081"
    depends_on:
      - mysql
      - redis

  product-service:
    build: ./product-service
    ports:
      - "8082:8082"
    depends_on:
      - mysql

  order-service:
    build: ./order-service
    ports:
      - "8083:8083"
    depends_on:
      - mysql
      - redis

  stock-service:
    build: ./stock-service
    ports:
      - "8084:8084"
    depends_on:
      - mysql
      - redis

volumes:
  mysql_data:
```

## 11. 依赖汇总 (Go Modules)

```
github.com/gin-gonic/gin          v1.9+      # Web 框架
gorm.io/gorm                      v1.25+     # ORM
gorm.io/driver/mysql               v1.5+      # MySQL 驱动
github.com/redis/go-redis/v9       v9.4+      # Redis 客户端
github.com/golang-jwt/jwt/v5       v5.2+      # JWT
github.com/spf13/viper             v1.18+     # 配置管理
go.uber.org/zap                    v1.27+     # 日志
github.com/go-playground/validator v10+       # 参数校验
github.com/swaggo/gin-swagger      v1.6+      # Swagger 文档
golang.org/x/crypto                latest     # bcrypt 密码加密
```

## 12. 技术架构总结

```
┌─────────────────────────────────────────────────────┐
│                    客户端层                           │
│              Web / App / 小程序                       │
└───────────────────────┬─────────────────────────────┘
                        │ HTTP/HTTPS
┌───────────────────────▼─────────────────────────────┐
│                  API Gateway (Gin)                    │
│            路由转发 / JWT认证 / 限流                   │
└───────────────────────┬─────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│                   微服务层 (Go + Gin)                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│  │用户服务  │ │商品服务  │ │订单服务  │ │库存服务│ │
│  │  Gin     │ │  Gin     │ │  Gin     │ │  Gin   │ │
│  │  GORM    │ │  GORM    │ │  GORM    │ │  GORM  │ │
│  └──────────┘ └──────────┘ └──────────┘ └────────┘ │
└───────────────────────┬─────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────┐
│                    数据层                             │
│         ┌───────────────┐  ┌───────────────┐        │
│         │  MySQL 8.0    │  │  Redis 7.0    │        │
│         │  持久化存储    │  │  缓存/会话    │        │
│         └───────────────┘  └───────────────┘        │
└─────────────────────────────────────────────────────┘
```
