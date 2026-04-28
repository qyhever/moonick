import { describe, expect, it } from "vitest";
import { devProxy } from "../../devProxy";

describe("vite dev proxy", () => {
  it("proxies api requests to backend server", () => {
    expect(devProxy).toMatchObject({
      "/api": {
        target: "http://127.0.0.1:6303",
        changeOrigin: true,
      },
    });
  });
});
