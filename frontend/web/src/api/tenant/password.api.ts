import { clearAuthSession, getAccessToken } from "../../utils/auth";
import type { ResettableOwner } from "../../types/shared/password.type";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";
async function request<T>(path: string, body?: object): Promise<T> {
  const response = await fetch(`${API_URL}/api/v1${path}`, {
    method: body ? "POST" : "GET",
    headers: { Authorization: `Bearer ${getAccessToken() || ""}`, "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (response.status === 401) {
    clearAuthSession();
    window.location.assign(path.startsWith("/super-admin") ? "/super-admin/login" : "/tenant/login");
    throw new Error("กรุณาเข้าสู่ระบบใหม่");
  }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "ดำเนินการไม่สำเร็จ");
  return result.data as T;
}
export const resetStaffPassword = (shopID: string, userID: string, password: string) => request<void>(`/tenant/shops/${shopID}/users/${userID}/password/reset`, { password });
export const resetOwnerPassword = (userID: string, password: string) => request<void>(`/super-admin/users/${userID}/password/reset`, { password });
export const listResettableOwners = (search = "") => request<ResettableOwner[]>(`/super-admin/users?search=${encodeURIComponent(search)}`);
export const changePassword = (current_password: string, new_password: string) => request<void>("/tenant/password/change", { current_password, new_password });
