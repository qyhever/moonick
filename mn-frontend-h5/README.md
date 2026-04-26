# mn-frontend-h5

## 简介

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

## 依赖安装与启动

```bash
cd mn-frontend-h5
npm install
npm run dev
```

## 测试与构建

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

## 后端契约约束

当前前端实现依赖以下后端协议：

- 接口前缀：`/api/v1/...`
- 行程状态：`active / full / closed / expired`
- 行程类型：`driver_post / passenger_post`
- 列表响应结构：`items / total / pageNum / pageSize`
- 统一业务响应结构：

```json
{
  "code": 1000,
  "message": "success",
  "data": {}
}
```

## 说明

- 当前前端没有接入 refresh 接口，`refresh` 仍是显式占位
- H5 的部分产品文案以当前后端已支持字段为准，没有扩展到未落地的价格、备注等字段
