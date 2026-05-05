import { create } from "zustand";

import {
  api,
  attachAuthStore,
  type ApiResponse,
  unwrapApiResponse,
} from "../../lib/http";

export type AdminProfile = {
  id: number;
  username: string;
  name: string;
  status: string;
};

type AuthPayload = {
  accessToken: string;
  refreshToken: string;
  admin?: AdminProfile | null;
};

type AuthState = {
  accessToken: string | null;
  refreshToken: string | null;
  admin: AdminProfile | null;
  login: (payload: { username: string; password: string }) => Promise<void>;
  refresh: () => Promise<void>;
  logout: () => void;
};

const AUTH_STORAGE_KEY = "mn-admin-auth";

function readStoredAuth() {
  if (typeof window === "undefined") {
    return {
      accessToken: null,
      refreshToken: null,
      admin: null,
    };
  }

  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return {
      accessToken: null,
      refreshToken: null,
      admin: null,
    };
  }

  try {
    const parsed = JSON.parse(raw) as {
      accessToken?: string | null;
      refreshToken?: string | null;
      admin?: AdminProfile | null;
    };

    return {
      accessToken: parsed.accessToken ?? null,
      refreshToken: parsed.refreshToken ?? null,
      admin: parsed.admin ?? null,
    };
  } catch {
    return {
      accessToken: null,
      refreshToken: null,
      admin: null,
    };
  }
}

function persistAuth(state: Pick<AuthState, "accessToken" | "refreshToken" | "admin">) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(state));
}

function clearStoredAuth() {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(AUTH_STORAGE_KEY);
}

function applyAuthPayload(
  set: (partial: Partial<AuthState>) => void,
  payload: AuthPayload,
) {
  const nextState = {
    accessToken: payload.accessToken,
    refreshToken: payload.refreshToken,
    admin: payload.admin ?? null,
  };

  persistAuth(nextState);
  set(nextState);
}

const initialState = readStoredAuth();

export const useAdminAuthStore = create<AuthState>((set, get) => ({
  ...initialState,
  login: async (payload) => {
    const response = await api.post<ApiResponse<AuthPayload>>("/admin/v1/auth/login", payload);
    applyAuthPayload(set, unwrapApiResponse(response.data));
  },
  refresh: async () => {
    const refreshToken = get().refreshToken;
    if (!refreshToken) {
      throw new Error("Missing refresh token");
    }

    const response = await api.post<ApiResponse<AuthPayload>>(
      "/admin/v1/auth/refresh",
      null,
      {
        skipAuthRefresh: true,
        useRefreshToken: true,
      },
    );
    applyAuthPayload(set, unwrapApiResponse(response.data));
  },
  logout: () => {
    clearStoredAuth();
    set({
      accessToken: null,
      refreshToken: null,
      admin: null,
    });
  },
}));

attachAuthStore(useAdminAuthStore);
