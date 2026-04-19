# 分布式微服务电商系统

## 项目概述

本项目是一个基于 Go 语言的分布式微服务电商系统，采用微服务架构将系统拆分为用户服务、商品服务、订单服务和库存服务四个核心模块，各服务独立部署、独立扩展，通过 RESTful API 进行服务间通信。

## 系统总体设计

### 1. 系统架构

系统采用微服务架构模式，将业务拆分为以下四个核心服务：

```
                          ┌─────────────┐
                          │  API Gateway│
                          │  (Gin HTTP) │
                          └──────┬──────┘
                                 │
            ┌────────────┬───────┴───────┬────────────┐
            ▼            ▼               ▼            ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  用户服务    │ │  商品服务    │ │  订单服务    │ │  库存服务    │
    │ User Service │ │Product Service│ │Order Service │ │Stock Service │
    │  :8081       │ │  :8082       │ │  :8083       │ │  :8084       │
    └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └──────┬───────┘
           │                │                │                │
           ▼                ▼                ▼                ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  用户数据库  │ │  商品数据库  │ │  订单数据库  │ │  库存数据库  │
    │  user_db     │ │ product_db   │ │  order_db    │ │  stock_db    │
    └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
                          MySQL 数据库集群
```

### 2. 服务说明

| 服务名称 | 端口 | 职责 |
|---------|------|------|
| 用户服务 (User Service) | 8081 | 用户注册、登录、认证、用户信息管理 |
| 商品服务 (Product Service) | 8082 | 商品的增删改查、分类管理 |
| 订单服务 (Order Service) | 8083 | 订单创建、查询、状态管理 |
| 库存服务 (Stock Service) | 8084 | 库存查询、扣减、补充 |

### 3. 技术栈

| 类别 | 技术选型 |
|------|---------|
| 编程语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis |
| 认证 | JWT (golang-jwt) |
| 配置管理 | Viper |
| 日志 | Zap |
| API 文档 | Swagger (swaggo) |
| 容器化 | Docker + Docker Compose |
| 服务注册发现 | Consul (HashiCorp) |
| API网关 | Kong |
| 流量治理 | gobreaker (熔断) + TokenBucket (限流) |
| 压力测试 | hey + wrk |

### 4. 项目目录结构

```
Distributed-Software-Course-Project/
├── README.md                   # 项目说明文档
├── doc/                        # 设计文档
│   ├── architecture.md         # 系统架构设计
│   ├── api.md                  # API 接口文档
│   ├── er-diagram.md           # 数据库 ER 图
│   └── tech-stack.md           # 技术栈选型说明
├── user-service/               # 用户服务
│   ├── cmd/                    # 程序入口
│   ├── internal/               # 内部代码
│   │   ├── handler/            # HTTP 处理器
│   │   ├── service/            # 业务逻辑层
│   │   ├── repository/         # 数据访问层
│   │   ├── model/              # 数据模型
│   │   └── middleware/         # 中间件
│   ├── config/                 # 配置文件
│   ├── go.mod
│   └── go.sum
├── product-service/            # 商品服务
│   ├── cmd/
│   ├── internal/
│   ├── config/
│   ├── go.mod
│   └── go.sum
├── order-service/              # 订单服务
│   ├── cmd/
│   ├── internal/
│   ├── config/
│   ├── go.mod
│   └── go.sum
├── stock-service/              # 库存服务
│   ├── cmd/
│   ├── internal/
│   ├── config/
│   ├── go.mod
│   └── go.sum
├── docker-compose.yml          # 容器编排
└── .gitignore
```

### 5. 服务间通信

- **同步通信**：服务间通过 HTTP RESTful API 进行同步调用
- **数据隔离**：每个服务拥有独立的数据库，保证数据自治
- **服务间调用**：订单服务在创建订单时需调用商品服务查询商品信息、调用库存服务执行库存扣减

### 6. 核心业务流程

**用户下单流程：**

```
用户 → 登录认证(用户服务) → 浏览商品(商品服务) → 创建订单(订单服务)
                                                        │
                                                        ├─→ 查询商品信息(商品服务)
                                                        ├─→ 扣减库存(库存服务)
                                                        └─→ 生成订单记录(订单服务)
```

## 服务注册发现与配置管理

### 技术选型说明

| Java技术栈 | Go替代方案 | 说明 |
|-----------|-----------|------|
| Nacos | **Consul** | HashiCorp开源的服务注册发现与配置管理 |
| Spring Cloud Gateway | **Kong** | 基于Nginx的高性能API网关 |
| Sentinel | **gobreaker** + **TokenBucket** | 熔断器和限流器实现流量治理 |
| JMeter | **hey** + **wrk** | Go编写的高性能HTTP压测工具 |

### Consul 服务注册发现

Consul 提供：
- **服务注册**：微服务启动时自动注册到Consul
- **服务发现**：通过DNS或HTTP API获取服务实例列表
- **健康检查**：自动检测不健康的服务实例并下线
- **配置管理**：Key/Value存储，支持动态配置

访问 Consul UI：`http://localhost:8500`

### Kong API网关

Kong 提供：
- **路由管理**：基于路径、主机名的动态路由
- **负载均衡**：支持轮询、加权、最少连接等算法
- **健康检查**：自动对上游服务进行健康检查
- **限流插件**：内置rate-limiting插件
- **缓存插件**：proxy-cache缓存响应

访问 Kong Admin API：`http://localhost:8001`
访问 Kong Proxy：`http://localhost:8000`

### 流量治理

#### 熔断器 (Circuit Breaker)

使用 `gobreaker` 库实现：
- 错误率超过50%或连续失败3次时打开熔断
- 熔断打开后30秒尝试恢复
- 防止级联故障传播

#### 限流器 (Rate Limiter)

使用自定义TokenBucket实现：
- 每秒允许100个请求
- 突发容量200个请求
- 超过限制返回429状态码

### 压力测试

使用 Go 编写的压测工具替代 JMeter：

```bash
# 进入测试目录
cd benchmark

# 运行限流测试
./run_hey_test.sh -t rate-limit -n 5000 -c 200

# 运行网关压测
./run_wrk_test.sh -d 30s -t 12 -c 400

# 运行完整测试场景
./test_scenarios.sh
```

## 详细设计文档

- [系统架构设计](doc/architecture.md)
- [API 接口文档](doc/api.md)
- [数据库 ER 图](doc/er-diagram.md)
- [技术栈选型说明](doc/tech-stack.md)

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 7.0+
- Docker & Docker Compose

### 启动步骤

```bash
# 1. 克隆项目
git clone <repository-url>
cd Distributed-Software-Course-Project

# 2. 启动所有服务（包括Consul和Kong）
docker-compose up -d --build

# 3. 检查服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f user-service-1
```

### 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Nginx | 80 | 前端静态资源 + API反向代理 |
| Kong Proxy | 8000 | API网关HTTP入口 |
| Kong Admin | 8001 | API网关管理接口 |
| Consul | 8500 | 服务注册发现 + 配置管理 |
| user-service-1 | 8081 | 用户服务实例1 |
| user-service-2 | 8082 | 用户服务实例2 |
| product-service-1 | 8083 | 商品服务实例1 |
| product-service-2 | 8084 | 商品服务实例2 |
| order-service-1 | 8085 | 订单服务实例1 |
| order-service-2 | 8086 | 订单服务实例2 |
| stock-service-1 | 8087 | 库存服务实例1 |
| stock-service-2 | 8088 | 库存服务实例2 |
| MySQL | 3307/3308/3309/3310/3311 | 数据库端口 |
| Redis | 6380 | 缓存服务 |

### API网关调用示例

```bash
# 通过Kong网关调用用户服务
curl -X POST http://localhost:8000/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123"}'

# 直接调用用户服务
curl -X POST http://localhost:8081/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123"}'

# 查看Consul注册的服务
curl http://localhost:8500/v1/catalog/services

# 查看Kong路由配置
curl http://localhost:8001/routes
```

## 许可证

本项目仅用于课程学习。
