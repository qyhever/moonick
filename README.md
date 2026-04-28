# 明叶同行

明叶同行是一个拼车信息平台，当前仓库包含 3 个子应用：

- `mn-backend`：Go + Gin 后端服务
- `mn-frontend-h5`：React + Vite 的 H5 业务端
- `mn-frontend-admin`：React + Vite + Ant Design 的 PC 管理端

## 当前实现范围

已完成的 v1 能力：

- H5 用户注册、登录、登录回跳
- H5 首页、行程详情、发布、编辑、我的发布、我的收藏、个人中心
- 用户头像上传、昵称更新、默认联系方式更新
- 后台管理员登录、登录守卫、首页看板
- 后台行程列表、行程详情、行程完整字段编辑
- 后台用户列表、用户只读详情
- 后端用户、管理员、行程、收藏、文件上传与 MySQL 持久化

当前统一约束：

- H5 接口前缀：`/api/v1/...`
- Admin 接口前缀：`/api/admin/v1/...`
- 行程状态：`active / full / closed / expired`
- 行程类型：`driver_post / passenger_post`

## 本地启动顺序

### 1. 启动后端

配置文件位于：

- `mn-backend/internal/config/dev.yml`
- 可选本地覆盖：`mn-backend/internal/config/dev.local.yml`

最小需要检查的配置：

- MySQL 地址、用户名、密码、库名
- JWT `secret`
- R2 文件上传配置
- 管理员账号种子：
  - `auth.admin.username`
  - `auth.admin.password`
  - `auth.admin.name`

MySQL 初始化：

```bash
cd mn-backend
mysql -h <host> -P <port> -u <user> -p <database> < docs/sql/001_init.sql
```

初始化 SQL 文件路径：

- `mn-backend/docs/sql/001_init.sql`

启动命令：

```bash
cd mn-backend
MOONICK_ENV=dev make dev
```

热重载方式：

```bash
cd mn-backend
./scripts/dev.sh hot
```

默认启动地址来自 `dev.yml`，当前仓库默认是：

```text
http://localhost:6303
```

### 2. 启动 H5

```bash
cd mn-frontend-h5
npm install
npm run dev
```

### 3. 启动 Admin

```bash
cd mn-frontend-admin
npm install
npm run dev
```

## 验证命令

后端：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

H5：

```bash
cd mn-frontend-h5
npm run test
npm run build
```

Admin：

```bash
cd mn-frontend-admin
npm run test
npm run build
```

## 相关文档

- 总体概览：`docs/technical/overview.md`
- H5 技术方案：`docs/technical/h5.md`
- Admin 技术方案：`docs/technical/admin.md`
- Backend 技术方案：`docs/technical/backend.md`
- 联调检查清单：`docs/technical/api-checklist.md`
