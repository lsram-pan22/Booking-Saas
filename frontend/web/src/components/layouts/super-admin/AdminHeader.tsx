
import { App, Layout, Avatar, Typography, Space, Dropdown } from "antd";
import { useState } from "react";
import {
  UserOutlined,
  DownOutlined,
  LogoutOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { getCurrentUser, logout } from "../../../utils/auth";

const { Header } = Layout;
const { Text } = Typography;

export default function AdminHeader() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [signingOut, setSigningOut] = useState(false);
  const user = getCurrentUser();

  const handleLogout = async () => {
    setSigningOut(true);
    try { await logout(); navigate("/super-admin/login", { replace: true }); }
    catch (error) { message.error(error instanceof Error ? error.message : "ออกจากระบบไม่สำเร็จ"); }
    finally { setSigningOut(false); }
  };

  return (
    <Header
      style={{
        background: "#7F867B",
        padding: "0 24px",
        display: "flex",
        justifyContent: "flex-end",
        alignItems: "center",
      }}
    >
      <Dropdown
        menu={{
          items: [
            {
              key: "profile",
              label: "Profile",
              icon: <UserOutlined />,
            },
            {
              type: "divider",
            },
            {
              key: "logout",
              label: "ออกจากระบบทุกอุปกรณ์",
              disabled: signingOut,
              icon: <LogoutOutlined />,
              danger: true,
              onClick: handleLogout,
            },
          ],
        }}
        placement="bottomRight"
        trigger={["click"]}
      >
        <Space style={{ cursor: "pointer" }}>
          <Avatar icon={<UserOutlined />} />

          <Text strong style={{ color: "#FFFFFF" }}>
            {user?.full_name || "Super Admin"}
          </Text>

          <DownOutlined style={{ color: "#FFFFFF" }} />
        </Space>
      </Dropdown>
    </Header>
  );
}
