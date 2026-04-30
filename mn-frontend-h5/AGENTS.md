# mn-frontend-h5

本文件作为 `mn-frontend-h5` 的协作入口。目标是让开发者和代理快速了解当前页面范围、依赖契约和验证方式。

## 导航

- 当前模块入口：当前文件 `AGENTS.md`
- 仓库总入口：`../AGENTS.md`
- 文档索引：`../docs/index.md`
- H5 技术方案：`../docs/technical/h5.md`
- 技术总览：`../docs/technical/overview.md`
- 联调清单：`../docs/technical/api-checklist.md`

## 模块简介

`mn-frontend-h5` 是明叶同行的移动端业务前端，基于 React + Vite。

当前已实现：

- 登录、注册、登录回跳
- 首页行程列表
- 行程详情
- 发布行程
- 编辑行程
- 我的发布
- 我的收藏
- 个人中心
- 昵称、默认联系方式管理
- 头像上传与失败回退

## 启动方式

```bash
cd mn-frontend-h5
npm install
npm run dev
```

本地开发默认通过 Vite 代理将 `/api` 转发到 `http://127.0.0.1:6303`。

## 验证命令

```bash
cd mn-frontend-h5
npm run test
npm run build
```

定向测试：

```bash
cd mn-frontend-h5
npm run test -- publish-form.test.tsx favorite-toggle.test.tsx avatar-upload.test.tsx
```

## 后端契约

当前前端实现依赖以下后端协议：

- 接口前缀：`/api/v1/...`
- 行程状态：`active` / `full` / `closed` / `expired`
- 行程类型：`driver_post` / `passenger_post`
- 列表响应结构：`items` / `total` / `pageNum` / `pageSize`
- 统一业务响应结构：

```json
{
  "code": 1000,
  "message": "success",
  "data": {}
}
```

## 当前实现边界

- H5 已接入 `refresh` 续期流程：受保护请求遇到 `1006 / 1007` 时会自动尝试 refresh，成功后重放原请求，失败则清理本地登录态
- H5 的部分产品文案以后端已支持字段为准，没有扩展到未落地的价格、备注等字段
- 已验证 H5 发布后，Admin 修改字段与状态，H5 刷新后可同步看到最新内容

## 协作建议

修改页面表单、筛选项或状态文案前，先核对后端枚举值和字段含义。涉及接口适配时，优先以根目录 `AGENTS.md` 和 `docs/technical/h5.md` 为准。

涉及 H5 UI 变更时，先在仓库根目录 `htmls/*.html` 中做静态稿验证视觉和布局，确认效果后再同步到 `mn-frontend-h5` 的 React 代码与样式。后续默认沿用这个“HTML 先行、H5 同步”的流程。
