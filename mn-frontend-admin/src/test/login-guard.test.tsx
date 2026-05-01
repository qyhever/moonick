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
