import { beforeEach, expect, it, vi } from "vitest";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("../lib/http", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../lib/http")>();

  return {
    ...actual,
    api: {
      post: mockPost,
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    },
  };
});

import { uploadAvatar } from "../features/trips/api";

beforeEach(() => {
  mockPost.mockReset();
});

it("uploads avatar through the users route without overriding multipart headers", async () => {
  mockPost.mockResolvedValue({
    data: {
      code: 1000,
      message: "success",
      data: {
        url: "https://cdn.example.com/avatar.png",
      },
    },
  });

  const file = new File(["avatar"], "avatar.png", { type: "image/png" });

  await expect(uploadAvatar(file)).resolves.toBe("https://cdn.example.com/avatar.png");
  expect(mockPost).toHaveBeenCalledTimes(1);

  const [url, formData, config] = mockPost.mock.calls[0];
  expect(url).toBe("/v1/users/avatar");
  expect(formData).toBeInstanceOf(FormData);
  expect((formData as FormData).get("file")).toBe(file);
  expect(config).toBeUndefined();
});
