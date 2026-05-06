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

type HomePageSnapshot = {
  trips: TripSummary[];
  total: number;
  pageNum: number;
  hasMore: boolean;
  filters: HomeFilters;
  scrollY: number;
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
const FILTER_DRAWER_TRANSITION_MS = 320;
let cachedHomePageSnapshot: HomePageSnapshot | null = null;
let shouldRestoreHomePageSnapshot = false;

function cloneHomePageSnapshot(snapshot: HomePageSnapshot): HomePageSnapshot {
  return {
    trips: [...snapshot.trips],
    total: snapshot.total,
    pageNum: snapshot.pageNum,
    hasMore: snapshot.hasMore,
    filters: { ...snapshot.filters },
    scrollY: snapshot.scrollY,
  };
}

function saveHomePageSnapshot(snapshot: HomePageSnapshot) {
  cachedHomePageSnapshot = cloneHomePageSnapshot(snapshot);
  shouldRestoreHomePageSnapshot = true;
}

function consumeHomePageSnapshot() {
  if (!shouldRestoreHomePageSnapshot || !cachedHomePageSnapshot) {
    return null;
  }

  shouldRestoreHomePageSnapshot = false;
  return cloneHomePageSnapshot(cachedHomePageSnapshot);
}

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
  const restoreSnapshotRef = useRef<HomePageSnapshot | null>(consumeHomePageSnapshot());
  const restoredSnapshot = restoreSnapshotRef.current;
  const drawerScrollLockRef = useRef<{
    scrollY: number;
    bodyOverflow: string;
    bodyPosition: string;
    bodyTop: string;
    bodyLeft: string;
    bodyRight: string;
    bodyWidth: string;
    htmlOverflow: string;
  } | null>(null);
  const accessToken = useAuthStore((state) => state.accessToken);
  const currentUser = useAuthStore((state) => state.user);
  const [trips, setTrips] = useState<TripSummary[]>(restoredSnapshot?.trips ?? []);
  const [total, setTotal] = useState(restoredSnapshot?.total ?? 0);
  const [pageNum, setPageNum] = useState(restoredSnapshot?.pageNum ?? 1);
  const [hasMore, setHasMore] = useState(restoredSnapshot?.hasMore ?? true);
  const [loading, setLoading] = useState(!restoredSnapshot);
  const [loadingMore, setLoadingMore] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState<HomeFilters>(restoredSnapshot?.filters ?? defaultFilters);
  const [isFilterDrawerMounted, setIsFilterDrawerMounted] = useState(false);
  const [isFilterDrawerOpen, setIsFilterDrawerOpen] = useState(false);
  const [isFilterDrawerClosing, setIsFilterDrawerClosing] = useState(false);
  const [draftDrawerFilters, setDraftDrawerFilters] = useState({
    fromText: restoredSnapshot?.filters.fromText ?? defaultFilters.fromText,
    toText: restoredSnapshot?.filters.toText ?? defaultFilters.toText,
    datePreset: restoredSnapshot?.filters.datePreset ?? defaultFilters.datePreset,
    onlyAvailable: restoredSnapshot?.filters.onlyAvailable ?? defaultFilters.onlyAvailable,
    negotiableOnly: restoredSnapshot?.filters.negotiableOnly ?? defaultFilters.negotiableOnly,
  });
  const pageNumRef = useRef(1);
  const tripsLengthRef = useRef(0);
  const drawerAnimationFrameRef = useRef<number | null>(null);
  const drawerCloseTimeoutRef = useRef<number | null>(null);
  const shouldSkipInitialFetchRef = useRef(Boolean(restoredSnapshot));
  const snapshotRef = useRef<HomePageSnapshot>({
    trips: restoredSnapshot?.trips ?? [],
    total: restoredSnapshot?.total ?? 0,
    pageNum: restoredSnapshot?.pageNum ?? 1,
    hasMore: restoredSnapshot?.hasMore ?? true,
    filters: restoredSnapshot?.filters ?? defaultFilters,
    scrollY: restoredSnapshot?.scrollY ?? 0,
  });

  useEffect(() => {
    pageNumRef.current = pageNum;
  }, [pageNum]);

  useEffect(() => {
    tripsLengthRef.current = trips.length;
  }, [trips.length]);

  useEffect(() => {
    snapshotRef.current = {
      trips,
      total,
      pageNum,
      hasMore,
      filters,
      scrollY: window.scrollY,
    };
  }, [filters, hasMore, pageNum, total, trips]);

  useEffect(() => {
    if (!restoredSnapshot) {
      return;
    }

    window.scrollTo({ top: restoredSnapshot.scrollY, left: 0, behavior: "auto" });
  }, [restoredSnapshot]);

  useEffect(() => {
    return () => {
      if (drawerAnimationFrameRef.current !== null) {
        window.cancelAnimationFrame(drawerAnimationFrameRef.current);
      }

      if (drawerCloseTimeoutRef.current !== null) {
        window.clearTimeout(drawerCloseTimeoutRef.current);
      }
    };
  }, []);

  useEffect(() => {
    function unlockBodyScroll() {
      const lockedState = drawerScrollLockRef.current;
      if (!lockedState) {
        return;
      }

      document.body.style.overflow = lockedState.bodyOverflow;
      document.body.style.position = lockedState.bodyPosition;
      document.body.style.top = lockedState.bodyTop;
      document.body.style.left = lockedState.bodyLeft;
      document.body.style.right = lockedState.bodyRight;
      document.body.style.width = lockedState.bodyWidth;
      document.documentElement.style.overflow = lockedState.htmlOverflow;
      drawerScrollLockRef.current = null;
      window.scrollTo({ top: lockedState.scrollY, left: 0, behavior: "auto" });
    }

    if (!isFilterDrawerMounted) {
      unlockBodyScroll();
      return;
    }

    const scrollY = window.scrollY;

    drawerScrollLockRef.current = {
      scrollY,
      bodyOverflow: document.body.style.overflow,
      bodyPosition: document.body.style.position,
      bodyTop: document.body.style.top,
      bodyLeft: document.body.style.left,
      bodyRight: document.body.style.right,
      bodyWidth: document.body.style.width,
      htmlOverflow: document.documentElement.style.overflow,
    };

    document.body.style.overflow = "hidden";
    document.body.style.position = "fixed";
    document.body.style.top = `-${scrollY}px`;
    document.body.style.left = "0";
    document.body.style.right = "0";
    document.body.style.width = "100%";
    document.documentElement.style.overflow = "hidden";

    return unlockBodyScroll;
  }, [isFilterDrawerMounted]);

  useEffect(() => {
    if (shouldSkipInitialFetchRef.current) {
      shouldSkipInitialFetchRef.current = false;
      return;
    }

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
    if (drawerCloseTimeoutRef.current !== null) {
      window.clearTimeout(drawerCloseTimeoutRef.current);
      drawerCloseTimeoutRef.current = null;
    }

    if (drawerAnimationFrameRef.current !== null) {
      window.cancelAnimationFrame(drawerAnimationFrameRef.current);
    }

    setDraftDrawerFilters({
      fromText: filters.fromText,
      toText: filters.toText,
      datePreset: filters.datePreset,
      onlyAvailable: filters.onlyAvailable,
      negotiableOnly: filters.negotiableOnly,
    });
    setIsFilterDrawerMounted(true);
    setIsFilterDrawerClosing(false);

    drawerAnimationFrameRef.current = window.requestAnimationFrame(() => {
      setIsFilterDrawerOpen(true);
      drawerAnimationFrameRef.current = null;
    });
  }

  function closeFilterDrawer() {
    setIsFilterDrawerOpen(false);
    setIsFilterDrawerClosing(true);

    if (drawerCloseTimeoutRef.current !== null) {
      window.clearTimeout(drawerCloseTimeoutRef.current);
    }

    drawerCloseTimeoutRef.current = window.setTimeout(() => {
      setIsFilterDrawerMounted(false);
      setIsFilterDrawerClosing(false);
      drawerCloseTimeoutRef.current = null;
    }, FILTER_DRAWER_TRANSITION_MS);
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

  function rememberCurrentPageState() {
    saveHomePageSnapshot({
      ...snapshotRef.current,
      scrollY: window.scrollY,
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
                <TripCard
                  key={trip.id}
                  disableLink={trip.status === "full"}
                  onNavigate={rememberCurrentPageState}
                  trip={trip}
                />
              ))}
            </div>
          </InfiniteScroll>
        )}
      </section>

      {isFilterDrawerMounted ? (
        <div
          className={
            isFilterDrawerOpen
              ? "filter-drawer is-open"
              : isFilterDrawerClosing
                ? "filter-drawer is-closing"
                : "filter-drawer"
          }
          role="dialog"
          aria-modal="true"
          aria-labelledby="home-filter-drawer-title"
          data-state={isFilterDrawerOpen ? "open" : isFilterDrawerClosing ? "closing" : "closed"}
        >
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
                    高级查询
                  </h2>
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
