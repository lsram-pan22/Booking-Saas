import { Button, Flex, Menu, Typography } from "antd";
import { CalendarOutlined, AppstoreOutlined, CreditCardOutlined, SettingOutlined, ShopOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from "@ant-design/icons";
import { useLocation, useMatch, useNavigate } from "react-router-dom";
import type { TenantSidebarProps } from "../../../types/tenant/layout.type";

export default function TenantSidebar({ collapsed = false, onNavigate, onToggle, profile }: TenantSidebarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const shopMatch = useMatch("/tenant/shops/:shopID/*");
  const staffShop = profile?.tenants.find(shop => shop.role === "staff" && shop.status === "active");
  return <>
    <Flex align="center" justify={collapsed ? "center" : "space-between"} style={{ height: 64, padding: collapsed ? 0 : "0 16px" }}>
      {!collapsed && <Typography.Title level={4} style={{ margin: 0, color: "white", whiteSpace: "nowrap" }}>Booking SaaS</Typography.Title>}
      {onToggle && <Button type="text" aria-label={collapsed ? "ขยายเมนู" : "ย่อเมนู"} icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={onToggle} style={{ color: "white", width: 40, height: 40, fontSize: 18 }} />}
    </Flex>
    <Menu theme="dark" mode="inline" inlineCollapsed={collapsed} selectedKeys={[location.pathname]} style={{ background: "#7F867B", borderInlineEnd: 0 }} onClick={({ key }) => { navigate(key); onNavigate?.(); }} items={!profile ? [] : profile.workspace_role === "staff" ? [
      ...(staffShop ? [{ key: `/tenant/shops/${staffShop.id}/bookings`, label: "การจอง", icon: <CalendarOutlined /> }] : []),
      { type: "divider" },
      { key: "/tenant/settings", label: "ตั้งค่าบัญชี", icon: <SettingOutlined /> },
    ] : shopMatch ? [
      { key: `/tenant/shops/${shopMatch.params.shopID}`, label: "ข้อมูลร้านค้า", icon: <ShopOutlined /> },
      { type: "divider" },
      { key: `/tenant/shops/${shopMatch.params.shopID}/bookings`, label: "การจอง", icon: <CalendarOutlined /> },
      { key: `/tenant/shops/${shopMatch.params.shopID}/services`, label: "บริการ", icon: <AppstoreOutlined /> },
    ] : [
      { key: "/tenant/dashboard", label: "ร้านค้าของฉัน", icon: <ShopOutlined /> },
      { type: "divider" },
      { key: "/tenant/packages", label: "แพ็กเกจ", icon: <CreditCardOutlined /> },
      { key: "/tenant/settings", label: "ตั้งค่า", icon: <SettingOutlined /> },
    ]} />
  </>;
}
