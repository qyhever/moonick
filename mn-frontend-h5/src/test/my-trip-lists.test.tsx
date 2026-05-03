import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { ReactNode } from "react";
import { beforeEach, vi } from "vitest";

const {
  mockGetMyFavorites,
  mockGetMyTrips,
  mockUpdateTripStatus,
} = vi.hoisted(() => ({
  mockGetMyFavorites: vi.fn(),
  mockGetMyTrips: vi.fn(),
  mockUpdateTripStatus: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getMyFavorites: mockGetMyFavorites,
  getMyTrips: mockGetMyTrips,
  updateTripStatus: mockUpdateTripStatus,
}));

vi.mock("react-infinite-scroll-component", () => ({
  default: ({
    children,
    next,
    hasMore,
    refreshFunction,
    dataLength,
    pullDownToRefreshContent,
    releaseToRefreshContent,
  }: {
    children: ReactNode;
    next: () => void;
    hasMore: boolean;
    refreshFunction?: () => void | Promise<void>;
    dataLength: number;
    pullDownToRefreshContent?: ReactNode;
    releaseToRefreshContent?: ReactNode;
  }) => (
    <div
      data-testid="infinite-scroll"
      data-has-more={String(hasMore)}
      data-length={String(dataLength)}
    >
      {pullDownToRefreshContent}
      {releaseToRefreshContent}
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

import MyFavoritesPage from "../features/trips/pages/MyFavoritesPage";
import MyTripsPage from "../features/trips/pages/MyTripsPage";

function createTrip(id: number, fromText: string, toText: string) {
  return {
    id,
    userId: id + 100,
    tripType: "driver_post" as const,
    fromText,
    toText,
    departureDate: "2026-05-01",
    departureTime: "09:00",
    seatCount: 3,
    isPriceNegotiable: true,
    status: "active" as const,
    favorited: false,
    unavailable: false,
  };
}

beforeEach(() => {
  mockGetMyFavorites.mockReset();
  mockGetMyTrips.mockReset();
  mockUpdateTripStatus.mockReset();
});

it("loads more favorites and appends them", async () => {
  const user = userEvent.setup();

  mockGetMyFavorites
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
      <MyFavoritesPage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetMyFavorites).toHaveBeenNthCalledWith(1, {
      pageNum: 1,
      pageSize: 10,
    });
  });

  expect(await screen.findByText("上海虹桥")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "触发加载更多" }));

  await waitFor(() => {
    expect(mockGetMyFavorites).toHaveBeenNthCalledWith(2, {
      pageNum: 2,
      pageSize: 10,
    });
  });

  expect(await screen.findByText("北京南站")).toBeInTheDocument();
  expect(screen.getByText("上海虹桥")).toBeInTheDocument();
});

it("refreshes favorites from page one and replaces old items", async () => {
  const user = userEvent.setup();

  mockGetMyFavorites
    .mockResolvedValueOnce({
      items: [createTrip(1, "上海虹桥", "杭州东站")],
      total: 2,
      pageNum: 1,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(3, "深圳北站", "广州南站")],
      total: 1,
      pageNum: 1,
      pageSize: 10,
    });

  render(
    <MemoryRouter>
      <MyFavoritesPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("上海虹桥")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "触发下拉刷新" }));

  await waitFor(() => {
    expect(mockGetMyFavorites).toHaveBeenNthCalledWith(2, {
      pageNum: 1,
      pageSize: 10,
    });
  });

  expect(await screen.findByText("深圳北站")).toBeInTheDocument();
  expect(screen.queryByText("上海虹桥")).not.toBeInTheDocument();
});

it("loads more published trips and appends them", async () => {
  const user = userEvent.setup();

  mockGetMyTrips
    .mockResolvedValueOnce({
      items: [createTrip(11, "苏州园区", "南京南站")],
      total: 2,
      pageNum: 1,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(12, "合肥南站", "武汉站")],
      total: 2,
      pageNum: 2,
      pageSize: 10,
    });

  render(
    <MemoryRouter>
      <MyTripsPage />
    </MemoryRouter>,
  );

  await waitFor(() => {
    expect(mockGetMyTrips).toHaveBeenNthCalledWith(1, {
      pageNum: 1,
      pageSize: 10,
    });
  });

  expect(await screen.findByText("苏州园区")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "触发加载更多" }));

  await waitFor(() => {
    expect(mockGetMyTrips).toHaveBeenNthCalledWith(2, {
      pageNum: 2,
      pageSize: 10,
    });
  });

  expect(await screen.findByText("合肥南站")).toBeInTheDocument();
  expect(screen.getByText("苏州园区")).toBeInTheDocument();
});

it("refreshes published trips from page one and replaces old items", async () => {
  const user = userEvent.setup();

  mockGetMyTrips
    .mockResolvedValueOnce({
      items: [createTrip(11, "苏州园区", "南京南站")],
      total: 2,
      pageNum: 1,
      pageSize: 10,
    })
    .mockResolvedValueOnce({
      items: [createTrip(13, "青岛北站", "济南西站")],
      total: 1,
      pageNum: 1,
      pageSize: 10,
    });

  render(
    <MemoryRouter>
      <MyTripsPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("苏州园区")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "触发下拉刷新" }));

  await waitFor(() => {
    expect(mockGetMyTrips).toHaveBeenNthCalledWith(2, {
      pageNum: 1,
      pageSize: 10,
    });
  });

  expect(await screen.findByText("青岛北站")).toBeInTheDocument();
  expect(screen.queryByText("苏州园区")).not.toBeInTheDocument();
});

it("shows release-to-refresh copy on favorites and published trip pages", async () => {
  mockGetMyFavorites.mockResolvedValue({
    items: [createTrip(1, "上海虹桥", "杭州东站")],
    total: 1,
    pageNum: 1,
    pageSize: 10,
  });
  mockGetMyTrips.mockResolvedValue({
    items: [createTrip(11, "苏州园区", "南京南站")],
    total: 1,
    pageNum: 1,
    pageSize: 10,
  });

  const { unmount } = render(
    <MemoryRouter>
      <MyFavoritesPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("松开立即刷新")).toBeInTheDocument();

  unmount();

  render(
    <MemoryRouter>
      <MyTripsPage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("松开立即刷新")).toBeInTheDocument();
});
