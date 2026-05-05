# H5 重置密码页面实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 `htmls/v1.html` 的重置密码部分同步到 `mn-frontend-h5`，新增独立“重置密码”页面、账户设置入口、前端校验和“暂未接入”提示。

**架构：** 保持 `AccountSettingsPage` 作为入口聚合页，在 `mn-frontend-h5/src/features/profile/pages` 下新增单一职责的 `PasswordResetPage`。路由层新增受保护页面，样式继续集中在 `src/styles/h5.css`，测试沿用 `src/test/profile-page.test.tsx` 的账户域验证方式。

**技术栈：** React 18、React Router 6、TypeScript、Ant Design Mobile Toast、Vitest、Testing Library

---

## 文件结构

**创建：**

- `mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx`
  - 独立承载重置密码页面 UI、受控表单、前端校验、Toast 提示

**修改：**

- `mn-frontend-h5/src/features/profile/pages/AccountSettingsPage.tsx`
  - 将“安全管理”区块的占位项替换为可跳转的“重置密码”入口
- `mn-frontend-h5/src/router/index.tsx`
  - 注册新的受保护路由 `/me/settings/password-reset`
- `mn-frontend-h5/src/components/MobileTabBar.tsx`
  - 为新路由配置顶部栏标题与返回兜底
- `mn-frontend-h5/src/styles/h5.css`
  - 补充重置密码页需要的 `password-reset-*` 样式
- `mn-frontend-h5/src/test/profile-page.test.tsx`
  - 为入口展示、页面渲染、表单校验、Toast 提示补测试

## 任务 1：先锁定路由和入口的失败测试

**文件：**

- 修改：`mn-frontend-h5/src/test/profile-page.test.tsx`
- 目标实现文件：`mn-frontend-h5/src/features/profile/pages/AccountSettingsPage.tsx`
- 目标实现文件：`mn-frontend-h5/src/router/index.tsx`

- [ ] **步骤 1：为账户设置页新增“重置密码”入口断言**

```tsx
it("renders reset password entry in account settings", async () => {
  mockGetCurrentUserProfile.mockResolvedValue({
    id: 1,
    email: "user@example.com",
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    defaultWechat: "wechat_01",
    defaultPhone: "13800138000",
  });

  render(
    <MemoryRouter>
      <AccountSettingsPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("重置密码")).toBeInTheDocument();
  expect(screen.getByText("定期更新登录密码，保护账户登录安全")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: /重置密码.*去修改/i })).toHaveAttribute("href", "/me/settings/password-reset");
});
```

- [ ] **步骤 2：为新路由页新增渲染测试**

```tsx
import { RouterProvider, createMemoryRouter } from "react-router-dom";
import { routes } from "../router";

it("renders password reset page via protected route", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  expect(await screen.findByRole("heading", { name: "重置密码" })).toBeInTheDocument();
  expect(screen.getByText("Password Reset")).toBeInTheDocument();
  expect(screen.getByLabelText("新密码")).toBeInTheDocument();
  expect(screen.getByLabelText("确认密码")).toBeInTheDocument();
});
```

- [ ] **步骤 3：运行定向测试，确认当前失败**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- 至少有 1 个新增断言失败
- 常见失败信号包括 `Unable to find an element with the text: 重置密码`
- 或路由 `/me/settings/password-reset` 未注册

- [ ] **步骤 4：最小实现入口与路由让测试转绿**

在 `AccountSettingsPage.tsx` 中，将当前安全项替换为链接结构：

```tsx
import { Link } from "react-router-dom";

<div className="profile-setting-list">
  <Link className="profile-setting-entry" to="/me/settings/password-reset">
    <div>
      <p className="profile-setting-item__label">重置密码</p>
      <p className="profile-setting-item__value">定期更新登录密码，保护账户登录安全</p>
    </div>
    <span className="profile-setting-entry__action">去修改</span>
  </Link>
</div>
```

在 `router/index.tsx` 中注册受保护路由：

```tsx
{
  path: "/me/settings/password-reset",
  element: (
    <RequireAuth>
      <PasswordResetPage />
    </RequireAuth>
  ),
}
```

先用最小占位页面通过路由测试：

```tsx
export default function PasswordResetPage() {
  return (
    <main className="h5-shell">
      <section className="page-panel">
        <h1>重置密码</h1>
        <p>Password Reset</p>
        <label className="field-block">
          <span>新密码</span>
          <input type="password" />
        </label>
        <label className="field-block">
          <span>确认密码</span>
          <input type="password" />
        </label>
      </section>
    </main>
  );
}
```

- [ ] **步骤 5：重新运行定向测试验证通过**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- 新增的入口与路由测试通过
- 后续涉及 UI 完整结构的断言仍可在后续任务继续扩展

- [ ] **步骤 6：Commit**

```bash
git add \
  mn-frontend-h5/src/features/profile/pages/AccountSettingsPage.tsx \
  mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx \
  mn-frontend-h5/src/router/index.tsx \
  mn-frontend-h5/src/test/profile-page.test.tsx
git commit -m "feat: add h5 password reset route"
```

## 任务 2：补齐重置密码页的完整静态结构

**文件：**

- 修改：`mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx`
- 修改：`mn-frontend-h5/src/components/MobileTabBar.tsx`
- 修改：`mn-frontend-h5/src/test/profile-page.test.tsx`

- [ ] **步骤 1：先写页面静态结构的失败测试**

```tsx
it("renders password reset page content from html draft", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  expect(await screen.findByText("请设置新的登录密码，让账户安全保持在你手里。")).toBeInTheDocument();
  expect(screen.getByText("这是一个重置密码场景，不需要输入当前密码。完成后，下次登录请使用新密码。")).toBeInTheDocument();
  expect(screen.getByText("安全提示")).toBeInTheDocument();
  expect(screen.getByText("设置建议")).toBeInTheDocument();
  expect(screen.getByText("不要重复使用旧密码")).toBeInTheDocument();
  expect(screen.getByText("避免与其他平台共用密码")).toBeInTheDocument();
});
```

- [ ] **步骤 2：运行测试验证静态内容断言失败**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- 新增关于 Hero、副标题或建议卡片的断言失败

- [ ] **步骤 3：实现完整页面结构并补顶部栏标题映射**

在 `PasswordResetPage.tsx` 中将最小占位替换为完整结构：

```tsx
export default function PasswordResetPage() {
  return (
    <main className="h5-shell h5-shell--profile">
      <section className="page-panel password-reset-panel">
        <section className="card password-reset-hero">
          <p className="password-reset-hero__eyebrow">Password Reset</p>
          <h1 className="password-reset-hero__title">
            请设置新的登录密码，
            <br />
            让账户安全保持在你手里。
          </h1>
          <p className="password-reset-hero__subtitle">
            这是一个重置密码场景，不需要输入当前密码。完成后，下次登录请使用新密码。
          </p>
        </section>

        <section className="card password-reset-form-card">
          <h2 className="form-title">重置密码</h2>
          <p className="form-subtitle">输入新的登录密码，并再次确认，避免因输入错误影响后续登录。</p>
          {/* 任务 3 再补状态与提交 */}
        </section>

        <section className="card password-reset-tips">
          <div className="section-header">
            <div>
              <h2 className="section-title">设置建议</h2>
              <p className="section-subtitle">保持轻量，但把安全提醒说清楚</p>
            </div>
          </div>
          <div className="password-reset-tip-list">
            <article className="password-reset-tip">
              <span className="password-reset-tip__icon">01</span>
              <div>
                <p className="password-reset-tip__title">不要重复使用旧密码</p>
                <p className="password-reset-tip__text">如果你在多个平台复用同一密码，单点泄露会放大整体账户风险。</p>
              </div>
            </article>
            <article className="password-reset-tip">
              <span className="password-reset-tip__icon">02</span>
              <div>
                <p className="password-reset-tip__title">避免与其他平台共用密码</p>
                <p className="password-reset-tip__text">为高频使用的平台设置独立密码，可以减少一个站点泄露后带来的连锁影响。</p>
              </div>
            </article>
          </div>
        </section>
      </section>
    </main>
  );
}
```

在 `MobileTabBar.tsx` 的 `getRouteFrame()` 中加入：

```tsx
if (pathname === "/me/settings/password-reset") {
  return {
    title: "重置密码",
    showTopBar: true,
    showTabBar: false,
    backFallback: "/me/settings",
  };
}
```

- [ ] **步骤 4：运行测试确认静态结构通过**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- 页面文案与建议卡片相关断言通过
- 路由页顶部栏标题从“返回”变为“重置密码”

- [ ] **步骤 5：Commit**

```bash
git add \
  mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx \
  mn-frontend-h5/src/components/MobileTabBar.tsx \
  mn-frontend-h5/src/test/profile-page.test.tsx
git commit -m "feat: sync h5 password reset page content"
```

## 任务 3：补前端校验与“暂未接入”提示

**文件：**

- 修改：`mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx`
- 修改：`mn-frontend-h5/src/test/profile-page.test.tsx`

- [ ] **步骤 1：先写校验与提示的失败测试**

```tsx
import userEvent from "@testing-library/user-event";

it("validates password reset form and shows pending toast", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await screen.findByRole("heading", { name: "重置密码" });

  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("请输入新的登录密码");

  await userEvent.type(screen.getByLabelText("新密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("请再次输入新的登录密码");

  await userEvent.type(screen.getByLabelText("确认密码"), "different123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("两次输入的密码不一致");
});
```

增加 Toast 断言时，在测试文件顶部 mock `antd-mobile`：

```tsx
const { mockToastShow } = vi.hoisted(() => ({
  mockToastShow: vi.fn(),
}));

vi.mock("antd-mobile", () => ({
  Toast: {
    show: mockToastShow,
  },
}));
```

并补成功路径测试：

```tsx
it("shows pending toast after valid password reset submit", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await userEvent.type(screen.getByLabelText("新密码"), "secret123");
  await userEvent.type(screen.getByLabelText("确认密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));

  expect(mockToastShow).toHaveBeenCalledWith({
    content: "暂未接入",
  });
});
```

- [ ] **步骤 2：运行测试，确认校验与 Toast 相关断言失败**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- `role="alert"` 未出现或文案不匹配
- `mockToastShow` 未被调用

- [ ] **步骤 3：实现受控表单、校验和 Toast**

在 `PasswordResetPage.tsx` 中补齐状态与提交逻辑：

```tsx
import type { FormEvent } from "react";
import { useState } from "react";
import { Toast } from "antd-mobile";

export default function PasswordResetPage() {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!password) {
      setError("请输入新的登录密码");
      return;
    }

    if (!confirmPassword) {
      setError("请再次输入新的登录密码");
      return;
    }

    if (password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }

    setError("");
    Toast.show({
      content: "暂未接入",
    });
  }
```

表单字段与错误区按以下方式接入：

```tsx
<form className="password-reset-form" onSubmit={handleSubmit}>
  <label className="field-block">
    <span>新密码</span>
    <input onChange={(event) => setPassword(event.target.value)} type="password" value={password} />
  </label>
  <label className="field-block">
    <span>确认密码</span>
    <input onChange={(event) => setConfirmPassword(event.target.value)} type="password" value={confirmPassword} />
  </label>
  <button className="primary-button" type="submit">确认</button>
  {error ? <p role="alert">{error}</p> : null}
</form>
```

- [ ] **步骤 4：运行定向测试确认通过**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- 校验错误提示按顺序通过
- `mockToastShow` 收到 `{ content: "暂未接入" }`
- 账户设置页原有测试保持通过

- [ ] **步骤 5：Commit**

```bash
git add \
  mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx \
  mn-frontend-h5/src/test/profile-page.test.tsx
git commit -m "feat: add h5 password reset validation"
```

## 任务 4：补齐页面样式并做最终验证

**文件：**

- 修改：`mn-frontend-h5/src/styles/h5.css`
- 修改：`mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx`

- [ ] **步骤 1：先补一个结构存在性的测试，防止样式类挂空**

```tsx
it("applies dedicated password reset layout classes", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await screen.findByRole("heading", { name: "重置密码" });

  expect(document.querySelector(".password-reset-panel")).not.toBeNull();
  expect(document.querySelector(".password-reset-hero")).not.toBeNull();
  expect(document.querySelector(".password-reset-form-card")).not.toBeNull();
  expect(document.querySelector(".password-reset-tip-list")).not.toBeNull();
});
```

- [ ] **步骤 2：运行测试确认类名断言失败或未完全覆盖**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
```

预期：

- 若任务 2 的类名尚未完整接入，则这里会失败
- 若测试已通过，可直接进入样式实现

- [ ] **步骤 3：在 `h5.css` 中补重置密码页样式**

新增一组与现有设计语言一致的样式：

```css
.password-reset-panel {
  display: grid;
  gap: 16px;
}

.password-reset-hero {
  display: grid;
  gap: 10px;
  padding: 20px;
  border-radius: 24px;
  background: linear-gradient(180deg, rgba(255, 245, 236, 0.96), rgba(250, 245, 239, 0.82));
  border: 1px solid rgba(227, 198, 176, 0.5);
}

.password-reset-hero__eyebrow {
  margin: 0;
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #a56f4c;
}

.password-reset-hero__title {
  margin: 0;
  font-size: 26px;
  line-height: 1.35;
}

.password-reset-form-card,
.password-reset-tips {
  display: grid;
  gap: 12px;
}

.password-reset-tip-list {
  display: grid;
  gap: 12px;
}

.password-reset-tip {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  gap: 12px;
  align-items: flex-start;
}
```

如果 `PasswordResetPage.tsx` 里按钮还未使用现有主按钮类，此时统一为：

```tsx
<button className="primary-button" type="submit">
  确认
</button>
```

- [ ] **步骤 4：运行完整验证**

运行：

```bash
cd mn-frontend-h5
npm run test -- src/test/profile-page.test.tsx
npm run build
```

预期：

- `src/test/profile-page.test.tsx` 全绿
- `vite build` 成功，无 TypeScript 错误

- [ ] **步骤 5：Commit**

```bash
git add \
  mn-frontend-h5/src/styles/h5.css \
  mn-frontend-h5/src/features/profile/pages/PasswordResetPage.tsx \
  mn-frontend-h5/src/test/profile-page.test.tsx
git commit -m "style: polish h5 password reset page"
```

## 自检映射

- 规格中的“账户设置页新增重置密码入口”对应任务 1
- 规格中的“同步 HTML 四层结构与文案”对应任务 2
- 规格中的“前端校验与暂未接入提示”对应任务 3
- 规格中的“样式同步与验证命令”对应任务 4

