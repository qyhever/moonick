# 明叶同行 H5 登录态 refresh 流程完成记录

日期：2026-04-29
分支：`feature`
范围：H5 用户端登录态自动续期闭环

## 一、本次完成内容

本轮已完成 P1 收口后的下一步任务第 2 项：

- 补齐 H5 登录态 `refresh` 流程

实际落地包括：

- 后端新增用户端 `POST /api/v1/auth/refresh`
- `AuthService` 新增基于 `refreshToken` 的续期能力
- H5 `auth` store 从“显式占位”改为真实调用 `refresh` 接口
- H5 请求层新增自动续期逻辑，支持：
  - 业务响应返回 `1006 / 1007` 时自动尝试 refresh
  - refresh 成功后自动重放原请求
  - refresh 失败后清理本地登录态
- 补充前后端自动化测试，覆盖 refresh 成功与失败分支

## 二、当前真实能力

当前 H5 登录态关于 refresh 的真实能力为：

- token 存储：`accessToken` 与 `refreshToken` 持久化在 `localStorage`
- 触发时机：受保护请求收到业务码 `1006`（需要登录）或 `1007`（无效 token）时
- 续期接口：`POST /api/v1/auth/refresh`
- 鉴权方式：使用 `Authorization: Bearer <refreshToken>`
- 成功行为：更新本地 token 与用户信息，并重放原请求
- 失败行为：清空本地登录态，后续由路由守卫重新要求登录

这意味着：

- H5 的 refresh 现在不再是“规格占位”，而是已接入真实前后端链路
- access token 失效后，常规受保护请求可以自动恢复，不需要用户手动重新登录

## 三、已同步文档

本次已同步以下文档：

- H5 协作入口：`mn-frontend-h5/AGENTS.md`
- 技术总览：`docs/technical/overview.md`
- 联调与验收清单：`docs/technical/api-checklist.md`

## 四、已验证结果

本轮已执行并通过：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...

cd mn-frontend-h5
npm run test
npm run build
```

结果：通过

补充说明：

- 本轮验证覆盖了后端 refresh 接口、H5 自动续期、请求重放和失败登出回退
- 尚未在本文档内声明“已完成浏览器人工联调”，该项已补入 `docs/technical/api-checklist.md`

## 五、下一步建议

接下来优先级可顺延为：

1. 做 H5 refresh 的浏览器人工联调
2. 视产品需求决定 H5 是否展示并编辑价格、备注
3. 对 Admin 做拆包优化，消掉构建 warning
