import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const {
  mockGetCurrentUserProfile,
  mockGetMyTrips,
  mockGetMyFavorites,
  mockUpdateUserContact,
  mockUpdateUserProfile,
} = vi.hoisted(() => ({
  mockGetCurrentUserProfile: vi.fn(),
  mockGetMyTrips: vi.fn(),
  mockGetMyFavorites: vi.fn(),
  mockUpdateUserContact: vi.fn(),
  mockUpdateUserProfile: vi.fn(),
}));

vi.mock("../features/profile/api", () => ({
  getCurrentUserProfile: mockGetCurrentUserProfile,
  updateUserContact: mockUpdateUserContact,
  updateUserProfile: mockUpdateUserProfile,
}));

vi.mock("../features/trips/api", () => ({
  getMyTrips: mockGetMyTrips,
  getMyFavorites: mockGetMyFavorites,
}));

vi.mock("../features/profile/components/AvatarUploader", () => ({
  default: function MockAvatarUploader() {
    return <div>AvatarUploader</div>;
  },
}));

import ProfilePage from "../features/profile/pages/ProfilePage";
import AccountSettingsPage from "../features/profile/pages/AccountSettingsPage";
import { useAuthStore } from "../store/auth";

beforeEach(() => {
  window.localStorage.clear();
  useAuthStore.setState({
    accessToken: "access-token",
    refreshToken: "refresh-token",
    user: {
      id: 1,
      phone: "13800138000",
      nickname: "初始用户",
      avatarUrl: "",
      status: "active",
      defaultWechat: "",
      defaultPhone: "13800138000",
    },
  });

  mockGetCurrentUserProfile.mockReset();
  mockGetMyTrips.mockReset();
  mockGetMyFavorites.mockReset();
  mockUpdateUserContact.mockReset();
  mockUpdateUserProfile.mockReset();
});

it("loads profile data once after login hydration", async () => {
  mockGetCurrentUserProfile.mockImplementation(async () => ({
    id: 1,
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    defaultWechat: "",
    defaultPhone: "13800138000",
  }));
  mockGetMyTrips.mockResolvedValue({
    items: [],
    total: 2,
    pageNum: 1,
    pageSize: 10,
  });
  mockGetMyFavorites.mockResolvedValue({
    items: [],
    total: 3,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <ProfilePage />
    </MemoryRouter>,
  );

  expect(await screen.findByRole("heading", { name: "测试用户" })).toBeInTheDocument();

  await waitFor(() => {
    expect(mockGetCurrentUserProfile).toHaveBeenCalledTimes(1);
    expect(mockGetMyTrips).toHaveBeenCalledTimes(1);
    expect(mockGetMyFavorites).toHaveBeenCalledTimes(1);
  });
});

it("renders the profile overview with a dedicated account settings entry", async () => {
  mockGetCurrentUserProfile.mockImplementation(async () => ({
    id: 1,
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    defaultWechat: "wechat_01",
    defaultPhone: "13800138000",
  }));
  mockGetMyTrips.mockResolvedValue({
    items: [],
    total: 12,
    pageNum: 1,
    pageSize: 10,
  });
  mockGetMyFavorites.mockResolvedValue({
    items: [],
    total: 5,
    pageNum: 1,
    pageSize: 10,
  });

  render(
    <MemoryRouter>
      <ProfilePage />
    </MemoryRouter>,
  );

  expect(await screen.findByText("常用服务")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: /账户设置/i })).toBeInTheDocument();
  expect(screen.getByText("信用评分")).toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "保存修改" })).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "退出登录" })).not.toBeInTheDocument();
  expect(screen.queryByText("账号安全")).not.toBeInTheDocument();
});

it("renders the dedicated account settings page and keeps edit actions working", async () => {
  mockGetCurrentUserProfile.mockResolvedValue({
    id: 1,
    phone: "13800138000",
    nickname: "测试用户",
    avatarUrl: "",
    status: "active",
    defaultWechat: "wechat_01",
    defaultPhone: "13800138000",
  });
  mockUpdateUserProfile.mockResolvedValue({ ok: true });
  mockUpdateUserContact.mockResolvedValue({ ok: true });

  render(
    <MemoryRouter>
      <AccountSettingsPage />
    </MemoryRouter>,
  );

  expect(await screen.findByRole("heading", { name: "账户设置" })).toBeInTheDocument();
  expect(screen.getByText("AvatarUploader")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "保存修改" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "退出登录" })).toBeInTheDocument();

  fireEvent.click(screen.getByRole("button", { name: "修改" }));
  fireEvent.change(screen.getByLabelText("修改昵称"), { target: { value: "新昵称" } });
  fireEvent.click(screen.getAllByRole("button", { name: "去管理" })[0]);
  fireEvent.change(screen.getByLabelText("默认手机号"), { target: { value: "13900001111" } });
  fireEvent.change(screen.getByLabelText("默认微信号"), { target: { value: "wechat_02" } });
  fireEvent.click(screen.getByRole("button", { name: "保存修改" }));

  await waitFor(() => {
    expect(mockUpdateUserProfile).toHaveBeenCalledWith("新昵称");
    expect(mockUpdateUserContact).toHaveBeenCalledWith("wechat_02", "13900001111");
  });
});
