import { clearAuthSession, getAccessToken } from "../utils/auth";

export interface Package {
  id: string;
  code: string;
  name: string;
  description: string;
  monthly_price_satang: number;
  branch_limit: number;
  user_limit: number;
  staff_management: "basic" | "full";
  reports: "basic" | "advanced";
  export_excel: boolean;
  multi_branch: boolean;
  api_webhook: boolean;
  active: boolean;
  sort_order: number;
}

export type PackageInput = Omit<Package, "id">;
const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

async function request<T>(path: string, method = "GET", body?: PackageInput): Promise<T> {
  const token = getAccessToken();
  if (!token) throw new Error("กรุณาเข้าสู่ระบบใหม่");
  const response = await fetch(`${API_URL}/api/super-admin${path}`, {
    method,
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (response.status === 401) {
    clearAuthSession();
    window.location.assign("/super-admin/login");
    throw new Error("Session หมดอายุ กรุณาเข้าสู่ระบบใหม่");
  }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "ไม่สามารถบันทึกข้อมูลได้");
  return result.data as T;
}

export const listPackages = () => request<Package[]>("/packages");
export const savePackage = (body: PackageInput, id?: string) =>
  request<Package>(id ? `/packages/${id}` : "/packages", id ? "PUT" : "POST", body);
