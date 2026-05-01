# H5 首页筛选抽屉收敛实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 H5 首页主筛选区收敛为仅保留“车找人 / 人找车”，并把其它筛选统一迁入抽屉。

**架构：** 保留 `tripType` 作为主区即时筛选；将 `fromText`、`toText` 与现有抽屉条件并入同一份抽屉草稿态，经“确定”统一提交。样式层面用一行类型 chip 加右侧 `LayoutGrid` 图标按钮替换当前主区多段结构。

**技术栈：** React、TypeScript、Vitest、Testing Library、lucide-react

---

### 任务 1：先改测试锁定新行为

**文件：**
- 修改：`mn-frontend-h5/src/test/home-page.test.tsx`

- [ ] 步骤 1：把首页快速筛选测试改为新交互
- [ ] 步骤 2：运行 `npm run test -- src/test/home-page.test.tsx`，确认旧实现下失败

### 任务 2：实现首页结构与抽屉状态调整

**文件：**
- 修改：`mn-frontend-h5/src/pages/HomePage.tsx`
- 修改：`mn-frontend-h5/src/styles/h5.css`

- [ ] 步骤 1：移除主区中的“全部”、起点、终点、查询按钮、整行文字触发器
- [ ] 步骤 2：加入 `LayoutGrid` 图标按钮，并将起点终点输入迁入抽屉
- [ ] 步骤 3：调整抽屉草稿态和提交逻辑，保证非类型筛选统一从抽屉确认生效

### 任务 3：验证

**文件：**
- 验证：`mn-frontend-h5/src/test/home-page.test.tsx`

- [ ] 步骤 1：运行 `npm run test -- src/test/home-page.test.tsx`
- [ ] 步骤 2：运行 `npm run build`
