import { clearAuthSession, getAccessToken } from "../../utils/auth";
import type { PlanOverview, PlanSelection } from "../../types/tenant/plan.type";
const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";
async function request<T>(path: string, body?: PlanSelection): Promise<T> {
  const response = await fetch(`${baseURL}/api/v1/tenant/subscription${path}`, { method: body ? "PUT" : "GET", headers: { Authorization: `Bearer ${getAccessToken() || ""}`, "Content-Type": "application/json" }, body: body ? JSON.stringify(body) : undefined });
  if (response.status === 401) { clearAuthSession(); window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบใหม่"); }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "จัดการแพ็กเกจไม่สำเร็จ");
  return result.data as T;
}
export const getPlanOverview = () => request<PlanOverview>("");
export const selectFreePlanScope = (body: PlanSelection) => request<void>("/selection", body);
