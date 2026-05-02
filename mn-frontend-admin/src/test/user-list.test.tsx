import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const { mockGetAdminUsers } = vi.hoisted(() => ({
  mockGetAdminUsers: vi.fn(),
}));

vi.mock("../features/users/api", () => ({
  getAdminUsers: mockGetAdminUsers,
}));

import UserListPage from "../features/users/UserListPage";

beforeEach(() => {
  mockGetAdminUsers.mockReset();
});

it("requests paginated users and reloads when page changes", async () => {
  mockGetAdminUsers.mockResolvedValue({
    items: [
      {
        id: 1,
        email: "user1@example.com",
        phone: "13800138000",
        nickname: "测试用户",
        status: "active",
        createdAt: "2026-05-02T09:45:03+08:00",
      },
    ],
    total: 21,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <UserListPage />
    </MemoryRouter>,
  );

  await screen.findByText("测试用户");
  expect(screen.getByText("2026-05-02 09:45:03")).toBeInTheDocument();
  expect(screen.getByText("正常")).toBeInTheDocument();
  expect(screen.queryByText("active")).not.toBeInTheDocument();

  expect(mockGetAdminUsers).toHaveBeenCalledWith({
    keyword: "",
    pageNum: 1,
    pageSize: 10,
  });

  mockGetAdminUsers.mockResolvedValue({
    items: [
      {
        id: 11,
        email: "user2@example.com",
        phone: "13900139000",
        nickname: "第二页用户",
        status: "active",
        createdAt: "2026-05-03T08:30:00+08:00",
      },
    ],
    total: 21,
    pageNum: 2,
    pageSize: 10,
  });

  await userEvent.click(screen.getByRole("listitem", { name: "2" }));

  await waitFor(() => {
    expect(mockGetAdminUsers).toHaveBeenLastCalledWith({
      keyword: "",
      pageNum: 2,
      pageSize: 10,
    });
  });
  expect(await screen.findByText("第二页用户")).toBeInTheDocument();
});
