import axios from "axios";

declare module "axios" {
  interface AxiosRequestConfig {
    _retry?: boolean;
    skipAuthRefresh?: boolean;
    useRefreshToken?: boolean;
  }

  interface InternalAxiosRequestConfig {
    _retry?: boolean;
    skipAuthRefresh?: boolean;
    useRefreshToken?: boolean;
  }
}

export const API_SUCCESS_CODE = 1000;
export const API_INVALID_TOKEN_CODE = 1007;

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

type AuthSnapshot = {
  accessToken: string | null;
  refreshToken: string | null;
};

type RequestConfigWithAuth = {
  _retry?: boolean;
  skipAuthRefresh?: boolean;
  useRefreshToken?: boolean;
};

export class ApiBusinessError extends Error {
  code: number;

  constructor(message: string, code: number) {
    super(message);
    this.name = "ApiBusinessError";
    this.code = code;
  }
}

const AUTH_STORAGE_KEY = "mn-admin-auth";

function readStoredAccessToken() {
  return readStoredAuth().accessToken;
}

function readStoredAuth() {
  if (typeof window === "undefined") {
    return {
      accessToken: null,
      refreshToken: null,
    };
  }

  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return {
      accessToken: null,
      refreshToken: null,
    };
  }

  try {
    const parsed = JSON.parse(raw) as AuthSnapshot;
    return {
      accessToken: parsed.accessToken ?? null,
      refreshToken: parsed.refreshToken ?? null,
    };
  } catch {
    return {
      accessToken: null,
      refreshToken: null,
    };
  }
}

export const api = axios.create();

let authStore: AuthStoreBinding | null = null;
let refreshPromise: Promise<void> | null = null;

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
  const auth = readStoredAuth();
  const requestConfig = config as typeof config & RequestConfigWithAuth;
  const token = requestConfig.useRefreshToken ? auth.refreshToken : auth.accessToken;

  if (token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

function shouldRefresh(response: ApiResponse<unknown>) {
  return response.code === API_INVALID_TOKEN_CODE;
}

async function refreshAuthOnce() {
  if (!authStore) {
    throw new Error("Auth store is not attached");
  }

  if (!refreshPromise) {
    refreshPromise = authStore
      .getState()
      .refresh()
      .finally(() => {
        refreshPromise = null;
      });
  }

  return refreshPromise;
}

api.interceptors.response.use(
  async (response) => {
    const payload = response.data as ApiResponse<unknown> | undefined;
    const config = (response.config ?? {}) as typeof response.config & RequestConfigWithAuth;

    if (!payload || config.skipAuthRefresh || config._retry || !shouldRefresh(payload)) {
      return response;
    }

    if (!authStore) {
      return response;
    }

    try {
      await refreshAuthOnce();
    } catch (error) {
      authStore.getState().logout();
      throw error;
    }

    return api.request({
      ...config,
      _retry: true,
    });
  },
  async (error) => {
    throw error;
  },
);
