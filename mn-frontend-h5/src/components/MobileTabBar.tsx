import { useEffect } from "react";
import { ChevronLeft, House, Plus, UserRound } from "lucide-react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";

import { useAuthStore } from "../store/auth";

type TabItem = {
  label: string;
  to: string;
  icon: typeof House;
  isActive: (pathname: string) => boolean;
};

type RouteFrame = {
  title: string;
  showTopBar: boolean;
  showTabBar: boolean;
  backFallback: string;
};

function getRouteFrame(pathname: string): RouteFrame {
  if (pathname === "/") {
    return {
      title: "",
      showTopBar: false,
      showTabBar: true,
      backFallback: "/",
    };
  }

  if (pathname === "/me/profile") {
    return {
      title: "",
      showTopBar: false,
      showTabBar: true,
      backFallback: "/",
    };
  }

  if (pathname === "/publish") {
    return {
      title: "发布行程",
      showTopBar: true,
      showTabBar: false,
      backFallback: "/",
    };
  }

  if (pathname === "/login") {
    return {
      title: "登录",
      showTopBar: false,
      showTabBar: false,
      backFallback: "/",
    };
  }

  if (pathname === "/register") {
    return {
      title: "注册",
      showTopBar: false,
      showTabBar: false,
      backFallback: "/login",
    };
  }

  if (pathname === "/me/trips") {
    return {
      title: "我的发布",
      showTopBar: true,
      showTabBar: false,
      backFallback: "/me/profile",
    };
  }

  if (pathname === "/me/favorites") {
    return {
      title: "我的收藏",
      showTopBar: true,
      showTabBar: false,
      backFallback: "/me/profile",
    };
  }

  if (pathname === "/me/settings") {
    return {
      title: "账户设置",
      showTopBar: true,
      showTabBar: false,
      backFallback: "/me/profile",
    };
  }

  if (pathname.endsWith("/edit")) {
    return {
      title: "编辑行程",
      showTopBar: true,
      showTabBar: false,
      backFallback: "/me/trips",
    };
  }

  if (pathname.startsWith("/trips/")) {
    return {
      title: "行程详情",
      showTopBar: true,
      showTabBar: false,
      backFallback: "/",
    };
  }

  return {
    title: "返回",
    showTopBar: true,
    showTabBar: false,
    backFallback: "/",
  };
}

export default function AppLayout() {
  const location = useLocation();
  const frame = getRouteFrame(location.pathname);
  const shouldShowTopBar = frame.showTopBar;

  return (
    <>
      <ScrollToTopOnRouteChange />
      {shouldShowTopBar ? <PageTopBar backFallback={frame.backFallback} title={frame.title} /> : null}
      <div className={shouldShowTopBar ? "app-route app-route--with-topbar" : "app-route"}>
        <Outlet />
      </div>
      {frame.showTabBar ? <MobileTabBar /> : null}
    </>
  );
}

function ScrollToTopOnRouteChange() {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: "auto" });
  }, [pathname]);

  return null;
}

function PageTopBar({ title, backFallback }: { title: string; backFallback: string }) {
  const navigate = useNavigate();

  function handleBack() {
    if (window.history.length > 1) {
      navigate(-1);
      return;
    }

    navigate(backFallback, { replace: true });
  }

  return (
    <header className="page-topbar-shell">
      <div className="page-topbar">
        <button
          aria-label="返回上一页"
          className="page-topbar__back"
          onClick={handleBack}
          type="button"
        >
          <ChevronLeft size={18} strokeWidth={2.4} />
        </button>
        <strong className="page-topbar__title">{title}</strong>
        <span aria-hidden="true" className="page-topbar__spacer" />
      </div>
    </header>
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
      icon: Plus,
      isActive: (currentPath) => currentPath === "/publish",
    },
    {
      label: "我的",
      to: accessToken ? "/me/profile" : "/login?redirect=%2Fme%2Fprofile",
      icon: UserRound,
      isActive: (currentPath) => currentPath.startsWith("/me/"),
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
