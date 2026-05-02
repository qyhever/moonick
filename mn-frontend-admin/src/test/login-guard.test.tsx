import { render, screen } from "@testing-library/react";
import { RouterProvider, createMemoryRouter } from "react-router-dom";
import { beforeEach } from "vitest";

import { useAdminAuthStore } from "../features/auth/store";
import { routes } from "../router";

function renderWithRouter(initialEntry: string) {
  const router = createMemoryRouter(routes, {
    initialEntries: [initialEntry],
  });

  return render(
    <RouterProvider
      router={router}
      future={{
        v7_startTransition: true,
      }}
    />
  );
}

beforeEach(() => {
  window.localStorage.clear();
  useAdminAuthStore.setState({
    accessToken: null,
    refreshToken: null,
    admin: null,
  });
});

it("redirects anonymous admin user to login", async () => {
  renderWithRouter("/dashboard");
  expect(await screen.findByRole("button", { name: /登\s*录/ })).toBeInTheDocument();
});

it("renders fixed admin shell layout styles for authenticated routes", async () => {
  useAdminAuthStore.setState({
    accessToken: "token",
    refreshToken: "refresh",
    admin: {
      id: 1,
      username: "admin",
      name: "管理员",
      status: "active",
    },
  });

  renderWithRouter("/dashboard");

  const sider = await screen.findByRole("complementary");
  const header = screen.getByRole("banner");
  const content = screen.getByTestId("admin-layout-content");

  expect(sider).toHaveStyle({
    position: "fixed",
    top: "0px",
    left: "0px",
    bottom: "0px",
  });
  expect(header).toHaveStyle({
    position: "fixed",
    top: "0px",
    left: "232px",
    right: "0px",
  });
  expect(content).toHaveStyle({
    marginTop: "64px",
    height: "calc(100vh - 64px)",
    overflowY: "auto",
  });
});
