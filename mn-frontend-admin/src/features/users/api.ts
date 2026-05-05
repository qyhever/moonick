import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";
import type { AdminTripListResponse } from "../trips/api";

export type AdminUserSummary = {
  id: number;
  email: string;
  phone: string;
  nickname: string;
  status: string;
  createdAt: string;
};

export type AdminUserListResponse = {
  items: AdminUserSummary[];
  total: number;
  pageNum: number;
  pageSize: number;
};

export type AdminUserDetail = {
  id: number;
  email: string;
  phone: string;
  nickname: string;
  avatarUrl: string;
  status: string;
  createdAt: string;
  defaultWechat: string;
  defaultPhone: string;
  publishedTripCount: number;
  favoriteCount: number;
};

export async function getAdminUsers(params?: { pageNum?: number; pageSize?: number; keyword?: string }) {
  const response = await api.get<ApiResponse<AdminUserListResponse>>("/admin/v1/users", { params });
  return unwrapApiResponse(response.data);
}

export async function getAdminUserDetail(id: string | number) {
  const response = await api.get<ApiResponse<AdminUserDetail>>(`/admin/v1/users/${id}`);
  return unwrapApiResponse(response.data);
}

export async function getAdminUserTrips(id: string | number, params?: { pageNum?: number; pageSize?: number }) {
  const response = await api.get<ApiResponse<AdminTripListResponse>>(`/admin/v1/users/${id}/trips`, {
    params,
  });
  return unwrapApiResponse(response.data);
}
