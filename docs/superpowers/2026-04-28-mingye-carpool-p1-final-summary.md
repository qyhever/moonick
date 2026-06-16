# 明叶同行 P1 MySQL 落库与后台完整编辑最终变更存档

日期：2026-04-28
分支：`feature`
范围：P1 MySQL 落库与后台完整编辑

## 一、最终状态

本轮 P1 计划已全部完成：

- 任务 1：初始化 SQL 与数据库连接入口
- 任务 2：用户与管理员仓储切换到 MySQL
- 任务 3：行程与收藏仓储切换到 MySQL
- 任务 4：后台完整行程编辑后端接口
- 任务 5：Admin 完整编辑页升级
- 任务 6：文档补充与全链路验证

所有任务均已完成以下闭环：

- 实现
- 规格审查
- 代码质量复审
- 主线程验证

## 二、实际落地能力

### 2.1 后端

后端当前已从内存演示态切换为真实 MySQL 持久化，核心数据链路包括：

- 用户
- 管理员
- 行程
- 收藏

已落地能力包括：

- `mn-backend/docs/sql/001_init.sql` 初始化脚本
- 数据库连接入口与管理员 seed upsert
- 用户注册、登录、资料更新、头像上传
- 管理员登录、看板、用户查询
- 行程创建、查询、编辑、状态更新
- 收藏创建、取消、查询
- 后台完整行程编辑

后台完整行程编辑当前支持字段：

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

当前后端还保留了一个兼容边界：

- `PUT /api/admin/v1/trips/:id` 同时兼容 legacy `{status}` 请求和完整编辑 payload

### 2.2 Admin 前端

Admin 端已从“只改状态”升级为“完整编辑表单 + 二次确认”。

实际落地包括：

- 行程详情页展示 `priceAmount`、`remark`
- 编辑页完整回填后端详情字段
- 保存时提交完整 payload
- 继续保留 `ConfirmSubmitButton` 二次确认
- `expired` 行程不再展示误导性的编辑入口
- `expired` 行程进入编辑页时不可保存

### 2.3 文档

文档已收敛到当前真实能力：

- 根目录 [AGENTS.md](/Users/await/apros/moonick/AGENTS.md)
- 后端 [mn-backend/AGENTS.md](/Users/await/apros/moonick/mn-backend/AGENTS.md)
- 联调清单 [docs/technical/api-checklist.md](/Users/await/apros/moonick/docs/technical/api-checklist.md)

已补充并修正：

- MySQL 初始化 SQL 导入方式
- `mn-backend/docs/sql/001_init.sql` 路径
- P1 联调闭环路径
- 后端真实 MySQL 持久化表述
- Admin 完整字段编辑表述

## 三、已验证结果

本轮主线程实际执行并通过：

### 3.1 后端

```bash
cd mn-backend && GOCACHE=/tmp/moonick-gocache go test ./...
```

结果：通过

### 3.2 H5

```bash
cd mn-frontend-h5 && npm run test && npm run build
```

结果：通过

### 3.3 Admin

```bash
cd mn-frontend-admin && npm run test && npm run build
```

结果：通过

补充说明：

- `mn-frontend-admin` 构建仍有 Vite chunk size warning
- 该 warning 不影响产物生成，不构成当前 P1 阻塞项

## 四、当前真实边界

本轮完成后，仍存在这些明确边界：

- H5 `refresh` 仍是占位逻辑，尚未接入完整续期流程
- H5 行程表单当前未接价格、备注等字段
- 行程过期任务当前只有任务骨架与 repository 能力，尚未接入实际调度启动链路
- Admin 构建仍存在 chunk size warning，后续可单独做拆包优化

## 五、建议提交范围

如果准备提交，这一轮建议纳入：

- `mn-backend/`
- `mn-frontend-admin/`
- `AGENTS.md`
- `docs/technical/api-checklist.md`
- `docs/superpowers/` 下与本轮 P1 相关的规格、计划、暂停摘要和最终存档

## 六、下一阶段建议

P1 收口后，下一阶段更合理的优先级是：

1. 接通行程过期任务的真实调度链路
2. 补 H5 登录态 refresh 流程
3. 视产品需求决定 H5 是否展示并编辑价格、备注
4. 对 Admin 做拆包优化，消掉构建 warning
