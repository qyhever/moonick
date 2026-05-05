import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockGetTripDetail, mockToggleFavorite, mockToastShow } = vi.hoisted(() => ({
  mockGetTripDetail: vi.fn(),
  mockToggleFavorite: vi.fn(),
  mockToastShow: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getTripDetail: mockGetTripDetail,
  toggleFavorite: mockToggleFavorite,
}));

vi.mock("antd-mobile", () => ({
  Toast: {
    show: mockToastShow,
  },
}));

import TripDetailPage from "../features/trips/pages/TripDetailPage";
import { useAuthStore } from "../store/auth";

beforeEach(() => {
  window.localStorage.clear();
  useAuthStore.setState({
    accessToken: null,
    refreshToken: null,
    user: null,
  });
  mockGetTripDetail.mockReset();
  mockToggleFavorite.mockReset();
  mockToastShow.mockReset();
  mockGetTripDetail.mockResolvedValue({
    id: 7,
    userId: 12,
    tripType: "driver_post",
    fromText: "上海虹桥",
    toText: "杭州东",
    departureDate: "2026-04-26",
    departureTime: "09:30",
    seatCount: 2,
    isPriceNegotiable: true,
    status: "active",
    favorited: false,
    unavailable: false,
    contactPhone: "13800138000",
    contactWechat: "mingye-trip",
    createdAt: "2026-04-25T12:00:00Z",
    updatedAt: "2026-04-25T12:00:00Z",
  });
  mockToggleFavorite.mockResolvedValue({
    favorited: true,
  });
});

it("toggles favorite state and shows success toast", async () => {
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: {
      id: 99,
      email: "user@example.com",
      phone: "13800138000",
      nickname: "测试用户",
      avatarUrl: "",
      status: "active",
      defaultWechat: "",
      defaultPhone: "13800138000",
    },
  });

  const router = createMemoryRouter(
    [
      { path: "/trips/:id", element: <TripDetailPage /> },
    ],
    { initialEntries: ["/trips/7"] },
  );

  render(
    <RouterProvider
      router={router}
      future={{
        v7_startTransition: true,
      }}
    />
  );

  const button = await screen.findByRole("button", { name: "收藏" });
  await userEvent.click(button);

  await waitFor(() => {
    expect(button).toHaveAttribute("data-favorited", "true");
  });
  expect(await screen.findByRole("button", { name: "已收藏" })).toBeInTheDocument();
  expect(mockToastShow).toHaveBeenCalledWith(
    expect.objectContaining({
      content: "收藏成功",
    }),
  );
});

it("shows cancel favorite success toast when toggle result is unfavorited", async () => {
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: {
      id: 99,
      email: "user@example.com",
      phone: "13800138000",
      nickname: "测试用户",
      avatarUrl: "",
      status: "active",
      defaultWechat: "",
      defaultPhone: "13800138000",
    },
  });
  mockGetTripDetail.mockResolvedValue({
    id: 7,
    userId: 12,
    tripType: "driver_post",
    fromText: "上海虹桥",
    toText: "杭州东",
    departureDate: "2026-04-26",
    departureTime: "09:30",
    seatCount: 2,
    isPriceNegotiable: true,
    status: "active",
    favorited: true,
    unavailable: false,
    contactPhone: "13800138000",
    contactWechat: "mingye-trip",
    createdAt: "2026-04-25T12:00:00Z",
    updatedAt: "2026-04-25T12:00:00Z",
  });
  mockToggleFavorite.mockResolvedValue({
    favorited: false,
  });

  const router = createMemoryRouter(
    [
      { path: "/trips/:id", element: <TripDetailPage /> },
    ],
    { initialEntries: ["/trips/7"] },
  );

  render(
    <RouterProvider
      router={router}
      future={{
        v7_startTransition: true,
      }}
    />
  );

  const button = await screen.findByRole("button", { name: "已收藏" });
  await userEvent.click(button);

  await waitFor(() => {
    expect(button).toHaveAttribute("data-favorited", "false");
  });
  expect(await screen.findByRole("button", { name: "收藏" })).toBeInTheDocument();
  expect(mockToastShow).toHaveBeenCalledWith(
    expect.objectContaining({
      content: "取消收藏成功",
    }),
  );
});

it("redirects guest to login instead of calling favorite api", async () => {
  const router = createMemoryRouter(
    [
      { path: "/trips/:id", element: <TripDetailPage /> },
      { path: "/login", element: <h1>登录</h1> },
    ],
    { initialEntries: ["/trips/7"] },
  );

  render(
    <RouterProvider
      router={router}
      future={{
        v7_startTransition: true,
      }}
    />
  );

  const button = await screen.findByRole("button", { name: "收藏" });
  await userEvent.click(button);

  expect(await screen.findByText("登录")).toBeInTheDocument();
  expect(router.state.location.pathname).toBe("/login");
  expect(router.state.location.search).toBe("?redirect=%2Ftrips%2F7");
  expect(mockToggleFavorite).not.toHaveBeenCalled();
});

it("shows publish success toast from route state", async () => {
  const router = createMemoryRouter(
    [{ path: "/trips/:id", element: <TripDetailPage /> }],
    {
      initialEntries: [
        {
          pathname: "/trips/7",
          state: { toast: "发布成功" },
        },
      ],
    },
  );

  render(
    <RouterProvider
      router={router}
      future={{
        v7_startTransition: true,
      }}
    />
  );

  await waitFor(() => {
    expect(mockToastShow).toHaveBeenCalledWith(
      expect.objectContaining({
        content: "发布成功",
      }),
    );
  });
});
