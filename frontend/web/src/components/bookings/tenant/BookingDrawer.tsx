import { useEffect, useState } from "react";
import { Alert, App, Button, Col, DatePicker, Drawer, Flex, Form, Input, Row, Select } from "antd";
import dayjs from "dayjs";
import { createBooking } from "../../../api/tenant/bookings.api";
import type { BookingDrawerProps, BookingForm } from "../../../types/tenant/booking.type";

export default function BookingDrawer({ shopID, date, services, staff, role, openingTime, closingTime, open, onClose, onCreated }: BookingDrawerProps) {
  const [form] = Form.useForm<BookingForm>();
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  const serviceID = Form.useWatch("service_id", form);
  const service = services.find(item => item.id === serviceID);
  useEffect(() => { if (open) { form.resetFields(); form.setFieldsValue({ date: dayjs(date), time: openingTime || "09:00", staff_member_id: staff.length === 1 ? staff[0].id : undefined }); } }, [open, date, form, staff, openingTime]);
  const submit = async (values: BookingForm) => {
    const bookingDate = values.date.format("YYYY-MM-DD");
    const startsAt = new Date(`${bookingDate}T${values.time}:00+07:00`);
    if (!Number.isFinite(startsAt.getTime()) || startsAt <= new Date()) { message.error("กรุณาเลือกวันและเวลาในอนาคต"); return; }
    if (openingTime && closingTime && service) {
      const minutes = (value: string) => { const [hour, minute] = value.split(":").map(Number); return hour * 60 + minute; };
      if (minutes(values.time) < minutes(openingTime) || minutes(values.time) + service.duration_minutes > minutes(closingTime)) { message.error("เวลาเริ่มและสิ้นสุดการจองต้องอยู่ในเวลาเปิด–ปิดร้าน"); return; }
    }
    setSaving(true);
    try {
      await createBooking(shopID, { service_id: values.service_id, staff_member_id: values.staff_member_id, customer_name: values.customer_name.trim(), customer_phone: values.customer_phone.trim(), note: (values.note || "").trim(), starts_at: startsAt.toISOString() });
      message.success("เพิ่มการจองเรียบร้อยแล้ว"); onClose(); onCreated(bookingDate);
    } catch (e) { message.error(e instanceof Error ? e.message : "เชื่อมต่อระบบไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title="เพิ่มการจอง" placement="right" size="min(560px, 100vw)" forceRender open={open} closable={!saving} keyboard={!saving} mask={{ closable: !saving }} onClose={() => { if (!saving) onClose(); }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={onClose}>ยกเลิก</Button><Button type="primary" form="create-booking-form" htmlType="submit" loading={saving}>บันทึกการจอง</Button></Flex>}>
    <Form<BookingForm> id="create-booking-form" form={form} layout="vertical" requiredMark={false} disabled={saving} onFinish={submit}>
      <Row gutter={16}>
        <Col span={24}><Form.Item name="customer_name" label="ชื่อลูกค้า" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อลูกค้า" }]}><Input maxLength={150} /></Form.Item></Col>
        <Col span={24}><Form.Item name="customer_phone" label="เบอร์โทรศัพท์" rules={[{ required: true, message: "กรุณากรอกเบอร์โทรศัพท์" }, { pattern: /^[+0-9() -]{7,30}$/, message: "รูปแบบเบอร์โทรศัพท์ไม่ถูกต้อง" }, { validator: (_, value: string) => { const digits = (value || "").replace(/\D/g, "").length; return !value || (digits >= 7 && digits <= 15) ? Promise.resolve() : Promise.reject(new Error("กรุณากรอกหมายเลข 7–15 หลัก")); } }]}><Input type="tel" maxLength={30} /></Form.Item></Col>
        <Col span={24}><Form.Item name="service_id" label="บริการ" rules={[{ required: true, message: "กรุณาเลือกบริการ" }]}><Select showSearch optionFilterProp="label" placeholder="เลือกบริการ" options={services.map(item => ({ value: item.id, label: `${item.name} (${item.duration_minutes} นาที)` }))} /></Form.Item></Col>
        <Col span={24}><Form.Item name="staff_member_id" label="ผู้ให้บริการ" rules={[{ required: true, message: "กรุณาเลือกผู้ให้บริการ" }]}><Select disabled={saving || role === "staff"} showSearch optionFilterProp="label" placeholder="เลือกผู้ให้บริการ" options={staff.map(item => ({ value: item.id, label: `${item.name}${item.role === "owner" ? " (เจ้าของร้าน)" : ""}` }))} /></Form.Item></Col>
        <Col xs={24} sm={12}><Form.Item name="date" label="วันที่ (เวลาไทย)" rules={[{ required: true, message: "กรุณาเลือกวันที่" }]}><DatePicker format="DD/MM/YYYY" placeholder="DD/MM/YYYY" style={{ width: "100%" }} /></Form.Item></Col>
        <Col xs={24} sm={12}><Form.Item name="time" label="เวลาเริ่ม" extra={openingTime && closingTime ? `ร้านเปิด ${openingTime}–${closingTime} น.` : undefined} rules={[{ required: true, message: "กรุณาเลือกเวลา" }]}><Input type="time" step={60} min={openingTime || undefined} max={closingTime === "24:00" ? "23:59" : closingTime || undefined} /></Form.Item></Col>
        <Col span={24}><Form.Item name="note" label="หมายเหตุ"><Input.TextArea rows={3} maxLength={2000} showCount /></Form.Item></Col>
        {service && <Col span={24}><Alert showIcon type="info" title={`${service.duration_minutes} นาที · ${(service.price_satang / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" })}`} description="การจองใหม่มีสถานะรอยืนยัน ระบบคำนวณเวลาสิ้นสุดและตรวจเวลาทับซ้อนก่อนบันทึก" /></Col>}
      </Row>
    </Form>
  </Drawer>;
}
