import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";
import type { AuthUser } from "../../store/auth";

type OkResponse = {
  ok: boolean;
};

export async function updateUserProfile(nickname: string) {
  const response = await api.put<ApiResponse<OkResponse>>("/api/v1/users/profile", { nickname });
  return unwrapApiResponse(response.data);
}

export async function updateUserContact(defaultWechat: string, defaultPhone: string) {
  const response = await api.put<ApiResponse<OkResponse>>("/api/v1/users/contact", {
    defaultWechat,
    defaultPhone,
  });
  return unwrapApiResponse(response.data);
}

export async function getCurrentUserProfile() {
  const response = await api.get<ApiResponse<AuthUser>>("/api/v1/users/me");
  return unwrapApiResponse(response.data);
}
