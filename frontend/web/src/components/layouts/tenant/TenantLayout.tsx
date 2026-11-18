import { useEffect, useState } from "react";
import { Alert, Drawer, Grid, Layout, Spin } from "antd";
import { Link, Navigate, Outlet } from "react-router-dom";
import TenantHeader from "./TenantHeader";
import TenantSidebar from "./TenantSidebar";
import { getTenantProfile } from "../../../api/tenant/account.api";
import type { TenantProfile, TenantSettingsResult } from "../../../types/tenant/account.type";
import { updateCurrentUser } from "../../../utils/auth";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";
import type { TenantShop } from "../../../types/tenant/shop.type";

export default function TenantLayout() {
  const screens = Grid.useBreakpoint();
  const [collapsed, setCollapsed] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [profile, setProfile] = useState<TenantProfile>();
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    getTenantProfile(controller.signal).then((data) => {
      if (!controller.signal.aborted) {
        updateCurrentUser(data.user);
        setProfile(data);
      }
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "โหลดข้อมูลไม่สำเร็จ");
    });
    return () => controller.abort();
  }, []);
  const desktop = Boolean(screens.lg);
  const updateUser = (user: TenantSettingsResult) => {
    if (!profile) return;
    const updatedUser = { ...profile.user, ...user };
    updateCurrentUser(updatedUser);
    setProfile((current) => current ? { ...current, user: { ...current.user, ...user } } : current);
  };
  const addShop = (shop: TenantShop) => {
    setProfile((current) => current ? { ...current, enabled_shop_count: current.enabled_shop_count + (current.tenants.some((item) => item.id === shop.id) ? 0 : 1), owned_shop_count: current.owned_shop_count + (current.tenants.some((item) => item.id === shop.id) ? 0 : 1), tenants: [...current.tenants.filter((item) => item.id !== shop.id), shop] } : current);
  };
  const updateShop = (shop: TenantShop) => {
    setProfile((current) => current ? { ...current, tenants: current.tenants.map((item) => item.id === shop.id ? shop : item) } : current);
  };
  if (profile?.must_change_password) return <Navigate to="/tenant/change-password" replace />;
  return <Layout style={{ position: "fixed", inset: 0, height: "100dvh", overflow: "hidden" }}>
    {desktop && <Layout.Sider width={250} collapsedWidth={80} collapsed={collapsed} trigger={null} style={{ background: "#7F867B", overflowY: "auto" }}><TenantSidebar profile={profile} collapsed={collapsed} onToggle={() => setCollapsed((value) => !value)} /></Layout.Sider>}
    <Drawer title="เมนูร้านค้า" placement="left" size={250} open={!desktop && menuOpen} onClose={() => setMenuOpen(false)} styles={{ body: { padding: 0, background: "#7F867B" } }}><TenantSidebar profile={profile} onNavigate={() => setMenuOpen(false)} /></Drawer>
    <Layout style={{ minWidth: 0 }}>
      {profile?.subscription?.selection_required && <Alert type="warning" showIcon title="แพ็กเกจเดิมหมดอายุ ตอนนี้ใช้ Free กรุณาเลือกร้านและพนักงานภายในโควตา" action={<Link to="/tenant/packages">เลือกสิทธิ์ Free</Link>} />}
      <TenantHeader user={profile?.user} subscription={profile?.workspace_role === "owner" ? profile.subscription : null} showToggle={!desktop} collapsed={!menuOpen} onToggle={() => setMenuOpen((value) => !value)} />
      <Layout.Content style={{ padding: screens.md ? 24 : 16, overflow: "auto", minHeight: 0, background: "#F3F5F2" }}>{profile ? <Outlet context={{ profile, error, updateUser, addShop, updateShop } satisfies TenantLayoutContext} /> : error ? <Alert type="error" showIcon title={error} /> : <Spin />}</Layout.Content>
    </Layout>
  </Layout>;
}
