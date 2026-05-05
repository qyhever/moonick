import { describe, expect, it } from "vitest";
import { devProxy } from "../../devProxy";

describe("vite dev proxy", () => {
  it("proxies api requests to backend server", () => {
    expect(devProxy).toMatchObject({
      "/mmoonick/api": {
        target: "http://127.0.0.1:6303",
        changeOrigin: true,
      },
    });
  });

  it("rewrites the prefixed path back to backend api path", () => {
    const proxy = devProxy["/mmoonick/api"];
    expect(proxy.rewrite?.("/mmoonick/api/v1/trips")).toBe("/api/v1/trips");
  });
});
