import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Toast } from "antd-mobile";

import { resetPassword, sendVerificationCode } from "../api";
import { useAuthStore } from "../../../store/auth";
import { useNavigate } from "react-router-dom";

const RESET_PASSWORD_CODE_TYPE = "reset_password";
const RESET_PASSWORD_CODE_STORAGE_KEY = "mn-h5-reset-password-code-expires-at";
const RESET_PASSWORD_CODE_COUNTDOWN_SECONDS = 60;

function readStoredCountdownDeadline() {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(RESET_PASSWORD_CODE_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  const expiresAt = Number(raw);
  if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) {
    window.localStorage.removeItem(RESET_PASSWORD_CODE_STORAGE_KEY);
    return null;
  }

  return expiresAt;
}

function persistCountdownDeadline(expiresAt: number | null) {
  if (typeof window === "undefined") {
    return;
  }

  if (expiresAt === null) {
    window.localStorage.removeItem(RESET_PASSWORD_CODE_STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(RESET_PASSWORD_CODE_STORAGE_KEY, String(expiresAt));
}

function getRemainingSeconds(expiresAt: number | null) {
  if (expiresAt === null) {
    return 0;
  }

  const remainingMs = expiresAt - Date.now();
  if (remainingMs <= 0) {
    return 0;
  }

  return Math.ceil(remainingMs / 1000);
}

export default function PasswordResetPage() {
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const email = useAuthStore((state) => state.user?.email ?? "");
  const [verificationCode, setVerificationCode] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [isSendingCode, setIsSendingCode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [countdownExpiresAt, setCountdownExpiresAt] = useState<number | null>(() =>
    readStoredCountdownDeadline(),
  );
  const [countdownSeconds, setCountdownSeconds] = useState(() =>
    getRemainingSeconds(readStoredCountdownDeadline()),
  );

  useEffect(() => {
    if (countdownExpiresAt === null) {
      setCountdownSeconds(0);
      return;
    }

    const timer = window.setInterval(() => {
      const nextSeconds = getRemainingSeconds(countdownExpiresAt);
      if (nextSeconds <= 0) {
        persistCountdownDeadline(null);
        setCountdownExpiresAt(null);
        setCountdownSeconds(0);
      } else {
        setCountdownSeconds(nextSeconds);
      }
    }, 1000);

    setCountdownSeconds(getRemainingSeconds(countdownExpiresAt));

    return () => {
      window.clearInterval(timer);
    };
  }, [countdownExpiresAt]);

  async function handleSendVerificationCode() {
    if (!email) {
      setError("未获取到当前登录邮箱，请重新登录后再试");
      return;
    }

    setError("");
    setIsSendingCode(true);

    try {
      const payload = await sendVerificationCode(email, RESET_PASSWORD_CODE_TYPE);
      if (!payload.sent) {
        throw new Error("验证码发送失败，请稍后重试");
      }

      const expiresAt = Date.now() + RESET_PASSWORD_CODE_COUNTDOWN_SECONDS * 1000;
      persistCountdownDeadline(expiresAt);
      setCountdownExpiresAt(expiresAt);
      setCountdownSeconds(RESET_PASSWORD_CODE_COUNTDOWN_SECONDS);
    } catch (error) {
      setError(error instanceof Error ? error.message : "验证码发送失败，请稍后重试");
    } finally {
      setIsSendingCode(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!password) {
      setError("请输入新的登录密码");
      return;
    }

    if (!confirmPassword) {
      setError("请再次输入新的登录密码");
      return;
    }

    if (password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }

    if (!verificationCode.trim()) {
      setError("请输入验证码");
      return;
    }

    if (!email) {
      setError("未获取到当前登录邮箱，请重新登录后再试");
      return;
    }

    setError("");
    setIsSubmitting(true);

    try {
      await resetPassword(email, verificationCode.trim(), password);
      persistCountdownDeadline(null);
      logout();
      Toast.show({
        content: "密码已重置，请重新登录",
      });
      navigate("/login", { replace: true });
    } catch (error) {
      setError(error instanceof Error ? error.message : "重置密码失败，请稍后重试");
    } finally {
      setIsSubmitting(false);
    }
  }

  const isSendDisabled = isSendingCode || countdownSeconds > 0;
  const sendButtonText =
    countdownSeconds > 0 ? `${countdownSeconds}s后重试` : isSendingCode ? "发送中..." : "发送验证码";

  return (
    <main className="h5-shell h5-shell--profile">
      <section className="page-panel password-reset-panel">
        <section className="card password-reset-hero">
          <p className="password-reset-hero__eyebrow">Password Reset</p>
          <h1 className="password-reset-hero__title">
            请设置新的登录密码，
            <br />
            让账户安全保持在你手里。
          </h1>
          <p className="password-reset-hero__subtitle">
            这是一个重置密码场景，不需要输入当前密码。完成后，下次登录请使用新密码。
          </p>
        </section>

        <section className="card password-reset-form-card auth-form">
          <h2 className="form-title">重置密码</h2>
          <p className="form-subtitle">输入新的登录密码，并再次确认，避免因输入错误影响后续登录。</p>
          <form className="password-reset-form" onSubmit={handleSubmit}>
            <label className="field-block">
              <span>新密码</span>
              <input
                onChange={(event) => setPassword(event.target.value)}
                type="password"
                value={password}
              />
            </label>
            <label className="field-block">
              <span>确认密码</span>
              <input
                onChange={(event) => setConfirmPassword(event.target.value)}
                type="password"
                value={confirmPassword}
              />
            </label>
            <div className="auth-code">
              <label className="auth-code__label" htmlFor="password-reset-verification-code">
                验证码
              </label>
              <div className="auth-code__control">
                <input
                  id="password-reset-verification-code"
                  inputMode="numeric"
                  onChange={(event) => setVerificationCode(event.target.value)}
                  placeholder="请输入验证码"
                  type="text"
                  value={verificationCode}
                />
                <button
                  className="auth-code__send"
                  disabled={isSendDisabled}
                  type="button"
                  onClick={handleSendVerificationCode}
                >
                  {sendButtonText}
                </button>
              </div>
              {countdownSeconds > 0 ? (
                <p className="auth-code__hint">没收到验证码？{countdownSeconds}s 后可重新发送</p>
              ) : null}
            </div>
            <button className="primary-button primary-button--block" disabled={isSubmitting} type="submit">
              {isSubmitting ? "提交中..." : "确认"}
            </button>
            {error ? <p role="alert">{error}</p> : null}
          </form>
        </section>
      </section>
    </main>
  );
}
