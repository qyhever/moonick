# 文档索引

本文档说明仓库内各类文档的职责边界、查阅顺序和维护规则。

## 推荐阅读顺序

1. 仓库入口：`../AGENTS.md`
2. 需求总览：`requirements/overview.md`
3. 技术总览：`technical/overview.md`
4. 联调与验收：`technical/api-checklist.md`
5. 子项目入口：
   - `../mn-backend/AGENTS.md`
   - `../mn-frontend-h5/AGENTS.md`
   - `../mn-frontend-admin/AGENTS.md`

## 目录职责

### `AGENTS.md`

用于协作入口和快速上手：

- 仓库结构
- 文档导航
- 最短启动路径
- 修改前必看事项

不承担完整 PRD、完整技术方案或完整回归步骤。

### `docs/requirements`

用于描述产品范围和需求边界：

- 做什么
- 不做什么
- 页面和角色范围
- 业务规则与产品口径

这里不写实现细节，不写具体代码结构。

### `docs/technical`

用于描述当前实现和共享技术契约：

- 接口前缀
- 枚举值
- 鉴权规则
- 数据模型
- 技术边界
- 联调与验收路径

这里是当前共享契约的主事实源。

### `docs/superpowers`

用于保存历史计划、暂停摘要、过程记录和阶段总结。

这里是过程归档区，不是当前规范来源。除非需要追溯决策过程，否则优先阅读 `requirements`、`technical` 和各级 `AGENTS.md`。

## 文档更新规则

### 变更类型与必改文档

| 变更类型 | 必须检查或更新的文档 |
|---|---|
| 产品范围、页面能力、角色权限边界 | `docs/requirements/*` |
| 接口、鉴权、枚举、数据模型、共享约束 | `docs/technical/*` |
| 启动、联调、验收步骤 | `docs/technical/api-checklist.md` |
| 仓库入口、阅读路径、协作说明 | 根目录或子目录 `AGENTS.md` |
| 历史过程沉淀 | `docs/superpowers/*` |

### 更新原则

- 同一条共享规则只保留一个主定义位置，其他文档只做引用或摘要。
- 当需求与实现暂时不一致时，先在对应文档显式写出边界，不要隐含处理。
- 修改共享契约后，至少回看一次根 `AGENTS.md` 和 `technical/overview.md`，确认导航和摘要没有过期。
