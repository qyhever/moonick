import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockGetAdminTripDetail, mockUpdateAdminTrip } = vi.hoisted(() => ({
  mockGetAdminTripDetail: vi.fn(),
  mockUpdateAdminTrip: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  getAdminTripDetail: mockGetAdminTripDetail,
  updateAdminTrip: mockUpdateAdminTrip,
}));

import TripEditPage from "../features/trips/TripEditPage";
import TripDetailPage from "../features/trips/TripDetailPage";

const tripDetail = {
  id: 3,
  userId: 9,
  tripType: "driver_post",
  fromText: "上海",
  toText: "杭州",
  departureDate: "2026-04-30",
  departureTime: "09:00",
  seatCount: 2,
  priceAmount: 88.5,
  isPriceNegotiable: true,
  contactWechat: "mingye",
  contactPhone: "13800138000",
  remark: "原始备注",
  status: "active",
  favorited: false,
  createdAt: "2026-04-25T10:00:00Z",
  updatedAt: "2026-04-25T10:00:00Z",
};

function renderTripEditPage() {
  render(
    <MemoryRouter initialEntries={["/trips/3/edit"]}>
      <Routes>
        <Route path="/trips/:id/edit" element={<TripEditPage />} />
        <Route path="/trips/:id" element={<div>trip detail</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

function renderTripDetailPage() {
  render(
    <MemoryRouter initialEntries={["/trips/3"]}>
      <Routes>
        <Route path="/trips/:id" element={<TripDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockGetAdminTripDetail.mockReset();
  mockUpdateAdminTrip.mockReset();
});

it("loads trip detail into the full edit form", async () => {
  mockGetAdminTripDetail.mockResolvedValue({
    ...tripDetail,
  });

  renderTripEditPage();

  expect(await screen.findByLabelText("出发地")).toHaveValue("上海");
  expect(screen.getByLabelText("目的地")).toHaveValue("杭州");
  expect(screen.getByLabelText("出发日期")).toHaveValue("2026-04-30");
  expect(screen.getByLabelText("出发时间")).toHaveValue("09:00");
  expect(screen.getByLabelText("费用金额")).toHaveValue("88.50");
  expect(screen.getByLabelText("备注")).toHaveValue("原始备注");
});

it("requires confirmation before saving trip edit", async () => {
  mockGetAdminTripDetail.mockResolvedValue({
    ...tripDetail,
  });

  renderTripEditPage();

  await userEvent.click(await screen.findByRole("button", { name: /保\s*存/ }));

  expect(await screen.findByText("确认保存修改吗？")).toBeInTheDocument();
  expect(mockUpdateAdminTrip).not.toHaveBeenCalled();
});

it("submits full payload only after confirmation", async () => {
  mockGetAdminTripDetail.mockResolvedValue({
    ...tripDetail,
  });
  mockUpdateAdminTrip.mockResolvedValue({
    ...tripDetail,
  });

  renderTripEditPage();

  await userEvent.clear(await screen.findByLabelText("出发地"));
  await userEvent.type(screen.getByLabelText("出发地"), "苏州");
  await userEvent.clear(screen.getByLabelText("费用金额"));
  await userEvent.type(screen.getByLabelText("费用金额"), "66.6");
  await userEvent.click(screen.getByRole("switch"));
  await userEvent.clear(screen.getByLabelText("备注"));
  await userEvent.type(screen.getByLabelText("备注"), "后台改价");

  await userEvent.click(screen.getByRole("button", { name: /保\s*存/ }));
  expect(await screen.findByText("确认保存修改吗？")).toBeInTheDocument();
  expect(mockUpdateAdminTrip).not.toHaveBeenCalled();

  await userEvent.click(screen.getByRole("button", { name: /确\s*认/ }));

  await waitFor(() => {
    expect(mockUpdateAdminTrip).toHaveBeenCalledWith("3", {
      tripType: "driver_post",
      fromText: "苏州",
      toText: "杭州",
      departureDate: "2026-04-30",
      departureTime: "09:00",
      seatCount: 2,
      priceAmount: 66.6,
      isPriceNegotiable: false,
      contactWechat: "mingye",
      contactPhone: "13800138000",
      remark: "后台改价",
      status: "active",
    });
  });
});

it("hides edit entry and disables saving for expired trip", async () => {
  const expiredTrip = {
    ...tripDetail,
    status: "expired",
  };
  mockGetAdminTripDetail.mockResolvedValue(expiredTrip);

  renderTripDetailPage();

  expect(await screen.findByText("行程详情 #3")).toBeInTheDocument();
  expect(screen.queryByRole("link", { name: "编辑行程信息" })).not.toBeInTheDocument();

  renderTripEditPage();

  expect(await screen.findByRole("button", { name: /保\s*存/ })).toBeDisabled();
});
