import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import TripCard from "../features/trips/components/TripCard";
import { getTrips, type TripSummary } from "../features/trips/api";
import { useAuthStore } from "../store/auth";

export default function HomePage() {
  const accessToken = useAuthStore((state) => state.accessToken);
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
      <section className="hero-card">
        <p className="eyebrow">明叶同行</p>
        <h1 className="hero-card__title">诚信拼车，轻量约车，适合手机端快速成交</h1>
        <p className="hero-card__subtitle">白底卡片流配合深绿主色，首轮围绕发布、收藏、个人中心做最小闭环。</p>
      </section>

      <section className="page-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">快捷入口</p>
            <h1>今天想做什么</h1>
          </div>
        </div>

        <div className="quick-links">
          <Link className="quick-link" to="/publish">
            <strong>发布新行程</strong>
            <span>→</span>
          </Link>
          <Link className="quick-link" to={accessToken ? "/me/trips" : "/login?redirect=%2Fme%2Ftrips"}>
            <strong>我的发布</strong>
            <span>→</span>
          </Link>
          <Link
            className="quick-link"
            to={accessToken ? "/me/favorites" : "/login?redirect=%2Fme%2Ffavorites"}
          >
            <strong>我的收藏</strong>
            <span>→</span>
          </Link>
          <Link
            className="quick-link"
            to={accessToken ? "/me/profile" : "/login?redirect=%2Fme%2Fprofile"}
          >
            <strong>{accessToken ? "个人中心" : "登录 / 注册"}</strong>
            <span>→</span>
          </Link>
        </div>
      </section>

      <section className="page-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">推荐行程</p>
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
