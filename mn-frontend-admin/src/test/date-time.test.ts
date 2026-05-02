import { describe, expect, it } from "vitest";

import { formatDateTime } from "../lib/dateTime";

describe("formatDateTime", () => {
  it("returns fallback for undefined input", () => {
    expect(formatDateTime(undefined)).toBe("-");
  });

  it("formats iso date time string", () => {
    expect(formatDateTime("2026-05-02T09:45:03+08:00")).toBe("2026-05-02 09:45:03");
  });
});
