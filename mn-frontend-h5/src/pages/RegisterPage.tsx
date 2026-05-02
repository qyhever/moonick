import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { isValidEmail } from "../lib/validation";
import { useAuthStore } from "../store/auth";

export default function RegisterPage() {
  const register = useAuthStore((state) => state.register);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const redirect = searchParams.get("redirect") || "/";

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

    setError("");

    try {
      await register({ email, password, confirmPassword });
      navigate(redirect, { replace: true });
    } catch (error) {
      setError(error instanceof Error ? error.message : "注册失败，请稍后重试");
    }
  }

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
