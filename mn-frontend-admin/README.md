# mn-frontend-admin

## 简介

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

## 依赖安装与启动

```bash
cd mn-frontend-admin
npm install
npm run dev
```

## 测试与构建

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

## 后端契约约束

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

## 当前边界

- 行程编辑当前收敛为“状态编辑”，因为后端尚未提供完整的后台全字段编辑接口
- 用户详情页当前是只读页，不支持封禁、删除、编辑
- 构建目前会有 Vite 的大包体积 warning，但不影响产物生成
