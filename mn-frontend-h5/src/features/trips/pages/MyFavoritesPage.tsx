import { useEffect, useRef, useState } from "react";
import InfiniteScroll from "react-infinite-scroll-component";
import TripCard from "../components/TripCard";
import { getMyFavorites, type TripSummary } from "../api";

const PAGE_SIZE = 10;

export default function MyFavoritesPage() {
  const [trips, setTrips] = useState<TripSummary[]>([]);
  const [pageNum, setPageNum] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");
  const pageNumRef = useRef(1);
  const tripsLengthRef = useRef(0);

  useEffect(() => {
    pageNumRef.current = pageNum;
  }, [pageNum]);

  useEffect(() => {
    tripsLengthRef.current = trips.length;
  }, [trips.length]);

  useEffect(() => {
    let active = true;

    async function loadFavorites() {
      setLoading(true);
      setError("");

      try {
        const favorites = await getMyFavorites({
          pageNum: 1,
          pageSize: PAGE_SIZE,
        });
        if (active) {
          setTrips(favorites.items);
          setPageNum(favorites.pageNum);
          setHasMore(favorites.items.length < favorites.total);
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

  async function loadNextPage() {
    if (loading || loadingMore || refreshing || !hasMore) {
      return;
    }

    setLoadingMore(true);

    try {
      const nextTrips = await getMyFavorites({
        pageNum: pageNumRef.current + 1,
        pageSize: PAGE_SIZE,
      });
      setTrips((current) => [...current, ...nextTrips.items]);
      setPageNum(nextTrips.pageNum);
      setHasMore(tripsLengthRef.current + nextTrips.items.length < nextTrips.total);
      setError("");
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "加载更多失败");
    } finally {
      setLoadingMore(false);
    }
  }

  async function refreshTrips() {
    if (loading || refreshing) {
      return;
    }

    setRefreshing(true);

    try {
      const nextTrips = await getMyFavorites({
        pageNum: 1,
        pageSize: PAGE_SIZE,
      });
      setTrips(nextTrips.items);
      setPageNum(nextTrips.pageNum);
      setHasMore(nextTrips.items.length < nextTrips.total);
      setError("");
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "刷新失败");
    } finally {
      setRefreshing(false);
    }
  }

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

        {!loading && trips.length > 0 ? (
          <InfiniteScroll
            dataLength={trips.length}
            hasMore={hasMore}
            next={loadNextPage}
            loader={loadingMore ? <p className="subtle-text">正在加载更多...</p> : undefined}
            pullDownToRefresh
            pullDownToRefreshContent={<p className="subtle-text">下拉刷新</p>}
            pullDownToRefreshThreshold={70}
            refreshFunction={refreshTrips}
            releaseToRefreshContent={<p className="subtle-text">松开立即刷新</p>}
          >
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
          </InfiniteScroll>
        ) : null}
      </section>
    </main>
  );
}
