import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Alert, Button, Card, Form, Input, Typography } from "antd";

import { useAdminAuthStore } from "./store";

type LoginFormValues = {
  username: string;
  password: string;
};

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAdminAuthStore((state) => state.login);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const redirectTo = new URLSearchParams(location.search).get("redirect") || "/dashboard";

  async function handleFinish(values: LoginFormValues) {
    setSubmitting(true);
    setError("");

    try {
      await login(values);
      navigate(redirectTo, { replace: true });
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "登录失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        background:
          "radial-gradient(circle at top, rgba(21, 94, 239, 0.14), transparent 28%), #f5f7fb",
        padding: 24,
      }}
    >
      <Card style={{ width: 420, borderRadius: 20 }}>
        <Typography.Paragraph style={{ marginBottom: 8, color: "#155eef", letterSpacing: 1 }}>
          明叶同行 Admin
        </Typography.Paragraph>
        <Typography.Title level={2} style={{ marginTop: 0 }}>
          管理员登录
        </Typography.Title>

        <Form layout="vertical" onFinish={handleFinish}>
          <Form.Item label="用户名" name="username" rules={[{ required: true, message: "请输入用户名" }]}>
            <Input autoComplete="username" placeholder="请输入管理员用户名" />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true, message: "请输入密码" }]}>
            <Input.Password autoComplete="current-password" placeholder="请输入密码" />
          </Form.Item>

          {error ? <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} /> : null}

          <Button block htmlType="submit" loading={submitting} size="large" type="primary">
            登录
          </Button>
        </Form>
      </Card>
    </div>
  );
}
