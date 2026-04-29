import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, vi } from "vitest";

const {
  mockGetCurrentUserProfile,
  mockGetMyTrips,
  mockGetMyFavorites,
} = vi.hoisted(() => ({
  mockGetCurrentUserProfile: vi.fn(),
  mockGetMyTrips: vi.fn(),
  mockGetMyFavorites: vi.fn(),
}));

vi.mock("../features/profile/api", () => ({
  getCurrentUserProfile: mockGetCurrentUserProfile,
  updateUserContact: vi.fn(),
  updateUserProfile: vi.fn(),
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

  expect(await screen.findByDisplayValue("测试用户")).toBeInTheDocument();

  await waitFor(() => {
    expect(mockGetCurrentUserProfile).toHaveBeenCalledTimes(1);
    expect(mockGetMyTrips).toHaveBeenCalledTimes(1);
    expect(mockGetMyFavorites).toHaveBeenCalledTimes(1);
  });
});
