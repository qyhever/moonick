import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import TripCard from "../features/trips/components/TripCard";
import { getTrips, type TripSummary } from "../features/trips/api";
import { useAuthStore } from "../store/auth";

export default function HomePage() {
  const accessToken = useAuthStore((state) => state.accessToken);
  const currentUser = useAuthStore((state) => state.user);
  const [trips, setTrips] = useState<TripSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;

    async function loadTrips() {
      setLoading(true);
      setError("");

      try {
        const nextTrips = await getTrips();
        if (active) {
          setTrips(nextTrips.items);
        }
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "首页加载失败");
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void loadTrips();

    return () => {
      active = false;
    };
  }, []);

  return (
    <main className="h5-shell">
      <section className="nav-bar" aria-label="首页导航">
        <div className="nav-brand">
          明叶<span>同行</span>
        </div>
        <div className="nav-actions">
          <div className="nav-pill" aria-hidden="true">
            新
          </div>
          <Link
            className="nav-avatar"
            to={accessToken ? "/me/profile" : "/login?redirect=%2Fme%2Fprofile"}
          >
            {accessToken ? (currentUser?.nickname?.slice(0, 1) ?? "我") : "登录"}
          </Link>
        </div>
      </section>

      <section className="hero-card">
        <p className="eyebrow">Warm Ride Board</p>
        <h1 className="hero-card__title">顺路信息更清楚一点，联系和成交就会更快一点。</h1>
        <p className="hero-card__subtitle">
          首页先展示发布、管理和收藏入口，再承接最新行程流，统一跟随 `htmls/v1.html` 的暖色卡片风格。
        </p>
      </section>

      <section className="page-panel">
        <div className="section-header">
          <div>
            <h2 className="section-title">快捷入口</h2>
            <p className="section-subtitle">把常用动作压缩成四个主入口</p>
          </div>
        </div>

        <div className="quick-links">
          <Link className="quick-link" to="/publish">
            <span className="quick-link__icon quick-link__icon--publish" aria-hidden="true">
              发
            </span>
            <span className="quick-link__text">发布行程</span>
          </Link>
          <Link className="quick-link" to={accessToken ? "/me/trips" : "/login?redirect=%2Fme%2Ftrips"}>
            <span className="quick-link__icon quick-link__icon--trips" aria-hidden="true">
              程
            </span>
            <span className="quick-link__text">我的发布</span>
          </Link>
          <Link
            className="quick-link"
            to={accessToken ? "/me/favorites" : "/login?redirect=%2Fme%2Ffavorites"}
          >
            <span className="quick-link__icon quick-link__icon--favorites" aria-hidden="true">
              藏
            </span>
            <span className="quick-link__text">我的收藏</span>
          </Link>
          <Link
            className="quick-link"
            to={accessToken ? "/me/profile" : "/login?redirect=%2Fme%2Fprofile"}
          >
            <span className="quick-link__icon quick-link__icon--profile" aria-hidden="true">
              我
            </span>
            <span className="quick-link__text">{accessToken ? "个人中心" : "登录注册"}</span>
          </Link>
        </div>
      </section>

      <section className="page-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">Latest Trips</p>
            <h1>附近最新顺路信息</h1>
          </div>
        </div>

        {loading ? <p className="subtle-text">正在加载首页行程...</p> : null}
        {error ? <p role="alert">{error}</p> : null}
        {!loading && !error && trips.length === 0 ? <p className="subtle-text">当前还没有可展示的行程。</p> : null}

        <div className="trip-list">
          {trips.map((trip) => (
            <TripCard key={trip.id} disableLink={trip.status === "full"} trip={trip} />
          ))}
        </div>
      </section>
    </main>
  );
}
