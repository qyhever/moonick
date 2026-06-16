# 明叶同行 P1 MySQL 落库与后台完整编辑设计

生成时间：2026-04-26
适用分支：`feature`
阶段目标：将当前后端从内存仓储切换为真实 MySQL 持久化，并补齐后台完整行程编辑链路。

## 1. 设计目标

本阶段只解决 3 类问题：

1. 当前 `mn-backend/internal/repository/mysql` 仍以轻量内存实现为主，服务重启后业务数据无法保留。
2. 新环境缺少可直接执行的数据库初始化脚本，后端启动依赖人工猜测表结构。
3. 当前 Admin 行程编辑页只支持修改 `status`，不满足“后台完整编辑核心行程字段”的需要。

本阶段完成后，系统应满足以下结果：

- 后端用户、管理员、行程、收藏数据可真实落库到 MySQL。
- 新环境可通过一份初始化 SQL 完成表结构准备。
- 后台可编辑行程核心字段，并保存到数据库。
- H5 与 Admin 现有已通过的交互闭环不被打碎。

## 2. 明确不做的事情

本阶段不包含以下内容：

- 不引入迁移框架。
- 不引入 ORM。
- 不补 H5 `refresh` 接口。
- 不做后台用户封禁、删除、恢复等运营动作。
- 不借机整体重构 controller/service 分层。
- 不把 H5 表单一次性扩到所有产品文档字段。

## 3. 总体方案

采用“手动初始化 SQL + 真实 MySQL repository + 后台完整行程编辑补齐”的最小稳定方案。

方案原则：

- repository 接口签名尽量不变。
- service 层错误语义尽量不变。
- controller 路由与业务码协议尽量不变。
- 现有前端已适配的接口结构优先保持兼容。

## 4. 数据库与初始化 SQL

### 4.1 目录位置

初始化 SQL 放在：

- `mn-backend/docs/sql/001_init.sql`

后续如果继续追加结构，按 `002_*.sql`、`003_*.sql` 递增。

### 4.2 表结构

#### `users`

字段：

- `id`
- `phone`
- `password_hash`
- `nickname`
- `avatar_url`
- `default_wechat`
- `default_phone`
- `status`
- `created_at`
- `updated_at`

约束：

- `phone` 唯一索引

#### `admins`

字段：

- `id`
- `username`
- `password_hash`
- `display_name`
- `status`
- `last_login_at`
- `created_at`
- `updated_at`

约束：

- `username` 唯一索引

#### `trips`

字段：

- `id`
- `publisher_user_id`
- `trip_type`
- `from_city_text`
- `to_city_text`
- `departure_date`
- `departure_time`
- `departure_at`
- `seat_count`
- `price_amount`
- `is_price_negotiable`
- `contact_wechat`
- `contact_phone`
- `remark`
- `status`
- `closed_reason`
- `deleted_at`
- `created_at`
- `updated_at`

关键约束：

- `trip_type` 只允许 `driver_post / passenger_post`
- `status` 只允许 `active / full / closed / expired`
- `seat_count` 限制 `1 ~ 6`
- `price_amount` 可空
- `remark` 可空
- `deleted_at` 可空

说明：

- `price_amount` 与 `remark` 是本阶段为后台完整编辑补齐的核心字段。
- `closed_reason` 先落库，不在本阶段扩成完整枚举体系。

#### `trip_favorites`

字段：

- `id`
- `user_id`
- `trip_id`
- `created_at`

约束：

- 唯一索引：`(user_id, trip_id)`

### 4.3 初始化策略

- SQL 只负责建表和索引，不写死业务数据。
- 管理员种子仍通过后端配置注入。
- AGENTS 文档与联调文档明确要求“先导入 SQL，再启动服务”。

## 5. Repository 落库方案

### 5.1 总体原则

`internal/repository/mysql` 保持现有仓储边界，不新增大规模抽象层。

本阶段继续保留以下仓储：

- `UserRepository`
- `AdminRepository`
- `TripRepository`
- `FavoriteRepository`

每个仓储改成真实 MySQL 读写。

### 5.2 数据访问方式

本阶段使用标准库 `database/sql`。

原因：

- 依赖最轻。
- SQL 行为最可控。
- 更容易在不引入新框架的情况下保持现有测试语义。

### 5.3 用户仓储

需要支持：

- `Create`
- `FindByPhone`
- `FindByID`
- `UpdateProfile`
- `UpdateContact`
- `UpdateAvatarURL`
- `List`
- `Count`

关键要求：

- 唯一索引冲突可被稳定映射为“重复注册”语义。
- 列表与计数要支撑后台用户查询。

### 5.4 管理员仓储

需要支持：

- `FindByUsername`
- `FindByID`
- 管理员种子初始化或更新

关键要求：

- 启动时如果配置了管理员账号，则执行“存在则更新，不存在则插入”。
- 登录查询统一走数据库，不再依赖纯内存管理员对象。

### 5.5 行程仓储

需要支持：

- `Create`
- `Update`
- `FindByID`
- `List`

关键要求：

- `departure_date`、`departure_time`、`departure_at` 同时落库。
- `price_amount`、`remark`、`closed_reason`、`deleted_at` 支持映射。
- 列表逻辑继续支撑：
  - H5 首页
  - 我的发布
  - 后台行程列表
  - 用户发布列表

### 5.6 收藏仓储

需要支持：

- `Exists`
- `Create`
- `Delete`
- `List`
- `Count`
- `CountByUser`

关键要求：

- 收藏占位逻辑仍保留在 service 层，不下沉到 repository。
- 唯一索引冲突语义应稳定。

## 6. 后台完整行程编辑

### 6.1 编辑范围

本阶段后台允许编辑的字段：

- `tripType`
- `fromText`
- `toText`
- `departureDate`
- `departureTime`
- `seatCount`
- `priceAmount`
- `isPriceNegotiable`
- `contactWechat`
- `contactPhone`
- `remark`
- `status`

### 6.2 路由策略

继续使用现有接口：

- `PUT /api/admin/v1/trips/:id`

但将语义从“只改状态”升级为“后台完整编辑核心字段”。

原因：

- 路由已存在。
- Admin 前端已有接入点。
- 不必额外引入一条临时路由。

### 6.3 后端请求模型

新增后台专用 request：

- `AdminUpdateTripDetailRequest`

不复用用户侧 `UpsertTripRequest`，原因如下：

- 后台和用户侧能力边界不同。
- 后台需要额外编辑 `status`。
- 后台要补齐 `priceAmount` 与 `remark` 等字段。

### 6.4 服务层策略

新增或升级后台完整编辑逻辑：

- `AdminService.UpdateTripDetail(...)`

不建议把完整编辑继续塞进当前只改状态的逻辑中。

这样可以保持：

- 旧语义边界清晰
- 完整编辑与纯状态变更各自独立
- 后续扩展更容易测试

### 6.5 校验规则

后台完整编辑沿用并加强当前行程校验：

- 起点不能为空
- 终点不能为空
- 起点和终点不能相同
- 出发日期不能为空
- 出发时间不能为空
- 出发时间不能早于当前时间
- `seatCount` 必须在 `1 ~ 6`
- `contactWechat` 和 `contactPhone` 至少填写一种
- `tripType` 只能是 `driver_post / passenger_post`
- `status` 只能是 `active / full / closed`
- `expired` 不能人工改回
- `priceAmount` 在 `isPriceNegotiable=false` 时必须合法
- `remark` 长度限制为 `<= 1000`

### 6.6 前端 Admin 编辑页

当前 [TripEditPage.tsx](/Users/await/apros/moonick/mn-frontend-admin/src/features/trips/TripEditPage.tsx) 只支持修改状态，本阶段升级为完整表单。

表单字段：

- `tripType`
- `fromText`
- `toText`
- `departureDate`
- `departureTime`
- `seatCount`
- `priceAmount`
- `isPriceNegotiable`
- `contactWechat`
- `contactPhone`
- `remark`
- `status`

交互要求：

- 页面初始化先加载详情并回填
- 保存前二次确认保留
- 保存成功后跳回详情页

实现取舍：

- `departureDate / departureTime` 本阶段继续采用字符串输入
- 不强制引入 `DatePicker / TimePicker`

原因：

- 与当前后端字符串格式完全对齐
- 测试更轻
- 避免在“先落库”阶段扩大前端复杂度

## 7. 实施顺序

按以下顺序推进：

1. 初始化 SQL 与文档
2. 数据库连接入口
3. 用户与管理员仓储落库
4. 行程与收藏仓储落库
5. 后台完整编辑后端接口
6. Admin 完整编辑页升级
7. 全量测试与联调验证

## 8. 风险控制

### 8.1 service 语义被仓储实现改变

控制方式：

- repository 切换时不先改 service 业务判断
- 维持“找不到 / 重复 / 权限 / 无效状态”等错误语义

### 8.2 前端响应结构被打破

控制方式：

- 后端 response 尽量兼容，新增字段只增不删
- H5 当前已稳定的响应结构不主动打散

### 8.3 管理员种子与数据库数据冲突

控制方式：

- 启动时执行“存在则更新，不存在则插入”
- 登录统一走数据库查询

### 8.4 后台编辑范围膨胀

控制方式：

- 只覆盖行程核心字段
- 不扩展到删除、封禁、恢复等运营动作

## 9. 验证策略

### 9.1 SQL 层

- 手动导入 `001_init.sql`
- 确认表与索引创建成功

### 9.2 Repository 层

至少覆盖：

- 用户创建、查询、更新
- 管理员查询
- 行程创建、查询、更新、列表
- 收藏新增、取消、列表

### 9.3 Service / API 层

执行：

```bash
cd mn-backend
GOCACHE=/tmp/moonick-gocache go test ./...
```

重点验证：

- 认证链路
- 用户资料链路
- 行程链路
- 后台完整编辑链路

### 9.4 前端层

H5：

```bash
cd mn-frontend-h5
npm run test
npm run build
```

Admin：

```bash
cd mn-frontend-admin
npm run test
npm run build
```

### 9.5 人工联调路径

至少走通以下链路：

1. 导入 SQL
2. 启动后端
3. H5 注册用户
4. H5 发布行程
5. Admin 登录并打开该行程
6. Admin 修改起点、终点、出发时间、人数、联系方式、备注、状态
7. H5 回看详情和我的发布，确认字段变化生效
8. H5 收藏该行程，确认前后台查询都正常

## 10. 设计结论

本阶段以“稳定落库”和“补齐后台完整编辑”为唯一目标。

最终结论如下：

- 用一份手动初始化 SQL 管理表结构
- 用 `database/sql` 替换当前内存仓储实现
- 保持现有路由、业务码、分页协议和状态枚举不变
- 将 Admin 行程编辑从“状态编辑”升级为“核心字段完整编辑”
- 通过后端全量测试、双前端测试构建和一条人工联调路径完成验收
