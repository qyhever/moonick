import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";

export type TripType = "driver_post" | "passenger_post";
export type TripStatus = "active" | "full" | "closed" | "expired";

export type TripSummary = {
  id: number;
  userId: number;
  tripType: TripType | string;
  fromText: string;
  toText: string;
  departureDate: string;
  departureTime: string;
  seatCount: number;
  isPriceNegotiable: boolean;
  status: TripStatus;
  favorited: boolean;
  unavailable: boolean;
};

export type TripDetail = TripSummary & {
  contactPhone: string;
  contactWechat: string;
  createdAt: string;
  updatedAt: string;
};

export type TripFormPayload = {
  tripType: TripType;
  fromText: string;
  toText: string;
  departureDate: string;
  departureTime: string;
  seatCount: number;
  isPriceNegotiable: boolean;
  contactPhone: string;
  contactWechat: string;
};

export type TripListResponse = {
  items: TripSummary[];
  total: number;
  pageNum: number;
  pageSize: number;
};

export type TripQuery = {
  pageNum?: number;
  pageSize?: number;
  tripType?: string;
  status?: string;
  keyword?: string;
  fromText?: string;
  toText?: string;
  datePreset?: "today" | "tomorrow";
};

type FavoriteToggleResponse = {
  favorited: boolean;
};

type AvatarUploadResponse = {
  url?: string;
  avatarUrl?: string;
};

export async function getTrips(query?: TripQuery) {
  const response = await api.get<ApiResponse<TripListResponse>>("/api/v1/trips", {
    params: query,
  });
  return unwrapApiResponse(response.data);
}

export async function getTripDetail(id: string | number) {
  const response = await api.get<ApiResponse<TripDetail>>(`/api/v1/trips/${id}`);
  return unwrapApiResponse(response.data);
}

export async function createTrip(payload: TripFormPayload) {
  const response = await api.post<ApiResponse<TripDetail>>("/api/v1/trips", payload);
  return unwrapApiResponse(response.data);
}

export async function updateTrip(id: string | number, payload: TripFormPayload) {
  const response = await api.put<ApiResponse<TripDetail>>(`/api/v1/trips/${id}`, payload);
  return unwrapApiResponse(response.data);
}

export async function updateTripStatus(id: string | number, status: TripStatus) {
  const response = await api.patch<ApiResponse<TripDetail>>(`/api/v1/trips/${id}/status`, { status });
  return unwrapApiResponse(response.data);
}

export async function toggleFavorite(id: string | number) {
  const response = await api.post<ApiResponse<FavoriteToggleResponse>>(
    `/api/v1/trips/${id}/favorite`,
  );
  return unwrapApiResponse(response.data);
}

export async function getMyTrips() {
  const response = await api.get<ApiResponse<TripListResponse>>("/api/v1/me/trips");
  return unwrapApiResponse(response.data);
}

export async function getMyFavorites() {
  const response = await api.get<ApiResponse<TripListResponse>>("/api/v1/me/favorites");
  return unwrapApiResponse(response.data);
}

export async function uploadAvatar(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  const response = await api.post<ApiResponse<AvatarUploadResponse>>("/api/v1/files/avatar", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });

  const data = unwrapApiResponse(response.data);
  return data.avatarUrl ?? data.url ?? "";
}
