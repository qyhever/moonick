# moonick Backend

本文件作为 `mn-backend` 的协作入口。目标是让开发者和代理快速理解服务职责、接口边界、配置项与验证方式。

## 导航

- 当前模块入口：当前文件 `AGENTS.md`
- 仓库总入口：`../AGENTS.md`
- 文档索引：`../docs/index.md`
- 后端技术方案：`../docs/technical/backend.md`
- 技术总览：`../docs/technical/overview.md`
- 联调清单：`../docs/technical/api-checklist.md`

## 模块简介

`mn-backend` 是明叶同行的统一后端，负责承载：

- H5 用户认证
- Admin 管理员认证
- 用户资料与头像上传
- 行程发布、查询、编辑、状态更新
- 收藏关系
- 后台看板、用户查询、行程查询
- 行程过期任务调度

## 路由边界

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

## 数据初始化

后端当前使用真实 MySQL 持久化，启动前需要导入初始化 SQL：

```bash
cd mn-backend
mysql -h <host> -P <port> -u <user> -p <database> < docs/sql/001_init.sql
```

初始化脚本路径：

- `docs/sql/001_init.sql`
- 如果库已存在旧版 `users` 表且仍是手机号登录结构，先执行 `docs/sql/004_migrate_users_email_auth.sql`

补充说明：

- 当前脚本已移除 `trips.remark` 的 `TEXT DEFAULT ''` 定义，兼容本地联调使用的 MySQL 版本
- 如果导入过程曾中途失败，重跑前先清空目标库或删除已创建表，避免留下半套结构

## 启动方式

开发启动：

```bash
cd mn-backend
MOONICK_ENV=dev make dev
```

热重载启动：

```bash
cd mn-backend
./scripts/dev.sh hot
```

## 验证命令

建议显式指定可写的 Go 缓存目录：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

## 当前实现边界

当前仓库已接入真实 MySQL 持久化，用户、管理员、行程、收藏等核心数据都通过数据库读写。当前已对齐的实现边界如下：

- H5 侧行程发布、编辑、状态切换走真实持久化链路
- Admin 侧行程编辑支持完整字段更新，同时兼容旧的“仅传状态”请求
- 服务进程启动后会立即执行一次行程过期扫描，并按分钟级周期继续执行
- 路由、鉴权边界、业务码协议与分页协议按 v1 契约保持稳定
- 启动日志当前不应传播包含 R2 敏感配置的完整原文，修复前需谨慎处理日志输出

## 协作建议

修改接口、枚举值或响应结构前，先确认是否会影响 `mn-frontend-h5` 和 `mn-frontend-admin`。涉及联调契约的改动，应同步检查根目录 `AGENTS.md` 与 `docs/technical/*` 文档。
