import { useState } from "react";
import { Alert, App, Button, Col, Drawer, Flex, Form, Input, Row, Select } from "antd";
import { createShop } from "../../../api/tenant/shops.api";
import { businessTypes } from "../../../types/tenant/shop.type";
import type { CreateShopDrawerProps, CreateShopRequest } from "../../../types/tenant/shop.type";

export default function CreateShopDrawer({ open, onClose, onCreated, shopLimit }: CreateShopDrawerProps) {
  const [form] = Form.useForm<CreateShopRequest>();
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  const submit = async (values: CreateShopRequest) => {
    setSaving(true);
    try {
      const shop = await createShop({ ...values, name: values.name.trim() });
      onCreated(shop);
      form.resetFields();
      onClose();
      message.success("เพิ่มร้านค้าเรียบร้อยแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "เชื่อมต่อระบบไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title="เพิ่มร้านค้า" placement="right" size="min(480px, 100vw)" open={open} onClose={() => { if (!saving) onClose(); }} closable={!saving} keyboard={!saving} mask={{ closable: !saving }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={onClose}>ยกเลิก</Button><Button type="primary" htmlType="submit" form="create-shop-form" loading={saving}>เพิ่มร้านค้า</Button></Flex>}>
    <Alert type="info" showIcon title="เริ่มต้นด้วยร้านค้าของคุณ" description={`แพ็กเกจของคุณรองรับ ${shopLimit} ร้านค้า จำนวนผู้ใช้งานต่อร้านเป็นไปตามแพ็กเกจ`} style={{ marginBottom: 24 }} />
    <Form<CreateShopRequest> id="create-shop-form" form={form} layout="vertical" onFinish={submit} disabled={saving} requiredMark={false}>
      <Row gutter={16}>
        <Col span={24}><Form.Item name="name" label="ชื่อร้านค้า / ธุรกิจ" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อร้านค้า" }]}><Input maxLength={150} placeholder="ชื่อร้านที่ลูกค้าจะเห็น" /></Form.Item></Col>
        <Col span={24}><Form.Item name="business_type" label="ประเภทธุรกิจ" rules={[{ required: true, message: "กรุณาเลือกประเภทธุรกิจ" }]}><Select placeholder="เลือกประเภทธุรกิจ" options={businessTypes} /></Form.Item></Col>
      </Row>
    </Form>
  </Drawer>;
}
