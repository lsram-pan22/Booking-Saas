import { clearAuthSession, getAccessToken } from "../../utils/auth";
import type { Booking, BookingList, BookingStatus, BookingStatusRequest, BookingCustomerRequest, BookingUpdateRequest, CreateBookingRequest } from "../../types/tenant/booking.type";

async function requestBooking<T>(shopID: string, path: string, method: string, body?: CreateBookingRequest | BookingStatusRequest | BookingCustomerRequest | BookingUpdateRequest, signal?: AbortSignal): Promise<T> {
  const token = getAccessToken();
  if (!token) throw new Error("กรุณาเข้าสู่ระบบใหม่");
  const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080";
  const response = await fetch(`${baseURL}/api/v1/tenant/shops/${encodeURIComponent(shopID)}/bookings${path}`, {
    method, signal, headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: body ? JSON.stringify(body) : undefined,
  });
  if (response.status === 401) { clearAuthSession(); window.location.replace("/tenant/login"); throw new Error("กรุณาเข้าสู่ระบบใหม่"); }
  const result = await response.json();
  if (!response.ok || !result.success) throw new Error(result.message || "ไม่สามารถดำเนินการกับการจองได้");
  return result.data as T;
}
export const getBookings = (shopID: string, date: string, signal?: AbortSignal) => requestBooking<BookingList>(shopID, `?date=${encodeURIComponent(date)}&days=7`, "GET", undefined, signal);
export const createBooking = (shopID: string, body: CreateBookingRequest) => requestBooking<Booking>(shopID, "", "POST", body);
export const setBookingStatus = (shopID: string, id: string, status: BookingStatus, remark?: string) => requestBooking<Booking>(shopID, `/${encodeURIComponent(id)}/status`, "PATCH", { status, remark });
export const updateBooking = (shopID: string, id: string, body: BookingUpdateRequest) => requestBooking<Booking>(shopID, `/${encodeURIComponent(id)}`, "PATCH", body);
