import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

const { mockGetDashboardSummary } = vi.hoisted(() => ({
  mockGetDashboardSummary: vi.fn(),
}));

vi.mock("../features/dashboard/api", () => ({
  getDashboardSummary: mockGetDashboardSummary,
}));

import DashboardPage from "../features/dashboard/DashboardPage";

it("renders four summary cards", async () => {
  mockGetDashboardSummary.mockResolvedValue({
    totalTrips: 18,
    totalUsers: 9,
    activeTrips: 5,
    expiredTrips: 2,
    totalFavorites: 11,
  });

  render(<DashboardPage />);

  expect(await screen.findByText("行程总数")).toBeInTheDocument();
  expect(await screen.findByText("用户总数")).toBeInTheDocument();
  expect(await screen.findByText("当前有效行程数")).toBeInTheDocument();
  expect(await screen.findByText("过期行程数")).toBeInTheDocument();
});
