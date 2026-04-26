import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockNavigate, mockCreateTrip } = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockCreateTrip: vi.fn(),
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();

  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock("../features/trips/api", () => ({
  createTrip: mockCreateTrip,
}));

import PublishPage from "../features/trips/pages/PublishPage";

beforeEach(() => {
  vi.spyOn(Date, "now").mockReturnValue(new Date("2026-04-25T12:00:00.000Z").getTime());
  mockNavigate.mockReset();
  mockCreateTrip.mockReset();
});

afterEach(() => {
  vi.restoreAllMocks();
});

it("blocks publish when departure is earlier than now", async () => {
  const user = userEvent.setup();

  render(
    <MemoryRouter>
      <PublishPage />
    </MemoryRouter>,
  );

  await user.type(screen.getByLabelText("起点"), "上海");
  await user.type(screen.getByLabelText("终点"), "杭州");
  await user.type(screen.getByLabelText("出发日期"), "2026-04-24");
  await user.type(screen.getByLabelText("出发时间"), "08:30");
  await user.type(screen.getByLabelText("手机号"), "13800138000");
  await user.click(screen.getByRole("button", { name: "发布" }));

  expect(await screen.findByText("出发时间不能早于当前时间")).toBeInTheDocument();
  expect(mockCreateTrip).not.toHaveBeenCalled();
  expect(mockNavigate).not.toHaveBeenCalled();
});

it("submits backend-compatible payload", async () => {
  const user = userEvent.setup();
  mockCreateTrip.mockResolvedValue({ id: 9 });

  render(
    <MemoryRouter>
      <PublishPage />
    </MemoryRouter>,
  );

  await user.click(screen.getByRole("button", { name: "人找车" }));
  await user.type(screen.getByLabelText("起点"), "上海");
  await user.type(screen.getByLabelText("终点"), "杭州");
  await user.type(screen.getByLabelText("出发日期"), "2026-04-26");
  await user.type(screen.getByLabelText("出发时间"), "08:30");
  await user.clear(screen.getByLabelText("人数"));
  await user.type(screen.getByLabelText("人数"), "2");
  await user.type(screen.getByLabelText("手机号"), "13800138000");
  await user.click(screen.getByRole("button", { name: "发布" }));

  expect(mockCreateTrip).toHaveBeenCalledWith({
    tripType: "passenger_post",
    fromText: "上海",
    toText: "杭州",
    departureDate: "2026-04-26",
    departureTime: "08:30",
    seatCount: 2,
    isPriceNegotiable: true,
    contactPhone: "13800138000",
    contactWechat: "",
  });
  expect(mockNavigate).toHaveBeenCalledWith("/trips/9");
});
