# Admin 固定壳布局实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 `mn-frontend-admin` 的侧边菜单和顶部 Header 改为固定布局，只让右侧内容区独立滚动。

**架构：** 保持现有 `AdminLayout` 路由壳结构不变，在同一文件中提取侧边栏宽度和 Header 高度常量，并通过 `position: fixed` 固定 `Sider` 与 `Header`。右侧 `Layout` 用左边距避让侧边栏，`Content` 用顶部边距、固定高度和 `overflowY: auto` 形成独立滚动容器。

**技术栈：** React 18、React Router 6、Ant Design 5、Vitest、Testing Library

---

### 任务 1：为固定布局补充失败测试

**文件：**
- 修改：`mn-frontend-admin/src/test/login-guard.test.tsx`
- 测试：`mn-frontend-admin/src/test/login-guard.test.tsx`

- [ ] **步骤 1：编写失败的测试**

```tsx
it("renders fixed admin shell layout styles for authenticated routes", async () => {
  window.history.pushState({}, "", "/dashboard");

  useAdminAuthStore.setState({
    accessToken: "token",
    refreshToken: "refresh",
    admin: { id: "admin-1", username: "admin", name: "管理员" },
  });

  renderWithRouter();

  const sider = await screen.findByRole("complementary");
  const header = screen.getByRole("banner");
  const content = screen.getByTestId("admin-layout-content");

  expect(sider).toHaveStyle({
    position: "fixed",
    top: "0px",
    left: "0px",
    bottom: "0px",
  });
  expect(header).toHaveStyle({
    position: "fixed",
    top: "0px",
    left: "232px",
    right: "0px",
  });
  expect(content).toHaveStyle({
    marginTop: "64px",
    height: "calc(100vh - 64px)",
    overflowY: "auto",
  });
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：`cd mn-frontend-admin && npm run test -- login-guard.test.tsx`
预期：FAIL，缺少 `admin-layout-content` 或样式断言不匹配。

- [ ] **步骤 3：编写最少实现代码**

```tsx
const siderWidth = 232;
const headerHeight = 64;

<Sider
  style={{
    position: "fixed",
    top: 0,
    left: 0,
    bottom: 0,
    height: "100vh",
    overflow: "auto",
  }}
  width={siderWidth}
>

<Header
  style={{
    position: "fixed",
    top: 0,
    left: siderWidth,
    right: 0,
    height: headerHeight,
  }}
>

<Content
  data-testid="admin-layout-content"
  style={{
    marginTop: headerHeight,
    height: `calc(100vh - ${headerHeight}px)`,
    overflowY: "auto",
  }}
>
```

- [ ] **步骤 4：运行测试验证通过**

运行：`cd mn-frontend-admin && npm run test -- login-guard.test.tsx`
预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add mn-frontend-admin/src/layout/AdminLayout.tsx mn-frontend-admin/src/test/login-guard.test.tsx
git commit -m "fix(admin): keep shell layout fixed while content scrolls"
```

### 任务 2：运行完整验证

**文件：**
- 无代码修改
- 测试：`mn-frontend-admin/src/test/*.test.ts*`

- [ ] **步骤 1：运行管理端完整测试**

运行：`cd mn-frontend-admin && npm run test`
预期：PASS，全部测试通过。

- [ ] **步骤 2：运行构建验证**

运行：`cd mn-frontend-admin && npm run build`
预期：PASS，TypeScript 与 Vite 构建成功。

- [ ] **步骤 3：记录结果并准备交付**

```text
记录新增布局测试覆盖点、固定布局实现方式，以及测试/构建命令的实际结果。
```
