# 明叶同行后端技术方案

## 文档导航

- 文档索引：`../index.md`
- 仓库协作入口：`../../AGENTS.md`
- 后端协作入口：`../../mn-backend/AGENTS.md`
- 技术总览：[overview.md](./overview.md)
- H5 技术方案：[h5.md](./h5.md)
- Admin 技术方案：[admin.md](./admin.md)
- 联调清单：[api-checklist.md](./api-checklist.md)

## 1. 文档目标

本文档用于指导 `mn-backend` 的详细设计与开发实现，覆盖：

- 模块划分
- 数据结构
- 接口设计
- 文件上传
- 定时任务
- 日志与错误处理
- 测试用例

---

## 2. 现状与扩展原则

### 2.1 当前基础

现有 `mn-backend` 已具备以下基础能力：

- Gin 路由骨架
- Viper 配置加载
- Zap + lumberjack 日志
- 统一响应结构
- 简单 controller / service / repository 分层

### 2.2 扩展原则

- 保留现有 Go + Gin + 配置 + 日志体系
- 在此基础上扩展业务模块，不重构为微服务
- 业务代码按领域拆分，避免所有逻辑堆积在 `app` 模块中

---

## 3. 推荐目录结构

```text
internal/
  api/
    router.go
  controller/
    auth_controller.go
    user_controller.go
    trip_controller.go
    favorite_controller.go
    admin_auth_controller.go
    admin_trip_controller.go
    admin_user_controller.go
    file_controller.go
    response.go
    codes.go
  middleware/
    auth_user.go
    auth_admin.go
    request_id.go
  service/
    auth_service.go
    user_service.go
    trip_service.go
    favorite_service.go
    admin_service.go
    file_service.go
  repository/
    mysql/
    persistence/
  model/
    entity/
    request/
    response/
  pkg/
    jwt/
    password/
    pagination/
    storage/
    timeutil/
  task/
    trip_expire_task.go
```

---

## 4. 数据库设计

### 4.1 用户表 `users`

| 字段 | 类型建议 | 说明 |
|------|----------|------|
| `id` | bigint PK | 主键 |
| `phone` | varchar(20) | 手机号，唯一 |
| `password_hash` | varchar(255) | 密码哈希 |
| `nickname` | varchar(64) | 昵称 |
| `avatar_url` | varchar(512) | 头像完整 URL |
| `default_wechat` | varchar(64) | 默认微信号 |
| `default_phone` | varchar(20) | 默认手机号 |
| `status` | varchar(20) | `active / disabled` |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：

- 唯一索引：`uk_users_phone(phone)`

### 4.2 管理员表 `admins`

| 字段 | 类型建议 | 说明 |
|------|----------|------|
| `id` | bigint PK | 主键 |
| `username` | varchar(64) | 账号，唯一 |
| `password_hash` | varchar(255) | 密码哈希 |
| `display_name` | varchar(64) | 展示名 |
| `status` | varchar(20) | `active / disabled` |
| `last_login_at` | datetime | 最近登录时间 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

### 4.3 行程表 `trips`

| 字段 | 类型建议 | 说明 |
|------|----------|------|
| `id` | bigint PK | 主键 |
| `publisher_user_id` | bigint | 发布用户 ID |
| `trip_type` | varchar(20) | `driver_post / passenger_post` |
| `from_city_text` | varchar(128) | 起点文本 |
| `to_city_text` | varchar(128) | 终点文本 |
| `departure_date` | date | 出发日期 |
| `departure_time` | time | 出发时间 |
| `departure_at` | datetime | 实际排序与过期判断字段 |
| `seat_count` | tinyint | 人数/座位数，范围 1~6 |
| `price_amount` | decimal(10,2) | 人均费用 |
| `is_price_negotiable` | tinyint | 是否面议 |
| `contact_wechat` | varchar(64) | 发布时微信号快照 |
| `contact_phone` | varchar(20) | 发布时手机号快照 |
| `remark` | varchar(1000) | 备注 |
| `status` | varchar(20) | `active / full / closed / expired` |
| `closed_reason` | varchar(32) | 关闭原因 |
| `deleted_at` | datetime null | 软删除时间 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引建议：

- `idx_trips_publisher_created(publisher_user_id, created_at desc)`
- `idx_trips_status_departure(status, departure_at)`
- `idx_trips_type_departure(trip_type, departure_at)`
- `idx_trips_created(created_at desc)`

### 4.4 收藏表 `trip_favorites`

| 字段 | 类型建议 | 说明 |
|------|----------|------|
| `id` | bigint PK | 主键 |
| `user_id` | bigint | 用户 ID |
| `trip_id` | bigint | 行程 ID |
| `created_at` | datetime | 收藏时间 |

索引建议：

- 唯一索引：`uk_trip_favorites_user_trip(user_id, trip_id)`
- 索引：`idx_trip_favorites_user_created(user_id, created_at desc)`

### 4.5 文件资源表 `file_assets`

| 字段 | 类型建议 | 说明 |
|------|----------|------|
| `id` | bigint PK | 主键 |
| `biz_type` | varchar(32) | 业务类型，如 `avatar` |
| `biz_id` | bigint null | 业务对象 ID |
| `storage_provider` | varchar(32) | `r2` |
| `bucket` | varchar(128) | 存储桶 |
| `object_key` | varchar(512) | 对象 key |
| `url` | varchar(512) | 完整 URL |
| `mime_type` | varchar(128) | 文件类型 |
| `size` | bigint | 文件大小 |
| `uploader_type` | varchar(20) | `user / admin / system` |
| `uploader_id` | bigint | 上传者 ID |
| `created_at` | datetime | 创建时间 |

---

## 5. 鉴权设计

### 5.1 认证模式

- 用户端与管理员端均采用 JWT 双令牌
- `accessToken` 2 小时
- `refreshToken` 72 小时
- 两端都为纯无状态 JWT
- 不使用 Redis，不做 token 会话存储

### 5.2 Claim 设计

| 字段 | 说明 |
|------|------|
| `sub` | 主体 ID |
| `role` | `user / admin` |
| `type` | `access / refresh` |
| `iat` | 签发时间 |
| `exp` | 过期时间 |

### 5.3 中间件要求

- 用户鉴权中间件只接受 `role=user` 的 `access` token
- 管理员鉴权中间件只接受 `role=admin` 的 `access` token
- 刷新接口只接受 `type=refresh` token

---

## 6. 服务模块设计

### 6.1 Auth 模块

职责：

- 用户注册、登录、刷新、获取当前登录态
- 管理员登录、刷新、获取当前登录态
- 密码哈希与校验
- JWT 签发与解析

### 6.2 User 模块

职责：

- 查询当前用户资料
- 修改昵称
- 修改默认联系方式
- 更新头像

### 6.3 Trip 模块

职责：

- 首页行程列表
- 行程详情
- 发布行程
- 编辑本人行程
- 修改本人行程状态
- 我的发布列表

### 6.4 Favorite 模块

职责：

- 收藏 toggle
- 我的收藏列表

### 6.5 Admin 模块

职责：

- 看板数据聚合
- 行程管理查询与编辑
- 用户列表与详情查询

### 6.6 File 模块

职责：

- 文件校验
- 上传对象存储
- 写入 `file_assets`
- 返回完整 URL

### 6.7 Task 模块

职责：

- 服务启动后立即执行一次过期扫描
- 定时扫描过期行程
- 更新状态为 `expired`
- 记录任务执行日志

---

## 7. 核心业务流程

### 7.1 用户注册

1. 校验手机号格式
2. 校验手机号唯一
3. 密码加密存储
4. 创建用户
5. 签发 `accessToken + refreshToken`
6. 返回用户信息与 token

### 7.2 用户登录

1. 通过手机号查询用户
2. 校验密码
3. 校验用户状态
4. 签发双 token
5. 返回用户信息与 token

### 7.3 Token 刷新

1. 校验 refresh token 的签名、类型、身份、有效期
2. 重新签发 access token
3. v1 不轮换 refresh token

### 7.4 发布行程

1. 校验登录态
2. 校验起点终点不同
3. 校验出发时间不早于当前时间
4. 校验微信号和手机号至少一项
5. 组装 `departure_at`
6. 写入 `trips`
7. 返回行程详情

### 7.5 收藏 toggle

1. 校验登录态
2. 校验行程存在且未软删除
3. 查询收藏关系
4. 存在则删除，不存在则新增
5. 返回最终收藏状态

### 7.6 行程过期任务

1. 服务启动后立即执行一次扫描，补齐停机期间遗留的过期数据
2. 之后按分钟级周期扫描 `status in (active, full)` 且 `departure_at < now`
3. 批量更新为 `expired`
4. 记录任务处理数量与异常信息

---

## 8. 接口设计

### 8.1 通用约定

统一响应结构：

```json
{
  "code": 1000,
  "message": "success",
  "data": {}
}
```

统一分页参数：

- `pageNum`
- `pageSize`

### 8.2 H5 认证接口

| 接口 | 说明 |
|------|------|
| `POST /api/v1/auth/register` | 注册 |
| `POST /api/v1/auth/login` | 登录 |
| `POST /api/v1/auth/refresh` | 刷新 access token |
| `POST /api/v1/auth/logout` | 退出登录 |
| `GET /api/v1/auth/me` | 获取当前用户信息 |

### 8.3 H5 用户接口

| 接口 | 说明 |
|------|------|
| `PUT /api/v1/users/profile` | 修改昵称 |
| `PUT /api/v1/users/contact` | 修改默认联系方式 |
| `POST /api/v1/users/avatar` | 上传头像 |

### 8.4 H5 行程接口

| 接口 | 说明 |
|------|------|
| `GET /api/v1/trips` | 行程分页列表 |
| `GET /api/v1/trips/:id` | 行程详情 |
| `POST /api/v1/trips` | 发布行程 |
| `PUT /api/v1/trips/:id` | 编辑本人行程 |
| `PATCH /api/v1/trips/:id/status` | 修改本人行程状态 |
| `GET /api/v1/me/trips` | 我的发布列表 |
| `GET /api/v1/me/favorites` | 我的收藏列表 |
| `POST /api/v1/trips/:id/favorite` | 收藏 toggle |

### 8.5 Admin 接口

| 接口 | 说明 |
|------|------|
| `POST /api/admin/v1/auth/login` | 管理员登录 |
| `POST /api/admin/v1/auth/refresh` | 管理员刷新 token |
| `POST /api/admin/v1/auth/logout` | 管理员退出 |
| `GET /api/admin/v1/auth/me` | 获取管理员信息 |
| `GET /api/admin/v1/dashboard/summary` | 看板概览 |
| `GET /api/admin/v1/trips` | 后台行程列表 |
| `GET /api/admin/v1/trips/:id` | 后台行程详情 |
| `PUT /api/admin/v1/trips/:id` | 后台编辑行程 |
| `GET /api/admin/v1/users` | 后台用户列表 |
| `GET /api/admin/v1/users/:id` | 后台用户详情 |
| `GET /api/admin/v1/users/:id/trips` | 用户发布行程列表 |

---

## 9. 错误码设计

建议在现有错误码基础上补齐以下业务码：

| 错误码 | 含义 |
|--------|------|
| `1000` | success |
| `1001` | 请求参数错误 |
| `1002` | 用户已存在 |
| `1003` | 用户不存在 |
| `1004` | 用户名或密码错误 |
| `1005` | 服务繁忙 |
| `1006` | 需要登录 |
| `1007` | 无效 token |
| `1008` | 资源已存在 |
| `1009` | 资源不存在 |
| `1010` | 权限不足 |
| `1011` | 手机号已注册 |
| `1012` | 起点和终点不能相同 |
| `1013` | 出发时间不能早于当前时间 |
| `1014` | 请填写至少一种联系方式 |
| `1015` | 行程状态不可操作 |
| `1016` | 无权操作该行程 |
| `1017` | 文件类型不支持 |
| `1018` | 文件大小超限 |

---

## 10. 日志设计

### 10.1 日志类型

- 访问日志
- 业务日志
- 错误日志
- 定时任务日志

### 10.2 关键字段

- `requestId`
- `userId` / `adminId`
- `role`
- `path`
- `method`
- `status`
- `costMs`
- `tripId`
- `expiredTrips`
- `error`

### 10.3 脱敏规则

- 手机号脱敏记录
- 微信号脱敏记录
- 密码、token 原文禁止入日志

---

## 11. 文件上传设计

### 11.1 对象存储

- 存储服务：Cloudflare R2
- 上传方式：服务端接收文件后上传对象存储
- 返回值：完整 URL + 文件 ID

### 11.2 校验规则

- 支持格式：`jpg / png / webp / heic`
- 文件大小最大 10 MB
- 上传失败必须返回明确错误码

---

## 12. 测试用例

### 12.1 认证

- 正确手机号密码可登录
- 错误密码登录失败
- access token 过期后 refresh 成功
- refresh token 过期后刷新失败
- 用户 token 访问 admin 接口被拒绝
- admin token 访问用户接口被拒绝

### 12.2 行程

- 起点终点相同不能发布
- 出发时间早于当前不能发布
- 联系方式都为空不能发布
- 本人可编辑自己的行程
- 非本人不可编辑
- 过期任务可将有效行程置为 `expired`
- 过期任务在服务启动后会立即执行一次
- 过期任务会在 `context` 结束后停止继续调度

### 12.3 收藏

- 未登录不能收藏
- 已收藏再次 toggle 后取消收藏
- 下线行程在收藏列表中返回不存在/已下线提示

### 12.4 上传

- 合法图片上传成功
- 非法类型上传失败
- 超过 10 MB 上传失败

---

## 13. 开发顺序建议

建议按以下顺序实现：

1. 数据库表结构与基础模型
2. JWT、密码、统一中间件
3. 用户注册登录与管理员登录
4. H5 行程与收藏
5. Admin 看板与管理接口
6. 头像上传
7. 过期任务
8. 测试与联调
