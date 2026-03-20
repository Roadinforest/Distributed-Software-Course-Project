# 容器化 + 负载均衡 + 动静分离 + Redis缓存实施方案

## 1. 文档目标

本文档用于指导当前课程项目完成以下目标：

1. 使用 Dockerfile 与 Docker Compose 启动数据库、后端服务、Nginx。
2. 将后端服务以多实例启动，通过 Nginx 做代理转发与负载均衡。
3. 配置动静分离，静态资源由 Nginx 直出，动态请求转发后端。
4. 使用 JMeter 对静态与动态接口分别压测，比较响应时间。
5. 为商品详情引入 Redis 缓存，并处理缓存穿透、击穿、雪崩问题。

## 2. 当前项目现状与约束

基于当前仓库结构，现状如下：

1. `user-service` 已完整实现且有 `Dockerfile`。
2. `product-service`、`order-service`、`stock-service` 目前为脚手架目录。
3. 根 `docker-compose.yml` 当前仅包含 MySQL 与 Redis。
4. 项目尚未包含 Nginx 配置与前端静态资源目录。

因此建议采用“先闭环、后扩展”的推进策略：

1. 第一阶段先完成 `user-service` 双实例 + Nginx + 动静分离 + JMeter。
2. 第二阶段再落地 `product-service` 商品详情 Redis 缓存。

## 3. 目标架构（课程实验版）

```text
Browser/JMeter
      |
      v
Nginx:80
  |- /           -> 静态资源 (HTML/CSS/JS)
  |- /static/*   -> 静态资源
  |- /api/*      -> upstream backend_pool
                     |- user-service-1
                     |- user-service-2

MySQL:3306  <---- user-service / product-service
Redis:6379  <---- user-service / product-service(cache)
```

## 4. 分阶段实施计划

## 4.1 Phase A: 容器化与编排

1. 为服务补齐 Dockerfile（至少补齐 `product-service`）。
2. 所有服务配置支持环境变量覆盖，避免写死 `localhost`。
3. 在 `docker-compose.yml` 中定义：
   - `mysql`
   - `redis`
   - `user-service-1`
   - `user-service-2`
   - `nginx`
4. 服务间访问统一使用容器网络名，如 `mysql`、`redis`。
5. 添加 `healthcheck` 与 `depends_on`（条件健康）保证启动顺序更稳定。

## 4.2 Phase B: Nginx 负载均衡

1. 在 `nginx` 中定义 upstream（后端实例池）。
2. 实验至少三种算法：
   - 轮询（默认 round robin）
   - 最少连接（least_conn）
   - IP 哈希（ip_hash）
3. `/api/` 代理到后端池，并透传 `X-Forwarded-For` 等头。
4. 后端日志增加实例标识，便于统计请求分布是否均衡。

## 4.3 Phase C: 动静分离

1. 新建简单前端页面（`index.html` + `style.css` + `app.js`）。
2. Nginx 配置：
   - `/`、`/static/` 走本地静态目录
   - `/api/` 转发后端服务
3. 静态资源加缓存头，例如：`Cache-Control: public, max-age=300`。

## 4.4 Phase D: 商品详情 Redis 缓存

采用 Cache Aside（旁路缓存）模式：

1. 读流程：先查 Redis，未命中则查 MySQL，再回填 Redis。
2. 写流程：先更新 MySQL，再删除对应缓存 key。
3. 缓存 key 示例：`product:detail:{id}`。

## 4.5 Phase E: 缓存三大问题治理

1. 缓存穿透：
   - 参数合法性校验（ID > 0，格式校验）
   - 空值缓存（短 TTL，例如 30~60 秒）
   - 可选：布隆过滤器拦截非法 key
2. 缓存击穿：
   - 热点 key 加互斥锁（SETNX + 过期时间）
   - 或逻辑过期 + 后台异步重建
3. 缓存雪崩：
   - TTL 增加随机抖动（如 `baseTTL + rand(0~120s)`）
   - 热点数据预热
   - Redis 故障时服务降级与限流，避免打爆数据库

## 4.6 Phase F: JMeter 压测

### 场景设计

1. 动态接口压测：`http://<host>:80/api/...`
2. 静态文件压测：`http://<host>:80/` 或 `/static/*`
3. 切换不同 Nginx 负载均衡算法后重复动态压测。

### 观察指标

1. Average / P95 / P99 响应时间
2. Throughput（吞吐）
3. Error Rate（错误率）
4. 各实例日志请求数（验证均衡性）

### 结论判定建议

1. 静态资源响应时间应显著低于动态接口。
2. 轮询与最少连接下，请求分布应大致均衡。
3. `ip_hash` 可能出现按来源 IP 粘性分布，不要求严格平均。

## 5. 关键配置示例（模板）

以下为示例模板，需按实际项目路径与镜像名调整。

### 5.1 docker-compose 示例（节选）

```yaml
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
    ports:
      - "3307:3306"

  redis:
    image: redis:7.2
    ports:
      - "6380:6379"

  user-service-1:
    build: ./user-service
    environment:
      USER_SVC_DATABASE_HOST: mysql
      USER_SVC_DATABASE_PORT: 3306
      USER_SVC_REDIS_HOST: redis
      USER_SVC_REDIS_PORT: 6379
      INSTANCE_ID: user-service-1
    depends_on:
      mysql:
        condition: service_started
      redis:
        condition: service_started

  user-service-2:
    build: ./user-service
    environment:
      USER_SVC_DATABASE_HOST: mysql
      USER_SVC_DATABASE_PORT: 3306
      USER_SVC_REDIS_HOST: redis
      USER_SVC_REDIS_PORT: 6379
      INSTANCE_ID: user-service-2
    depends_on:
      mysql:
        condition: service_started
      redis:
        condition: service_started

  nginx:
    image: nginx:1.27-alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ./frontend:/usr/share/nginx/html:ro
    depends_on:
      - user-service-1
      - user-service-2
```

### 5.2 Nginx upstream 示例（节选）

```nginx
upstream backend_pool {
    # round robin: 默认，不写算法
    server user-service-1:8081;
    server user-service-2:8081;

    # 切换算法时，注释/替换为以下之一：
    # least_conn;
    # ip_hash;
}

server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /static/ {
        expires 5m;
        add_header Cache-Control "public, max-age=300";
        try_files $uri =404;
    }

    location /api/ {
        proxy_pass http://backend_pool/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 6. 验收清单

1. `docker compose up -d --build` 后，MySQL/Redis/Nginx/后端实例全部启动。
2. `http://<host>:80/` 可访问静态页面。
3. `http://<host>:80/api/...` 可访问后端接口。
4. 切换三种负载均衡算法后，JMeter 指标有可对比结果。
5. 后端日志能看出实例请求分布。
6. 商品详情缓存命中后响应时间明显下降。
7. 对不存在商品、高并发热点、批量过期场景有对应防护表现。

## 7. 风险与说明

1. 当前仓库中 `product/order/stock` 尚未完整实现，建议先完成 `user-service` 闭环实验。
2. 本方案是课程实验级，不包含生产级容灾、数据库高可用自动切换与灰度发布。
3. 若课程要求“后端两个不同端口（8081/8082）”的展示，可通过宿主机端口映射体现，容器内仍可统一监听 8081。

## 8. 建议文档与目录落位

1. 本文档：`doc/container-lb-cache-jmeter-plan.md`
2. Nginx 配置：`nginx/nginx.conf`、`nginx/conf.d/default.conf`
3. 前端静态：`frontend/index.html`、`frontend/static/*`
4. 后续压测报告建议新增：`doc/performance-test-report.md`
