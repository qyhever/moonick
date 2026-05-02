import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
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
  useAuthStore.setState({
    accessToken: null,
    refreshToken: null,
    user: null,
  });
  mockPost.mockReset();
  mockPost.mockImplementation(async (url: string, payload?: { email: string }) => {
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
