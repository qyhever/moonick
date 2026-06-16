# 明叶同行 v1 首版闭环最终摘要

生成时间：2026-04-26
分支：`feature`

## 当前结论

本轮计划中的 8 个任务已经推进到可联调、可验证、可提交的状态。

当前实际完成情况：

- 任务 1：完成
- 任务 2：完成
- 任务 3：完成
- 任务 4：完成
- 任务 5：完成
- 任务 6：完成
- 任务 7：完成
- 任务 8：完成

这里的“完成”是指：

- 代码已落地到当前工作区
- 自动化测试已实际执行
- H5 和 Admin 均已完成构建验证
- 启动说明与联调文档已补齐

## 本轮交付范围

### 1. 后端 `mn-backend`

已落地能力：

- JWT 基础设施
- H5 用户注册、登录、`/auth/me`
- 管理员登录、`/auth/me`
- 用户资料查询、昵称更新、联系方式更新
- 头像上传
- 行程列表、详情、创建、更新、我的发布
- 收藏 toggle、我的收藏
- 后台看板摘要
- 后台行程列表、详情、状态更新
- 后台用户列表、详情、用户行程列表
- 行程过期任务骨架

当前统一协议：

- H5 接口：`/api/v1/...`
- Admin 接口：`/api/admin/v1/...`
- 行程状态：`active / full / closed / expired`
- 行程类型：`driver_post / passenger_post`
- 分页参数：`pageNum / pageSize`
- token 字段：`accessToken / refreshToken`

### 2. H5 前端 `mn-frontend-h5`

已落地页面与能力：

- 登录页
- 注册页
- 首页
- 行程详情页
- 发布行程页
- 编辑行程页
- 我的发布列表
- 我的收藏列表
- 个人中心页

已闭环行为：

- 未登录访问受保护页面时跳转登录
- 登录成功后按 `redirect` 回跳
- 发布页按真实后端契约提交
- 本人可编辑和关闭自己的行程
- 我的发布支持设为满员、关闭
- 我的收藏支持展示失效行程占位
- 个人中心支持更新昵称、默认手机号、默认微信号
- 头像上传支持本地预览与失败回退

### 3. Admin 前端 `mn-frontend-admin`

已落地页面与能力：

- 管理员登录页
- 登录守卫
- 首页看板
- 行程列表页
- 行程详情页
- 行程编辑页
- 用户列表页
- 用户详情页

已闭环行为：

- 未登录访问后台会自动跳转登录页
- 首页看板读取后台真实摘要字段
- 行程列表支持基础筛选
- 行程详情可查看
- 行程编辑支持状态更新，并在保存前二次确认
- 用户详情为只读页，不提供封禁、删除等越界操作

## 文档补充

本轮新增或更新了以下联调与启动文档：

- `AGENTS.md`
- `mn-backend/AGENTS.md`
- `mn-frontend-h5/AGENTS.md`
- `mn-frontend-admin/AGENTS.md`
- `docs/technical/api-checklist.md`

此外，历史过程文档保留：

- `docs/superpowers/plans/2026-04-24-mingye-carpool-v1-implementation.md`
- `docs/superpowers/2026-04-25-mingye-carpool-progress-summary.md`
- `docs/superpowers/2026-04-25-mingye-carpool-task5-pause-summary.md`

## 实际验证结果

### 后端

已执行：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

结果：

- PASS

说明：

- 当前环境下 Go 默认缓存目录无写权限，因此需要显式指定 `GOCACHE=/tmp/moonick-gocache`

### H5

已执行：

```bash
cd mn-frontend-h5
npm run test -- publish-form.test.tsx favorite-toggle.test.tsx avatar-upload.test.tsx
npm run test
npm run build
```

结果：

- PASS

### Admin

已执行：

```bash
cd mn-frontend-admin
npm run test -- login-guard.test.tsx dashboard-page.test.tsx
npm run test -- trip-edit.test.tsx user-readonly.test.tsx
npm run test
npm run build
```

结果：

- PASS

说明：

- `npm run build` 当前会有 Vite 的大包体积 warning，但构建成功，不是阻塞项

## 当前已知边界

### 1. 后端边界

- 当前 repository 仍以轻量内存实现为主
- 后续切换真实 MySQL 时，需要保持：
  - 路由边界不变
  - 业务码语义不变
  - 分页协议不变
  - 行程状态枚举不变

### 2. H5 边界

- `refresh` 接口尚未接入，当前仍为显式占位
- 当前 H5 表单未扩展到价格、备注等未落地字段

### 3. Admin 边界

- 当前后台行程编辑只支持状态更新
- 原因是后端当前只提供了后台状态更新接口，并未提供完整字段编辑能力
- 后台构建存在 chunk size warning，后续可单独做拆包优化

## 当前工作区说明

与本轮实现无关但仍在工作区中的内容：

- `.codex/config.toml`

当前环境限制：

- 无法稳定写入 `.git/index.lock`
- 无法代为执行提交
- 本轮也无法代为清理部分构建目录

## 建议提交方式

如果按模块拆分提交，建议：

```bash
git add mn-backend AGENTS.md mn-backend/AGENTS.md docs/technical/api-checklist.md docs/superpowers
git commit -m "feat: 完成明叶同行后端与联调文档"

git add mn-frontend-h5
git commit -m "feat: 完成 H5 业务端首版闭环"

git add mn-frontend-admin
git commit -m "feat: 完成管理后台首版骨架和业务页"
```

如果一次性提交，建议：

```bash
git add AGENTS.md mn-backend docs/technical/api-checklist.md docs/superpowers mn-frontend-h5 mn-frontend-admin
git commit -m "feat: 完成明叶同行 v1 首版闭环"
```

## 建议的下一步

如果后续继续推进，优先级建议如下：

1. 把后端 repository 从内存实现切到真实 MySQL
2. 为 Admin 补完整字段级行程编辑接口
3. 为 H5 接入 refresh 流程
4. 为 Admin 做构建拆包，消除 chunk size warning
