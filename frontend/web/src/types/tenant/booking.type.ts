import type { Dayjs } from "dayjs";

export type BookingStatus = "pending" | "confirmed" | "completed" | "cancelled";
export interface Booking {
  can_manage: boolean;
  id: string;
  tenant_id: string;
  service_id: string;
  staff_member_id: string;
  customer_name: string;
  customer_phone: string;
  note: string;
  service_name: string;
  staff_name: string;
  duration_minutes: number;
  price_satang: number;
  starts_at: string;
  ends_at: string;
  status: BookingStatus;
  remark: string;
}
export interface BookingService {
  id: string;
  name: string;
  duration_minutes: number;
  price_satang: number;
}
export interface BookingStaff {
  id: string;
  name: string;
  role: "owner" | "staff";
}
export interface BookingList {
  opening_time: string | null;
  closing_time: string | null;
  role: "owner" | "staff";
  shop_name: string;
  shop_status: string;
  items: Booking[];
  services: BookingService[];
  staff: BookingStaff[];
}
export interface CreateBookingRequest {
  service_id: string;
  staff_member_id: string;
  customer_name: string;
  customer_phone: string;
  note: string;
  starts_at: string;
}
export interface BookingForm extends Omit<CreateBookingRequest, "starts_at"> {
  date: Dayjs;
  time: string;
}
export interface BookingDrawerProps {
  openingTime: string | null;
  closingTime: string | null;
  role: "owner" | "staff";
  shopID: string;
  date: string;
  services: BookingService[];
  staff: BookingStaff[];
  open: boolean;
  onClose: () => void;
  onCreated: (date: string) => void;
}
export interface BookingListProps {
  shopID: string;
  date: string;
  onDateChange: (date: string) => void;
}
export interface BookingStatusRequest {
  status: BookingStatus;
  remark?: string;
}
export type BookingCustomerRequest = Pick<Booking, "customer_name" | "customer_phone" | "note">;
export interface BookingUpdateRequest extends BookingCustomerRequest {
  service_id: string;
  starts_at: string;
}
export interface EditBookingForm extends BookingCustomerRequest {
  service_id: string;
  date: Dayjs;
  time: string;
}
export interface EditBookingDrawerProps {
  services: BookingService[];
  openingTime: string | null;
  closingTime: string | null;
  shopID: string;
  booking: Booking;
  onClose: () => void;
  onSaved: (booking: Booking) => void;
}
export interface CancelBookingForm {
  remark: string;
}
export interface CancelBookingDrawerProps {
  shopID: string;
  booking: Booking;
  onClose: () => void;
  onCancelled: (booking: Booking) => void;
}
export interface BookingCalendarProps {
  date: string;
  items: Booking[];
  openingTime: string | null;
  closingTime: string | null;
  onSelect: (booking: Booking) => void;
}
export interface CalendarAppointment {
  booking: Booking;
  start: number;
  end: number;
  lane: number;
}
export const bookingStatuses: Record<BookingStatus, { label: string; color: string }> = {
  pending: { label: "รอยืนยัน", color: "gold" },
  confirmed: { label: "ยืนยันแล้ว", color: "blue" },
  completed: { label: "เสร็จสิ้น", color: "green" },
  cancelled: { label: "ยกเลิก", color: "default" },
};
