import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, RouterProvider, createMemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const {
  mockToastShow,
  mockGetCurrentUserProfile,
  mockGetMyTrips,
  mockGetMyFavorites,
  mockResetPassword,
  mockSendVerificationCode,
  mockUpdateUserContact,
  mockUpdateUserProfile,
} = vi.hoisted(() => ({
  mockToastShow: vi.fn(),
  mockGetCurrentUserProfile: vi.fn(),
  mockGetMyTrips: vi.fn(),
  mockGetMyFavorites: vi.fn(),
  mockResetPassword: vi.fn(),
  mockSendVerificationCode: vi.fn(),
  mockUpdateUserContact: vi.fn(),
  mockUpdateUserProfile: vi.fn(),
}));

vi.mock("antd-mobile", async (importOriginal) => {
  const actual = await importOriginal<typeof import("antd-mobile")>();

  return {
    ...actual,
    Toast: {
      ...actual.Toast,
      show: mockToastShow,
    },
  };
});

vi.mock("../features/profile/api", () => ({
  getCurrentUserProfile: mockGetCurrentUserProfile,
  resetPassword: mockResetPassword,
  sendVerificationCode: mockSendVerificationCode,
  updateUserContact: mockUpdateUserContact,
  updateUserProfile: mockUpdateUserProfile,
}));

vi.mock("../features/trips/api", () => ({
  getMyTrips: mockGetMyTrips,
  getMyFavorites: mockGetMyFavorites,
}));

vi.mock("../features/profile/components/AvatarUploader", () => ({
  default: function MockAvatarUploader() {
    return <div>AvatarUploader</div>;
  },
}));

import ProfilePage from "../features/profile/pages/ProfilePage";
import AccountSettingsPage from "../features/profile/pages/AccountSettingsPage";
import { routes } from "../router";
import { useAuthStore } from "../store/auth";

beforeEach(() => {
  window.localStorage.clear();
  window.scrollTo = vi.fn();
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: {
      id: 1,
      email: "user@example.com",
      phone: "13800138000",
      nickname: "初始用户",
      avatarUrl: "",
      status: "active",
      defaultWechat: "",
      defaultPhone: "13800138000",
    },
  });

  mockGetCurrentUserProfile.mockReset();
  mockGetMyTrips.mockReset();
  mockGetMyFavorites.mockReset();
  mockResetPassword.mockReset();
  mockSendVerificationCode.mockReset();
  mockUpdateUserContact.mockReset();
  mockUpdateUserProfile.mockReset();
  mockToastShow.mockReset();

  mockSendVerificationCode.mockResolvedValue({ sent: true });
  mockResetPassword.mockResolvedValue({ ok: true });
});

it("loads profile data once after login hydration", async () => {
  mockGetCurrentUserProfile.mockImplementation(async () => ({
    id: 1,
    email: "user@example.com",
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    defaultWechat: "",
    defaultPhone: "13800138000",
  }));
  mockGetMyTrips.mockResolvedValue({
    items: [],
    total: 2,
    pageNum: 1,
    pageSize: 10,
  });
  mockGetMyFavorites.mockResolvedValue({
    items: [],
    total: 3,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <ProfilePage />
    </MemoryRouter>,
  );

  expect(await screen.findByRole("heading", { name: "测试用户" })).toBeInTheDocument();

  await waitFor(() => {
    expect(mockGetCurrentUserProfile).toHaveBeenCalledTimes(1);
    expect(mockGetMyTrips).toHaveBeenCalledTimes(1);
    expect(mockGetMyFavorites).toHaveBeenCalledTimes(1);
  });
});

it("renders the profile overview with a dedicated account settings entry", async () => {
  mockGetCurrentUserProfile.mockImplementation(async () => ({
    id: 1,
    email: "user@example.com",
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    defaultWechat: "wechat_01",
    defaultPhone: "13800138000",
  }));
  mockGetMyTrips.mockResolvedValue({
    items: [],
    total: 12,
    pageNum: 1,
    pageSize: 10,
  });
  mockGetMyFavorites.mockResolvedValue({
    items: [],
    total: 5,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <ProfilePage />
    </MemoryRouter>,
  );

  expect(await screen.findByRole("link", { name: /账户设置/i })).toBeInTheDocument();
  expect(screen.queryByText("已实名")).not.toBeInTheDocument();
  expect(screen.queryByText("信用评分")).not.toBeInTheDocument();
  expect(screen.queryByText("证件信息")).not.toBeInTheDocument();
  expect(screen.queryByText("钱包余额")).not.toBeInTheDocument();
  expect(screen.queryByText("常用服务")).not.toBeInTheDocument();
  expect(screen.queryByText("常用乘客")).not.toBeInTheDocument();
  expect(screen.queryByText("客服帮助")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "保存修改" })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "退出登录" })).not.toBeInTheDocument();
  expect(screen.queryByText("账号安全")).not.toBeInTheDocument();
});

it("renders fallback asset instead of plain text when profile avatar fails to load", async () => {
  mockGetCurrentUserProfile.mockResolvedValue({
    id: 1,
    email: "user@example.com",
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "https://cdn.example.com/broken-avatar.png",
    status: "active",
    defaultWechat: "wechat_01",
    defaultPhone: "13800138000",
  });
  mockGetMyTrips.mockResolvedValue({
    items: [],
    total: 12,
    pageNum: 1,
    pageSize: 10,
  });
  mockGetMyFavorites.mockResolvedValue({
    items: [],
    total: 5,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <ProfilePage />
    </MemoryRouter>,
  );

  fireEvent.error(await screen.findByAltText("当前头像"));

  expect(await screen.findByAltText("头像加载失败")).toHaveAttribute("src", "/image-fail.svg");
  expect(screen.queryByText("/image-fail.svg")).not.toBeInTheDocument();
});

it("uses default avatar asset in account settings when profile avatar is empty", async () => {
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
  mockUpdateUserProfile.mockResolvedValue({ ok: true });
  mockUpdateUserContact.mockResolvedValue({ ok: true });

  render(
    <MemoryRouter>
      <AccountSettingsPage />
    </MemoryRouter>,
  );

  expect(await screen.findByAltText("当前头像")).toHaveAttribute("src", "/image-default.svg");
});

it("loads account settings profile when access token exists but user is null", async () => {
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: null,
  });
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

  expect(await screen.findByText("测试用户")).toBeInTheDocument();

  await waitFor(() => {
    expect(mockGetCurrentUserProfile).toHaveBeenCalledTimes(1);
  });
  expect(screen.getByText("user@example.com")).toBeInTheDocument();
});

it("renders the dedicated account settings page and keeps edit actions working", async () => {
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
  mockUpdateUserProfile.mockResolvedValue({ ok: true });
  mockUpdateUserContact.mockResolvedValue({ ok: true });

  render(
    <MemoryRouter>
      <AccountSettingsPage />
    </MemoryRouter>,
  );

  expect(await screen.findByRole("heading", { name: "账户设置" })).toBeInTheDocument();
  expect(screen.getByText("AvatarUploader")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "保存修改" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "退出登录" })).toBeInTheDocument();

  fireEvent.click(screen.getByRole("button", { name: "修改" }));
  fireEvent.change(screen.getByLabelText("修改昵称"), { target: { value: "新昵称" } });
  fireEvent.click(screen.getAllByRole("button", { name: "去管理" })[0]);
  fireEvent.change(screen.getByLabelText("默认手机号"), { target: { value: "13900001111" } });
  fireEvent.change(screen.getByLabelText("默认微信号"), { target: { value: "wechat_02" } });
  fireEvent.click(screen.getByRole("button", { name: "保存修改" }));

  await waitFor(() => {
    expect(mockUpdateUserProfile).toHaveBeenCalledWith("新昵称");
    expect(mockUpdateUserContact).toHaveBeenCalledWith("wechat_02", "13900001111");
  });
});

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

it("renders password reset page via protected route", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  expect(await screen.findByText("重置密码", { selector: ".page-topbar__title" })).toBeInTheDocument();
  expect(screen.getByText("Password Reset")).toBeInTheDocument();
  expect(screen.getByLabelText("验证码")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "发送验证码" })).toBeInTheDocument();
  expect(screen.getByLabelText("新密码")).toBeInTheDocument();
  expect(screen.getByLabelText("确认密码")).toBeInTheDocument();
});

it("renders password reset page content from html draft", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  expect(await screen.findByText("请设置新的登录密码，让账户安全保持在你手里。")).toBeInTheDocument();
  expect(
    screen.getByText("这是一个重置密码场景，不需要输入当前密码。完成后，下次登录请使用新密码。"),
  ).toBeInTheDocument();
  expect(screen.getByText("重置密码", { selector: ".form-title" })).toBeInTheDocument();
  expect(screen.getByText("输入新的登录密码，并再次确认，避免因输入错误影响后续登录。")).toBeInTheDocument();
  expect(screen.getByPlaceholderText("请输入验证码")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "确认" })).toBeInTheDocument();
});

it("applies dedicated password reset layout classes", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await screen.findByText("重置密码", { selector: ".page-topbar__title" });

  expect(document.querySelector(".password-reset-panel")).not.toBeNull();
  expect(document.querySelector(".password-reset-hero")).not.toBeNull();
  expect(document.querySelector(".password-reset-form-card")).not.toBeNull();
  expect(document.querySelector(".auth-code")).not.toBeNull();
});

it("validates password reset form", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await screen.findByText("重置密码", { selector: ".page-topbar__title" });

  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("请输入新的登录密码");

  await userEvent.type(screen.getByLabelText("新密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("请再次输入新的登录密码");

  await userEvent.type(screen.getByLabelText("确认密码"), "different123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("两次输入的密码不一致");

  await userEvent.clear(screen.getByLabelText("确认密码"));
  await userEvent.type(screen.getByLabelText("确认密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));
  expect(screen.getByRole("alert")).toHaveTextContent("请输入验证码");
});

it("sends reset verification code with dedicated type and starts countdown", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await userEvent.click(screen.getByRole("button", { name: "发送验证码" }));

  await waitFor(() => {
    expect(mockSendVerificationCode).toHaveBeenCalledWith(
      "user@example.com",
      "reset_password",
    );
  });

  expect(screen.getByRole("button", { name: "60s后重试" })).toBeDisabled();
  expect(screen.getByText("没收到验证码？60s 后可重新发送")).toBeInTheDocument();
});

it("resets password then logs user out and redirects to login", async () => {
  const router = createMemoryRouter(routes, {
    initialEntries: ["/me/settings/password-reset"],
  });

  render(<RouterProvider router={router} />);

  await userEvent.type(screen.getByLabelText("验证码"), "123456");
  await userEvent.type(screen.getByLabelText("新密码"), "secret123");
  await userEvent.type(screen.getByLabelText("确认密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "确认" }));

  await waitFor(() => {
    expect(mockResetPassword).toHaveBeenCalledWith(
      "user@example.com",
      "123456",
      "secret123",
    );
  });

  expect(mockToastShow).toHaveBeenCalledWith({
    content: "密码已重置，请重新登录",
  });
  expect(useAuthStore.getState().accessToken).toBeNull();
  expect(useAuthStore.getState().refreshToken).toBeNull();
  expect(useAuthStore.getState().user).toBeNull();

  await waitFor(() => {
    expect(router.state.location.pathname).toBe("/login");
  });
});
