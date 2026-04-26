import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import TripCard from "../components/TripCard";
import { getMyTrips, updateTripStatus, type TripSummary } from "../api";

export default function MyTripsPage() {
  const [trips, setTrips] = useState<TripSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  useEffect(() => {
    let active = true;

    async function loadTrips() {
      setLoading(true);
      setError("");

      try {
        const nextTrips = await getMyTrips();
        if (active) {
          setTrips(nextTrips.items);
        }
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "我的发布加载失败");
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

  async function handleStatusChange(tripId: number, status: "full" | "closed") {
    setUpdatingId(tripId);
    setError("");

    try {
      const updated = await updateTripStatus(tripId, status);
      setTrips((current) => current.map((item) => (item.id === tripId ? updated : item)));
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "状态更新失败");
    } finally {
      setUpdatingId(null);
    }
  }

  return (
    <main className="h5-shell">
      <section className="page-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">我的发布</p>
            <h1>已发布行程</h1>
          </div>
          <Link className="secondary-link" to="/publish">
            新建
          </Link>
        </div>

        {loading ? <p className="subtle-text">正在加载...</p> : null}
        {error ? <p role="alert">{error}</p> : null}
        {!loading && !error && trips.length === 0 ? <p className="subtle-text">还没有发布任何行程。</p> : null}

        <div className="trip-list">
          {trips.map((trip) => (
            <TripCard
              key={trip.id}
              trip={trip}
              footer={
                <div className="inline-actions">
                  <Link className="secondary-link" to={`/trips/${trip.id}/edit`}>
                    编辑
                  </Link>
                  {trip.status === "active" ? (
                    <>
                      <button
                        className="secondary-link secondary-link--button"
                        disabled={updatingId === trip.id}
                        onClick={() => void handleStatusChange(trip.id, "full")}
                        type="button"
                      >
                        设为满员
                      </button>
                      <button
                        className="secondary-link secondary-link--button"
                        disabled={updatingId === trip.id}
                        onClick={() => void handleStatusChange(trip.id, "closed")}
                        type="button"
                      >
                        关闭
                      </button>
                    </>
                  ) : null}
                </div>
              }
            />
          ))}
        </div>
      </section>
    </main>
  );
}
