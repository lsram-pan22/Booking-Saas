import { useState } from "react";
import { App, Button, Col, Drawer, Flex, Form, Input, Row, Typography } from "antd";
import { setBookingStatus } from "../../../api/tenant/bookings.api";
import type { CancelBookingDrawerProps, CancelBookingForm } from "../../../types/tenant/booking.type";

export default function CancelBookingDrawer({ shopID, booking, onClose, onCancelled }: CancelBookingDrawerProps) {
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  const submit = async (values: CancelBookingForm) => {
    setSaving(true);
    try {
      const updated = await setBookingStatus(shopID, booking.id, "cancelled", values.remark.trim());
      onCancelled(updated); onClose(); message.success("ยกเลิกการจองเรียบร้อยแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "ยกเลิกการจองไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title="ยกเลิกการจอง" placement="right" size="min(480px, 100vw)" open closable={!saving} keyboard={!saving} mask={{ closable: !saving }} onClose={() => { if (!saving) onClose(); }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={onClose}>ย้อนกลับ</Button><Button type="primary" danger htmlType="submit" form="cancel-booking-form" loading={saving}>ยืนยันยกเลิกการจอง</Button></Flex>}>
    <Row gutter={[16, 16]}>
      <Col span={24}><Typography.Text strong>{booking.customer_name} · {booking.service_name}</Typography.Text><Typography.Paragraph type="secondary">เมื่อยกเลิก ช่วงเวลานี้จะเปิดให้จองใหม่ได้</Typography.Paragraph></Col>
      <Col span={24}><Form<CancelBookingForm> id="cancel-booking-form" layout="vertical" onFinish={submit} disabled={saving} requiredMark={false}>
        <Form.Item name="remark" label="เหตุผลการยกเลิก" rules={[{ required: true, whitespace: true, message: "กรุณาระบุเหตุผลการยกเลิก" }, { max: 1000, message: "ระบุได้ไม่เกิน 1,000 ตัวอักษร" }]}><Input.TextArea autoFocus rows={5} maxLength={1000} showCount placeholder="เช่น ลูกค้าขอเลื่อนนัด หรือผู้ให้บริการไม่สะดวก" /></Form.Item>
      </Form></Col>
    </Row>
  </Drawer>;
}
