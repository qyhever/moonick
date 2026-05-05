import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";
import type { AuthUser } from "../../store/auth";

type OkResponse = {
  ok: boolean;
};

type VerificationCodeResponse = {
  sent: boolean;
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

export async function sendVerificationCode(email: string, type: string) {
  const response = await api.post<ApiResponse<VerificationCodeResponse>>("/api/v1/auth/code", {
    email,
    type,
  });
  return unwrapApiResponse(response.data);
}

export async function resetPassword(email: string, code: string, password: string) {
  const response = await api.post<ApiResponse<OkResponse>>("/api/v1/auth/password/reset", {
    email,
    code,
    password,
  });
  return unwrapApiResponse(response.data);
}
