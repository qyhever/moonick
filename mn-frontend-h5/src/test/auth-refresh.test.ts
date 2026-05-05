import "@testing-library/jest-dom/vitest";
import type { AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import { beforeEach, expect, it } from "vitest";

import { api } from "../lib/http";
import { useAuthStore } from "../store/auth";

type MockResponse = AxiosResponse<unknown>;

function writeAuth(accessToken: string | null, refreshToken: string | null) {
  const payload = {
    accessToken,
    refreshToken,
    user: null,
  };

  if (accessToken || refreshToken) {
    window.localStorage.setItem("mn-h5-auth", JSON.stringify(payload));
  } else {
    window.localStorage.removeItem("mn-h5-auth");
  }

  useAuthStore.setState({
    accessToken,
    refreshToken,
    user: null,
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

it("refreshes login state once and retries the original request", async () => {
  writeAuth("expired-access-token", "refresh-token");

  let protectedCalls = 0;
  let refreshCalls = 0;

  api.defaults.adapter = async (config: InternalAxiosRequestConfig): Promise<MockResponse> => {
    if (config.url === "/v1/protected") {
      protectedCalls += 1;

      if (protectedCalls === 1) {
        expect(getAuthorizationHeader(config)).toBe("Bearer expired-access-token");
        return buildResponse(config, {
          code: 1007,
          message: "无效的token",
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

    if (config.url === "/v1/auth/refresh") {
      refreshCalls += 1;
      expect(getAuthorizationHeader(config)).toBe("Bearer refresh-token");

      return buildResponse(config, {
        code: 1000,
        message: "success",
        data: {
          accessToken: "refreshed-access-token",
          refreshToken: "refreshed-refresh-token",
          user: {
            id: 1,
            email: "user@example.com",
            phone: "13800138000",
            nickname: "测试用户",
            avatarUrl: "",
            status: "active",
            defaultWechat: "",
            defaultPhone: "13800138000",
          },
        },
      });
    }

    throw new Error(`Unexpected URL: ${config.url}`);
  };

  const response = await api.get("/v1/protected");

  expect(response.data.data).toEqual({ ok: true });
  expect(protectedCalls).toBe(2);
  expect(refreshCalls).toBe(1);
  expect(useAuthStore.getState().accessToken).toBe("refreshed-access-token");
  expect(useAuthStore.getState().refreshToken).toBe("refreshed-refresh-token");
});

it("logs out when refresh also fails", async () => {
  writeAuth("expired-access-token", "refresh-token");

  api.defaults.adapter = async (config: InternalAxiosRequestConfig): Promise<MockResponse> => {
    if (config.url === "/v1/protected") {
      return buildResponse(config, {
        code: 1007,
        message: "无效的token",
        data: null,
      });
    }

    if (config.url === "/v1/auth/refresh") {
      return buildResponse(config, {
        code: 1006,
        message: "需要登录",
        data: null,
      });
    }

    throw new Error(`Unexpected URL: ${config.url}`);
  };

  await expect(api.get("/v1/protected")).rejects.toThrow();
  expect(useAuthStore.getState().accessToken).toBeNull();
  expect(useAuthStore.getState().refreshToken).toBeNull();
  expect(window.localStorage.getItem("mn-h5-auth")).toBeNull();
});
