import "@testing-library/jest-dom/vitest";
import type { AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import { beforeEach, expect, it } from "vitest";

import { api } from "../lib/http";
import { useAdminAuthStore } from "../features/auth/store";

type MockResponse = AxiosResponse<unknown>;

function writeAuth(accessToken: string | null, refreshToken: string | null) {
  const payload = {
    accessToken,
    refreshToken,
    admin: null,
  };

  if (accessToken || refreshToken) {
    window.localStorage.setItem("mn-admin-auth", JSON.stringify(payload));
  } else {
    window.localStorage.removeItem("mn-admin-auth");
  }

  useAdminAuthStore.setState({
    accessToken,
    refreshToken,
    admin: null,
  });
}

function getAuthorizationHeader(config: AxiosRequestConfig) {
  const headers = config.headers as
    | {
        Authorization?: string;
      }
    | undefined;

  return headers?.Authorization ?? null;
}

function buildResponse(config: InternalAxiosRequestConfig, data: unknown): MockResponse {
  return {
    status: 200,
    statusText: "OK",
    config,
    headers: {},
    data,
  };
}

beforeEach(() => {
  window.localStorage.clear();
  writeAuth(null, null);
  api.defaults.adapter = undefined;
});

it("refreshes admin auth once and retries the original request when code is 1007", async () => {
  writeAuth("expired-access-token", "refresh-token");

  let protectedCalls = 0;
  let refreshCalls = 0;

  api.defaults.adapter = async (config: InternalAxiosRequestConfig): Promise<MockResponse> => {
    if (config.url === "/api/admin/v1/protected") {
      protectedCalls += 1;

      if (protectedCalls === 1) {
        expect(getAuthorizationHeader(config)).toBe("Bearer expired-access-token");
        return buildResponse(config, {
          code: 1007,
          message: "无效 token",
          data: null,
        });
      }

      expect(getAuthorizationHeader(config)).toBe("Bearer refreshed-access-token");
      return buildResponse(config, {
        code: 1000,
        message: "success",
        data: {
          ok: true,
        },
      });
    }

    if (config.url === "/api/admin/v1/auth/refresh") {
      refreshCalls += 1;
      expect(getAuthorizationHeader(config)).toBe("Bearer refresh-token");

      return buildResponse(config, {
        code: 1000,
        message: "success",
        data: {
          accessToken: "refreshed-access-token",
          refreshToken: "refreshed-refresh-token",
          admin: {
            id: 1,
            username: "admin",
            name: "管理员",
            status: "active",
          },
        },
      });
    }

    throw new Error(`Unexpected URL: ${config.url}`);
  };

  const response = await api.get("/api/admin/v1/protected");

  expect(response.data.data).toEqual({ ok: true });
  expect(protectedCalls).toBe(2);
  expect(refreshCalls).toBe(1);
  expect(useAdminAuthStore.getState().accessToken).toBe("refreshed-access-token");
  expect(useAdminAuthStore.getState().refreshToken).toBe("refreshed-refresh-token");
});

it("logs out admin user when refresh also fails", async () => {
  writeAuth("expired-access-token", "refresh-token");

  api.defaults.adapter = async (config: InternalAxiosRequestConfig): Promise<MockResponse> => {
    if (config.url === "/api/admin/v1/protected") {
      return buildResponse(config, {
        code: 1007,
        message: "无效 token",
        data: null,
      });
    }

    if (config.url === "/api/admin/v1/auth/refresh") {
      return buildResponse(config, {
        code: 1007,
        message: "refresh token 无效",
        data: null,
      });
    }

    throw new Error(`Unexpected URL: ${config.url}`);
  };

  await expect(api.get("/api/admin/v1/protected")).rejects.toThrow();
  expect(useAdminAuthStore.getState().accessToken).toBeNull();
  expect(useAdminAuthStore.getState().refreshToken).toBeNull();
  expect(window.localStorage.getItem("mn-admin-auth")).toBeNull();
});
