# 明叶同行

本文件是仓库级协作入口，用于帮助开发者和代理快速定位代码、文档和联调路径。

## 快速导航

- 文档索引：`docs/index.md`
- 需求总览：`docs/requirements/overview.md`
- 技术总览：`docs/technical/overview.md`
- 联调与验收：`docs/technical/api-checklist.md`
- 后端入口：`mn-backend/AGENTS.md`
- H5 入口：`mn-frontend-h5/AGENTS.md`
- Admin 入口：`mn-frontend-admin/AGENTS.md`

## 项目结构

- `mn-backend`：Go + Gin 后端服务
- `mn-frontend-h5`：React + Vite 的 H5 业务端
- `mn-frontend-admin`：React + Vite + Ant Design 的 PC 管理端
- `docs/requirements`：产品需求与范围边界
- `docs/technical`：技术方案、共享契约、联调清单
- `docs/superpowers`：历史计划、过程记录与归档，不作为当前事实源

## 当前共享契约

- H5 接口前缀：`/api/v1/...`
- Admin 接口前缀：`/api/admin/v1/...`
- 行程状态：`active` / `full` / `closed` / `expired`
- 行程类型：`driver_post` / `passenger_post`

共享契约的主定义以 `docs/technical/*` 为准；如与其他文档不一致，应优先修正文档而不是自行猜测。

## 最短启动路径

1. 先看 `docs/technical/api-checklist.md`
2. 按清单导入 `mn-backend/docs/sql/001_init.sql`
3. 启动后端：`cd mn-backend && MOONICK_ENV=dev make dev`
4. 启动 H5：`cd mn-frontend-h5 && npm install && npm run dev`
5. 启动 Admin：`cd mn-frontend-admin && npm install && npm run dev`

## 修改前必看

- 改需求范围、页面能力、非功能边界：看 `docs/requirements/*`
- 改接口、枚举、鉴权、数据模型、联调流程：看 `docs/technical/*`
- 改某个子项目实现：先看对应子目录 `AGENTS.md`

## 维护约定

- `AGENTS.md` 只保留入口、导航、最小协作约定，不重复承载完整技术细节。
- 需求文档回答“要做什么”，技术文档回答“怎么实现”，联调清单回答“怎么验证”。
- 历史计划和总结只放在 `docs/superpowers`，不作为当前规范来源。
