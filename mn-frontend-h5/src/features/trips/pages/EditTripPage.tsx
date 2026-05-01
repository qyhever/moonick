import type { ChangeEvent, FormEvent } from "react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";

import { getTripDetail, updateTrip, type TripFormPayload, type TripType } from "../api";

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
  remark: string;
};

const emptyForm: TripFormState = {
  tripType: "driver_post",
  fromText: "",
  toText: "",
  departureDate: "",
  departureTime: "",
  seatCount: "1",
  isPriceNegotiable: true,
  contactPhone: "",
  contactWechat: "",
  remark: "",
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
    remark: form.remark.trim(),
  };
}

function getDepartureTimestamp(form: TripFormState) {
  return new Date(`${form.departureDate}T${form.departureTime}:00`).getTime();
}

export default function EditTripPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [form, setForm] = useState<TripFormState>(emptyForm);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let active = true;

    async function loadTrip() {
      setLoading(true);
      setError("");

      try {
        const trip = await getTripDetail(id);
        if (!active) {
          return;
        }

        setForm({
          tripType: trip.tripType === "passenger_post" ? "passenger_post" : "driver_post",
          fromText: trip.fromText,
          toText: trip.toText,
          departureDate: trip.departureDate,
          departureTime: trip.departureTime,
          seatCount: String(trip.seatCount),
          isPriceNegotiable: trip.isPriceNegotiable,
          contactPhone: trip.contactPhone,
          contactWechat: trip.contactWechat,
          remark: trip.remark,
        });
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "行程加载失败");
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
      const updated = await updateTrip(id, buildPayload(form));
      navigate(`/trips/${updated.id}`);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "更新失败，请稍后再试");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="h5-shell">
      <section className="page-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">编辑行程</p>
            <h1>更新已有发布</h1>
          </div>
          <Link className="secondary-link" to="/me/trips">
            我的发布
          </Link>
        </div>

        {loading ? <p className="subtle-text">正在加载表单...</p> : null}

        {!loading ? (
          <form className="stack-form" onSubmit={handleSubmit}>
            <div className="field-block">
              <span>行程类型</span>
              <div className="segmented-row">
                <button
                  className={form.tripType === "driver_post" ? "segment-button is-active" : "segment-button"}
                  onClick={() => setForm((current) => ({ ...current, tripType: "driver_post" }))}
                  type="button"
                >
                  车找人
                </button>
                <button
                  className={form.tripType === "passenger_post" ? "segment-button is-active" : "segment-button"}
                  onClick={() => setForm((current) => ({ ...current, tripType: "passenger_post" }))}
                  type="button"
                >
                  人找车
                </button>
              </div>
            </div>
            <label className="field-block">
              <span>起点</span>
              <input name="fromText" value={form.fromText} onChange={updateField} />
            </label>
            <label className="field-block">
              <span>终点</span>
              <input name="toText" value={form.toText} onChange={updateField} />
            </label>
            <div className="field-grid">
              <label className="field-block">
                <span>出发日期</span>
                <input name="departureDate" value={form.departureDate} onChange={updateField} />
              </label>
              <label className="field-block">
                <span>出发时间</span>
                <input name="departureTime" value={form.departureTime} onChange={updateField} />
              </label>
            </div>
            <label className="field-block">
              <span>人数</span>
              <input name="seatCount" value={form.seatCount} onChange={updateField} />
            </label>
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
              <input name="contactPhone" value={form.contactPhone} onChange={updateField} />
            </label>
            <label className="field-block">
              <span>微信号</span>
              <input name="contactWechat" value={form.contactWechat} onChange={updateField} />
            </label>
            <label className="field-block">
              <span>备注</span>
              <textarea
                name="remark"
                value={form.remark}
                onChange={updateField}
                placeholder="可填写出发时间弹性、上下车说明、行李要求等"
              />
            </label>
            {error ? <p role="alert">{error}</p> : null}
            <button className="primary-button" disabled={submitting} type="submit">
              保存
            </button>
          </form>
        ) : null}
      </section>
    </main>
  );
}
