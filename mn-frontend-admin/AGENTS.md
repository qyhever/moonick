# mn-frontend-admin

本文件作为 `mn-frontend-admin` 的协作入口。目标是让开发者和代理快速掌握管理端能力范围、接口依赖和当前实现边界。

## 导航

- 当前模块入口：当前文件 `AGENTS.md`
- 仓库总入口：`../AGENTS.md`
- 文档索引：`../docs/index.md`
- Admin 技术方案：`../docs/technical/admin.md`
- 技术总览：`../docs/technical/overview.md`
- 联调清单：`../docs/technical/api-checklist.md`

## 模块简介

`mn-frontend-admin` 是明叶同行的 PC 管理端，基于 React + Vite + Ant Design。

当前已实现：

- 管理员登录页
- 未登录守卫与跳转登录
- 首页运营看板
- 行程列表
- 行程详情
- 行程状态编辑
- 用户列表
- 用户只读详情

## 启动方式

```bash
cd mn-frontend-admin
npm install
npm run dev
```

## 验证命令

```bash
cd mn-frontend-admin
npm run test
npm run build
```

定向测试：

```bash
cd mn-frontend-admin
npm run test -- login-guard.test.tsx dashboard-page.test.tsx
npm run test -- trip-edit.test.tsx user-readonly.test.tsx
```

## 后端契约

当前前端实现依赖以下后端协议：

- 接口前缀：`/api/admin/v1/...`
- 管理员登录接口：`POST /api/admin/v1/auth/login`
- 看板接口：`GET /api/admin/v1/dashboard/summary`
- 行程编辑当前只支持状态更新

看板当前展示的真实字段是：

- `totalTrips`
- `totalUsers`
- `activeTrips`
- `expiredTrips`
- `totalFavorites`

## 当前实现边界

- 行程编辑当前收敛为“状态编辑”，因为后端尚未提供完整的后台全字段编辑接口
- 用户详情页当前是只读页，不支持封禁、删除、编辑
- 构建目前会有 Vite 的大包体积 warning，但不影响产物生成

## 协作建议

修改管理端表格、详情字段或操作按钮前，先确认后端管理接口是否已提供对应能力。涉及联调时，优先核对根目录 `AGENTS.md`、`docs/technical/admin.md` 和后端实际路由。
