import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RouterProvider, createMemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("../lib/http", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../lib/http")>();

  return {
    ...actual,
    api: {
      post: mockPost,
      request: vi.fn(),
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    },
    attachAuthStore: vi.fn(),
  };
});

import { routes } from "../router";
import { useAuthStore } from "../store/auth";

function renderWithRouter(initialEntry: string) {
  const router = createMemoryRouter(routes, {
    initialEntries: [initialEntry],
  });

  return {
    router,
    ...render(
      <RouterProvider
        router={router}
        future={{
          v7_startTransition: true,
        }}
      />
    ),
  };
}

beforeEach(() => {
  window.localStorage.clear();
  window.scrollTo = vi.fn();
  useAuthStore.setState({
    accessToken: null,
    refreshToken: null,
    user: null,
  });
  mockPost.mockReset();
  mockPost.mockImplementation(async (url: string, payload?: { email: string; password?: string; code?: string }) => {
    if (url === "/api/v1/auth/refresh") {
      return {
        data: {
          code: 1000,
          message: "success",
          data: {
            accessToken: "refreshed-access-token",
            refreshToken: "refreshed-refresh-token",
            user: {
              id: 1,
              email: "user@example.com",
              phone: "13800138000",
              nickname: "测试用户",
              avatarUrl: "",
              status: "active",
              defaultWechat: "",
              defaultPhone: "13800138000",
            },
          },
        },
      };
    }

    if (url === "/api/v1/auth/code" && payload) {
      return {
        data: {
          code: 1000,
          message: "success",
          data: {
            sent: true,
          },
        },
      };
    }

    if (url === "/api/v1/auth/register" && payload) {
      return {
        data: {
          code: 1000,
          message: "success",
          data: {
            accessToken: "access-token",
            refreshToken: "refresh-token",
            user: {
              id: 1,
              email: payload.email,
              phone: "",
              nickname: "测试用户",
              avatarUrl: "",
              status: "active",
              defaultWechat: "",
              defaultPhone: "",
            },
          },
        },
      };
    }

    if (url !== "/api/v1/auth/login" || !payload) {
      throw new Error(`Unexpected URL: ${url}`);
    }

    return {
      data: {
        code: 1000,
        message: "success",
        data: {
            accessToken: "access-token",
            refreshToken: "refresh-token",
            user: {
              id: 1,
              email: payload.email,
              phone: "",
              nickname: "测试用户",
              avatarUrl: "",
              status: "active",
              defaultWechat: "",
              defaultPhone: "",
            },
          },
        },
    };
  });
});

it("redirects guest to login and jumps back after login", async () => {
  renderWithRouter("/publish");

  expect(await screen.findByText("登录")).toBeInTheDocument();

  await userEvent.type(screen.getByLabelText("邮箱"), "user@example.com");
  await userEvent.type(screen.getByLabelText("密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));

  expect(await screen.findByRole("heading", { name: "发一条顺路行程，快速被附近同城的人看到" })).toBeInTheDocument();
});

it("does not show route topbar on login or register pages", async () => {
  const { unmount } = renderWithRouter("/login");

  expect(screen.queryByRole("button", { name: "返回上一页" })).not.toBeInTheDocument();

  unmount();
  renderWithRouter("/register");

  expect(screen.queryByRole("button", { name: "返回上一页" })).not.toBeInTheDocument();
});

it("redirects logged-in user away from login page to home", async () => {
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: {
      id: 1,
      email: "user@example.com",
      phone: "13800138000",
      nickname: "测试用户",
      avatarUrl: "",
      status: "active",
      defaultWechat: "",
      defaultPhone: "13800138000",
    },
  });

  const { router } = renderWithRouter("/login");

  await vi.waitFor(() => {
    expect(router.state.location.pathname).toBe("/");
  });
});

it("redirects logged-in user away from register page to home", async () => {
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: {
      id: 1,
      email: "user@example.com",
      phone: "13800138000",
      nickname: "测试用户",
      avatarUrl: "",
      status: "active",
      defaultWechat: "",
      defaultPhone: "13800138000",
    },
  });

  const { router } = renderWithRouter("/register");

  await vi.waitFor(() => {
    expect(router.state.location.pathname).toBe("/");
  });
});

it("shows backend business error on login failure instead of crashing on empty data", async () => {
  mockPost.mockResolvedValueOnce({
    data: {
      code: 1004,
      message: "用户名或密码错误",
      data: null,
    },
  });

  renderWithRouter("/publish");

  await userEvent.type(screen.getByLabelText("邮箱"), "user@example.com");
  await userEvent.type(screen.getByLabelText("密码"), "wrong-password");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));

  expect(await screen.findByRole("alert")).toHaveTextContent("用户名或密码错误");
  expect(screen.queryByText("发布行程")).not.toBeInTheDocument();
  expect(mockPost).toHaveBeenCalledTimes(1);
});

it("blocks login submit when email format is invalid", async () => {
  renderWithRouter("/login");

  await userEvent.type(screen.getByLabelText("邮箱"), "invalid-email");
  await userEvent.type(screen.getByLabelText("密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));

  expect(await screen.findByRole("alert")).toHaveTextContent("请输入有效的邮箱地址");
  expect(mockPost).not.toHaveBeenCalled();
});

it("does not render verification code field on login page but keeps it on register page", async () => {
  const { unmount } = renderWithRouter("/login");

  expect(screen.queryByLabelText("验证码")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "发送验证码" })).not.toBeInTheDocument();

  unmount();
  renderWithRouter("/register");

  expect(screen.getByLabelText("验证码")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "发送验证码" })).toHaveClass("auth-code__send");
});

it("renders forgot password link on login page and carries current email to reset flow", async () => {
  renderWithRouter("/login");

  const forgotPasswordLink = screen.getByRole("link", { name: "忘记密码？" });
  expect(forgotPasswordLink).toHaveAttribute("href", "/password-reset");

  await userEvent.type(screen.getByLabelText("邮箱"), "user@example.com");

  expect(screen.getByRole("link", { name: "忘记密码？" })).toHaveAttribute(
    "href",
    "/password-reset?email=user%40example.com",
  );
});

it("allows guest to open password reset page directly", async () => {
  const { router } = renderWithRouter("/password-reset?email=user%40example.com");

  expect(await screen.findByText("重置密码", { selector: ".page-topbar__title" })).toBeInTheDocument();
  expect(screen.getByDisplayValue("user@example.com")).toBeInTheDocument();
  expect(router.state.location.pathname).toBe("/password-reset");
});

it("sends register verification code and restores countdown after remount", async () => {
  const { unmount } = renderWithRouter("/register");

  fireEvent.change(screen.getByLabelText("邮箱"), {
    target: { value: "user@example.com" },
  });
  fireEvent.click(screen.getByRole("button", { name: "发送验证码" }));

  expect(mockPost).toHaveBeenCalledWith("/api/v1/auth/code", {
    email: "user@example.com",
    type: "register",
  });
  await waitFor(() => {
    expect(screen.getByRole("button", { name: "60s后重试" })).toBeDisabled();
  });

  unmount();
  renderWithRouter("/register");
  expect(screen.getByRole("button", { name: "60s后重试" })).toBeDisabled();

  window.localStorage.setItem("mn-h5-register-code-expires-at", String(Date.now() - 1000));
  unmount();
  renderWithRouter("/register");

  expect(screen.getByRole("button", { name: "发送验证码" })).toBeEnabled();
});

it("shows resend helper text only during register code countdown", async () => {
  const { unmount } = renderWithRouter("/register");

  fireEvent.change(screen.getByLabelText("邮箱"), {
    target: { value: "user@example.com" },
  });
  fireEvent.click(screen.getByRole("button", { name: "发送验证码" }));

  expect(await screen.findByText("没收到验证码？60s 后可重新发送")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "60s后重试" })).toBeDisabled();

  window.localStorage.setItem("mn-h5-register-code-expires-at", String(Date.now() - 1000));
  unmount();
  renderWithRouter("/register");

  expect(screen.queryByText("没收到验证码？60s 后可重新发送")).not.toBeInTheDocument();
});

it("blocks register submit when email format is invalid", async () => {
  mockPost.mockImplementation(async (url: string) => {
    if (url === "/api/v1/auth/refresh") {
      return {
        data: {
          code: 1000,
          message: "success",
          data: {
            accessToken: "refreshed-access-token",
            refreshToken: "refreshed-refresh-token",
            user: {
              id: 1,
              email: "user@example.com",
              phone: "13800138000",
              nickname: "测试用户",
              avatarUrl: "",
              status: "active",
              defaultWechat: "",
              defaultPhone: "13800138000",
            },
          },
        },
      };
    }

    throw new Error(`Unexpected URL: ${url}`);
  });

  renderWithRouter("/register");

  await userEvent.type(screen.getByLabelText("邮箱"), "invalid-email");
  await userEvent.type(screen.getByLabelText("密码"), "secret123");
  await userEvent.type(screen.getByLabelText("确认密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "注册" }));

  expect(await screen.findByRole("alert")).toHaveTextContent("请输入有效的邮箱地址");
  expect(mockPost).not.toHaveBeenCalled();
});

it("submits register code with email and password to backend", async () => {
  const { router } = renderWithRouter("/register?redirect=%2Fpublish");

  await userEvent.type(screen.getByLabelText("邮箱"), "user@example.com");
  await userEvent.type(screen.getByLabelText("密码"), "secret123");
  await userEvent.type(screen.getByLabelText("确认密码"), "secret123");
  await userEvent.type(screen.getByLabelText("验证码"), "795232");
  await userEvent.click(screen.getByRole("button", { name: "注册" }));

  await waitFor(() => {
    expect(mockPost).toHaveBeenCalledWith("/api/v1/auth/register", {
      email: "user@example.com",
      password: "secret123",
      code: "795232",
    });
  });

  await waitFor(() => {
    expect(router.state.location.pathname).toBe("/publish");
  });
});

it("refreshes token pair through backend refresh endpoint", async () => {
  useAuthStore.setState({
    accessToken: "expired-access-token",
    refreshToken: "refresh-token",
    user: null,
  });

  await expect(useAuthStore.getState().refresh()).resolves.toBeUndefined();
  expect(useAuthStore.getState().accessToken).toBe("refreshed-access-token");
  expect(useAuthStore.getState().refreshToken).toBe("refreshed-refresh-token");
  expect(mockPost).toHaveBeenCalledWith(
    "/api/v1/auth/refresh",
    null,
    expect.objectContaining({
      skipAuthRefresh: true,
      useRefreshToken: true,
    }),
  );
});
