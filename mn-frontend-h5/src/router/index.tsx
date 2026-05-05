import type { ReactNode } from "react";
import { Navigate, type RouteObject, createBrowserRouter, useLocation } from "react-router-dom";

import AppLayout from "../components/MobileTabBar";
import AccountSettingsPage from "../features/profile/pages/AccountSettingsPage";
import PasswordResetPage from "../features/profile/pages/PasswordResetPage";
import ProfilePage from "../features/profile/pages/ProfilePage";
import EditTripPage from "../features/trips/pages/EditTripPage";
import MyFavoritesPage from "../features/trips/pages/MyFavoritesPage";
import MyTripsPage from "../features/trips/pages/MyTripsPage";
import PublishPage from "../features/trips/pages/PublishPage";
import TripDetailPage from "../features/trips/pages/TripDetailPage";
import HomePage from "../pages/HomePage";
import LoginPage from "../pages/LoginPage";
import RegisterPage from "../pages/RegisterPage";
import { useAuthStore } from "../store/auth";

function RequireAuth({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((state) => state.accessToken);
  const location = useLocation();

  if (!accessToken) {
    const redirect = `${location.pathname}${location.search}`;
    return <Navigate replace to={`/login?redirect=${encodeURIComponent(redirect)}`} />;
  }

  return <>{children}</>;
}

function RequireGuest({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((state) => state.accessToken);

  if (accessToken) {
    return <Navigate replace to="/" />;
  }

  return <>{children}</>;
}

export const routes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { path: "/", element: <HomePage /> },
      {
        path: "/login",
        element: (
          <RequireGuest>
            <LoginPage />
          </RequireGuest>
        ),
      },
      {
        path: "/register",
        element: (
          <RequireGuest>
            <RegisterPage />
          </RequireGuest>
        ),
      },
      { path: "/trips/:id", element: <TripDetailPage /> },
      {
        path: "/publish",
        element: (
          <RequireAuth>
            <PublishPage />
          </RequireAuth>
        ),
      },
      {
        path: "/trips/:id/edit",
        element: (
          <RequireAuth>
            <EditTripPage />
          </RequireAuth>
        ),
      },
      {
        path: "/me/trips",
        element: (
          <RequireAuth>
            <MyTripsPage />
          </RequireAuth>
        ),
      },
      {
        path: "/me/favorites",
        element: (
          <RequireAuth>
            <MyFavoritesPage />
          </RequireAuth>
        ),
      },
      {
        path: "/me/profile",
        element: (
          <RequireAuth>
            <ProfilePage />
          </RequireAuth>
        ),
      },
      {
        path: "/me/settings",
        element: (
          <RequireAuth>
            <AccountSettingsPage />
          </RequireAuth>
        ),
      },
      {
        path: "/me/settings/password-reset",
        element: (
          <RequireAuth>
            <PasswordResetPage />
          </RequireAuth>
        ),
      },
    ],
  },
];

export const appRouter = createBrowserRouter(routes);
