import { clearAuthSession, getAccessToken } from "../../utils/auth";
import type { ServiceList, ServiceRequest, ShopService } from "../../types/tenant/service.type";

async function requestService<T>(shopID: string, method: string, serviceID?: string, body?: ServiceRequest, signal?: AbortSignal): Promise<T> {
  const token = getAccessToken();
  if (!token) throw new Error("กรุณาเข้าสู่ระบบใหม่");
  const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";
  const response = await fetch(`${baseURL}/api/v1/tenant/shops/${encodeURIComponent(shopID)}/services${serviceID ? `/${encodeURIComponent(serviceID)}` : ""}`, {
    method, signal, headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: body ? JSON.stringify(body) : undefined,
  });
  if (response.status === 401) { clearAuthSession(); window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบใหม่"); }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "ไม่สามารถดำเนินการกับบริการได้");
  return result.data as T;
}

export const getServices = (shopID: string, signal?: AbortSignal) => requestService<ServiceList>(shopID, "GET", undefined, undefined, signal);
export const saveService = (shopID: string, body: ServiceRequest, serviceID?: string) => requestService<ShopService>(shopID, serviceID ? "PATCH" : "POST", serviceID, body);
