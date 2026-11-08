import { useEffect, useState } from "react";
import { Alert, App, Button, Card, Col, DatePicker, Descriptions, Drawer, Flex, Popconfirm, Row, Segmented, Space, Table, Tag, Typography } from "antd";
import dayjs from "dayjs";
import { EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { useParams } from "react-router-dom";
import { getBookings, setBookingStatus } from "../../../api/tenant/bookings.api";
import BookingDrawer from "../../../components/bookings/tenant/BookingDrawer";
import CancelBookingDrawer from "../../../components/bookings/tenant/CancelBookingDrawer";
import EditBookingDrawer from "../../../components/bookings/tenant/EditBookingDrawer";
import BookingCalendar from "../../../components/bookings/tenant/BookingCalendar";
import { bookingStatuses } from "../../../types/tenant/booking.type";
import type { Booking, BookingList, BookingListProps, BookingStatus } from "../../../types/tenant/booking.type";

const localDate = (date = new Date()) => new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Bangkok", year: "numeric", month: "2-digit", day: "2-digit" }).format(date);
const formatTime = (value: string) => new Date(value).toLocaleString("en-GB", { timeZone: "Asia/Bangkok", day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit", hour12: false });

function WeeklyBookings({ shopID, date, onDateChange }: BookingListProps) {
  const [data, setData] = useState<BookingList>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const [open, setOpen] = useState(false);
  const [savingID, setSavingID] = useState("");
  const [cancelling, setCancelling] = useState<Booking>();
  const [editingBooking, setEditingBooking] = useState<Booking>();
  const [view, setView] = useState("calendar");
  const [selectedID, setSelectedID] = useState("");
  const selected = data?.items.find(item => item.id === selectedID);
  const { message } = App.useApp();
  useEffect(() => {
    const controller = new AbortController();
    getBookings(shopID, date, controller.signal).then(result => { if (!controller.signal.aborted) setData(result); }).catch(e => {
      if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "โหลดการจองไม่สำเร็จ");
    }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [shopID, date, retry]);
  const reload = () => { setLoading(true); setError(""); setRetry(value => value + 1); };
  const canManage = data?.shop_status === "active" && !error && !loading;
  const changeStatus = async (booking: Booking, status: BookingStatus) => {
    if (!canManage || !booking.can_manage || savingID) return;
    setSavingID(booking.id);
    try {
      const updated = await setBookingStatus(shopID, booking.id, status);
      setData(previous => previous ? { ...previous, items: previous.items.map(item => item.id === updated.id ? updated : item) } : previous);
      message.success("เปลี่ยนสถานะเรียบร้อยแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "เปลี่ยนสถานะไม่สำเร็จ"); reload(); }
    finally { setSavingID(""); }
  };
  const editButton = (item: Booking) => item.can_manage && canManage && ["pending", "confirmed"].includes(item.status) ? <Button type="text" icon={<EditOutlined />} aria-label="แก้ไขการจอง" title="แก้ไขการจอง" disabled={!!savingID} onClick={() => { setSelectedID(""); setEditingBooking(item); }} /> : null;
  const actions = (item: Booking) => item.can_manage && canManage ? <Space wrap>
    {item.status === "pending" && <Button disabled={!!savingID} loading={savingID === item.id} onClick={() => void changeStatus(item, "confirmed")}>ยืนยัน</Button>}
    {item.status === "confirmed" && <Popconfirm title="ยืนยันว่าบริการเสร็จสิ้นแล้ว?" onConfirm={() => changeStatus(item, "completed")} okText="เสร็จสิ้น" cancelText="ย้อนกลับ"><Button disabled={!!savingID} loading={savingID === item.id}>เสร็จสิ้น</Button></Popconfirm>}
    {["pending", "confirmed"].includes(item.status) && <Button danger disabled={!!savingID} onClick={() => { setSelectedID(""); setCancelling(item); }}>ยกเลิก</Button>}
  </Space> : null;
  return <Row gutter={[24, 16]}>
    <Col span={24}><Flex justify="space-between" align="center" gap={16} wrap>
      <div><Typography.Title level={3} style={{ margin: 0 }}>การจอง</Typography.Title></div>
      <Space wrap><DatePicker aria-label="วันที่จอง" format="DD/MM/YYYY" placeholder="DD/MM/YYYY" value={dayjs(date)} allowClear={false} disabled={!!savingID} onChange={value => { if (value) onDateChange(value.format("YYYY-MM-DD")); }} /><Button aria-label="โหลดการจองใหม่" icon={<ReloadOutlined />} loading={loading} disabled={!!savingID} onClick={reload} />{canManage && <Button type="primary" icon={<PlusOutlined />} disabled={!data.services.length || !data.staff.length || !!savingID} onClick={() => setOpen(true)}>เพิ่มการจอง</Button>}</Space>
    </Flex></Col>
    <Col span={24}><Flex justify="space-between" gap={16} wrap>
      <Space><Button disabled={!!savingID} onClick={() => onDateChange(dayjs(date).subtract(7, "day").format("YYYY-MM-DD"))}>7 วันก่อนหน้า</Button><Button disabled={!!savingID} onClick={() => onDateChange(localDate())}>วันนี้</Button><Button disabled={!!savingID} onClick={() => onDateChange(dayjs(date).add(7, "day").format("YYYY-MM-DD"))}>7 วันถัดไป</Button></Space>
      <Segmented value={view} onChange={setView} options={[{ label: "ปฏิทิน 7 วัน", value: "calendar" }, { label: "รายการ", value: "list" }]} />
    </Flex></Col>
    <Col span={24}><Typography.Text strong>ตั้งแต่วันที่ : {dayjs(date).format("DD/MM/YYYY")} – {dayjs(date).add(6, "day").format("DD/MM/YYYY")}</Typography.Text></Col>
    {canManage && (!data.services.length || !data.staff.length) && <Col span={24}><Alert type="info" showIcon title="กรุณาเพิ่มบริการและผู้ให้บริการที่เปิดใช้งานก่อนสร้างการจอง" /></Col>}
    {data?.shop_status === "inactive" && <Col span={24}><Alert type="warning" showIcon title="ร้านค้าปิดใช้งาน ดูรายการจองได้อย่างเดียว" /></Col>}
    <Col span={24}>{error ? <Alert type="error" showIcon title={error} action={<Button onClick={reload} loading={loading}>ลองใหม่</Button>} /> : <Card>
      {view === "calendar" ? <Card loading={loading} variant="borderless" styles={{ body: { padding: 0 } }}><BookingCalendar date={date} items={data?.items || []} openingTime={data?.opening_time || null} closingTime={data?.closing_time || null} onSelect={item => setSelectedID(item.id)} /></Card> : <Table<Booking> rowKey="id" loading={loading} dataSource={data?.items || []} scroll={{ x: 1050 }} locale={{ emptyText: "ไม่มีการจองในช่วง 7 วันที่เลือก" }} columns={[
        { title: "วันที่/เวลา", key: "time", width: 185, render : (_, item) => <><Typography.Text>{formatTime(item.starts_at)}</Typography.Text><br /><Typography.Text type="secondary">ถึง {formatTime(item.ends_at)}</Typography.Text></> },
        { title: "ลูกค้า", key: "customer", render : (_, item) => <><Typography.Text strong>{item.customer_name}</Typography.Text><br /><Typography.Text type="secondary">{item.customer_phone}</Typography.Text></> },
        { title: "บริการ", key: "service", render : (_, item) => <>{item.service_name}<br /><Typography.Text type="secondary">{item.duration_minutes} นาที · {(item.price_satang / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" })}</Typography.Text></> },
        { title: "ผู้ให้บริการ", dataIndex: "staff_name" },
        { title: "สถานะ", dataIndex: "status", width: 120, render : (status: BookingStatus) => <Tag color={bookingStatuses[status].color}>{bookingStatuses[status].label}</Tag> },
        ...(canManage ? [{ title: "จัดการ", key: "actions", width: 200, render : (_: unknown, item: Booking) => actions(item) }] : []),
        ...(canManage ? [{ title: "", key: "edit", width: 56, fixed: "right" as const, align: "center" as const, render: (_: unknown, item: Booking) => editButton(item) }] : []),
      ]} expandable={{ rowExpandable: item => !!item.note || !!item.remark, expandedRowRender : item => <>
        {item.note && <Typography.Paragraph style={{ whiteSpace: "pre-wrap" }}>หมายเหตุ : {item.note}</Typography.Paragraph>}
        {item.remark && <Typography.Paragraph style={{ whiteSpace: "pre-wrap", marginBottom: 0 }}><Typography.Text strong>เหตุผลการยกเลิก : </Typography.Text>{item.remark}</Typography.Paragraph>}
      </> }} />}
    </Card>}</Col>
    <Drawer title="รายละเอียดการจอง" extra={selected ? editButton(selected) : undefined} placement="right" size="min(520px, 100vw)" open={!!selected} onClose={() => { if (!savingID) setSelectedID(""); }} closable={!savingID} keyboard={!savingID} mask={{ closable: !savingID }} footer={selected && canManage && selected.can_manage ? actions(selected) : undefined}>
      {selected && <Descriptions column={1} items={[
        { key: "status", label: "สถานะ", children: <Tag color={bookingStatuses[selected.status].color}>{bookingStatuses[selected.status].label}</Tag> },
        { key: "customer", label: "ลูกค้า", children: selected.customer_name },
        { key: "phone", label: "เบอร์โทรศัพท์", children: selected.customer_phone },
        { key: "service", label: "บริการ", children: selected.service_name },
        { key: "staff", label: "ผู้ให้บริการ", children: selected.staff_name },
        { key: "start", label: "เริ่ม (เวลาไทย)", children: formatTime(selected.starts_at) },
        { key: "end", label: "สิ้นสุด (เวลาไทย)", children: formatTime(selected.ends_at) },
        { key: "duration", label: "ระยะเวลา", children: `${selected.duration_minutes} นาที` },
        { key: "price", label: "ราคา", children: (selected.price_satang / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" }) },
        ...(selected.note ? [{ key: "note", label: "หมายเหตุ", children: <span style={{ whiteSpace: "pre-wrap" }}>{selected.note}</span> }] : []),
        ...(selected.remark ? [{ key: "remark", label: "เหตุผลการยกเลิก", children: <span style={{ whiteSpace: "pre-wrap" }}>{selected.remark}</span> }] : []),
      ]} />}
    </Drawer>
    {data && <BookingDrawer shopID={shopID} date={date} services={data.services} staff={data.staff} role={data.role} openingTime={data.opening_time} closingTime={data.closing_time} open={open} onClose={() => setOpen(false)} onCreated={createdDate => { if (createdDate >= date && createdDate < dayjs(date).add(7, "day").format("YYYY-MM-DD")) reload(); else onDateChange(createdDate); }} />}
    {cancelling && cancelling.can_manage && canManage && <CancelBookingDrawer key={cancelling.id} shopID={shopID} booking={cancelling} onClose={() => setCancelling(undefined)} onCancelled={updated => setData(previous => previous ? { ...previous, items: previous.items.map(item => item.id === updated.id ? updated : item) } : previous)} />}
    {editingBooking && editingBooking.can_manage && canManage && <EditBookingDrawer key={editingBooking.id} shopID={shopID} booking={editingBooking} services={data?.services || []} openingTime={data?.opening_time || null} closingTime={data?.closing_time || null} onClose={() => setEditingBooking(undefined)} onSaved={updated => { const updatedDate = localDate(new Date(updated.starts_at)); if (updatedDate >= date && updatedDate < dayjs(date).add(7, "day").format("YYYY-MM-DD")) reload(); else onDateChange(updatedDate); }} />}
  </Row>;
}

export default function Bookings() {
  const { shopID } = useParams();
  const [date, setDate] = useState(localDate);
  return shopID ? <WeeklyBookings key={`${shopID}:${date}`} shopID={shopID} date={date} onDateChange={setDate} /> : <Alert type="error" title="ไม่พบร้านค้า" />;
}
