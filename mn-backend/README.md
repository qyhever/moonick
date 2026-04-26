# moonick Backend

## 简介

`mn-backend` 是明叶同行的统一后端，负责承载：

- H5 用户认证
- Admin 管理员认证
- 用户资料与头像上传
- 行程发布、查询、编辑、状态更新
- 收藏关系
- 后台看板、用户查询、行程查询
- 行程过期任务骨架

## 路由域

用户端接口：

- `/api/v1/auth/*`
- `/api/v1/trips/*`
- `/api/v1/me/*`
- `/api/v1/users/*`
- `/api/v1/files/*`

管理端接口：

- `/api/admin/v1/auth/*`
- `/api/admin/v1/dashboard/*`
- `/api/admin/v1/trips/*`
- `/api/admin/v1/users/*`

## 本地配置

默认开发配置文件：

- `internal/config/dev.yml`

可选本地覆盖：

- `internal/config/dev.local.yml`

当前开发环境默认端口：

```yaml
server:
  port: 6303
```

如果需要启用管理员登录，确保配置：

```yaml
auth:
  admin:
    username: admin
    password: your-password
    name: 管理员
```

如果使用环境变量，关键项包括：

- `MOONICK_ENV`
- `MOONICK_AUTH_ADMIN_USERNAME`
- `MOONICK_AUTH_ADMIN_PASSWORD`
- `MOONICK_AUTH_ADMIN_NAME`

## 启动

编译并启动：

```bash
cd mn-backend
MOONICK_ENV=dev make dev
```

热重载：

```bash
cd mn-backend
./scripts/dev.sh hot
```

## 测试

在当前环境下，建议显式指定可写的 Go 缓存目录：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

## 当前实现说明

当前仓库中的 repository 仍以轻量内存实现为主，接口边界、错误语义和测试已经按 v1 路由契约对齐。后续如果切到真实 MySQL 持久化，应保持：

- 业务码协议不变
- 路由与鉴权边界不变
- 行程状态与分页协议不变
