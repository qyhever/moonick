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

  return render(<RouterProvider router={router} />);
}

beforeEach(() => {
  window.localStorage.clear();
  useAuthStore.setState({
    accessToken: null,
    refreshToken: null,
    user: null,
  });
  mockPost.mockReset();
  mockPost.mockImplementation(async (url: string, payload: { phone: string }) => {
    if (url !== "/api/v1/auth/login") {
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
            phone: payload.phone,
            nickname: "测试用户",
            avatarUrl: "",
            status: "active",
            defaultWechat: "",
            defaultPhone: payload.phone,
          },
        },
      },
    };
  });
});

it("redirects guest to login and jumps back after login", async () => {
  renderWithRouter("/publish");

  expect(await screen.findByText("登录")).toBeInTheDocument();

  await userEvent.type(screen.getByLabelText("手机号"), "13800138000");
  await userEvent.type(screen.getByLabelText("密码"), "secret123");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));

  expect(await screen.findByText("发布行程")).toBeInTheDocument();
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

  await userEvent.type(screen.getByLabelText("手机号"), "13800138000");
  await userEvent.type(screen.getByLabelText("密码"), "wrong-password");
  await userEvent.click(screen.getByRole("button", { name: "登录" }));

  expect(await screen.findByRole("alert")).toHaveTextContent("用户名或密码错误");
  expect(screen.queryByText("发布行程")).not.toBeInTheDocument();
  expect(mockPost).toHaveBeenCalledTimes(1);
});

it("keeps refresh as a placeholder and does not request a missing endpoint", async () => {
  useAuthStore.setState({
    accessToken: "expired-access-token",
    refreshToken: "refresh-token",
    user: null,
  });

  await expect(useAuthStore.getState().refresh()).rejects.toThrow();
  expect(mockPost).not.toHaveBeenCalled();
});
