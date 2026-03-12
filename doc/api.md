# API 接口文档

## 通用约定

### 基础信息

| 项目 | 说明 |
|------|------|
| 协议 | HTTP/HTTPS |
| 数据格式 | JSON |
| 字符编码 | UTF-8 |
| 认证方式 | JWT Bearer Token |

### 请求头

```
Content-Type: application/json
Authorization: Bearer <jwt_token>    # 需要认证的接口
```

### 统一响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {}
}
```

### 响应状态码

| code | 说明 |
|------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 / Token 无效 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 1. 用户服务 (User Service) — :8081

### 1.1 用户注册

```
POST /api/v1/users/register
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-32位 |
| password | string | 是 | 密码，6-64位 |
| email | string | 否 | 邮箱地址 |
| phone | string | 否 | 手机号 |

**请求示例：**

```json
{
    "username": "zhangsan",
    "password": "123456",
    "email": "zhangsan@example.com",
    "phone": "13800138000"
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "注册成功",
    "data": {
        "id": 1,
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "created_at": "2026-03-12T10:00:00Z"
    }
}
```

### 1.2 用户登录

```
POST /api/v1/users/login
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**请求示例：**

```json
{
    "username": "zhangsan",
    "password": "123456"
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "登录成功",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_at": "2026-03-13T10:00:00Z",
        "user": {
            "id": 1,
            "username": "zhangsan"
        }
    }
}
```

### 1.3 获取用户信息

```
GET /api/v1/users/profile
```

**请求头：** 需要 Authorization

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "username": "zhangsan",
        "email": "zhangsan@example.com",
        "phone": "13800138000",
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z"
    }
}
```

### 1.4 修改用户信息

```
PUT /api/v1/users/profile
```

**请求头：** 需要 Authorization

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 否 | 邮箱地址 |
| phone | string | 否 | 手机号 |

**请求示例：**

```json
{
    "email": "new_email@example.com",
    "phone": "13900139000"
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "修改成功",
    "data": null
}
```

---

## 2. 商品服务 (Product Service) — :8082

### 2.1 创建商品

```
POST /api/v1/products
```

**请求头：** 需要 Authorization

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 商品名称 |
| description | string | 否 | 商品描述 |
| price | float64 | 是 | 商品价格（元），精确到分 |
| category_id | int | 是 | 分类ID |
| images | string | 否 | 商品图片URL |

**请求示例：**

```json
{
    "name": "Go语言编程指南",
    "description": "一本全面的Go语言教程",
    "price": 59.90,
    "category_id": 1,
    "images": "https://example.com/book.jpg"
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "id": 1,
        "name": "Go语言编程指南",
        "description": "一本全面的Go语言教程",
        "price": 59.90,
        "category_id": 1,
        "images": "https://example.com/book.jpg",
        "status": 1,
        "created_at": "2026-03-12T10:00:00Z"
    }
}
```

### 2.2 获取商品详情

```
GET /api/v1/products/:id
```

**路径参数：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 商品ID |

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "name": "Go语言编程指南",
        "description": "一本全面的Go语言教程",
        "price": 59.90,
        "category_id": 1,
        "category_name": "图书",
        "images": "https://example.com/book.jpg",
        "status": 1,
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:00:00Z"
    }
}
```

### 2.3 商品列表（分页）

```
GET /api/v1/products?page=1&page_size=10&category_id=1&keyword=Go
```

**查询参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 10，最大 100 |
| category_id | int | 否 | 按分类筛选 |
| keyword | string | 否 | 搜索关键词 |

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 50,
        "page": 1,
        "page_size": 10,
        "items": [
            {
                "id": 1,
                "name": "Go语言编程指南",
                "price": 59.90,
                "category_name": "图书",
                "images": "https://example.com/book.jpg",
                "status": 1
            }
        ]
    }
}
```

### 2.4 修改商品

```
PUT /api/v1/products/:id
```

**请求头：** 需要 Authorization

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 商品名称 |
| description | string | 否 | 商品描述 |
| price | float64 | 否 | 商品价格 |
| category_id | int | 否 | 分类ID |
| images | string | 否 | 商品图片URL |
| status | int | 否 | 状态：1-上架 0-下架 |

**响应示例：**

```json
{
    "code": 200,
    "message": "修改成功",
    "data": null
}
```

### 2.5 删除商品

```
DELETE /api/v1/products/:id
```

**请求头：** 需要 Authorization

**响应示例：**

```json
{
    "code": 200,
    "message": "删除成功",
    "data": null
}
```

### 2.6 创建分类

```
POST /api/v1/categories
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 分类名称 |
| parent_id | int | 否 | 父分类ID，0为顶级分类 |

**请求示例：**

```json
{
    "name": "图书",
    "parent_id": 0
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "id": 1,
        "name": "图书",
        "parent_id": 0
    }
}
```

### 2.7 获取分类列表

```
GET /api/v1/categories
```

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "name": "图书",
            "parent_id": 0,
            "children": [
                {
                    "id": 3,
                    "name": "编程类",
                    "parent_id": 1
                }
            ]
        },
        {
            "id": 2,
            "name": "电子产品",
            "parent_id": 0
        }
    ]
}
```

---

## 3. 订单服务 (Order Service) — :8083

### 3.1 创建订单

```
POST /api/v1/orders
```

**请求头：** 需要 Authorization

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| items | array | 是 | 订单商品列表 |
| items[].product_id | int | 是 | 商品ID |
| items[].quantity | int | 是 | 购买数量 |
| address | string | 是 | 收货地址 |

**请求示例：**

```json
{
    "items": [
        {
            "product_id": 1,
            "quantity": 2
        },
        {
            "product_id": 3,
            "quantity": 1
        }
    ],
    "address": "北京市海淀区中关村大街1号"
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "下单成功",
    "data": {
        "order_no": "ORD20260312100000001",
        "user_id": 1,
        "total_amount": 189.70,
        "status": 0,
        "address": "北京市海淀区中关村大街1号",
        "items": [
            {
                "product_id": 1,
                "product_name": "Go语言编程指南",
                "price": 59.90,
                "quantity": 2,
                "subtotal": 119.80
            },
            {
                "product_id": 3,
                "product_name": "数据结构与算法",
                "price": 69.90,
                "quantity": 1,
                "subtotal": 69.90
            }
        ],
        "created_at": "2026-03-12T10:00:00Z"
    }
}
```

### 3.2 获取订单详情

```
GET /api/v1/orders/:order_no
```

**请求头：** 需要 Authorization

**路径参数：**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单编号 |

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "order_no": "ORD20260312100000001",
        "user_id": 1,
        "total_amount": 189.70,
        "status": 1,
        "status_text": "已支付",
        "address": "北京市海淀区中关村大街1号",
        "items": [
            {
                "product_id": 1,
                "product_name": "Go语言编程指南",
                "price": 59.90,
                "quantity": 2,
                "subtotal": 119.80
            }
        ],
        "created_at": "2026-03-12T10:00:00Z",
        "updated_at": "2026-03-12T10:30:00Z"
    }
}
```

### 3.3 用户订单列表

```
GET /api/v1/orders?page=1&page_size=10&status=1
```

**请求头：** 需要 Authorization

**查询参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 10 |
| status | int | 否 | 按状态筛选 |

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 5,
        "page": 1,
        "page_size": 10,
        "items": [
            {
                "order_no": "ORD20260312100000001",
                "total_amount": 189.70,
                "status": 1,
                "status_text": "已支付",
                "item_count": 2,
                "created_at": "2026-03-12T10:00:00Z"
            }
        ]
    }
}
```

### 3.4 更新订单状态

```
PUT /api/v1/orders/:order_no/status
```

**请求头：** 需要 Authorization

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int | 是 | 目标状态 |

**订单状态流转：**

| status | 说明 | 可流转至 |
|--------|------|---------|
| 0 | 待支付 | 1（已支付）/ 4（已取消） |
| 1 | 已支付 | 2（已发货） |
| 2 | 已发货 | 3（已完成） |
| 3 | 已完成 | — |
| 4 | 已取消 | — |

**请求示例：**

```json
{
    "status": 1
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "状态更新成功",
    "data": null
}
```

### 3.5 取消订单

```
POST /api/v1/orders/:order_no/cancel
```

**请求头：** 需要 Authorization

**响应示例：**

```json
{
    "code": 200,
    "message": "订单已取消",
    "data": null
}
```

> 取消订单时会调用库存服务恢复库存。

---

## 4. 库存服务 (Stock Service) — :8084

### 4.1 设置库存

```
POST /api/v1/stocks
```

**请求头：** 需要 Authorization

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品ID |
| quantity | int | 是 | 初始库存数量 |

**请求示例：**

```json
{
    "product_id": 1,
    "quantity": 500
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "库存设置成功",
    "data": {
        "product_id": 1,
        "quantity": 500
    }
}
```

### 4.2 查询库存

```
GET /api/v1/stocks/:product_id
```

**路径参数：**

| 字段 | 类型 | 说明 |
|------|------|------|
| product_id | int | 商品ID |

**响应示例：**

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "product_id": 1,
        "quantity": 498
    }
}
```

### 4.3 扣减库存

```
POST /api/v1/stocks/deduct
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品ID |
| quantity | int | 是 | 扣减数量 |

**请求示例：**

```json
{
    "product_id": 1,
    "quantity": 2
}
```

**响应示例（成功）：**

```json
{
    "code": 200,
    "message": "扣减成功",
    "data": {
        "product_id": 1,
        "remaining": 496
    }
}
```

**响应示例（库存不足）：**

```json
{
    "code": 400,
    "message": "库存不足",
    "data": null
}
```

### 4.4 恢复库存

```
POST /api/v1/stocks/restore
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| product_id | int | 是 | 商品ID |
| quantity | int | 是 | 恢复数量 |

**请求示例：**

```json
{
    "product_id": 1,
    "quantity": 2
}
```

**响应示例：**

```json
{
    "code": 200,
    "message": "库存恢复成功",
    "data": {
        "product_id": 1,
        "remaining": 498
    }
}
```

---

## 5. 接口汇总

| 服务 | 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|------|
| 用户 | POST | /api/v1/users/register | 用户注册 | 否 |
| 用户 | POST | /api/v1/users/login | 用户登录 | 否 |
| 用户 | GET | /api/v1/users/profile | 获取用户信息 | 是 |
| 用户 | PUT | /api/v1/users/profile | 修改用户信息 | 是 |
| 商品 | POST | /api/v1/products | 创建商品 | 是 |
| 商品 | GET | /api/v1/products/:id | 获取商品详情 | 否 |
| 商品 | GET | /api/v1/products | 商品列表 | 否 |
| 商品 | PUT | /api/v1/products/:id | 修改商品 | 是 |
| 商品 | DELETE | /api/v1/products/:id | 删除商品 | 是 |
| 商品 | POST | /api/v1/categories | 创建分类 | 是 |
| 商品 | GET | /api/v1/categories | 分类列表 | 否 |
| 订单 | POST | /api/v1/orders | 创建订单 | 是 |
| 订单 | GET | /api/v1/orders/:order_no | 订单详情 | 是 |
| 订单 | GET | /api/v1/orders | 订单列表 | 是 |
| 订单 | PUT | /api/v1/orders/:order_no/status | 更新订单状态 | 是 |
| 订单 | POST | /api/v1/orders/:order_no/cancel | 取消订单 | 是 |
| 库存 | POST | /api/v1/stocks | 设置库存 | 是 |
| 库存 | GET | /api/v1/stocks/:product_id | 查询库存 | 否 |
| 库存 | POST | /api/v1/stocks/deduct | 扣减库存 | 内部调用 |
| 库存 | POST | /api/v1/stocks/restore | 恢复库存 | 内部调用 |
