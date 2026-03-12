# 数据库 ER 图设计文档

## 1. 数据库规划

每个微服务拥有独立数据库，实现数据自治：

| 服务 | 数据库名 | 包含表 |
|------|---------|--------|
| 用户服务 | user_db | users |
| 商品服务 | product_db | products, categories |
| 订单服务 | order_db | orders, order_items |
| 库存服务 | stock_db | stocks |

## 2. ER 关系图

```
┌──────────────────── user_db ────────────────────┐
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │              users                        │    │
│  ├──────────────────────────────────────────┤    │
│  │ PK  id           BIGINT AUTO_INCREMENT    │    │
│  │     username     VARCHAR(32)  UNIQUE, NOT NULL │
│  │     password     VARCHAR(128) NOT NULL     │    │
│  │     email        VARCHAR(128)              │    │
│  │     phone        VARCHAR(20)               │    │
│  │     status       TINYINT  DEFAULT 1        │    │
│  │     created_at   DATETIME                  │    │
│  │     updated_at   DATETIME                  │    │
│  └──────────────────────────────────────────┘    │
└──────────────────────────────────────────────────┘


┌──────────────────── product_db ──────────────────┐
│                                                   │
│  ┌──────────────────────────────────────────┐     │
│  │            categories                     │     │
│  ├──────────────────────────────────────────┤     │
│  │ PK  id           BIGINT AUTO_INCREMENT    │     │
│  │     name         VARCHAR(64)  NOT NULL    │     │
│  │     parent_id    BIGINT  DEFAULT 0        │     │
│  │     created_at   DATETIME                 │     │
│  │     updated_at   DATETIME                 │     │
│  └────────────────────┬─────────────────────┘     │
│                       │ 1                          │
│                       │                            │
│                       │ N                          │
│  ┌────────────────────┴─────────────────────┐     │
│  │              products                     │     │
│  ├──────────────────────────────────────────┤     │
│  │ PK  id           BIGINT AUTO_INCREMENT    │     │
│  │ FK  category_id  BIGINT                   │     │
│  │     name         VARCHAR(128) NOT NULL    │     │
│  │     description  TEXT                     │     │
│  │     price        DECIMAL(10,2) NOT NULL   │     │
│  │     images       VARCHAR(512)             │     │
│  │     status       TINYINT  DEFAULT 1       │     │
│  │     created_at   DATETIME                 │     │
│  │     updated_at   DATETIME                 │     │
│  └──────────────────────────────────────────┘     │
└───────────────────────────────────────────────────┘


┌──────────────────── order_db ────────────────────┐
│                                                   │
│  ┌──────────────────────────────────────────┐     │
│  │               orders                      │     │
│  ├──────────────────────────────────────────┤     │
│  │ PK  id           BIGINT AUTO_INCREMENT    │     │
│  │     order_no     VARCHAR(32)  UNIQUE      │     │
│  │     user_id      BIGINT       NOT NULL    │     │
│  │     total_amount DECIMAL(12,2) NOT NULL   │     │
│  │     status       TINYINT  DEFAULT 0       │     │
│  │     address      VARCHAR(256)             │     │
│  │     created_at   DATETIME                 │     │
│  │     updated_at   DATETIME                 │     │
│  └────────────────────┬─────────────────────┘     │
│                       │ 1                          │
│                       │                            │
│                       │ N                          │
│  ┌────────────────────┴─────────────────────┐     │
│  │            order_items                    │     │
│  ├──────────────────────────────────────────┤     │
│  │ PK  id           BIGINT AUTO_INCREMENT    │     │
│  │ FK  order_id     BIGINT       NOT NULL    │     │
│  │     product_id   BIGINT       NOT NULL    │     │
│  │     product_name VARCHAR(128) NOT NULL    │     │
│  │     price        DECIMAL(10,2) NOT NULL   │     │
│  │     quantity     INT          NOT NULL    │     │
│  │     subtotal     DECIMAL(12,2) NOT NULL   │     │
│  │     created_at   DATETIME                 │     │
│  └──────────────────────────────────────────┘     │
└───────────────────────────────────────────────────┘


┌──────────────────── stock_db ────────────────────┐
│                                                   │
│  ┌──────────────────────────────────────────┐     │
│  │               stocks                      │     │
│  ├──────────────────────────────────────────┤     │
│  │ PK  id           BIGINT AUTO_INCREMENT    │     │
│  │     product_id   BIGINT  UNIQUE, NOT NULL │     │
│  │     quantity     INT     NOT NULL DEFAULT 0│     │
│  │     created_at   DATETIME                 │     │
│  │     updated_at   DATETIME                 │     │
│  └──────────────────────────────────────────┘     │
└───────────────────────────────────────────────────┘
```

## 3. 跨服务实体关系图

```
用户(users)              商品(products)            库存(stocks)
    │                        │                        │
    │ user_id                │ product_id             │ product_id
    │                        │                        │
    ▼                        ▼                        ▼
┌──────────┐  包含商品  ┌────────────┐  对应库存  ┌──────────┐
│  orders  │───────────▶│order_items │◀───────────│  stocks  │
│  订单表  │  1:N       │ 订单明细表 │  通过       │  库存表  │
└──────────┘            └────────────┘  product_id └──────────┘
     │                       │
     │                       │
     └───── 1 : N ───────────┘

说明：
  users    1 ──── N  orders       （一个用户可以有多个订单）
  orders   1 ──── N  order_items  （一个订单可以有多个商品项）
  categories 1 ── N  products     （一个分类下有多个商品）
  products 1 ──── 1  stocks       （一个商品对应一条库存记录）
```

> **注意**：由于微服务架构的数据隔离原则，上述跨服务的关系（如 orders 中的 user_id 引用 users 表）不使用数据库外键约束，而是在应用层通过服务间调用保证数据一致性。

## 4. 详细表结构

### 4.1 users — 用户表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 用户ID |
| username | VARCHAR(32) | UNIQUE, NOT NULL | 用户名 |
| password | VARCHAR(128) | NOT NULL | 密码（bcrypt加密存储） |
| email | VARCHAR(128) | | 邮箱 |
| phone | VARCHAR(20) | | 手机号 |
| status | TINYINT | DEFAULT 1 | 状态：1-正常 0-禁用 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引：**
- `idx_username` — username 唯一索引
- `idx_phone` — phone 普通索引

```sql
CREATE TABLE `users` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(32) NOT NULL,
    `password` VARCHAR(128) NOT NULL,
    `email` VARCHAR(128) DEFAULT '',
    `phone` VARCHAR(20) DEFAULT '',
    `status` TINYINT NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`username`),
    KEY `idx_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.2 categories — 商品分类表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 分类ID |
| name | VARCHAR(64) | NOT NULL | 分类名称 |
| parent_id | BIGINT | DEFAULT 0 | 父分类ID，0表示顶级 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

```sql
CREATE TABLE `categories` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(64) NOT NULL,
    `parent_id` BIGINT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.3 products — 商品表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 商品ID |
| category_id | BIGINT | NOT NULL | 分类ID |
| name | VARCHAR(128) | NOT NULL | 商品名称 |
| description | TEXT | | 商品描述 |
| price | DECIMAL(10,2) | NOT NULL | 价格（元） |
| images | VARCHAR(512) | | 图片URL |
| status | TINYINT | DEFAULT 1 | 状态：1-上架 0-下架 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引：**
- `idx_category_id` — category_id 普通索引
- `idx_status` — status 普通索引
- `idx_name` — name 普通索引（支持搜索）

```sql
CREATE TABLE `products` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `category_id` BIGINT NOT NULL,
    `name` VARCHAR(128) NOT NULL,
    `description` TEXT,
    `price` DECIMAL(10,2) NOT NULL,
    `images` VARCHAR(512) DEFAULT '',
    `status` TINYINT NOT NULL DEFAULT 1,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_status` (`status`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.4 orders — 订单表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 订单ID |
| order_no | VARCHAR(32) | UNIQUE, NOT NULL | 订单编号 |
| user_id | BIGINT | NOT NULL | 用户ID |
| total_amount | DECIMAL(12,2) | NOT NULL | 订单总金额 |
| status | TINYINT | DEFAULT 0 | 0-待支付 1-已支付 2-已发货 3-已完成 4-已取消 |
| address | VARCHAR(256) | | 收货地址 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引：**
- `idx_order_no` — order_no 唯一索引
- `idx_user_id` — user_id 普通索引
- `idx_status` — status 普通索引

```sql
CREATE TABLE `orders` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `order_no` VARCHAR(32) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `total_amount` DECIMAL(12,2) NOT NULL,
    `status` TINYINT NOT NULL DEFAULT 0,
    `address` VARCHAR(256) DEFAULT '',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_order_no` (`order_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.5 order_items — 订单明细表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 明细ID |
| order_id | BIGINT | NOT NULL | 订单ID |
| product_id | BIGINT | NOT NULL | 商品ID |
| product_name | VARCHAR(128) | NOT NULL | 商品名称（冗余快照） |
| price | DECIMAL(10,2) | NOT NULL | 下单时商品单价（快照） |
| quantity | INT | NOT NULL | 购买数量 |
| subtotal | DECIMAL(12,2) | NOT NULL | 小计金额 |
| created_at | DATETIME | NOT NULL | 创建时间 |

**索引：**
- `idx_order_id` — order_id 普通索引

```sql
CREATE TABLE `order_items` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `order_id` BIGINT NOT NULL,
    `product_id` BIGINT NOT NULL,
    `product_name` VARCHAR(128) NOT NULL,
    `price` DECIMAL(10,2) NOT NULL,
    `quantity` INT NOT NULL,
    `subtotal` DECIMAL(12,2) NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.6 stocks — 库存表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 库存ID |
| product_id | BIGINT | UNIQUE, NOT NULL | 商品ID |
| quantity | INT | NOT NULL, DEFAULT 0 | 当前库存数量 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

**索引：**
- `idx_product_id` — product_id 唯一索引

```sql
CREATE TABLE `stocks` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `product_id` BIGINT NOT NULL,
    `quantity` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_product_id` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

## 5. 设计说明

### 5.1 数据隔离
每个服务拥有独立数据库，不跨库访问。订单表中的 `user_id` 和 `product_id` 不使用物理外键，通过应用层保证引用完整性。

### 5.2 数据冗余（快照）
`order_items` 中的 `product_name` 和 `price` 是下单时的商品快照，即使后续商品信息变更，历史订单数据不受影响。

### 5.3 软删除
所有表通过 `status` 字段实现逻辑状态管理，不做物理删除，确保数据可追溯。

### 5.4 时间字段
所有表统一包含 `created_at` 和 `updated_at` 字段，由数据库自动维护。
