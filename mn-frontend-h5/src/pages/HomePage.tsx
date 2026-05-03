import { useEffect, useRef, useState } from "react";
import { Menu } from "lucide-react";
import { Link } from "react-router-dom";
import InfiniteScroll from "react-infinite-scroll-component";

import TripCard from "../features/trips/components/TripCard";
import { getTrips, type TripQuery, type TripSummary } from "../features/trips/api";
import { useAuthStore } from "../store/auth";

type HomeTripTypeFilter = "driver_post" | "passenger_post";
type HomeDatePreset = "all" | "today" | "tomorrow";

type HomeFilters = {
  tripType: HomeTripTypeFilter;
  fromText: string;
  toText: string;
  datePreset: HomeDatePreset;
  onlyAvailable: boolean;
  negotiableOnly: boolean;
};

const defaultFilters: HomeFilters = {
  tripType: "driver_post",
  fromText: "",
  toText: "",
  datePreset: "all",
  onlyAvailable: false,
  negotiableOnly: false,
};

const tripTypeOptions: Array<{ label: string; value: HomeTripTypeFilter }> = [
  { label: "车找人", value: "driver_post" },
  { label: "人找车", value: "passenger_post" },
];

const datePresetOptions: Array<{ label: string; value: HomeDatePreset }> = [
  { label: "全部", value: "all" },
  { label: "今天", value: "today" },
  { label: "明天", value: "tomorrow" },
];

const PAGE_SIZE = 10;

function buildTripQuery(filters: HomeFilters, pageNum: number, pageSize = PAGE_SIZE): TripQuery {
  const query: TripQuery = {
    pageNum,
    pageSize,
  };

  query.tripType = filters.tripType;

  if (filters.fromText) {
    query.fromText = filters.fromText;
  }

  if (filters.toText) {
    query.toText = filters.toText;
  }

  if (filters.datePreset !== "all") {
    query.datePreset = filters.datePreset;
  }

  return query;
}

function applyClientFilters(trips: TripSummary[], filters: HomeFilters) {
  return trips.filter((trip) => {
    if (filters.onlyAvailable && trip.status !== "active") {
      return false;
    }

    if (filters.negotiableOnly && !trip.isPriceNegotiable) {
      return false;
    }

    return true;
  });
}

export default function HomePage() {
  const accessToken = useAuthStore((state) => state.accessToken);
  const currentUser = useAuthStore((state) => state.user);
  const [trips, setTrips] = useState<TripSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [pageNum, setPageNum] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState<HomeFilters>(defaultFilters);
  const [isFilterDrawerOpen, setIsFilterDrawerOpen] = useState(false);
  const [draftDrawerFilters, setDraftDrawerFilters] = useState({
    fromText: defaultFilters.fromText,
    toText: defaultFilters.toText,
    datePreset: defaultFilters.datePreset,
    onlyAvailable: defaultFilters.onlyAvailable,
    negotiableOnly: defaultFilters.negotiableOnly,
  });
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

    async function loadFirstPage() {
      setLoading(true);
      setError("");

      try {
        const nextTrips = await getTrips(buildTripQuery(filters, 1));
        if (active) {
          setTrips(nextTrips.items);
          setTotal(nextTrips.total);
          setPageNum(nextTrips.pageNum);
          setHasMore(nextTrips.items.length < nextTrips.total);
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

    void loadFirstPage();

    return () => {
      active = false;
    };
  }, [filters.datePreset, filters.fromText, filters.toText, filters.tripType]);

  async function loadNextPage() {
    if (loading || loadingMore || refreshing || !hasMore) {
      return;
    }

    setLoadingMore(true);

    try {
      const nextTrips = await getTrips(buildTripQuery(filters, pageNumRef.current + 1));
      setTrips((current) => [...current, ...nextTrips.items]);
      setTotal(nextTrips.total);
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
      const nextTrips = await getTrips(buildTripQuery(filters, 1));
      setTrips(nextTrips.items);
      setTotal(nextTrips.total);
      setPageNum(nextTrips.pageNum);
      setHasMore(nextTrips.items.length < nextTrips.total);
      setError("");
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "刷新失败");
    } finally {
      setRefreshing(false);
    }
  }

  function updateQuickFilter<K extends keyof HomeFilters>(key: K, value: HomeFilters[K]) {
    setFilters((current) => ({
      ...current,
      [key]: value,
    }));
  }

  function openFilterDrawer() {
    setDraftDrawerFilters({
      fromText: filters.fromText,
      toText: filters.toText,
      datePreset: filters.datePreset,
      onlyAvailable: filters.onlyAvailable,
      negotiableOnly: filters.negotiableOnly,
    });
    setIsFilterDrawerOpen(true);
  }

  function closeFilterDrawer() {
    setIsFilterDrawerOpen(false);
  }

  function applyDrawerFilters() {
    setFilters((current) => ({
      ...current,
      fromText: draftDrawerFilters.fromText.trim(),
      toText: draftDrawerFilters.toText.trim(),
      datePreset: draftDrawerFilters.datePreset,
      onlyAvailable: draftDrawerFilters.onlyAvailable,
      negotiableOnly: draftDrawerFilters.negotiableOnly,
    }));
    setIsFilterDrawerOpen(false);
  }

  function resetDrawerFilters() {
    setDraftDrawerFilters({
      fromText: defaultFilters.fromText,
      toText: defaultFilters.toText,
      datePreset: defaultFilters.datePreset,
      onlyAvailable: defaultFilters.onlyAvailable,
      negotiableOnly: defaultFilters.negotiableOnly,
    });
  }

  const visibleTrips = applyClientFilters(trips, filters);
  const visibleTotal = visibleTrips.length;

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
          专为有拼车、搭车、返乡、跨城出行需求的车主和乘客打造。
        </p>
      </section>

      <section className="page-panel">
        <div>
          <div>
            <p className="eyebrow">Trip Board</p>
          </div>
          <div className="home-filter-primary-row">
            <div className="home-filter-chip-row" role="group" aria-label="行程类型筛选">
              {tripTypeOptions.map((option) => (
                <button
                  key={option.value}
                  className={filters.tripType === option.value ? "home-filter-chip is-active" : "home-filter-chip"}
                  onClick={() => updateQuickFilter("tripType", option.value)}
                  type="button"
                >
                  {option.label}
                </button>
              ))}
            </div>

            <button
              aria-label="打开筛选抽屉"
              className="home-filter-icon-trigger"
              onClick={openFilterDrawer}
              type="button"
            >
              <Menu size={18} strokeWidth={2.2} />
            </button>
          </div>
          <div className="section-link">{visibleTotal} 条结果</div>
        </div>

        {loading && trips.length === 0 ? <p className="subtle-text">正在加载首页行程...</p> : null}
        {error && trips.length === 0 ? <p role="alert">{error}</p> : null}
        {!loading && !error && visibleTrips.length === 0 ? <p className="subtle-text">当前还没有可展示的行程。</p> : null}

        {error && trips.length === 0 ? null : (
          <InfiniteScroll
            dataLength={trips.length}
            endMessage={!loading && visibleTrips.length > 0 ? <p className="subtle-text">已经到底了</p> : undefined}
            hasMore={hasMore}
            loader={<p className="subtle-text">正在加载更多...</p>}
            next={loadNextPage}
            pullDownToRefresh
            pullDownToRefreshContent={<p className="subtle-text">下拉刷新</p>}
            pullDownToRefreshThreshold={70}
            refreshFunction={refreshTrips}
            releaseToRefreshContent={<p className="subtle-text">松开立即刷新</p>}
          >
            <div className="trip-list">
              {visibleTrips.map((trip) => (
                <TripCard key={trip.id} disableLink={trip.status === "full"} trip={trip} />
              ))}
            </div>
          </InfiniteScroll>
        )}
      </section>

      {isFilterDrawerOpen ? (
        <div className="filter-drawer" role="dialog" aria-modal="true" aria-labelledby="home-filter-drawer-title">
          <button
            aria-label="关闭筛选抽屉"
            className="filter-drawer__backdrop"
            onClick={closeFilterDrawer}
            type="button"
          />
          <div className="filter-drawer__sheet">
            <div className="filter-drawer__content" data-testid="home-filter-drawer-content">
              <div className="filter-drawer__handle" aria-hidden="true" />
              <div className="filter-drawer__header">
                <div>
                  <h2 className="section-title" id="home-filter-drawer-title">
                    筛选抽屉
                  </h2>
                  <p className="section-subtitle">时间与低频条件收进这里，首页保持轻量</p>
                </div>
                <button className="filter-drawer__text-button" onClick={closeFilterDrawer} type="button">
                  关闭
                </button>
              </div>

              <div className="filter-drawer__group">
                <div className="home-filter-grid">
                  <label className="home-filter-field">
                    <span>起点</span>
                    <input
                      aria-label="起点"
                      name="fromText"
                      onChange={(event) =>
                        setDraftDrawerFilters((current) => ({
                          ...current,
                          fromText: event.target.value,
                        }))
                      }
                      placeholder="如：上海虹桥"
                      value={draftDrawerFilters.fromText}
                    />
                  </label>
                  <label className="home-filter-field">
                    <span>终点</span>
                    <input
                      aria-label="终点"
                      name="toText"
                      onChange={(event) =>
                        setDraftDrawerFilters((current) => ({
                          ...current,
                          toText: event.target.value,
                        }))
                      }
                      placeholder="如：杭州东站"
                      value={draftDrawerFilters.toText}
                    />
                  </label>
                </div>
              </div>

              <div className="filter-drawer__group">
                <p className="filter-drawer__label">时间</p>
                <div className="home-filter-chip-row" role="group" aria-label="时间筛选">
                  {datePresetOptions.map((option) => (
                    <button
                      key={option.value}
                      className={
                        draftDrawerFilters.datePreset === option.value
                          ? "home-filter-chip is-active"
                          : "home-filter-chip"
                      }
                      onClick={() =>
                        setDraftDrawerFilters((current) => ({
                          ...current,
                          datePreset: option.value,
                        }))
                      }
                      type="button"
                    >
                      {option.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="filter-drawer__group">
                <p className="filter-drawer__label">更多条件</p>
                <div className="home-filter-chip-row">
                  <button
                    className={draftDrawerFilters.onlyAvailable ? "home-filter-chip is-active" : "home-filter-chip"}
                    onClick={() =>
                      setDraftDrawerFilters((current) => ({
                        ...current,
                        onlyAvailable: !current.onlyAvailable,
                      }))
                    }
                    type="button"
                  >
                    仅看可约
                  </button>
                  <button
                    className={draftDrawerFilters.negotiableOnly ? "home-filter-chip is-active" : "home-filter-chip"}
                    onClick={() =>
                      setDraftDrawerFilters((current) => ({
                        ...current,
                        negotiableOnly: !current.negotiableOnly,
                      }))
                    }
                    type="button"
                  >
                    可议价
                  </button>
                </div>
              </div>
            </div>

            <div className="filter-drawer__actions">
              <button className="secondary-link secondary-link--button" onClick={resetDrawerFilters} type="button">
                重置
              </button>
              <button className="primary-button" onClick={applyDrawerFilters} type="button">
                确定
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </main>
  );
}
