import axios from "axios";

export const API_SUCCESS_CODE = 1000;

export type ApiResponse<T> = {
  code: number;
  message: string;
  data: T | null;
};

type AuthStoreBinding = {
  getState: () => {
    refresh: () => Promise<void>;
    logout: () => void;
  };
};

const AUTH_STORAGE_KEY = "mn-h5-auth";

export class ApiBusinessError extends Error {
  code: number;

  constructor(message: string, code: number) {
    super(message);
    this.name = "ApiBusinessError";
    this.code = code;
  }
}

export class RefreshUnavailableError extends Error {
  constructor(message = "刷新登录态尚未接入后端接口") {
    super(message);
    this.name = "RefreshUnavailableError";
  }
}

function readStoredAccessToken() {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as { accessToken?: string | null };
    return parsed.accessToken ?? null;
  } catch {
    return null;
  }
}

export const api = axios.create();

let authStore: AuthStoreBinding | null = null;

export function attachAuthStore(store: AuthStoreBinding) {
  authStore = store;
}

export function unwrapApiResponse<T>(response: ApiResponse<T>): T {
  if (response.code !== API_SUCCESS_CODE) {
    throw new ApiBusinessError(response.message || "请求失败", response.code);
  }

  if (response.data === null) {
    throw new ApiBusinessError(response.message || "响应数据为空", response.code);
  }

  return response.data;
}

api.interceptors.request.use((config) => {
  const accessToken = readStoredAccessToken();

  if (accessToken) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${accessToken}`;
  }

  return config;
});
