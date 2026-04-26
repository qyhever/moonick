import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, vi } from "vitest";

const { mockUploadAvatar } = vi.hoisted(() => ({
  mockUploadAvatar: vi.fn(),
}));

vi.mock("../features/trips/api", () => ({
  uploadAvatar: mockUploadAvatar,
}));

import AvatarUploader from "../features/profile/components/AvatarUploader";

function makeFile(name: string) {
  return new File(["avatar"], name, { type: "image/png" });
}

beforeEach(() => {
  mockUploadAvatar.mockReset();
  mockUploadAvatar.mockRejectedValue(new Error("upload failed"));
  Object.defineProperty(URL, "createObjectURL", {
    writable: true,
    value: vi.fn(() => "blob:preview"),
  });
  Object.defineProperty(URL, "revokeObjectURL", {
    writable: true,
    value: vi.fn(),
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

it("restores previous avatar when upload fails", async () => {
  render(<AvatarUploader initialUrl="https://cdn.example.com/a.png" />);

  await userEvent.upload(screen.getByLabelText("上传头像"), makeFile("b.png"));

  expect(mockUploadAvatar).toHaveBeenCalledTimes(1);
  expect(await screen.findByAltText("当前头像")).toHaveAttribute(
    "src",
    "https://cdn.example.com/a.png",
  );
});
