import { useEffect, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";

import Toast from "../../../components/Toast";
import { getTripDetail, toggleFavorite, updateTripStatus, type TripDetail } from "../api";
import { useAuthStore } from "../../../store/auth";

const statusLabelMap: Record<TripDetail["status"], string> = {
  active: "可约",
  full: "已满",
  closed: "已关闭",
  expired: "已过期",
};

const tripTypeLabelMap: Record<string, string> = {
  driver_post: "车找人",
  passenger_post: "人找车",
};

function formatDeparture(date: string, time: string) {
  const departure = new Date(`${date}T${time}:00`);

  if (Number.isNaN(departure.getTime())) {
    return `${date} ${time}`;
  }

  return departure.toLocaleString("zh-CN", {
    month: "long",
    day: "numeric",
    weekday: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function TripDetailPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const routeToast = (location.state as { toast?: string } | null)?.toast ?? "";
  const accessToken = useAuthStore((state) => state.accessToken);
  const currentUser = useAuthStore((state) => state.user);
  const [trip, setTrip] = useState<TripDetail | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [toast, setToast] = useState("");
  const isOwner = !!trip && currentUser?.id === trip.userId;

  useEffect(() => {
    if (!toast) {
      return undefined;
    }

    const timer = window.setTimeout(() => {
      setToast("");
    }, 2000);

    return () => {
      window.clearTimeout(timer);
    };
  }, [toast]);

  useEffect(() => {
    let active = true;

    async function loadTrip() {
      setLoading(true);
      setError("");

      try {
        const detail = await getTripDetail(id);
        if (active) {
          setTrip(detail);
        }
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "详情加载失败");
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void loadTrip();

    return () => {
      active = false;
    };
  }, [id]);

  useEffect(() => {
    if (!routeToast) {
      return;
    }

    setToast(routeToast);
    navigate(`${location.pathname}${location.search}`, { replace: true, state: null });
  }, [location.pathname, location.search, navigate, routeToast]);

  async function handleToggleFavorite() {
    if (!trip || trip.status !== "active" || submitting) {
      return;
    }

    if (!accessToken) {
      const redirect = `${location.pathname}${location.search}`;
      navigate(`/login?redirect=${encodeURIComponent(redirect)}`);
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      const result = await toggleFavorite(trip.id);
      setTrip((current) => (current ? { ...current, favorited: result.favorited } : current));
      setToast("收藏成功");
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "收藏失败，请稍后再试");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleCloseTrip() {
    if (!trip || !isOwner || submitting) {
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      const updated = await updateTripStatus(trip.id, "closed");
      setTrip(updated);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "关闭失败，请稍后再试");
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return (
      <main className="h5-shell">
        <section className="page-panel">
          <p className="subtle-text">正在载入行程详情...</p>
        </section>
      </main>
    );
  }

  if (!trip) {
    return (
      <main className="h5-shell">
        <section className="page-panel">
          <h1>行程详情</h1>
          <p role="alert">{error || "未找到该行程"}</p>
        </section>
      </main>
    );
  }

  return (
    <main className="h5-shell">
      {toast ? <Toast message={toast} /> : null}

      <section className="hero-card hero-card--compact">
        <p className="eyebrow">行程详情</p>
        <h1 className="hero-card__title">
          {trip.fromText}
          <span> → </span>
          {trip.toText}
        </h1>
        <p className="hero-card__subtitle">{formatDeparture(trip.departureDate, trip.departureTime)}</p>
        <div className="detail-chip-row">
          <span className={`status-pill status-pill--${trip.status}`}>{statusLabelMap[trip.status]}</span>
          <span className="detail-chip">{tripTypeLabelMap[trip.tripType] ?? trip.tripType}</span>
          <span className="detail-chip">{trip.seatCount} 人</span>
          <span className="detail-chip">{trip.isPriceNegotiable ? "费用面议" : "费用未标注"}</span>
        </div>
      </section>

      <section className="page-panel">
        <div className="detail-section">
          <h2>联系信息</h2>
          <dl className="detail-list">
            <div>
              <dt>手机号</dt>
              <dd>{trip.contactPhone || "未填写"}</dd>
            </div>
            <div>
              <dt>微信号</dt>
              <dd>{trip.contactWechat || "未填写"}</dd>
            </div>
          </dl>
        </div>

        {error ? <p role="alert">{error}</p> : null}

        <div className="action-row">
          {isOwner ? (
            <>
              <Link className="secondary-link" to={`/trips/${trip.id}/edit`}>
                编辑
              </Link>
              <button
                type="button"
                className="primary-button primary-button--ghost"
                disabled={trip.status === "closed" || trip.status === "expired" || submitting}
                onClick={handleCloseTrip}
              >
                关闭行程
              </button>
            </>
          ) : (
            <>
              <button
                type="button"
                className="primary-button primary-button--ghost"
                data-favorited={trip.favorited}
                disabled={trip.status !== "active" || submitting}
                onClick={handleToggleFavorite}
              >
                收藏
              </button>
              {trip.contactPhone ? (
                <a className="secondary-link" href={`tel:${trip.contactPhone}`}>
                  联系 TA
                </a>
              ) : (
                <span className="secondary-link">微信联系</span>
              )}
            </>
          )}
        </div>
      </section>
    </main>
  );
}
