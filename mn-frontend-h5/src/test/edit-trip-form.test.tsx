import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockNavigate, mockGetTripDetail, mockUpdateTrip, mockUseParams } = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockGetTripDetail: vi.fn(),
  mockUpdateTrip: vi.fn(),
  mockUseParams: vi.fn(),
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();

  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => mockUseParams(),
  };
});

vi.mock("../features/trips/api", () => ({
  getTripDetail: mockGetTripDetail,
  updateTrip: mockUpdateTrip,
}));

import EditTripPage from "../features/trips/pages/EditTripPage";

beforeEach(() => {
  vi.spyOn(Date, "now").mockReturnValue(new Date("2026-04-25T12:00:00.000Z").getTime());
  mockNavigate.mockReset();
  mockGetTripDetail.mockReset();
  mockUpdateTrip.mockReset();
  mockUseParams.mockReturnValue({ id: "12" });
});

afterEach(() => {
  vi.restoreAllMocks();
});

it("loads and submits remark in edit form", async () => {
  const user = userEvent.setup();
  mockGetTripDetail.mockResolvedValue({
    id: 12,
    userId: 7,
    tripType: "driver_post",
    fromText: "上海",
    toText: "杭州",
    departureDate: "2026-04-26",
    departureTime: "08:30",
    seatCount: 2,
    isPriceNegotiable: true,
    status: "active",
    favorited: false,
    unavailable: false,
    contactPhone: "13800138000",
    contactWechat: "moonick",
    remark: "原备注",
    createdAt: "2026-04-20T12:00:00Z",
    updatedAt: "2026-04-20T12:00:00Z",
  });
  mockUpdateTrip.mockResolvedValue({ id: 12 });

  render(
    <MemoryRouter>
      <EditTripPage />
    </MemoryRouter>,
  );

  expect(await screen.findByDisplayValue("原备注")).toBeInTheDocument();

  await user.clear(screen.getByLabelText("备注"));
  await user.type(screen.getByLabelText("备注"), " 更新后的备注 ");
  await user.click(screen.getByRole("button", { name: "保存" }));

  await waitFor(() => {
    expect(mockUpdateTrip).toHaveBeenCalledWith("12", {
      tripType: "driver_post",
      fromText: "上海",
      toText: "杭州",
      departureDate: "2026-04-26",
      departureTime: "08:30",
      seatCount: 2,
      isPriceNegotiable: true,
      contactPhone: "13800138000",
      contactWechat: "moonick",
      remark: "更新后的备注",
    });
  });
  expect(mockNavigate).toHaveBeenCalledWith("/trips/12");
});
