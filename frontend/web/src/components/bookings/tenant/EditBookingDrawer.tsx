import { useState } from "react";
import { Alert, App, Button, Col, DatePicker, Drawer, Flex, Form, Input, Row, Select } from "antd";
import dayjs from "dayjs";
import { updateBooking } from "../../../api/tenant/bookings.api";
import type { EditBookingDrawerProps, EditBookingForm } from "../../../types/tenant/booking.type";

const thaiDate = (value: string) => new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Bangkok", year: "numeric", month: "2-digit", day: "2-digit" }).format(new Date(value));
const thaiTime = (value: string) => new Date(value).toLocaleTimeString("en-GB", { timeZone: "Asia/Bangkok", hour: "2-digit", minute: "2-digit", hour12: false });

export default function EditBookingDrawer({ shopID, booking, services, openingTime, closingTime, onClose, onSaved }: EditBookingDrawerProps) {
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm<EditBookingForm>();
  const { message } = App.useApp();
  const serviceID = Form.useWatch("service_id", form) || booking.service_id;
  const date = Form.useWatch("date", form);
  const time = Form.useWatch("time", form);
  const originalService = { id: booking.service_id, name: booking.service_name, duration_minutes: booking.duration_minutes, price_satang: booking.price_satang };
  const selectedService = serviceID === booking.service_id ? originalService : services.find(item => item.id === serviceID);
  const options = [originalService, ...services.filter(item => item.id !== booking.service_id)];
  const previewStart = date && time ? new Date(`${date.format("YYYY-MM-DD")}T${time}:00+07:00`) : new Date(booking.starts_at);
  const previewEnd = selectedService && Number.isFinite(previewStart.getTime()) ? new Date(previewStart.getTime() + selectedService.duration_minutes * 60000) : undefined;
  const submit = async (values: EditBookingForm) => {
    const startsAt = new Date(`${values.date.format("YYYY-MM-DD")}T${values.time}:00+07:00`);
    if (!selectedService || !Number.isFinite(startsAt.getTime())) { message.error("กรุณาตรวจสอบบริการ วันและเวลา"); return; }
    const moved = startsAt.getTime() !== new Date(booking.starts_at).getTime();
    if (moved && startsAt <= new Date()) { message.error("วันและเวลาใหม่ต้องอยู่ในอนาคต"); return; }
    if ((moved || values.service_id !== booking.service_id) && openingTime && closingTime) {
      const minutes = (value: string) => { const [hour, minute] = value.split(":").map(Number); return hour * 60 + minute; };
      if (minutes(values.time) < minutes(openingTime) || minutes(values.time) + selectedService.duration_minutes > minutes(closingTime)) { message.error("เวลาเริ่มและสิ้นสุดการจองต้องอยู่ในเวลาเปิด–ปิดร้าน"); return; }
    }
    setSaving(true);
    try {
      const updated = await updateBooking(shopID, booking.id, { customer_name: values.customer_name.trim(), customer_phone: values.customer_phone.trim(), note: (values.note || "").trim(), service_id: values.service_id, starts_at: startsAt.toISOString() });
      onSaved(updated); onClose(); message.success("แก้ไขการจองเรียบร้อยแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "แก้ไขการจองไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title="แก้ไขการจอง" placement="right" size="min(560px, 100vw)" open closable={!saving} keyboard={!saving} mask={{ closable: !saving }} onClose={() => { if (!saving) onClose(); }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={onClose}>ยกเลิก</Button><Button type="primary" htmlType="submit" form="edit-booking-form" loading={saving}>บันทึก</Button></Flex>}>
    <Form<EditBookingForm> id="edit-booking-form" form={form} layout="vertical" requiredMark={false} disabled={saving} onFinish={submit} initialValues={{ customer_name: booking.customer_name, customer_phone: booking.customer_phone, note: booking.note, service_id: booking.service_id, date: dayjs(thaiDate(booking.starts_at)), time: thaiTime(booking.starts_at) }}>
      <Row gutter={16}>
        <Col span={24}><Form.Item name="customer_name" label="ชื่อลูกค้า" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อลูกค้า" }, { max: 150, message: "ระบุชื่อได้ไม่เกิน 150 ตัวอักษร" }]}><Input maxLength={150} /></Form.Item></Col>
        <Col span={24}><Form.Item name="customer_phone" label="เบอร์โทรศัพท์" rules={[{ required: true, message: "กรุณากรอกเบอร์โทรศัพท์" }, { pattern: /^[+0-9() -]{7,30}$/, message: "รูปแบบเบอร์โทรศัพท์ไม่ถูกต้อง" }, { validator: (_, value: string) => { const digits = (value || "").replace(/\D/g, "").length; return !value || (digits >= 7 && digits <= 15) ? Promise.resolve() : Promise.reject(new Error("กรุณากรอกหมายเลข 7–15 หลัก")); } }]}><Input type="tel" maxLength={30} /></Form.Item></Col>
        <Col span={24}><Form.Item name="service_id" label="บริการ" rules={[{ required: true, message: "กรุณาเลือกบริการ" }]}><Select showSearch optionFilterProp="label" options={options.map(item => ({ value: item.id, label: `${item.name} (${item.duration_minutes} นาที)${item.id === booking.service_id ? " · บริการเดิม" : ""}` }))} /></Form.Item></Col>
        <Col xs={24} sm={12}><Form.Item name="date" label="วันที่ (เวลาไทย)" rules={[{ required: true, message: "กรุณาเลือกวันที่" }]}><DatePicker format="DD/MM/YYYY" placeholder="DD/MM/YYYY" style={{ width: "100%" }} /></Form.Item></Col>
        <Col xs={24} sm={12}><Form.Item name="time" label="เวลาเริ่ม" extra={openingTime && closingTime ? `ร้านเปิด ${openingTime}–${closingTime} น.` : undefined} rules={[{ required: true, message: "กรุณาเลือกเวลา" }]}><Input type="time" step={60} /></Form.Item></Col>
        <Col span={24}><Form.Item name="note" label="หมายเหตุ" rules={[{ max: 2000, message: "ระบุหมายเหตุได้ไม่เกิน 2,000 ตัวอักษร" }]}><Input.TextArea rows={3} maxLength={2000} showCount /></Form.Item></Col>
        {selectedService && <Col span={24}><Alert type="info" showIcon title={`${selectedService.duration_minutes} นาที · ${(selectedService.price_satang / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" })}`} description={previewEnd ? `สิ้นสุด ${new Intl.DateTimeFormat("en-GB", { timeZone: "Asia/Bangkok", day: "2-digit", month: "2-digit", year: "numeric", hour: "2-digit", minute: "2-digit", hour12: false }).format(previewEnd)} น.` : undefined} /></Col>}
      </Row>
    </Form>
  </Drawer>;
}
