import { create } from "zustand";

import {
  type ApiResponse,
  api,
  attachAuthStore,
  unwrapApiResponse,
} from "../lib/http";

export type AuthUser = {
  id: number;
  email: string;
  phone: string;
  nickname: string;
  avatarUrl: string;
  status: string;
  defaultWechat: string;
  defaultPhone: string;
};

type AuthPayload = {
  accessToken: string;
  refreshToken: string;
  user?: AuthUser | null;
};

type AuthResponse<T> = {
  code: number;
  message: string;
  data: T | null;
};

type Credentials = {
  email: string;
  password: string;
};

type RegisterPayload = Credentials & {
  confirmPassword: string;
};

type AuthState = {
  accessToken: string | null;
  refreshToken: string | null;
  user: AuthUser | null;
  login: (payload: Credentials) => Promise<void>;
  register: (payload: RegisterPayload) => Promise<void>;
  refresh: () => Promise<void>;
  logout: () => void;
  setUser: (user: AuthUser | null) => void;
};

const AUTH_STORAGE_KEY = "mn-h5-auth";

function readStoredAuth() {
  if (typeof window === "undefined") {
    return {
      accessToken: null,
      refreshToken: null,
      user: null,
    };
  }

  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return {
      accessToken: null,
      refreshToken: null,
      user: null,
    };
  }

  try {
    const parsed = JSON.parse(raw) as {
      accessToken?: string | null;
      refreshToken?: string | null;
      user?: AuthUser | null;
    };

    return {
      accessToken: parsed.accessToken ?? null,
      refreshToken: parsed.refreshToken ?? null,
      user: parsed.user ?? null,
    };
  } catch {
    return {
      accessToken: null,
      refreshToken: null,
      user: null,
    };
  }
}

function persistAuth(state: Pick<AuthState, "accessToken" | "refreshToken" | "user">) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(
    AUTH_STORAGE_KEY,
    JSON.stringify({
      accessToken: state.accessToken,
      refreshToken: state.refreshToken,
      user: state.user,
    }),
  );
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
    user: payload.user ?? null,
  };

  persistAuth(nextState);
  set(nextState);
}

const initialState = readStoredAuth();

export const useAuthStore = create<AuthState>((set, get) => ({
  ...initialState,
  login: async (payload) => {
    const res = await api.post<AuthResponse<AuthPayload>>("/api/v1/auth/login", payload);
    applyAuthPayload(set, unwrapApiResponse(res.data));
  },
  register: async (payload) => {
    const res = await api.post<AuthResponse<AuthPayload>>("/api/v1/auth/register", {
      email: payload.email,
      password: payload.password,
    });
    applyAuthPayload(set, unwrapApiResponse(res.data));
  },
  refresh: async () => {
    const refreshToken = get().refreshToken;
    if (!refreshToken) {
      throw new Error("Missing refresh token");
    }

    const res = await api.post<AuthResponse<AuthPayload>>(
      "/api/v1/auth/refresh",
      null,
      {
        skipAuthRefresh: true,
        useRefreshToken: true,
      },
    );
    applyAuthPayload(set, unwrapApiResponse(res.data));
  },
  logout: () => {
    clearStoredAuth();
    set({
      accessToken: null,
      refreshToken: null,
      user: null,
    });
  },
  setUser: (user) => {
    const current = get();
    const nextState = {
      accessToken: current.accessToken,
      refreshToken: current.refreshToken,
      user,
    };
    persistAuth(nextState);
    set({ user });
  },
}));

attachAuthStore(useAuthStore);
