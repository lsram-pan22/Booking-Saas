import { Alert, Button, Flex, Tag, Typography, theme } from "antd";
import dayjs from "dayjs";
import { bookingStatuses } from "../../../types/tenant/booking.type";
import type { BookingCalendarProps, CalendarAppointment } from "../../../types/tenant/booking.type";

const minutes = (value: string) => { const [hour, minute] = value.split(":").map(Number); return hour * 60 + minute; };
const clock = (value: number) => `${String(Math.floor(value / 60)).padStart(2, "0")}:${String(value % 60).padStart(2, "0")}`;

export default function BookingCalendar({ date, items, openingTime, closingTime, onSelect }: BookingCalendarProps) {
  const { token } = theme.useToken();
  const fromMinute = minutes(openingTime || "00:00");
  const toMinute = minutes(closingTime || "24:00");
  const range = toMinute - fromMinute;
  const timelineWidth = Math.max(720, range * 2);
  const ticks = Array.from({ length: Math.ceil(range / 60) }, (_, index) => fromMinute + index * 60);
  const days = Array.from({ length: 7 }, (_, index) => {
    const day = dayjs(date).add(index, "day");
    const midnight = new Date(`${day.format("YYYY-MM-DD")}T00:00:00+07:00`).getTime();
    const opening = midnight + fromMinute * 60000;
    const closing = midnight + toMinute * 60000;
    const daily = items.filter(item => item.status !== "cancelled" && new Date(item.starts_at).getTime() < midnight + 86400000 && new Date(item.ends_at).getTime() > midnight);
    const laneEnds: number[] = [];
    const appointments: CalendarAppointment[] = daily.map(booking => ({ booking, start: Math.max(opening, new Date(booking.starts_at).getTime()), end: Math.min(closing, new Date(booking.ends_at).getTime()), lane: 0 }))
      .filter(item => item.end > item.start).sort((a, b) => a.start - b.start || a.end - b.end || a.booking.id.localeCompare(b.booking.id)).map(item => {
        let lane = laneEnds.findIndex(end => end <= item.start);
        if (lane === -1) lane = laneEnds.length;
        laneEnds[lane] = item.end;
        return { ...item, lane };
      });
    const outside = daily.filter(item => new Date(item.starts_at).getTime() < opening || new Date(item.ends_at).getTime() > closing).length;
    return { day, opening, appointments, outside, laneCount: laneEnds.length, height: Math.max(76, laneEnds.length * 64 + 4) };
  });
  return <>
    <Flex gap={8} wrap style={{ marginBottom: 16 }}>{Object.entries(bookingStatuses).filter(([status]) => status !== "cancelled").map(([status, info]) => <Tag key={status} color={info.color}>{info.label}</Tag>)}<Typography.Text type="secondary">เวลาไทย · {openingTime || "00:00"}–{closingTime || "24:00"} · คลิกการจองเพื่อดูรายละเอียด</Typography.Text></Flex>
    {(!openingTime || !closingTime) && <Alert type="info" showIcon title="ยังไม่ได้ตั้งเวลาเปิด–ปิดร้าน จึงแสดงเวลา 00:00–24:00" description="เจ้าของร้านตั้งค่าได้ที่ ข้อมูลร้านค้า → แก้ไขข้อมูลร้านค้า" style={{ marginBottom: 16 }} />}
    {days.some(day => day.outside > 0) && <Alert type="warning" showIcon title="มีการจองเดิมอยู่นอกเวลาเปิด–ปิดปัจจุบัน" description="ปฏิทินแสดงเฉพาะส่วนที่อยู่ในเวลาร้าน ดูข้อมูลทั้งหมดได้ในมุมมองรายการ" style={{ marginBottom: 16 }} />}
    <div role="region" aria-label="ปฏิทินการจอง 7 วัน วันแนวตั้ง เวลาแนวนอน" tabIndex={0} style={{ overflow: "auto", maxHeight: "70vh", border: `1px solid ${token.colorBorderSecondary}`, borderRadius: token.borderRadius }}>
      <div style={{ minWidth: timelineWidth + 150 }}>
        <Flex style={{ position: "sticky", top: 0, zIndex: 4, height: 44, background: token.colorBgContainer, borderBottom: `1px solid ${token.colorBorderSecondary}` }}>
          <div style={{ width: 150, flexShrink: 0, padding: 12, boxSizing: "border-box", position: "sticky", left: 0, background: token.colorBgContainer, zIndex: 5 }}>วันที่ / เวลา</div>
          <div style={{ flex: 1, position: "relative", minWidth: timelineWidth }}>
            {ticks.map(tick => <Typography.Text key={tick} type="secondary" style={{ position: "absolute", left: `${(tick - fromMinute) / range * 100}%`, top: 12, paddingLeft: 6 }}>{clock(tick)}</Typography.Text>)}
            <Typography.Text type="secondary" style={{ position: "absolute", right: 4, top: 12 }}>{clock(toMinute)}</Typography.Text>
          </div>
        </Flex>
        {days.map(({ day, opening, appointments, height, outside, laneCount }) => <Flex key={day.format("YYYY-MM-DD")} style={{ borderBottom: `1px solid ${token.colorBorderSecondary}` }}>
          <div style={{ width: 150, flexShrink: 0, padding: "14px", boxSizing: "border-box", position: "sticky", left: 0, zIndex: 2, background: token.colorBgContainer }}>
            <Typography.Text strong>{["อาทิตย์", "จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์"][day.day()]}</Typography.Text><br /><Typography.Text>{day.format("DD/MM/YYYY")}</Typography.Text>
            {outside > 0 && <div><Typography.Text type="warning">นอกเวลา {outside} รายการ</Typography.Text></div>}
          </div>
          <div style={{ flex: 1, minWidth: timelineWidth, height, position: "relative", background: token.colorBgContainer }}>
            {Array.from({ length: Math.ceil(range / 30) }, (_, index) => <div key={index} style={{ position: "absolute", left: `${index * 30 / range * 100}%`, top: 0, bottom: 0, borderLeft: `1px ${index % 2 ? "dashed" : "solid"} ${token.colorBorderSecondary}` }} />)}
            {appointments.map(({ booking, start, end, lane }) => {
              const color = booking.status === "pending" ? token.colorWarning : booking.status === "confirmed" ? token.colorPrimary : token.colorSuccess;
              const background = booking.status === "pending" ? token.colorWarningBg : booking.status === "confirmed" ? token.colorPrimaryBg : token.colorSuccessBg;
              const label = `ลูกค้า : ${booking.customer_name} · คนรับงาน : ${booking.staff_name} · รายการ : ${booking.service_name}`;
              return <Button key={booking.id} aria-label={label} title={label} onClick={() => onSelect(booking)} style={{ position: "absolute", top: (height - laneCount * 64) / 2 + lane * 64 + 2, left: `${(start - opening) / 60000 / range * 100}%`, width: `${(end - start) / 60000 / range * 100}%`, minWidth: 2, height: 60, padding: "2px 6px", fontSize: 12, lineHeight: "18px", display: "flex", flexDirection: "column", justifyContent: "center", gap: 0, overflow: "hidden", textAlign: "center", borderColor: color, borderLeftWidth: 4, background, color: token.colorText }}>
                <div style={{ width: "100%", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}><strong>ลูกค้า: {booking.customer_name}</strong></div>
                <div style={{ width: "100%", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>คนรับงาน: {booking.staff_name}</div>
                <div style={{ width: "100%", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>รายการ: {booking.service_name}</div>
              </Button>;
            })}
          </div>
        </Flex>)}
      </div>
    </div>
  </>;
}
