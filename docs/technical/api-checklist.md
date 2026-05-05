# 明叶同行联调与验收检查清单

## 文档导航

- 文档索引：`../index.md`
- 仓库协作入口：`../../AGENTS.md`
- 后端协作入口：`../../mn-backend/AGENTS.md`
- H5 协作入口：`../../mn-frontend-h5/AGENTS.md`
- Admin 协作入口：`../../mn-frontend-admin/AGENTS.md`
- 技术总览：[overview.md](./overview.md)
- 后端技术方案：[backend.md](./backend.md)
- H5 技术方案：[h5.md](./h5.md)
- Admin 技术方案：[admin.md](./admin.md)

## 文档目的

本清单用于记录当前仓库已经落地并可验证的联调路径，不包含尚未实现的需求点。

---

## 一、启动前检查

- [ ] 已准备 MySQL，并按后端配置可正常连接
- [ ] 已导入初始化 SQL：

```bash
cd mn-backend
mysql -h <host> -P <port> -u <user> -p <database> < docs/sql/001_init.sql
```

说明：

- `mn-backend/docs/sql/001_init.sql` 当前已去掉 `trips.remark` 的 `TEXT DEFAULT ''` 定义，兼容本地联调使用的 MySQL 版本
- 如果目标库已经存在旧版手机号登录结构的 `users` 表，需额外执行 `mn-backend/docs/sql/004_migrate_users_email_auth.sql`
- 如果导入过程曾中途失败，重跑前先清空目标库或删除已创建表，避免留下半套结构

- [ ] 已确认初始化脚本路径：`mn-backend/docs/sql/001_init.sql`
- [ ] 已检查 `mn-backend/internal/config/dev.yml`
- [ ] 已按需准备 `mn-backend/internal/config/dev.local.yml` 本地覆盖配置
- [ ] 已配置 JWT `secret`
- [ ] 已配置 R2 上传参数
- [ ] 已配置管理员账号种子：
  - [ ] `auth.admin.username`
  - [ ] `auth.admin.password`
  - [ ] `auth.admin.name`

---

## 二、本地启动顺序

- [ ] 启动后端

```bash
cd mn-backend
MOONICK_ENV=dev make dev
```

说明：

- 后端启动后应立即看到一条行程过期任务执行日志
- 该任务后续会按分钟级周期继续执行

- [ ] 启动 H5

```bash
cd mn-frontend-h5
npm install
npm run dev
```

说明：

- H5 开发服务默认通过 Vite 代理转发 `/api` 到 `http://127.0.0.1:6303`
- 当前本地联调入口为 `http://127.0.0.1:5080`

- [ ] 启动 Admin

```bash
cd mn-frontend-admin
npm install
npm run dev
```

说明：

- Admin 开发服务默认通过 Vite 代理转发 `/api` 到 `http://127.0.0.1:6303`
- 当前本地联调入口为 `http://127.0.0.1:5090`

---

## 三、自动化验证

### 3.1 后端

- [ ] 执行后端测试：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

### 3.2 H5

- [ ] 执行 H5 测试与构建：

```bash
cd mn-frontend-h5
npm run test
npm run build
```

### 3.3 Admin

- [ ] 执行 Admin 测试与构建：

```bash
cd mn-frontend-admin
npm run test
npm run build
```

说明：

- `mn-frontend-admin` 构建目前存在大包体积 warning，但构建成功

---

## 四、人工回归清单

### 4.1 后端过期任务

- [x] 后端启动后，日志中可看到一次行程过期任务执行记录
- [x] 准备 1 条 `status=active` 或 `status=full` 且出发时间早于当前时间的行程
- [x] 等待 1 分钟内下一次扫描或重启后端
- [x] 该行程状态可被更新为 `expired`

验收记录（2026-04-29）：

- 本地启动命令：`cd mn-backend && MOONICK_ENV=dev make dev`
- 启动日志可见行程过期任务执行记录：`trip expire task completed {"expiredTrips": 0}`
- 按清单准备过期测试数据后，重启后端触发启动即扫描
- 数据库中目标行程状态已从 `active/full` 更新为 `expired`

### 4.2 H5 游客路径

- [ ] 游客访问首页正常
- [ ] 游客访问行程详情页正常
- [ ] 游客点击发布页会跳转登录页

### 4.3 H5 登录与注册

- [ ] 用户可注册
- [ ] 用户可登录
- [ ] 已登录用户可发送重置密码验证码
- [ ] 已登录用户可重置密码，并在成功后重新登录
- [ ] 未登录访问受保护页面时，登录后能按 `redirect` 回跳
- [x] access token 失效后，H5 受保护请求可自动调用 `POST /api/v1/auth/refresh` 并重放原请求
- [x] refresh token 失效后，H5 会清理本地登录态并重新要求登录

验收记录（2026-04-29）：

- 联调入口：`http://localhost:5080/me/profile`
- 测试账号：`15927700475 / secret123`
- 成功分支：
  - 手动篡改 `Local Storage -> mn-h5-auth.accessToken`
  - 刷新受保护页面后，`Network` 面板可见 `POST /api/v1/auth/refresh`
  - 页面保持登录态，仍停留在 `/me/profile`
  - `mn-h5-auth` 中的 `accessToken` 与 `refreshToken` 均被替换为新值
- 失败分支：
  - 手动篡改 `Local Storage -> mn-h5-auth.accessToken` 与 `refreshToken`
  - 刷新受保护页面后，页面跳转到 `/login?redirect=%2Fme%2Fprofile`
  - `Local Storage` 中的 `mn-h5-auth` 已被清理

### 4.4 H5 行程

- [ ] 发布页能校验：
  - [ ] 起点和终点不能相同
  - [ ] 出发时间不能早于当前时间
  - [ ] 至少填写一种联系方式
  - [ ] 人数范围必须在 `1 ~ 6`
- [ ] 发布成功后跳转详情页
- [ ] 本人可编辑自己的行程
- [ ] 本人可把自己的行程设为满员或关闭
- [ ] 他人行程可收藏
- [ ] 已满或已关闭行程不会允许继续收藏

### 4.5 H5 个人中心

- [ ] 头像上传成功时能更新展示
- [ ] 上传失败时会回退原头像
- [ ] 昵称更新正常
- [ ] 默认手机号更新正常
- [ ] 默认微信号更新正常
- [ ] 我的发布数量与我的收藏数量能展示

## 五、Admin 人工回归清单

### 5.1 登录与守卫

- [ ] 未登录访问 `/dashboard` 会跳转 `/login`
- [ ] 管理员登录成功后进入后台首页

### 5.2 看板

- [ ] 首页展示：
  - [ ] 行程总数
  - [ ] 用户总数
  - [ ] 当前有效行程数
  - [ ] 过期行程数
  - [ ] 收藏总数

### 5.3 行程管理

- [ ] 行程列表可打开
- [ ] 可按关键字和状态筛选
- [ ] 行程详情可查看
- [ ] 行程编辑页可打开
- [ ] 点击保存前会出现二次确认
- [ ] 可编辑完整字段：
  - [ ] 行程类型
  - [ ] 出发地
  - [ ] 目的地
  - [ ] 出发日期
  - [ ] 出发时间
  - [ ] 座位数
  - [ ] 价格
  - [ ] 是否可议价
  - [ ] 联系微信
  - [ ] 联系手机号
  - [ ] 备注
  - [ ] 行程状态（`active / full / closed`）

### 5.4 用户管理

- [ ] 用户列表可打开
- [ ] 用户详情可查看
- [ ] 用户详情页不出现封禁、删除等越界操作

### 5.5 P1 联调闭环

- [ ] H5 发布一条新行程成功
- [ ] Admin 打开该行程详情并进入编辑页
- [ ] Admin 完整修改至少 1 组 H5 可见字段（如出发地、目的地、出发日期、出发时间、座位数、状态）并保存成功
- [ ] 返回 H5 详情页或我的发布页，可看到对应字段的最新内容

---

## 六、当前已知边界

- H5 行程表单当前未包含价格、备注等未落地字段
- Admin 构建仍存在 chunk size warning，后续可单独做拆包优化
- H5 详情页不会自动热更新后台改动，当前需要手动刷新后才能看到最新行程内容
- Admin 行程编辑页的“人数/座位数”控件当前交互偏弱，自动化输入稳定性一般
- 后端启动日志当前存在明文打印 R2 配置的风险，修复前不要传播包含敏感信息的完整启动日志

## 七、当前已验证基线（2026-04-29）

本轮本地联调已完成以下验证：

- 初始化 SQL 可成功导入
- 后端可按 `MOONICK_ENV=dev make dev` 正常启动
- 管理员种子账号 `admin / admin123` 可正常登录 Admin 接口
- H5 与 Admin 的 Vite 开发代理可正常转发 `/api` 请求到后端
- H5 与 Admin 的测试、构建命令可执行成功
- `P1 联调闭环` 已验证通过：
  - H5 成功发布行程 `#2`
  - Admin 可编辑 H5 可见字段并保存成功
  - H5 刷新后可看到后台修改后的终点、时间、人数
  - Admin 将状态改为 `full` 后，H5 刷新可显示“已满”
- 后端过期任务已验证通过：
  - 启动日志可见任务执行记录
  - 过期测试行程可自动转为 `expired`
- H5 `refresh` 鉴权链路已验证通过：
  - `accessToken` 失效后可自动调用 `POST /api/v1/auth/refresh` 并重放请求
  - `refreshToken` 失效后会清理本地登录态并跳回登录页
