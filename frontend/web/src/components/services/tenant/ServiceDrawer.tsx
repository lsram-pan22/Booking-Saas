import { useEffect, useState } from "react";
import { App, Button, Col, Drawer, Flex, Form, Input, InputNumber, Row, Switch } from "antd";
import { saveService } from "../../../api/tenant/services.api";
import type { ServiceDrawerProps, ServiceForm } from "../../../types/tenant/service.type";

export default function ServiceDrawer({ shopID, open, service, onClose, onSaved }: ServiceDrawerProps) {
  const [form] = Form.useForm<ServiceForm>();
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  useEffect(() => {
    if (open) form.setFieldsValue(service ? { ...service, price_baht: service.price_satang / 100, active: service.status === "active" } : { name: "", description: "", duration_minutes: 30, price_baht: 0, active: true });
  }, [form, open, service]);
  const submit = async (values: ServiceForm) => {
    setSaving(true);
    try {
      const saved = await saveService(shopID, { name: values.name.trim(), description: (values.description || "").trim(), duration_minutes: values.duration_minutes, price_satang: Math.round(values.price_baht * 100), status: values.active ? "active" : "inactive" }, service?.id);
      onSaved(saved); onClose(); message.success(service ? "แก้ไขบริการเรียบร้อยแล้ว" : "เพิ่มบริการเรียบร้อยแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "เชื่อมต่อระบบไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title={service ? "แก้ไขบริการ" : "เพิ่มบริการ"} placement="right" size="min(520px, 100vw)" forceRender open={open} onClose={() => { if (!saving) onClose(); }} closable={!saving} keyboard={!saving} mask={{ closable: !saving }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={onClose}>ยกเลิก</Button><Button type="primary" form="shop-service-form" htmlType="submit" loading={saving}>บันทึก</Button></Flex>}>
    <Form<ServiceForm> id="shop-service-form" form={form} layout="vertical" disabled={saving} onFinish={submit} requiredMark={false}>
      <Row gutter={16}>
        <Col span={24}><Form.Item name="name" label="ชื่อบริการ" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อบริการ" }]}><Input maxLength={150} /></Form.Item></Col>
        <Col span={24}><Form.Item name="description" label="รายละเอียด"><Input.TextArea rows={4} maxLength={2000} showCount /></Form.Item></Col>
        <Col xs={24} sm={12}><Form.Item name="duration_minutes" label="ระยะเวลา (นาที)" rules={[{ required: true, message: "กรุณาระบุระยะเวลา" }]}><InputNumber min={1} max={1440} precision={0} style={{ width: "100%" }} /></Form.Item></Col>
        <Col xs={24} sm={12}><Form.Item name="price_baht" label="ราคา (บาท)" rules={[{ required: true, message: "กรุณาระบุราคา" }]}><InputNumber min={0} max={1000000} precision={2} style={{ width: "100%" }} /></Form.Item></Col>
        <Col span={24}><Form.Item name="active" label="สถานะบริการ" valuePropName="checked"><Switch checkedChildren="เปิด" unCheckedChildren="ปิด" /></Form.Item></Col>
      </Row>
    </Form>
  </Drawer>;
}
