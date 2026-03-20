# 容器化 + 负载均衡 + 动静分离 + Redis缓存 + 读写分离 - 实施记录

## 概述

本文档记录了分布式软件课程项目中容器化、负载均衡、动静分离、Redis缓存和读写分离的实现过程。

---

## 一、项目结构

```
Distributed-Software-Course-Project/
├── docker-compose.yml           # 容器编排配置
├── nginx/
│   ├── nginx.conf               # Nginx主配置
│   └── conf.d/
│       └── default.conf         # 站点配置（负载均衡+动静分离）
├── frontend/                    # 前端静态资源
│   ├── index.html
│   └── static/
│       ├── css/style.css
│       └── js/app.js
├── user-service/                # 用户服务（已存在）
│   ├── Dockerfile
│   ├── cmd/main.go
│   └── config/config.yaml
└── product-service/             # 商品服务（新建）
    ├── Dockerfile
    ├── cmd/main.go
    ├── config/config.yaml
    └── internal/
        ├── config/config.go
        ├── bootstrap/bootstrap.go
        ├── model/product.go
        ├── repository/product_repository.go
        ├── service/product_service.go
        ├── handler/product_handler.go
        ├── router/router.go
        └── pkg/response/response.go
```

---

## 二、实施步骤

### 步骤1: 检查现有配置

检查了以下文件：
- `user-service/cmd/main.go` - 确认支持 INSTANCE_ID 环境变量
- `user-service/internal/router/router.go` - 确认有 /healthz 端点
- `user-service/internal/bootstrap/bootstrap.go` - 确认支持 Redis 连接
- `user-service/config/config.yaml` - 确认配置文件存在
- `docker-compose.yml` - 已有基础配置

### 步骤2: 创建前端静态文件

创建了以下文件实现动静分离：

1. **frontend/index.html** - 主页面
   - 用户注册/登录表单
   - 商品查询功能
   - 负载均衡测试功能

2. **frontend/static/css/style.css** - 样式文件
   - 响应式布局
   - 美观的UI设计

3. **frontend/static/js/app.js** - JavaScript脚本
   - API调用封装
   - 负载均衡测试逻辑

### 步骤3: 配置Nginx动静分离

修改 `nginx/conf.d/default.conf`：

```nginx
# 用户服务后端
upstream user_backend {
    server user-service-1:8081;
    server user-service-2:8081;
    # 切换算法: least_conn; 或 ip_hash;
}

# 商品服务后端
upstream product_backend {
    server product-service-1:8083;
    server product-service-2:8083;
}

server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    # 静态文件 - 动静分离
    location /static/ {
        try_files $uri =404;
        expires 5m;
        add_header Cache-Control "public, max-age=300";
    }

    # 前端首页
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 用户API代理
    location /api/v1/users/ {
        proxy_pass http://user_backend;
        # ... 代理配置
    }

    # 商品API代理
    location /api/v1/products/ {
        proxy_pass http://product_backend;
        # ... 代理配置
    }
}
```

### 步骤4: 创建product-service商品服务

创建了完整的商品服务，实现Redis缓存：

1. **项目结构搭建**
   ```bash
   mkdir -p product-service/cmd
   mkdir -p product-service/internal/{config,bootstrap,model,repository,service,handler,router,pkg/response}
   mkdir -p product-service/config
   ```

2. **核心文件创建**

   - `config/config.go` - 配置加载，支持环境变量
   - `bootstrap/bootstrap.go` - 初始化MySQL和Redis连接
   - `model/product.go` - 商品数据模型
   - `repository/product_repository.go` - 数据访问层
   - `service/product_service.go` - 业务逻辑层（包含缓存处理）
   - `handler/product_handler.go` - HTTP处理层
   - `router/router.go` - 路由配置
   - `cmd/main.go` - 入口文件

3. **Dockerfile创建**
   ```dockerfile
   FROM golang:1.23-alpine AS builder
   WORKDIR /app
   COPY . .
   RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o product-service ./cmd

   FROM alpine:3.20
   WORKDIR /app
   COPY --from=builder /app/product-service /app/product-service
   COPY --from=builder /app/config /app/config
   EXPOSE 8083
   CMD ["/app/product-service"]
   ```

### 步骤5: Redis缓存实现

在 `service/product_service.go` 中实现了完整的缓存策略：

#### 缓存穿透防护
```go
// 1. 参数合法性校验
if id == 0 {
    return nil, fmt.Errorf("invalid product id: %d", id)
}

// 2. 空值缓存（短TTL）
if err.Error() == "record not found" {
    s.setNullCache(ctx, cacheKey)  // TTL = 60秒
}
```

#### 缓存击穿防护
```go
// 分布式锁
lockKey := lockKeyPrefix + fmt.Sprint(id)
locked, err := s.redis.SetNX(ctx, lockKey, "1", lockExpire).Result()

if locked {
    // 获取锁成功，查询数据库并写入缓存
    // ...
    s.redis.Del(ctx, lockKey)  // 释放锁
} else {
    // 未获取到锁，等待其他请求加载完成
}
```

#### 缓存雪崩防护
```go
// TTL随机抖动
ttl := s.config.Cache.TTL
if s.config.Cache.EnableRandom {
    ttl += rand.Intn(120)  // 0-120秒随机延迟
}
if ttl > s.config.Cache.MaxTTL {
    ttl = s.config.Cache.MaxTTL
}
```

### 步骤6: 更新docker-compose.yml

完整的容器编排配置：

```yaml
services:
  mysql:                      # 用户数据库
    image: mysql:8.0
    ports:
      - "3307:3306"

  mysql-product:             # 商品数据库
    image: mysql:8.0
    ports:
      - "3308:3306"

  redis:                     # Redis缓存
    image: redis:7.2
    ports:
      - "6380:6379"

  user-service-1:             # 用户服务实例1
    build: ./user-service
    ports:
      - "8081:8081"

  user-service-2:             # 用户服务实例2
    build: ./user-service
    ports:
      - "8082:8081"

  product-service-1:          # 商品服务实例1
    build: ./product-service
    ports:
      - "8083:8083"

  product-service-2:          # 商品服务实例2
    build: ./product-service
    ports:
      - "8084:8083"

  nginx:                     # Nginx负载均衡
    image: nginx:1.27-alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d/default.conf:/etc/nginx/conf.d/default.conf:ro
      - ./frontend:/usr/share/nginx/html:ro
```

---

## 三、关键配置说明

### 负载均衡算法切换

编辑 `nginx/conf.d/default.conf`：

| 算法 | 配置 | 说明 |
|------|------|------|
| 轮询 | (默认) | 依次分配请求 |
| least_conn | `least_conn;` | 分配给连接数最少的 |
| ip_hash | `ip_hash;` | 基于IP的会话保持 |

### 缓存配置

编辑 `product-service/config/config.yaml`：

```yaml
cache:
  ttl: 300           # 基础TTL(秒)
  max_ttl: 600       # 最大TTL(秒)
  null_ttl: 60       # 空值缓存TTL(秒)
  enable_random: true  # 启用随机TTL防止雪崩
```

---

## 四、启动与测试

### 启动服务
```bash
# 在项目根目录执行
docker compose up -d --build

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f
```

### 测试URL
- 静态页面: http://localhost/
- 用户API: http://localhost/api/v1/users/login
- 商品API: http://localhost/api/v1/products/1

### JMeter压测
参考 `doc/container-lb-cache-jmeter-plan.md` 中的压测方案进行测试。

---

## 五、文件变更清单

### 新增文件
- `frontend/index.html`
- `frontend/static/css/style.css`
- `frontend/static/js/app.js`
- `product-service/Dockerfile`
- `product-service/go.mod`
- `product-service/go.sum`
- `product-service/cmd/main.go`
- `product-service/config/config.yaml`
- `product-service/internal/config/config.go`
- `product-service/internal/bootstrap/bootstrap.go`
- `product-service/internal/model/product.go`
- `product-service/internal/repository/product_repository.go`
- `product-service/internal/service/product_service.go`
- `product-service/internal/handler/product_handler.go`
- `product-service/internal/router/router.go`
- `product-service/internal/pkg/response/response.go`

### 修改文件
- `docker-compose.yml` - 添加product-service和配置
- `nginx/conf.d/default.conf` - 添加动静分离和商品服务代理

---

## 六、注意事项

1. **Docker Desktop**: 确保Docker守护进程正常运行
2. **端口占用**: 确保80、3307、3308、6380、8081-8084端口可用
3. **首次启动**: 首次构建可能需要较长时间下载依赖
4. **健康检查**: 所有服务配置了healthcheck，确保依赖服务启动后再启动

---

## 七、后续优化建议

1. 添加JMeter压测脚本和测试报告
2. 实现商品详情的热点数据预热
3. 添加监控和日志收集
4. 实现服务注册与发现

---

## 八、MySQL读写分离实施

### 8.1 概述

MySQL读写分离通过主从复制架构实现：
- **主库(Master)**: 处理所有写操作(CREATE/UPDATE/DELETE)
- **从库(Slave)**: 处理所有读操作(SELECT)

### 8.2 Docker Compose配置

#### 主库配置
```yaml
mysql-master:
  image: mysql:8.0
  command: [
    "--server-id=1",
    "--log-bin=mysql-bin",
    "--binlog-format=ROW",
    "--gtid-mode=ON",
    "--enforce-gtid-consistency=ON"
  ]
```

#### 从库配置
```yaml
mysql-slave:
  image: mysql:8.0
  command: [
    "--server-id=2",
    "--relay-log=relay-bin",
    "--read-only=1",
    "--gtid-mode=ON",
    "--enforce-gtid-consistency=ON"
  ]
```

### 8.3 代码实现

#### 配置文件更新
在 `config.go` 中添加从库配置：
```go
type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    // ... 主库配置

    // 读写分离 - 从库
    ReadHost     string `mapstructure:"read_host"`
    ReadPort     int    `mapstructure:"read_port"`
    ReadUser     string `mapstructure:"read_user"`
    ReadPassword string `mapstructure:"read_password"`
    ReadDBName   string `mapstructure:"read_dbname"`
}
```

#### Repository层实现
```go
type ProductRepository struct {
    db     *gorm.DB  // 主库 - 写操作
    dbRead *gorm.DB  // 从库 - 读操作
}

// 读操作走从库
func (r *ProductRepository) FindByID(id uint) (*model.Product, error) {
    return r.dbRead.First(&product, id).Error
}

// 写操作走主库
func (r *ProductRepository) Create(product *model.Product) error {
    return r.db.Create(product).Error
}
```

### 8.4 环境变量配置

```yaml
product-service-1:
  environment:
    # 写库(主库)
    PRODUCT_SVC_DATABASE_HOST: mysql-master
    # 读库(从库)
    PRODUCT_SVC_DATABASE_READ_HOST: mysql-slave
```

### 8.5 测试读写分离

启动服务后，可通过以下方式验证：

1. **查看日志输出**:
   - 写操作日志: `[product-service-1] Write to MASTER for creating product`
   - 读操作日志: `[product-service-1] Read from SLAVE for product 1`

2. **访问状态接口**:
   ```bash
   curl http://localhost:8083/api/v1/db-stats
   # 返回: {"read_mode":"slave","write_mode":"master"}
   ```

3. **验证主从复制**:
   ```bash
   # 连接从库查看复制状态
   docker exec -it <slave-container> mysql -uroot -proot123 -e "SHOW SLAVE STATUS\G"
   ```

### 8.6 端口映射

| 服务 | 端口 |
|------|------|
| MySQL主库 | 3308 |
| MySQL从库 | 3309 |
| Redis | 6380 |

---

## 九、缓存与读写分离结合

### 9.1 请求处理流程

```
客户端请求
    │
    ▼
┌─────────────────────┐
│  查询Redis缓存      │──是──▶ 返回缓存数据
└─────────────────────┘
    │否
    ▼
┌─────────────────────┐
│  获取分布式锁       │──失败──▶ 等待后重试
└─────────────────────┘
    │成功
    ▼
┌─────────────────────┐
│  查MySQL从库(读)   │  ←── SELECT操作
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│  写入Redis缓存      │
└─────────────────────┘
    │
    ▼
释放锁，返回数据
```

### 9.2 写操作处理

```
写请求(POST/PUT/DELETE)
    │
    ▼
┌─────────────────────┐
│  MySQL主库执行     │  ←── CREATE/UPDATE/DELETE
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│  清除Redis缓存      │
└─────────────────────┘
    │
    ▼
返回响应
```

---

## 十、文件变更清单(读写分离)

### 新增文件
- `mysql/setup-replication.sql` - 主库复制用户配置
- `mysql/setup-slave.sh` - 从库复制配置脚本

### 修改文件
- `docker-compose.yml` - 添加MySQL主从配置
- `product-service/internal/config/config.go` - 添加读写分离配置
- `product-service/internal/bootstrap/bootstrap.go` - 初始化双数据库连接
- `product-service/internal/repository/product_repository.go` - 读写分离实现
- `product-service/internal/handler/product_handler.go` - 添加db-stats接口
- `product-service/internal/router/router.go` - 添加测试路由
- `product-service/cmd/main.go` - 传递双数据库连接
- `doc/architecture.md` - 更新架构文档
