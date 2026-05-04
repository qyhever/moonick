import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, vi } from "vitest";
import type { ReactNode } from "react";

const { mockGetTrips } = vi.hoisted(() => ({
  mockGetTrips: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getTrips: mockGetTrips,
}));

vi.mock("react-infinite-scroll-component", () => ({
  default: ({
    children,
    next,
    hasMore,
    refreshFunction,
    dataLength,
  }: {
    children: ReactNode;
    next: () => void;
    hasMore: boolean;
    refreshFunction?: () => void | Promise<void>;
    dataLength: number;
  }) => (
    <div
      data-testid="infinite-scroll"
      data-has-more={String(hasMore)}
      data-length={String(dataLength)}
    >
      <button
        type="button"
        onClick={() => {
          if (hasMore) {
            next();
          }
        }}
      >
        触发加载更多
      </button>
      <button
        type="button"
        onClick={() => {
          void refreshFunction?.();
        }}
      >
        触发下拉刷新
      </button>
      {children}
    </div>
  ),
}));

import AppLayout from "../components/MobileTabBar";
import HomePage from "../pages/HomePage";
import { useAuthStore } from "../store/auth";

function createTrip(
  id: number,
  fromText: string,
  toText: string,
  overrides?: Partial<{
    tripType: "driver_post" | "passenger_post";
    status: "active" | "full" | "closed" | "expired";
    isPriceNegotiable: boolean;
  }>,
) {
  return {
    id,
    userId: id + 100,
    tripType: overrides?.tripType ?? "driver_post",
    fromText,
    toText,
    departureDate: "2026-05-01",
    departureTime: "09:00",
    seatCount: 3,
    isPriceNegotiable: overrides?.isPriceNegotiable ?? true,
    status: overrides?.status ?? "active",
    favorited: false,
    unavailable: false,
  };
}

const defaultTripListResponse = {
  items: [
    createTrip(1, "上海虹桥", "杭州东站"),
    createTrip(2, "苏州园区", "南京南站", {
      tripType: "passenger_post",
      status: "full",
      isPriceNegotiable: false,
    }),
  ],
  total: 2,
  pageNum: 1,
  pageSize: 10,
};

beforeEach(() => {
  window.localStorage.clear();
  window.scrollTo = vi.fn();
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

it("loads next page and appends trips when infinite scroll requests more", async () => {
  const user = userEvent.setup();

  mockGetTrips
    .mockResolvedValueOnce({
      items: [createTrip(1, "上海虹桥", "杭州东站")],
      total: 2,
      pageNum: 1,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(2, "北京南站", "天津西站")],
      total: 2,
      pageNum: 2,
      pageSize: 10,
    });

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("上海虹桥")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "触发加载更多" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenLastCalledWith({
      pageNum: 2,
      pageSize: 10,
      tripType: "driver_post",
    });
  });

  expect(await screen.findByText("北京南站")).toBeInTheDocument();
  expect(screen.getByText("上海虹桥")).toBeInTheDocument();
});

it("refreshes trips from page one and replaces old items", async () => {
  const user = userEvent.setup();

  mockGetTrips
    .mockResolvedValueOnce({
      items: [createTrip(1, "上海虹桥", "杭州东站")],
      total: 1,
      pageNum: 1,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(3, "北京南站", "天津西站")],
      total: 1,
      pageNum: 1,
      pageSize: 10,
    });

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("上海虹桥")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "触发下拉刷新" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenLastCalledWith({
      pageNum: 1,
      pageSize: 10,
      tripType: "driver_post",
    });
  });

  expect(await screen.findByText("北京南站")).toBeInTheDocument();
  expect(screen.queryByText("上海虹桥")).not.toBeInTheDocument();
});

it("resets pagination to page one when filters change after loading more", async () => {
  const user = userEvent.setup();

  mockGetTrips
    .mockResolvedValueOnce({
      items: [createTrip(1, "上海虹桥", "杭州东站")],
      total: 20,
      pageNum: 1,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(2, "北京南站", "天津西站")],
      total: 20,
      pageNum: 2,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(4, "深圳北站", "广州南站", { tripType: "passenger_post" })],
      total: 1,
      pageNum: 1,
      pageSize: 10,
    });

  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenNthCalledWith(1, {
      pageNum: 1,
      pageSize: 10,
      tripType: "driver_post",
    });
  });

  await user.click(screen.getByRole("button", { name: "触发加载更多" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenNthCalledWith(2, {
      pageNum: 2,
      pageSize: 10,
      tripType: "driver_post",
    });
  });

  await user.click(screen.getByRole("button", { name: "人找车" }));

  await waitFor(() => {
    expect(mockGetTrips).toHaveBeenLastCalledWith({
      pageNum: 1,
      pageSize: 10,
      tripType: "passenger_post",
    });
  });

  expect(await screen.findByText("深圳北站")).toBeInTheDocument();
});

it("disables further load-more when all trips are loaded", async () => {
  render(
    <MemoryRouter>
      <HomePage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(screen.getByTestId("infinite-scroll")).toHaveAttribute("data-has-more", "false");
  });
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
