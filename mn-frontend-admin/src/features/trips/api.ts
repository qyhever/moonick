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
  priceAmount: number;
  remark: string;
  contactWechat: string;
  contactPhone: string;
  createdAt: string;
  updatedAt: string;
};

export type AdminTripUpdatePayload = {
  tripType: string;
  fromText: string;
  toText: string;
  departureDate: string;
  departureTime: string;
  seatCount: number;
  priceAmount: number;
  isPriceNegotiable: boolean;
  contactWechat: string;
  contactPhone: string;
  remark: string;
  status: AdminTripStatus;
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
    "/admin/v1/trips",
    withQuery(params),
  );
  return unwrapApiResponse(response.data);
}

export async function getAdminTripDetail(id: string | number) {
  const response = await api.get<ApiResponse<AdminTripDetail>>(`/admin/v1/trips/${id}`);
  return unwrapApiResponse(response.data);
}

export async function updateAdminTrip(id: string | number, payload: AdminTripUpdatePayload) {
  const response = await api.put<ApiResponse<AdminTripDetail>>(`/admin/v1/trips/${id}`, payload);
  return unwrapApiResponse(response.data);
}
