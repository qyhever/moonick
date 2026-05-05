import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import UserAvatar from "../features/profile/components/UserAvatar";

it("uses default avatar asset when src is empty", () => {
  render(
    <UserAvatar
      alt="当前头像"
      defaultSrc="/image-default.svg"
      fallback={<img alt="头像加载失败" src="/image-fail.svg" />}
      src=""
    />,
  );

  expect(screen.getByAltText("当前头像")).toHaveAttribute("src", "/image-default.svg");
});

it("shows fallback content when avatar image fails to load", async () => {
  render(<UserAvatar alt="当前头像" fallback="测" src="https://cdn.example.com/avatar.png" />);

  fireEvent.error(screen.getByAltText("当前头像"));

  expect(await screen.findByText("测")).toBeInTheDocument();
});

it("does not render a fallback asset path as plain text", async () => {
  render(
    <UserAvatar
      alt="当前头像"
      fallback={<img alt="头像加载失败" src="/image-fail.svg" />}
      src="https://cdn.example.com/avatar.png"
    />,
  );

  fireEvent.error(screen.getByAltText("当前头像"));

  expect(await screen.findByAltText("头像加载失败")).toHaveAttribute("src", "/image-fail.svg");
  expect(screen.queryByText("/image-fail.svg")).not.toBeInTheDocument();
});
