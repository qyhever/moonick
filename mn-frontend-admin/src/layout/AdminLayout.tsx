import { DashboardOutlined, TeamOutlined, CarOutlined, LogoutOutlined } from "@ant-design/icons";
import { Button, Layout, Menu, Space, Typography } from "antd";
import type { MenuProps } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import { useAdminAuthStore } from "../features/auth/store";

const { Header, Content, Sider } = Layout;

const menuItems: MenuProps["items"] = [
  {
    key: "/dashboard",
    icon: <DashboardOutlined />,
    label: "首页",
  },
  {
    key: "/trips",
    icon: <CarOutlined />,
    label: "行程管理",
  },
  {
    key: "/users",
    icon: <TeamOutlined />,
    label: "用户管理",
  },
];

export default function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const admin = useAdminAuthStore((state) => state.admin);
  const logout = useAdminAuthStore((state) => state.logout);

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Sider theme="light" width={232}>
        <div style={{ padding: "24px 20px 12px" }}>
          <Typography.Text strong style={{ color: "#155eef" }}>
            明叶同行 Admin
          </Typography.Text>
        </div>
        <Menu
          items={menuItems}
          mode="inline"
          onClick={({ key }) => navigate(key)}
          selectedKeys={[location.pathname]}
          style={{ borderInlineEnd: 0 }}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            background: "#fff",
            padding: "0 24px",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <Typography.Title level={4} style={{ margin: 0 }}>
            后台管理
          </Typography.Title>
          <Space>
            <Typography.Text>{admin?.name || admin?.username || "管理员"}</Typography.Text>
            <Button
              icon={<LogoutOutlined />}
              onClick={() => {
                logout();
                navigate("/login", { replace: true });
              }}
              type="text"
            >
              退出
            </Button>
          </Space>
        </Header>
        <Content style={{ padding: 24, background: "#f5f7fb" }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
