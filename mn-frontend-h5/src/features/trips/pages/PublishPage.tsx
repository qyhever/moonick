import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import { createTrip, type TripFormPayload, type TripType } from "../api";

type TripFormState = {
  tripType: TripType;
  fromText: string;
  toText: string;
  departureDate: string;
  departureTime: string;
  seatCount: string;
  isPriceNegotiable: boolean;
  contactPhone: string;
  contactWechat: string;
};

const initialForm: TripFormState = {
  tripType: "driver_post",
  fromText: "",
  toText: "",
  departureDate: "",
  departureTime: "",
  seatCount: "1",
  isPriceNegotiable: true,
  contactPhone: "",
  contactWechat: "",
};

function buildPayload(form: TripFormState): TripFormPayload {
  return {
    tripType: form.tripType,
    fromText: form.fromText.trim(),
    toText: form.toText.trim(),
    departureDate: form.departureDate.trim(),
    departureTime: form.departureTime.trim(),
    seatCount: Number(form.seatCount) || 1,
    isPriceNegotiable: form.isPriceNegotiable,
    contactPhone: form.contactPhone.trim(),
    contactWechat: form.contactWechat.trim(),
  };
}

function getDepartureTimestamp(form: TripFormState) {
  return new Date(`${form.departureDate}T${form.departureTime}:00`).getTime();
}

export default function PublishPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<TripFormState>(initialForm);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function updateField(event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) {
    const { name, value } = event.target;
    setForm((current) => ({ ...current, [name]: value }));
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!form.fromText.trim() || !form.toText.trim() || !form.departureDate.trim() || !form.departureTime.trim()) {
      setError("请完整填写出发信息");
      return;
    }

    if (form.fromText.trim() === form.toText.trim()) {
      setError("起点和终点不能相同");
      return;
    }

    const seatCount = Number(form.seatCount);
    if (!Number.isInteger(seatCount) || seatCount < 1 || seatCount > 6) {
      setError("人数范围为 1 到 6 人");
      return;
    }

    if (getDepartureTimestamp(form) < Date.now()) {
      setError("出发时间不能早于当前时间");
      return;
    }

    if (!form.contactWechat.trim() && !form.contactPhone.trim()) {
      setError("请填写至少一种联系方式");
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      const created = await createTrip(buildPayload(form));
      navigate(`/trips/${created.id}`, {
        state: { toast: "发布成功" },
      });
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "发布失败，请稍后再试");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="h5-shell">
      <section className="hero-card hero-card--compact">
        <p className="eyebrow">发布行程</p>
        <h1 className="hero-card__title">发一条顺路行程，快速被附近同城的人看到</h1>
        <p className="hero-card__subtitle">延续暖色卡片式录入体验，只保留最关键字段和校验闭环。</p>
      </section>

      <section className="page-panel">
        <form className="stack-form" onSubmit={handleSubmit}>
          <div className="field-block">
            <span>行程类型</span>
            <div className="segmented-row">
              <button
                className={form.tripType === "driver_post" ? "segment-button is-active" : "segment-button"}
                name="tripType"
                onClick={() => setForm((current) => ({ ...current, tripType: "driver_post" }))}
                type="button"
              >
                车找人
              </button>
              <button
                className={form.tripType === "passenger_post" ? "segment-button is-active" : "segment-button"}
                name="tripType"
                onClick={() => setForm((current) => ({ ...current, tripType: "passenger_post" }))}
                type="button"
              >
                人找车
              </button>
            </div>
          </div>

          <label className="field-block">
            <span>起点</span>
            <input name="fromText" value={form.fromText} onChange={updateField} placeholder="如：上海虹桥站" />
          </label>

          <label className="field-block">
            <span>终点</span>
            <input name="toText" value={form.toText} onChange={updateField} placeholder="如：杭州东站" />
          </label>

          <div className="field-grid">
            <label className="field-block">
              <span>出发日期</span>
              <input
                name="departureDate"
                value={form.departureDate}
                onChange={updateField}
                placeholder="YYYY-MM-DD"
              />
            </label>
            <label className="field-block">
              <span>出发时间</span>
              <input
                name="departureTime"
                value={form.departureTime}
                onChange={updateField}
                placeholder="HH:mm"
              />
            </label>
          </div>

          <div className="field-grid">
            <label className="field-block">
              <span>人数</span>
              <input name="seatCount" value={form.seatCount} onChange={updateField} />
            </label>
          </div>

          <label className="check-row">
            <input
              checked={form.isPriceNegotiable}
              onChange={(event) => {
                setForm((current) => ({ ...current, isPriceNegotiable: event.target.checked }));
              }}
              type="checkbox"
            />
            <span>费用面议</span>
          </label>

          <label className="field-block">
            <span>手机号</span>
            <input
              name="contactPhone"
              value={form.contactPhone}
              onChange={updateField}
              placeholder="至少填写一种联系方式"
            />
          </label>

          <label className="field-block">
            <span>微信号</span>
            <input
              name="contactWechat"
              value={form.contactWechat}
              onChange={updateField}
              placeholder="可选"
            />
          </label>

          {error ? <p role="alert">{error}</p> : null}

          <div className="action-row action-row--single">
            <button className="primary-button primary-button--block" disabled={submitting} type="submit">
              发布
            </button>
          </div>
        </form>
      </section>
    </main>
  );
}
