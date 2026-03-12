# User Service

## Run locally

```bash
cd user-service
go mod tidy
go run ./cmd
```

Default address: `http://localhost:8081`

## API

- `POST /api/v1/users/register`
- `POST /api/v1/users/login`
- `GET /api/v1/users/profile` (requires `Authorization: Bearer <token>`)
- `PUT /api/v1/users/profile` (requires `Authorization: Bearer <token>`)

## Dependencies

- MySQL 8.0 (`user_db`)
- Redis 7.0+

在根节点执行

```bash
docker compose up -d mysql redis
```


## 用 curl 做最小闭环验收
1）注册：
```bash
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"zhangsan","password":"123456","email":"zhangsan@example.com","phone":"13800138000"}'
```
预期：code 为 200，message 类似 register success。


2）登录：
```bash
curl -X POST http://localhost:8081/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"zhangsan","password":"123456"}'
```
预期：返回 token 和 expires_at。


3）带 token 获取 profile：
```bash
curl http://localhost:8081/api/v1/users/profile \
  -H "Authorization: Bearer <上一步token>"
```
预期：200 且能看到当前用户信息。