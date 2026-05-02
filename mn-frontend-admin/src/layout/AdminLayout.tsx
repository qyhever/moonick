import { DashboardOutlined, TeamOutlined, CarOutlined, LogoutOutlined } from "@ant-design/icons";
import { useRef } from "react";
import { Button, FloatButton, Layout, Menu, Modal, Space, Typography } from "antd";
import type { MenuProps } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import { useAdminAuthStore } from "../features/auth/store";

const { Header, Content, Sider } = Layout;
const siderWidth = 232;
const headerHeight = 64;

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
  const contentRef = useRef<HTMLDivElement | null>(null);
  const [modal, contextHolder] = Modal.useModal();
  const admin = useAdminAuthStore((state) => state.admin);
  const logout = useAdminAuthStore((state) => state.logout);

  const handleLogoutConfirm = () => {
    modal.confirm({
      title: "确认退出登录吗？",
      okText: "确认退出",
      cancelText: "取消",
      onOk: () => {
        logout();
        navigate("/login", { replace: true });
      },
    });
  };

  return (
    <Layout style={{ minHeight: "100vh", overflow: "hidden" }}>
      {contextHolder}
      <Sider
        theme="light"
        width={siderWidth}
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          bottom: 0,
          height: "100vh",
          overflow: "auto",
        }}
      >
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

      <Layout style={{ marginLeft: siderWidth, minHeight: "100vh" }}>
        <Header
          style={{
            background: "#fff",
            padding: "0 24px",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            position: "fixed",
            top: 0,
            left: siderWidth,
            right: 0,
            height: headerHeight,
            zIndex: 10,
          }}
        >
          <Typography.Title level={4} style={{ margin: 0 }}>
            后台管理
          </Typography.Title>
          <Space>
            <Typography.Text>{admin?.name || admin?.username || "管理员"}</Typography.Text>
            <Button icon={<LogoutOutlined />} onClick={handleLogoutConfirm} type="text">
              退出
            </Button>
          </Space>
        </Header>
        <Content
          data-testid="admin-layout-content"
          ref={contentRef}
          style={{
            padding: 24,
            background: "#f5f7fb",
            marginTop: headerHeight,
            height: `calc(100vh - ${headerHeight}px)`,
            overflowY: "auto",
          }}
        >
          <Outlet />
          <FloatButton.BackTop target={() => contentRef.current ?? window} />
        </Content>
      </Layout>
    </Layout>
  );
}
