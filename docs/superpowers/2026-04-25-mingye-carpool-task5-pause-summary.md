# 明叶同行任务 5 暂停前变更摘要

生成时间：2026-04-25
分支：`feature`
阶段：任务 5（H5 核心业务页）修正中

## 当前结论

- 任务 1 已完成
- 任务 2 已完成
- 任务 3 已完成
- 任务 4 已完成
- 任务 5 已完成一轮实现后的契约对齐修正，但尚未重新执行测试与构建验证

这意味着：

- 当前工作区里的任务 5 代码已经不是最初子代理交付版本
- 我已按后端真实接口契约重写了任务 5 的关键页面和 API 封装
- 由于用户要求先暂停，我还没有重新运行 `npm run test` 与 `npm run build`
- 因此当前不能宣称任务 5 已闭环通过

## 触发本轮修正的原因

在恢复任务 5 后，我重新执行并确认了以下事实：

- `mn-frontend-h5` 的测试与构建当时可以通过
- 但前端实现与当前后端契约存在实质性偏差，属于集成风险，不是样式或命名问题

已确认的不一致点包括：

- 行程状态前端写成了 `draft / active / completed / cancelled`，后端实际为 `active / full / closed / expired`
- 前端把列表接口当成 `TripSummary[]` 使用，后端实际返回 `items / total / pageNum / pageSize`
- 前端详情和列表依赖了不存在的字段：
  - `departureAt`
  - `seatsAvailable`
  - `pricePerSeat`
  - `note`
  - `publisher`
- 前端发布/编辑 payload 使用了错误字段：
  - `seatsAvailable`
  - `pricePerSeat`
  - `note`
- 后端当前真实字段是：
  - `tripType`
  - `departureDate`
  - `departureTime`
  - `seatCount`
  - `isPriceNegotiable`
  - `contactWechat`
  - `contactPhone`
- 个人中心页只做了头像上传，没有把昵称、默认联系方式管理补到任务 5 所需的最小闭环

## 本轮已完成的修正

### 1. 行程 API 与类型对齐后端契约

已修改：

- `mn-frontend-h5/src/features/trips/api.ts`

主要变更：

- 新增 `TripType`，按当前后端实现使用 `driver_post / passenger_post`
- `TripStatus` 改为 `active / full / closed / expired`
- `TripSummary` 改为对齐后端字段：
  - `userId`
  - `tripType`
  - `departureDate`
  - `departureTime`
  - `seatCount`
  - `isPriceNegotiable`
  - `favorited`
  - `unavailable`
- `TripDetail` 改为补齐：
  - `contactPhone`
  - `contactWechat`
  - `createdAt`
  - `updatedAt`
- 列表接口改为消费 `TripListResponse`
- 发布/编辑 payload 改为：
  - `tripType`
  - `fromText`
  - `toText`
  - `departureDate`
  - `departureTime`
  - `seatCount`
  - `isPriceNegotiable`
  - `contactPhone`
  - `contactWechat`

### 2. 首页、详情、发布、编辑、我的发布、我的收藏改为真实字段

已修改：

- `mn-frontend-h5/src/pages/HomePage.tsx`
- `mn-frontend-h5/src/features/trips/components/TripCard.tsx`
- `mn-frontend-h5/src/features/trips/pages/TripDetailPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/PublishPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/EditTripPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/MyTripsPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/MyFavoritesPage.tsx`

主要变更：

- 首页改为从 `TripListResponse.items` 取数据
- `TripCard` 不再依赖不存在的 `publisher / pricePerSeat / note`
- 卡片显示逻辑改为基于：
  - `tripType`
  - `departureDate + departureTime`
  - `seatCount`
  - `isPriceNegotiable`
  - `status`
  - `favorited`
  - `unavailable`
- 首页对 `full` 行程禁用点击，符合技术方案里的“已满禁用”要求
- 发布页增加并校验：
  - 行程类型
  - 人数范围 `1 ~ 6`
  - 出发时间不能早于当前时间
  - 至少填写一种联系方式
- 编辑页按真实详情字段回填表单，并复用发布页同一套核心校验
- 详情页区分本人/他人：
  - 本人可编辑、关闭行程
  - 他人可收藏、联系
- 我的发布页补了状态更新操作：
  - 设为满员
  - 关闭
- 我的收藏页补了失效收藏展示：
  - 当 `unavailable = true` 时展示“该行程已下线或不存在”

### 3. 个人中心补到最小可用闭环

已新增：

- `mn-frontend-h5/src/features/profile/api.ts`

已修改：

- `mn-frontend-h5/src/features/profile/pages/ProfilePage.tsx`
- `mn-frontend-h5/src/features/profile/components/AvatarUploader.tsx`
- `mn-frontend-h5/src/store/auth.ts`

主要变更：

- 个人中心新增真实接口调用：
  - `GET /api/v1/users/me`
  - `PUT /api/v1/users/profile`
  - `PUT /api/v1/users/contact`
- 页面现在支持：
  - 更新昵称
  - 更新默认手机号
  - 更新默认微信号
  - 拉取并展示我的发布数量
  - 拉取并展示我的收藏数量
- `auth store` 新增 `setUser`，用于同步头像、昵称、联系方式等本地登录态
- 头像上传增加：
  - `accept` 类型限制
  - `10 MB` 大小限制
  - 上传失败回退原头像预览

### 4. 测试同步更新

已修改：

- `mn-frontend-h5/src/test/publish-form.test.tsx`
- `mn-frontend-h5/src/test/favorite-toggle.test.tsx`

主要变更：

- 发布测试新增断言，验证前端提交的是后端兼容 payload
- 收藏测试 mock 数据改为真实详情结构

## 本轮涉及文件

- `mn-frontend-h5/src/store/auth.ts`
- `mn-frontend-h5/src/features/trips/api.ts`
- `mn-frontend-h5/src/features/trips/components/TripCard.tsx`
- `mn-frontend-h5/src/features/trips/pages/PublishPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/EditTripPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/TripDetailPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/MyTripsPage.tsx`
- `mn-frontend-h5/src/features/trips/pages/MyFavoritesPage.tsx`
- `mn-frontend-h5/src/pages/HomePage.tsx`
- `mn-frontend-h5/src/features/profile/api.ts`
- `mn-frontend-h5/src/features/profile/pages/ProfilePage.tsx`
- `mn-frontend-h5/src/features/profile/components/AvatarUploader.tsx`
- `mn-frontend-h5/src/test/publish-form.test.tsx`
- `mn-frontend-h5/src/test/favorite-toggle.test.tsx`
- `mn-frontend-h5/src/styles/h5.css`

## 当前未验证项

本轮修正完成后，以下命令还没有重新执行：

```bash
cd mn-frontend-h5 && npm run test -- publish-form.test.tsx favorite-toggle.test.tsx avatar-upload.test.tsx
cd mn-frontend-h5 && npm run test
cd mn-frontend-h5 && npm run build
```

因此当前状态应定义为：

- 已修改
- 未验证
- 未完成任务 5 闭环

## 下次恢复建议

建议从以下顺序继续：

1. 先执行任务 5 的定向测试：

```bash
cd mn-frontend-h5 && npm run test -- publish-form.test.tsx favorite-toggle.test.tsx avatar-upload.test.tsx
```

2. 再执行全量前端测试：

```bash
cd mn-frontend-h5 && npm run test
```

3. 最后执行构建验证：

```bash
cd mn-frontend-h5 && npm run build
```

4. 若验证失败，优先修正任务 5 本轮引入的问题，再进入规格审查与代码质量审查
5. 若验证通过，再继续任务 5 的规格审查和代码质量审查闭环

## 备注

- 当前环境仍然不能稳定写入 `.git/index.lock`，无法代为提交 commit
- 本次文档只记录任务 5 暂停前的增量变化，不替代更早的整体进展存档：
  - `docs/superpowers/2026-04-25-mingye-carpool-progress-summary.md`
