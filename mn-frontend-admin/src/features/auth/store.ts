import { create } from "zustand";

import { api, type ApiResponse, unwrapApiResponse } from "../../lib/http";

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

const initialState = readStoredAuth();

export const useAdminAuthStore = create<AuthState>((set) => ({
  ...initialState,
  login: async (payload) => {
    const response = await api.post<ApiResponse<AuthPayload>>("/api/admin/v1/auth/login", payload);
    const data = unwrapApiResponse(response.data);
    const nextState = {
      accessToken: data.accessToken,
      refreshToken: data.refreshToken,
      admin: data.admin ?? null,
    };
    persistAuth(nextState);
    set(nextState);
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
