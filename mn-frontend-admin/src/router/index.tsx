import type { ReactNode } from "react";
import { Navigate, RouterProvider, createBrowserRouter, type RouteObject, useLocation } from "react-router-dom";
import { Card } from "antd";

import LoginPage from "../features/auth/LoginPage";
import DashboardPage from "../features/dashboard/DashboardPage";
import { useAdminAuthStore } from "../features/auth/store";
import TripDetailPage from "../features/trips/TripDetailPage";
import TripEditPage from "../features/trips/TripEditPage";
import TripListPage from "../features/trips/TripListPage";
import UserDetailPage from "../features/users/UserDetailPage";
import UserListPage from "../features/users/UserListPage";
import AdminLayout from "../layout/AdminLayout";

function RequireAdminAuth({ children }: { children: ReactNode }) {
  const accessToken = useAdminAuthStore((state) => state.accessToken);
  const location = useLocation();

  if (!accessToken) {
    const redirect = `${location.pathname}${location.search}`;
    return <Navigate replace to={`/login?redirect=${encodeURIComponent(redirect)}`} />;
  }

  return <>{children}</>;
}

export const routes: RouteObject[] = [
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: (
      <RequireAdminAuth>
        <AdminLayout />
      </RequireAdminAuth>
    ),
    children: [
      { index: true, element: <Navigate replace to="/dashboard" /> },
      { path: "dashboard", element: <DashboardPage /> },
      { path: "trips", element: <TripListPage /> },
      { path: "trips/:id", element: <TripDetailPage /> },
      { path: "trips/:id/edit", element: <TripEditPage /> },
      { path: "users", element: <UserListPage /> },
      { path: "users/:id", element: <UserDetailPage /> },
    ],
  },
];

const appRouter = createBrowserRouter(routes, {
  basename: import.meta.env.BASE_URL,
});

export function AppRouter() {
  return (
    <RouterProvider
      router={appRouter}
      future={{
        v7_startTransition: true,
      }}
    />
  )
}

export default appRouter;
