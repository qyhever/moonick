import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { type ApiResponse, api, unwrapApiResponse } from "../lib/http";
import { isValidEmail } from "../lib/validation";
import { useAuthStore } from "../store/auth";

const REGISTER_CODE_EXPIRES_AT_STORAGE_KEY = "mn-h5-register-code-expires-at";
const REGISTER_CODE_COUNTDOWN_SECONDS = 60;

type RegisterCodeResponse = {
  sent: boolean;
};

function readStoredCountdownDeadline() {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(REGISTER_CODE_EXPIRES_AT_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  const expiresAt = Number(raw);
  if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) {
    window.localStorage.removeItem(REGISTER_CODE_EXPIRES_AT_STORAGE_KEY);
    return null;
  }

  return expiresAt;
}

function persistCountdownDeadline(expiresAt: number | null) {
  if (typeof window === "undefined") {
    return;
  }

  if (expiresAt === null) {
    window.localStorage.removeItem(REGISTER_CODE_EXPIRES_AT_STORAGE_KEY);
    return;
  }

  window.localStorage.setItem(REGISTER_CODE_EXPIRES_AT_STORAGE_KEY, String(expiresAt));
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

export default function RegisterPage() {
  const register = useAuthStore((state) => state.register);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [verificationCode, setVerificationCode] = useState("");
  const [error, setError] = useState("");
  const [isSendingCode, setIsSendingCode] = useState(false);
  const [countdownExpiresAt, setCountdownExpiresAt] = useState<number | null>(() =>
    readStoredCountdownDeadline(),
  );
  const [countdownSeconds, setCountdownSeconds] = useState(() =>
    getRemainingSeconds(readStoredCountdownDeadline()),
  );
  const redirect = searchParams.get("redirect") || "/";

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

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!isValidEmail(email)) {
      setError("请输入有效的邮箱地址");
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

    setError("");

    try {
      await register({ email, password, confirmPassword, code: verificationCode.trim() });
      persistCountdownDeadline(null);
      navigate(redirect, { replace: true });
    } catch (error) {
      setError(error instanceof Error ? error.message : "注册失败，请稍后重试");
    }
  }

  async function handleSendVerificationCode() {
    if (!isValidEmail(email)) {
      setError("请输入有效的邮箱地址");
      return;
    }

    setError("");
    setIsSendingCode(true);

    try {
      const response = await api.post<ApiResponse<RegisterCodeResponse>>("/api/v1/auth/code", {
        email,
        type: "register",
      });
      const payload = unwrapApiResponse(response.data);
      if (!payload.sent) {
        throw new Error("验证码发送失败，请稍后重试");
      }
      const expiresAt = Date.now() + REGISTER_CODE_COUNTDOWN_SECONDS * 1000;
      persistCountdownDeadline(expiresAt);
      setCountdownExpiresAt(expiresAt);
      setCountdownSeconds(REGISTER_CODE_COUNTDOWN_SECONDS);
    } catch (error) {
      setError(error instanceof Error ? error.message : "验证码发送失败，请稍后重试");
    } finally {
      setIsSendingCode(false);
    }
  }

  const isSendDisabled = isSendingCode || countdownSeconds > 0;
  const sendButtonText =
    countdownSeconds > 0 ? `${countdownSeconds}s后重试` : isSendingCode ? "发送中..." : "发送验证码";

  return (
    <main className="h5-shell h5-shell--auth">
      <section className="auth-card">
        <p className="eyebrow">Create Account</p>
        <h1>注册账号</h1>
        <p className="auth-card__subtitle">完成基础注册后即可发布顺路信息、收藏行程并维护默认联系方式。</p>
        <form className="auth-form" noValidate onSubmit={handleSubmit}>
          <label>
            邮箱
            <input type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
          </label>
          <label>
            密码
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>
          <label>
            确认密码
            <input
              type="password"
              value={confirmPassword}
              onChange={(event) => setConfirmPassword(event.target.value)}
            />
          </label>
          <div className="auth-code">
            <label className="auth-code__label" htmlFor="register-verification-code">
              验证码
            </label>
            <div className="auth-code__control">
              <input
                id="register-verification-code"
                inputMode="numeric"
                placeholder="请输入验证码"
                type="text"
                value={verificationCode}
                onChange={(event) => setVerificationCode(event.target.value)}
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
          {error ? <p role="alert">{error}</p> : null}
          <div className="auth-actions">
            <button className="primary-button" type="submit">
              注册
            </button>
          </div>
        </form>
        <p className="auth-footer">
          已有账号？
          <Link to={`/login?redirect=${encodeURIComponent(redirect)}`}> 去登录</Link>
        </p>
      </section>
    </main>
  );
}
