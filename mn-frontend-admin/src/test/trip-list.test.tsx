import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockGetAdminTrips } = vi.hoisted(() => ({
  mockGetAdminTrips: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getAdminTrips: mockGetAdminTrips,
}));

import TripListPage from "../features/trips/TripListPage";

beforeEach(() => {
  mockGetAdminTrips.mockReset();
});

it("requests paginated trips and reloads when page changes", async () => {
  mockGetAdminTrips.mockResolvedValue({
    items: [
      {
        id: 1,
        userId: 10,
        tripType: "driver_post",
        fromText: "广州",
        toText: "深圳",
        departureDate: "2026-05-02",
        departureTime: "09:00",
        seatCount: 3,
        isPriceNegotiable: false,
        status: "active",
        favorited: false,
      },
    ],
    total: 21,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <TripListPage />
    </MemoryRouter>,
  );

  await screen.findByText("广州 → 深圳");
  expect(screen.getByText("车找人")).toBeInTheDocument();
  expect(screen.getByText("可约")).toBeInTheDocument();
  expect(screen.queryByText("driver_post")).not.toBeInTheDocument();
  expect(screen.queryByText("active")).not.toBeInTheDocument();

  expect(mockGetAdminTrips).toHaveBeenCalledWith({
    pageNum: 1,
    pageSize: 10,
  });

  mockGetAdminTrips.mockResolvedValue({
    items: [
      {
        id: 11,
        userId: 11,
        tripType: "passenger_post",
        fromText: "上海",
        toText: "杭州",
        departureDate: "2026-05-03",
        departureTime: "10:00",
        seatCount: 2,
        isPriceNegotiable: true,
        status: "full",
        favorited: false,
      },
    ],
    total: 21,
    pageNum: 2,
    pageSize: 10,
  });

  await userEvent.click(screen.getByRole("listitem", { name: "2" }));

  await waitFor(() => {
    expect(mockGetAdminTrips).toHaveBeenLastCalledWith({
      pageNum: 2,
      pageSize: 10,
    });
  });
  expect(await screen.findByText("上海 → 杭州")).toBeInTheDocument();
  expect(screen.getByText("人找车")).toBeInTheDocument();
  expect(screen.getByText("已满")).toBeInTheDocument();
});
