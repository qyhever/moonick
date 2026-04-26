import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockGetTripDetail, mockToggleFavorite } = vi.hoisted(() => ({
  mockGetTripDetail: vi.fn(),
  mockToggleFavorite: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getTripDetail: mockGetTripDetail,
  toggleFavorite: mockToggleFavorite,
}));

import TripDetailPage from "../features/trips/pages/TripDetailPage";

beforeEach(() => {
  mockGetTripDetail.mockReset();
  mockToggleFavorite.mockReset();
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

it("toggles favorite state without toast", async () => {
  render(
    <MemoryRouter initialEntries={["/trips/7"]}>
      <Routes>
        <Route path="/trips/:id" element={<TripDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );

  const button = await screen.findByRole("button", { name: "收藏" });
  await userEvent.click(button);

  await waitFor(() => {
    expect(button).toHaveAttribute("data-favorited", "true");
  });
});
