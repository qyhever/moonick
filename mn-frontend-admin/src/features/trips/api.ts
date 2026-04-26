import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";

export type AdminTripStatus = "active" | "full" | "closed" | "expired";

export type AdminTripSummary = {
  id: number;
  userId: number;
  tripType: string;
  fromText: string;
  toText: string;
  departureDate: string;
  departureTime: string;
  seatCount: number;
  isPriceNegotiable: boolean;
  status: AdminTripStatus;
  favorited: boolean;
  unavailable?: boolean;
};

export type AdminTripDetail = AdminTripSummary & {
  contactWechat: string;
  contactPhone: string;
  createdAt: string;
  updatedAt: string;
};

export type AdminTripListResponse = {
  items: AdminTripSummary[];
  total: number;
  pageNum: number;
  pageSize: number;
};

export type AdminTripQuery = {
  pageNum?: number;
  pageSize?: number;
  tripType?: string;
  status?: string;
  keyword?: string;
};

function withQuery(params?: Record<string, string | number | undefined>) {
  return {
    params,
  };
}

export async function getAdminTrips(params?: AdminTripQuery) {
  const response = await api.get<ApiResponse<AdminTripListResponse>>(
    "/api/admin/v1/trips",
    withQuery(params),
  );
  return unwrapApiResponse(response.data);
}

export async function getAdminTripDetail(id: string | number) {
  const response = await api.get<ApiResponse<AdminTripDetail>>(`/api/admin/v1/trips/${id}`);
  return unwrapApiResponse(response.data);
}

export async function updateAdminTrip(id: string | number, status: Exclude<AdminTripStatus, "expired">) {
  const response = await api.put<ApiResponse<AdminTripDetail>>(`/api/admin/v1/trips/${id}`, { status });
  return unwrapApiResponse(response.data);
}
