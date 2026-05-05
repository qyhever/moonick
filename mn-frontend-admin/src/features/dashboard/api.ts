import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";

export type DashboardSummary = {
  totalUsers: number;
  totalTrips: number;
  activeTrips: number;
  expiredTrips: number;
  totalFavorites: number;
};

export async function getDashboardSummary() {
  const response = await api.get<ApiResponse<DashboardSummary>>("/admin/v1/dashboard/summary");
  return unwrapApiResponse(response.data);
}
