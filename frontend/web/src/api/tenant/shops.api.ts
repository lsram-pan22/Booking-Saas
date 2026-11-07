import { clearAuthSession, getAccessToken } from "../../utils/auth";
import type { AddShopMemberRequest, CreateShopRequest, UpdateShopRequest, ShopDetails, ShopMember, ShopMembers, MemberStatusRequest, MemberStatusResult } from "../../types/tenant/shop.type";

async function requestShop<T = ShopDetails>(path: string, method: string, body?: CreateShopRequest | AddShopMemberRequest | MemberStatusRequest, signal?: AbortSignal): Promise<T> {
  const token = getAccessToken();
  if (!token) throw new Error("กรุณาเข้าสู่ระบบใหม่");
  const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";
  const response = await fetch(`${baseURL}/api/v1/tenant/shops${path}`, {
    method, headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: body ? JSON.stringify(body) : undefined, signal,
  });
  if (response.status === 401) { clearAuthSession(); window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบใหม่"); }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "ไม่สามารถดำเนินการกับร้านค้าได้");
  return result.data as T;
}

export const createShop = (body: CreateShopRequest) => requestShop("", "POST", body);
export const getShop = (id: string, signal?: AbortSignal) => requestShop(`/${encodeURIComponent(id)}`, "GET", undefined, signal);
export const updateShop = (id: string, body: UpdateShopRequest) => requestShop(`/${encodeURIComponent(id)}`, "PATCH", body);
export const getShopMembers = (id: string, signal?: AbortSignal) => requestShop<ShopMembers>(`/${encodeURIComponent(id)}/members`, "GET", undefined, signal);
export const addShopMember = (id: string, body: AddShopMemberRequest) => requestShop<ShopMember>(`/${encodeURIComponent(id)}/members`, "POST", body);
export const setShopMemberStatus = (shopID: string, memberID: string, body: MemberStatusRequest) => requestShop<MemberStatusResult>(`/${encodeURIComponent(shopID)}/members/${encodeURIComponent(memberID)}`, "PATCH", body);
