import { Heart, House, SquarePen, UserRound } from "lucide-react";
import { NavLink, Outlet, useLocation } from "react-router-dom";

import { useAuthStore } from "../store/auth";

type TabItem = {
  label: string;
  to: string;
  icon: typeof House;
  isActive: (pathname: string) => boolean;
};

const HIDDEN_TABBAR_PREFIXES = ["/login", "/register", "/trips/"];
const HIDDEN_TABBAR_SUFFIXES = ["/edit"];

export default function AppLayout() {
  const location = useLocation();

  const shouldHideTabBar =
    HIDDEN_TABBAR_PREFIXES.some((prefix) => location.pathname.startsWith(prefix)) ||
    HIDDEN_TABBAR_SUFFIXES.some((suffix) => location.pathname.endsWith(suffix));

  return (
    <>
      <Outlet />
      {shouldHideTabBar ? null : <MobileTabBar />}
    </>
  );
}

function MobileTabBar() {
  const accessToken = useAuthStore((state) => state.accessToken);
  const pathname = useLocation().pathname;

  const tabs: TabItem[] = [
    {
      label: "首页",
      to: "/",
      icon: House,
      isActive: (currentPath) => currentPath === "/",
    },
    {
      label: "发布",
      to: "/publish",
      icon: SquarePen,
      isActive: (currentPath) => currentPath === "/publish",
    },
    {
      label: "收藏",
      to: accessToken ? "/me/favorites" : "/login?redirect=%2Fme%2Ffavorites",
      icon: Heart,
      isActive: (currentPath) => currentPath.startsWith("/me/favorites"),
    },
    {
      label: "我的",
      to: accessToken ? "/me/profile" : "/login?redirect=%2Fme%2Fprofile",
      icon: UserRound,
      isActive: (currentPath) =>
        currentPath.startsWith("/me/") && !currentPath.startsWith("/me/favorites"),
    },
  ];

  return (
    <nav className="tab-bar" aria-label="底部导航">
      {tabs.map((tab) => {
        const Icon = tab.icon;
        const isActive = tab.isActive(pathname);

        return (
          <NavLink
            key={tab.label}
            aria-current={isActive ? "page" : undefined}
            className={`tab-item${isActive ? " active" : ""}`}
            to={tab.to}
          >
            <span className="tab-item__icon" aria-hidden="true">
              <Icon size={20} strokeWidth={2} />
            </span>
            <span>{tab.label}</span>
          </NavLink>
        );
      })}
    </nav>
  );
}
