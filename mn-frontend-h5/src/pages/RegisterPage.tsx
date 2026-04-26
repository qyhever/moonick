import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import { useAuthStore } from "../store/auth";

export default function RegisterPage() {
  const register = useAuthStore((state) => state.register);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const redirect = searchParams.get("redirect") || "/";

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (password !== confirmPassword) {
      setError("两次输入的密码不一致");
      return;
    }

    setError("");

    try {
      await register({ phone, password, confirmPassword });
      navigate(redirect, { replace: true });
    } catch (error) {
      setError(error instanceof Error ? error.message : "注册失败，请稍后重试");
    }
  }

  return (
    <main>
      <h1>注册</h1>
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
        <label>
          确认密码
          <input
            type="password"
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
          />
        </label>
        {error ? <p role="alert">{error}</p> : null}
        <button type="submit">注册</button>
      </form>
      <Link to={`/login?redirect=${encodeURIComponent(redirect)}`}>去登录</Link>
    </main>
  );
}
