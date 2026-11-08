import type { TenantProfile } from "../types/tenant/account.type";

export function normalizeTenantProfile(profile: TenantProfile): TenantProfile {
  const role = profile.workspace_role || profile.user?.role;
  if ((role !== "owner" && role !== "staff") || !profile.user?.id || !Array.isArray(profile.tenants)) {
    throw new Error("ข้อมูลสิทธิ์ Tenant ไม่ครบ กรุณารีสตาร์ต backend แล้วลองเข้าสู่ระบบใหม่");
  }
  return {
    ...profile,
    workspace_role: role,
    user: { ...profile.user, role, system_role: "user" },
  };
}

export function getTenantLandingPath(profile: TenantProfile): string {
  if (profile.workspace_role === "owner") return "/tenant/dashboard";
  const shop = profile.tenants.find((item) => item.role === "staff" && item.status === "active");
  return shop ? `/tenant/shops/${shop.id}/bookings` : "/tenant/staff";
}
