import "@testing-library/jest-dom/vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockGetTrips } = vi.hoisted(() => ({
  mockGetTrips: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getTrips: mockGetTrips,
}));

import AppLayout from "../components/MobileTabBar";
import HomePage from "../pages/HomePage";
import { useAuthStore } from "../store/auth";

const defaultTripListResponse = {
  items: [
    {
      id: 1,
      userId: 2,
      tripType: "driver_post" as const,
      fromText: "上海虹桥",
      toText: "杭州东站",
      departureDate: "2026-05-01",
      departureTime: "09:00",
      seatCount: 3,
      isPriceNegotiable: true,
      status: "active" as const,
      favorited: false,
      unavailable: false,
    },
    {
      id: 2,
      userId: 3,
      tripType: "passenger_post" as const,
      fromText: "苏州园区",
      toText: "南京南站",
      departureDate: "2026-05-02",
      departureTime: "14:00",
      seatCount: 1,
      isPriceNegotiable: false,
      status: "full" as const,
      favorited: false,
      unavailable: false,
    },
  ],
  total: 2,
  pageNum: 1,
  pageSize: 10,
};

beforeEach(() => {
  window.localStorage.clear();
  useAuthStore.setState({
    accessToken: "",
    refreshToken: "",
    user: null,
  });
  mockGetTrips.mockReset();
  mockGetTrips.mockResolvedValue(defaultTripListResponse);
});

afterEach(() => {
  vi.useRealTimers();
});

it("loads trips on home page mount", async () => {
  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledWith({
      pageNum: 1,
      pageSize: 10,
      tripType: "driver_post",
    });
  });

  expect(await screen.findByText("上海虹桥")).toBeInTheDocument();
});

it("re-queries trips when switching trip type chips", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledTimes(1);
  });

  await user.click(screen.getByRole("button", { name: "人找车" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenLastCalledWith({
      pageNum: 1,
      pageSize: 10,
      tripType: "passenger_post",
    });
  });
});

it("applies start and destination filters from drawer after confirmation", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledTimes(1);
  });

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));
  await user.type(screen.getByLabelText("起点"), "上海");
  await user.type(screen.getByLabelText("终点"), "杭州");

  mockGetTrips.mockClear();
  await user.click(screen.getByRole("button", { name: "确定" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledWith({
      pageNum: 1,
      pageSize: 10,
      tripType: "driver_post",
      fromText: "上海",
      toText: "杭州",
    });
  });
});

it("applies date preset from drawer after confirmation", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledTimes(1);
  });

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));
  await user.click(screen.getByRole("button", { name: "今天" }));

  expect(mockGetTrips).toHaveBeenCalledTimes(1);

  await user.click(screen.getByRole("button", { name: "确定" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenLastCalledWith({
      pageNum: 1,
      pageSize: 10,
      tripType: "driver_post",
      datePreset: "today",
    });
  });

  expect(screen.queryByText("时间与低频条件收进这里，首页保持轻量")).not.toBeInTheDocument();
});

it("applies real drawer conditions and supports reset and close", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("上海虹桥")).toBeInTheDocument();
  expect(screen.getByText("苏州园区")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));
  await user.click(screen.getByRole("button", { name: "仅看可约" }));
  await user.click(screen.getByRole("button", { name: "可议价" }));
  await user.click(screen.getByRole("button", { name: "确定" }));

  expect(screen.getByText("上海虹桥")).toBeInTheDocument();
  expect(screen.queryByText("苏州园区")).not.toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));
  await user.click(screen.getByRole("button", { name: "重置" }));
  await user.click(screen.getByRole("button", { name: "确定" }));

  expect(await screen.findByText("苏州园区")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));
  await user.click(screen.getByRole("button", { name: "关闭" }));
  expect(screen.queryByText("时间与低频条件收进这里，首页保持轻量")).not.toBeInTheDocument();
});

it("renders drawer actions outside the scrollable drawer content", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledTimes(1);
  });

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));

  const drawerContent = screen.getByTestId("home-filter-drawer-content");
  const confirmButton = screen.getByRole("button", { name: "确定" });
  const resetButton = screen.getByRole("button", { name: "重置" });

  expect(drawerContent).not.toContainElement(confirmButton);
  expect(drawerContent).not.toContainElement(resetButton);
});

it("keeps filter drawer available when rendered inside app layout with bottom tab bar", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter initialEntries={["/"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<HomePage />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenCalledTimes(1);
  });

  expect(screen.getByRole("navigation", { name: "底部导航" })).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "打开筛选抽屉" }));

  expect(screen.getByRole("button", { name: "确定" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "重置" })).toBeInTheDocument();
});
