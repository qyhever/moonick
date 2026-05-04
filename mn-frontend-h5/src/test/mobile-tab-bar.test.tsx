import "@testing-library/jest-dom/vitest";
import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, vi } from "vitest";

import AppLayout from "../components/MobileTabBar";
import { useAuthStore } from "../store/auth";

beforeEach(() => {
  window.localStorage.clear();
  window.scrollTo = vi.fn();
  useAuthStore.setState({
    accessToken: null,
    refreshToken: null,
    user: null,
  });
});

it("does not show favorites tab in bottom navigation", () => {
  render(
    <MemoryRouter initialEntries={["/"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<div>首页内容</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  expect(screen.getByRole("navigation", { name: "底部导航" })).toBeInTheDocument();
  expect(screen.queryByRole("link", { name: "收藏" })).not.toBeInTheDocument();
});

it("shows tab bar only on home and profile pages", () => {
  render(
    <MemoryRouter initialEntries={["/publish"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/publish" element={<div>发布页</div>} />
          <Route path="/me/profile" element={<div>我的页面</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  expect(screen.queryByRole("navigation", { name: "底部导航" })).not.toBeInTheDocument();
  expect(screen.getByRole("button", { name: "返回上一页" })).toBeInTheDocument();
  expect(screen.getByText("发布行程")).toBeInTheDocument();

  cleanup();

  render(
    <MemoryRouter initialEntries={["/me/profile"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/publish" element={<div>发布页</div>} />
          <Route path="/me/profile" element={<div>我的页面</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  expect(screen.getByRole("navigation", { name: "底部导航" })).toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "返回上一页" })).not.toBeInTheDocument();
});

it("resets window scroll to top on route changes handled by app layout", async () => {
  const scrollToSpy = vi.mocked(window.scrollTo);

  render(
    <MemoryRouter initialEntries={["/"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<div>首页内容</div>} />
          <Route path="/trips/:id" element={<div>详情页</div>} />
          <Route path="/publish" element={<div>发布页</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  expect(scrollToSpy).toHaveBeenCalledTimes(1);
  expect(scrollToSpy).toHaveBeenLastCalledWith({ top: 0, left: 0, behavior: "auto" });

  cleanup();

  render(
    <MemoryRouter initialEntries={["/publish"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<div>首页内容</div>} />
          <Route path="/publish" element={<div>发布页</div>} />
          <Route path="/trips/:id" element={<div>详情页</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  expect(scrollToSpy).toHaveBeenCalledTimes(2);
  expect(scrollToSpy).toHaveBeenLastCalledWith({ top: 0, left: 0, behavior: "auto" });

  cleanup();

  render(
    <MemoryRouter initialEntries={["/trips/7"]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/" element={<div>首页内容</div>} />
          <Route path="/publish" element={<div>发布页</div>} />
          <Route path="/trips/:id" element={<div>详情页</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );

  expect(scrollToSpy).toHaveBeenCalledTimes(3);
  expect(scrollToSpy).toHaveBeenLastCalledWith({ top: 0, left: 0, behavior: "auto" });
});
