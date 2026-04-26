import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { useAuthStore } from "../store/auth";

export default function LoginPage() {
  const login = useAuthStore((state) => state.login);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const redirect = searchParams.get("redirect") || "/";

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    try {
      await login({ phone, password });
      navigate(redirect, { replace: true });
    } catch (error) {
      setError(error instanceof Error ? error.message : "登录失败，请稍后重试");
    }
  }

  return (
    <main>
      <h1>手机号登录</h1>
      <form onSubmit={handleSubmit}>
        <label>
          手机号
          <input value={phone} onChange={(event) => setPhone(event.target.value)} />
        </label>
        <label>
          密码
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>
        {error ? <p role="alert">{error}</p> : null}
        <button type="submit">登录</button>
      </form>
      <Link to={`/register?redirect=${encodeURIComponent(redirect)}`}>去注册</Link>
    </main>
  );
}
