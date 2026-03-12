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
- Docker & Docker Compose（可选）

### 启动步骤

```bash
# 1. 克隆项目
git clone <repository-url>
cd Distributed-Software-Course-Project

# 2. 启动基础设施（MySQL、Redis）
docker-compose up -d mysql redis

# 3. 启动各服务
cd user-service && go run cmd/main.go &
cd product-service && go run cmd/main.go &
cd order-service && go run cmd/main.go &
cd stock-service && go run cmd/main.go &
```

## 许可证

本项目仅用于课程学习。
