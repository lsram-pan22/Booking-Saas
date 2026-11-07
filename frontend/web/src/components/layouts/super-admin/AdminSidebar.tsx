
import { Button, Flex, Layout, Menu, Typography } from "antd";
import {
  DashboardOutlined,
  ShopOutlined,
  UserOutlined,
  AppstoreOutlined,
  CreditCardOutlined,
  BarChartOutlined,
  SettingOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined,
} from "@ant-design/icons";
import { useLocation, useNavigate } from "react-router-dom";
import type { AdminSidebarProps } from "../../../types/super-admin/layout.type";

const { Sider } = Layout;
const { Title } = Typography;

const menuItems = [
  {
    key: "/super-admin/dashboard",
    icon: <DashboardOutlined />,
    label: "Dashboard",
  },
  {
    key: "/super-admin/businesses",
    icon: <ShopOutlined />,
    label: "Businesses",
  },
  {
    key: "/super-admin/users",
    icon: <UserOutlined />,
    label: "Users",
  },
  {
    key: "/super-admin/packages",
    icon: <AppstoreOutlined />,
    label: "Packages",
  },
  {
    key: "/super-admin/subscriptions",
    icon: <CreditCardOutlined />,
    label: "Subscriptions",
  },
  {
    key: "/super-admin/reports",
    icon: <BarChartOutlined />,
    label: "Reports",
  },
  {
    key: "/super-admin/settings",
    icon: <SettingOutlined />,
    label: "Settings",
  },
];

export default function AdminSidebar({
  collapsed,
  onToggle,
}: AdminSidebarProps) {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <Sider
      theme="dark"
      width={250}
      collapsedWidth={80}
      collapsible
      trigger={null}
      collapsed={collapsed}
      style={{
        background: "#7F867B",
      }}
    >
      <Flex
        align="center"
        justify={collapsed ? "center" : "space-between"}
        style={{
          height: 64,
          padding: collapsed ? 0 : "0 16px",
        }}
      >
        {!collapsed && (
          <Title
            level={4}
            style={{
              color: "#ffffff",
              margin: 0,
              whiteSpace: "nowrap",
            }}
          >
            Booking SaaS
          </Title>
        )}

        <Button
          type="text"
          icon={
            collapsed
              ? <MenuUnfoldOutlined />
              : <MenuFoldOutlined />
          }
          onClick={onToggle}
          style={{
            color: "#ffffff",
            width: 40,
            height: 40,
            fontSize: 18,
          }}
        />
      </Flex>

      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
        onClick={({ key }) => navigate(key)}
        style={{
          background: "#7F867B",
          borderInlineEnd: 0,
        }}
      />
    </Sider>
  );
}
