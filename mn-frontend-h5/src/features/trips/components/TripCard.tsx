import type { ReactNode } from "react";
import { Link } from "react-router-dom";

import type { TripSummary } from "../api";

const statusLabelMap: Record<TripSummary["status"], string> = {
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
    month: "numeric",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

type TripCardProps = {
  trip: TripSummary;
  footer?: ReactNode;
  disableLink?: boolean;
};

function TripCardBody({ trip }: { trip: TripSummary }) {
  return (
    <>
      <div className="trip-card__route">
        <div className="trip-card__point">
          <p className="trip-card__label">起点</p>
          <h3 className="trip-card__city">{trip.fromText || "--"}</h3>
        </div>
        <div className="trip-card__connector" aria-hidden="true">
          <span className="trip-card__dot" />
          <span className="trip-card__line" />
          <span className="trip-card__dot" />
        </div>
        <div className="trip-card__point">
          <p className="trip-card__label">终点</p>
          <h3 className="trip-card__city">{trip.toText || "--"}</h3>
        </div>
      </div>
      <div className="trip-card__meta">
        <span className="tag">{tripTypeLabelMap[trip.tripType] ?? trip.tripType}</span>
        <span className="tag">{trip.seatCount} 人</span>
        <span className="tag">{trip.isPriceNegotiable ? "费用面议" : "费用未标注"}</span>
        {trip.favorited ? <span className="tag trip-card__favorite">已收藏</span> : null}
        {trip.unavailable ? <span className="tag trip-card__favorite">已失效</span> : null}
      </div>
    </>
  );
}

export default function TripCard({ trip, footer, disableLink = false }: TripCardProps) {
  const blocked = disableLink || trip.unavailable;

  return (
    <article className="trip-card">
      <span className={`trip-card__accent trip-card__accent--${trip.status}`} aria-hidden="true" />
      <div className="trip-card__inner">
        <div className="trip-card__topline">
          <span className={`status-pill status-pill--${trip.status}`}>{statusLabelMap[trip.status]}</span>
          <span className="trip-card__time">{formatDeparture(trip.departureDate, trip.departureTime)}</span>
        </div>
        {blocked ? (
          <div className="trip-card__link trip-card__link--disabled" aria-disabled="true">
            <TripCardBody trip={trip} />
          </div>
        ) : (
          <Link className="trip-card__link" to={`/trips/${trip.id}`}>
            <TripCardBody trip={trip} />
          </Link>
        )}
        {footer ? <div className="trip-card__footer">{footer}</div> : null}
      </div>
    </article>
  );
}
