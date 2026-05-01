import { useEffect, useState } from "react";
import TripCard from "../components/TripCard";
import { getMyFavorites, type TripSummary } from "../api";

export default function MyFavoritesPage() {
  const [trips, setTrips] = useState<TripSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;

    async function loadFavorites() {
      setLoading(true);
      setError("");

      try {
        const favorites = await getMyFavorites();
        if (active) {
          setTrips(favorites.items);
        }
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "收藏列表加载失败");
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void loadFavorites();

    return () => {
      active = false;
    };
  }, []);

  return (
    <main className="h5-shell">
      <section className="page-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">我的收藏</p>
            <h1>收藏的顺路行程</h1>
          </div>
        </div>

        {loading ? <p className="subtle-text">正在加载...</p> : null}
        {error ? <p role="alert">{error}</p> : null}
        {!loading && !error && trips.length === 0 ? <p className="subtle-text">还没有收藏任何行程。</p> : null}

        <div className="trip-list">
          {trips.map((trip) => (
            <TripCard
              key={trip.id}
              disableLink={trip.unavailable}
              trip={trip}
              footer={
                trip.unavailable ? <p className="subtle-text">该行程已下线或不存在</p> : undefined
              }
            />
          ))}
        </div>
      </section>
    </main>
  );
}
