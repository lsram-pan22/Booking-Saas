import { Avatar, Button, Dropdown, Flex, Layout, Space, Typography } from "antd";
import { DownOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined, UserOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { getCurrentUser, logout } from "../../../utils/auth";
import type { TenantHeaderProps } from "../../../types/tenant/layout.type";

export default function TenantHeader({ collapsed, onToggle, showToggle = true, user: profileUser }: TenantHeaderProps) {
  const navigate = useNavigate();
  const user = profileUser || getCurrentUser();
  return <Layout.Header style={{ background: "#7F867B", padding: "0 16px", flexShrink: 0 }}>
    <Flex align="center" justify={showToggle ? "space-between" : "flex-end"} gap={12} style={{ height: "100%" }}>
      {showToggle && <Button type="text" aria-label={collapsed ? "เปิดเมนู" : "ย่อเมนู"} icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={onToggle} style={{ color: "white", fontSize: 18 }} />}
      <Flex align="center" gap={12} style={{ minWidth: 0 }}>
      <Dropdown placement="bottomRight" trigger={["click"]} menu={{ items: [{ key: "logout", label: "ออกจากระบบ", icon: <LogoutOutlined />, danger: true }], onClick: ({ key }) => { if (key === "logout") { logout(); navigate("/tenant/login", { replace: true }); } } }}>
        <Button type="text" style={{ color: "white", maxWidth: "80%" }}>
          <Space style={{ minWidth: 0 }}><Avatar size="small" icon={<UserOutlined />} /><Typography.Text ellipsis style={{ color: "white", maxWidth: 180 }}>{user?.full_name || "บัญชีของฉัน"}</Typography.Text><DownOutlined /></Space>
        </Button>
      </Dropdown>
      </Flex>
    </Flex>
  </Layout.Header>;
}
