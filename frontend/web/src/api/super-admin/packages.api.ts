import type { Package, PackageInput } from "../../types/super-admin/package.type";
import { clearAuthSession, getAccessToken } from "../../utils/auth";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

async function request<T>(path: string, method = "GET", body?: PackageInput): Promise<T> {
  const token = getAccessToken();
  if (!token) throw new Error("กรุณาเข้าสู่ระบบใหม่");
  const response = await fetch(`${API_URL}/api/v1/super-admin${path}`, {
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
