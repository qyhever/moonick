import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { vi } from "vitest";

const { mockGetAdminUserDetail, mockGetAdminUserTrips } = vi.hoisted(() => ({
  mockGetAdminUserDetail: vi.fn(),
  mockGetAdminUserTrips: vi.fn(),
}));

vi.mock("../features/users/api", () => ({
  getAdminUserDetail: mockGetAdminUserDetail,
  getAdminUserTrips: mockGetAdminUserTrips,
}));

import UserDetailPage from "../features/users/UserDetailPage";

it("does not render destructive actions on user detail page", async () => {
  mockGetAdminUserDetail.mockResolvedValue({
    id: 8,
    email: "user8@example.com",
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    createdAt: "2026-05-02T09:45:03+08:00",
    defaultWechat: "mingye-user",
    defaultPhone: "13800138000",
    publishedTripCount: 4,
    favoriteCount: 2,
  });
  mockGetAdminUserTrips.mockResolvedValue({
    items: [],
    total: 0,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter initialEntries={["/users/8"]}>
      <Routes>
        <Route path="/users/:id" element={<UserDetailPage />} />
      </Routes>
    </MemoryRouter>,
  );

  expect(await screen.findByText("基本资料")).toBeInTheDocument();
  expect(screen.getByText("正常")).toBeInTheDocument();
  expect(screen.getByText("2026-05-02 09:45:03")).toBeInTheDocument();
  expect(screen.queryByText("active")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "封禁" })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "删除" })).not.toBeInTheDocument();
});
