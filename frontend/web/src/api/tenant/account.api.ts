import type { TenantRegistration, TenantRegistrationResult } from "../../types/tenant/account.type";
import type { TenantLoginRequest, TenantProfile, TenantSession } from "../../types/tenant/account.type";
import { clearAuthSession, getAccessToken } from "../../utils/auth";
import type { TenantSettingsForm, TenantSettingsResult } from "../../types/tenant/account.type";
import { normalizeTenantProfile } from "../../utils/tenant-access";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export async function updateTenantProfile(body: TenantSettingsForm): Promise<TenantSettingsResult> {
  const token = getAccessToken();
  if (!token) throw new Error("กรุณาเข้าสู่ระบบใหม่");
  const response = await fetch(`${API_URL}/api/v1/tenant/me`, {
    method: "PATCH", headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: JSON.stringify(body),
  });
  if (response.status === 401 || response.status === 403) {
    clearAuthSession(); window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบใหม่");
  }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "บันทึกข้อมูลไม่สำเร็จ");
  return result.data as TenantSettingsResult;
}

export async function loginTenant(body: TenantLoginRequest): Promise<TenantSession> {
  const response = await fetch(`${API_URL}/api/v1/tenant/login`, {
    method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body),
  });
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "เข้าสู่ระบบไม่สำเร็จ");
  if (!result.data?.access_token) throw new Error("ไม่พบข้อมูล session");
  return { ...result.data, ...normalizeTenantProfile(result.data) } as TenantSession;
}

export async function getTenantProfile(signal?: AbortSignal): Promise<TenantProfile> {
  const token = getAccessToken();
  if (!token) { window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบ"); }
  const response = await fetch(`${API_URL}/api/v1/tenant/me`, { headers: { Authorization: `Bearer ${token}` }, signal });
  if (response.status === 401 || response.status === 403) {
    clearAuthSession(); window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบใหม่");
  }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "โหลดข้อมูลร้านค้าไม่สำเร็จ");
  return normalizeTenantProfile(result.data as TenantProfile);
}

export async function registerTenant(body: TenantRegistration): Promise<TenantRegistrationResult> {
  const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";
  const response = await fetch(`${baseURL}/api/v1/tenant/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "สมัครสมาชิกไม่สำเร็จ");
  return result.data as TenantRegistrationResult;
}
