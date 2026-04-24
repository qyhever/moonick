# 明叶同行技术方案设计说明

## 1. 背景

本设计说明基于以下需求文档整理：

- `docs/requirements/overview.md`
- `docs/requirements/h5.md`
- `docs/requirements/admin.md`

项目当前形态为：

- `mn-backend`：已有 Go + Gin 基础骨架
- `mn-frontend-h5`：待实现
- `mn-frontend-admin`：待实现

目标是输出一套可直接指导开发的技术方案文档，而不是仅停留在架构概念层。

---

## 2. 设计结论

本次技术方案采用以下结构：

- `docs/technical/overview.md`
- `docs/technical/backend.md`
- `docs/technical/h5.md`
- `docs/technical/admin.md`

设计原则如下：

- 延续现有后端技术栈：`Go + Gin`
- 后端为单体分层架构，不拆微服务
- H5 与 Admin 共用同一后端服务，但路由域和鉴权域严格隔离
- 认证统一采用纯无状态 JWT 双令牌
- `accessToken` 有效期 2 小时，`refreshToken` 有效期 72 小时
- v1 不使用 Redis，不做黑名单，不做服务端会话存储
- 存储层采用 `MySQL + Cloudflare R2`

---

## 3. 核心补缺结论

为确保需求可直接开发，以下内容作为默认补缺规则纳入正式技术文档：

- H5 补充“编辑行程页”
- H5 若保留“删除”交互，后端按软删除/下线语义设计，不做物理删除
- 历史行程中的联系方式保存发布快照
- `expired` 为系统态，由定时任务维护
- 首页列表按发布时间倒序
- 我的收藏按收藏时间倒序
- 用户与管理员均采用独立 token 域，不能混用

---

## 4. 方案范围

纳入范围：

- 用户注册登录
- 管理员登录
- 行程发布、浏览、详情、编辑、关闭
- 收藏
- 个人中心
- 后台看板
- 后台行程管理
- 后台用户查询
- 头像上传
- 过期行程定时任务

不纳入范围：

- 支付
- 地图 POI 标准化
- 评价
- 多角色权限
- 投诉工单
- 复杂 BI 图表
- Redis 会话治理

---

## 5. 规格自检结果

### 5.1 占位符检查

无 `TODO`、`待定`、空章节。

### 5.2 一致性检查

- 认证方案统一为双令牌纯无状态 JWT
- 数据存储统一为 MySQL + R2
- H5、Admin、Backend 的边界已分别拆开，没有交叉冲突

### 5.3 范围检查

当前文档范围适合直接进入开发计划与实施，不需要进一步拆成多个完全独立项目。

### 5.4 模糊性检查

以下原需求中的模糊项已在技术文档中明确：

- 删除行程的实现语义
- 编辑行程页面补充
- 联系方式是否跟随用户资料变更
- `expired` 状态的维护方式
- 收藏列表与首页列表排序规则

---

## 6. 交付物

本轮交付物为以下正式技术文档：

- [总览技术方案](/Users/await/apros/moonick/docs/technical/overview.md)
- [后端技术方案](/Users/await/apros/moonick/docs/technical/backend.md)
- [H5 技术方案](/Users/await/apros/moonick/docs/technical/h5.md)
- [Admin 技术方案](/Users/await/apros/moonick/docs/technical/admin.md)

以上文档可作为后续开发、任务拆分、接口联调和测试设计依据。
