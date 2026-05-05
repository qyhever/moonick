import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { isValidEmail } from "../lib/validation";
import { useAuthStore } from "../store/auth";

export default function LoginPage() {
  const login = useAuthStore((state) => state.login);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const redirect = searchParams.get("redirect") || "/";
  const trimmedEmail = email.trim();
  const forgotPasswordTarget = trimmedEmail
    ? `/password-reset?email=${encodeURIComponent(trimmedEmail)}`
    : "/password-reset";

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!isValidEmail(email)) {
      setError("请输入有效的邮箱地址");
      return;
    }

    setError("");

    try {
      await login({ email, password });
      navigate(redirect, { replace: true });
    } catch (error) {
      setError(error instanceof Error ? error.message : "登录失败，请稍后重试");
    }
  }

  return (
    <main className="h5-shell h5-shell--auth">
      <section className="auth-card">
        <p className="eyebrow">Welcome Back</p>
        <h1>邮箱登录</h1>
        <p className="auth-card__subtitle">继续使用明叶同行，优先查看附近最新行程、收藏记录和个人发布。</p>
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
          <div className="auth-form__meta">
            <Link className="auth-form__meta-link" to={forgotPasswordTarget}>
              忘记密码？
            </Link>
          </div>
          {error ? <p role="alert">{error}</p> : null}
          <div className="auth-actions">
            <button className="primary-button" type="submit">
              登录
            </button>
          </div>
        </form>
        <p className="auth-footer">
          还没有账号？
          <Link to={`/register?redirect=${encodeURIComponent(redirect)}`}> 去注册</Link>
        </p>
      </section>
    </main>
  );
}
