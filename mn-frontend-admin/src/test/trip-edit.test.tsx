import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { vi } from "vitest";

const { mockGetAdminTripDetail, mockUpdateAdminTrip } = vi.hoisted(() => ({
  mockGetAdminTripDetail: vi.fn(),
  mockUpdateAdminTrip: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getAdminTripDetail: mockGetAdminTripDetail,
  updateAdminTrip: mockUpdateAdminTrip,
}));

import TripEditPage from "../features/trips/TripEditPage";

it("requires confirmation before saving trip edit", async () => {
  mockGetAdminTripDetail.mockResolvedValue({
    id: 3,
    userId: 9,
    tripType: "driver_post",
    fromText: "上海",
    toText: "杭州",
    departureDate: "2026-04-30",
    departureTime: "09:00",
    seatCount: 2,
    isPriceNegotiable: true,
    contactWechat: "mingye",
    contactPhone: "13800138000",
    status: "active",
    favorited: false,
    createdAt: "2026-04-25T10:00:00Z",
    updatedAt: "2026-04-25T10:00:00Z",
  });

  render(
    <MemoryRouter initialEntries={["/trips/3/edit"]}>
      <Routes>
        <Route path="/trips/:id/edit" element={<TripEditPage />} />
      </Routes>
    </MemoryRouter>,
  );

  await userEvent.click(await screen.findByRole("button", { name: /保\s*存/ }));

  expect(await screen.findByText("确认保存修改吗？")).toBeInTheDocument();
  expect(mockUpdateAdminTrip).not.toHaveBeenCalled();
});
