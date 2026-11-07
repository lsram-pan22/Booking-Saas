import { useState } from "react";
import { Alert, App, Button, Col, Drawer, Flex, Form, Input, Row, Segmented, Tag } from "antd";
import { addShopMember } from "../../../api/tenant/shops.api";
import type { AddMemberDrawerProps, AddMemberForm } from "../../../types/tenant/shop.type";

export default function AddMemberDrawer({ shopID, open, onClose, onAdded }: AddMemberDrawerProps) {
  const [form] = Form.useForm<AddMemberForm>();
  const mode = Form.useWatch("mode", form) || "new";
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  const submit = async (values: AddMemberForm) => {
    setSaving(true);
    try {
      const email = values.email.trim().toLowerCase();
      const member = await addShopMember(shopID, values.mode === "existing" ? { mode: "existing", email } : {
        mode: "new", username: values.username?.trim().toLowerCase() || "", email, full_name: values.full_name?.trim() || "", phone: values.phone || "", password: values.password || "",
      });
      onAdded(member); form.resetFields(); onClose(); message.success("เพิ่มผู้ใช้งานในร้านแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "เพิ่มผู้ใช้งานไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title="เพิ่มผู้ใช้งานของร้าน" placement="right" size="min(520px, 100vw)" open={open} forceRender onClose={() => { if (!saving) onClose(); }} closable={!saving} keyboard={!saving} mask={{ closable: !saving }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={onClose}>ยกเลิก</Button><Button type="primary" htmlType="submit" form="add-shop-member-form" loading={saving}>เพิ่มผู้ใช้งาน</Button></Flex>}>
    <Alert type="info" showIcon title="เพิ่มเป็นพนักงานของร้าน" description="บัญชีใหม่เข้าสู่ระบบด้วย Username หรืออีเมลและรหัสผ่านที่กำหนด ส่วนบัญชีที่มีอยู่ใช้ข้อมูลและรหัสผ่านเดิม" style={{ marginBottom: 24 }} />
    <Form<AddMemberForm> id="add-shop-member-form" form={form} layout="vertical" initialValues={{ mode: "new" }} onFinish={submit} disabled={saving} requiredMark={false}>
      <Row gutter={16}>
        <Col span={24}><Form.Item name="mode" label="ประเภทบัญชี"><Segmented block disabled={saving} options={[{ value: "new", label: "สร้างบัญชีใหม่" }, { value: "existing", label: "บัญชีที่มีอยู่" }]} /></Form.Item></Col>
        <Col span={24}><Form.Item name="email" label="อีเมล" rules={[{ required: true, message: "กรุณากรอกอีเมล" }, { type: "email", message: "อีเมลไม่ถูกต้อง" }]}><Input type="email" autoComplete="off" maxLength={254} placeholder="staff@example.com" /></Form.Item></Col>
        {mode === "new" && <>
          <Col span={24}><Form.Item name="username" label="Username" preserve={false} extra="ใช้ตัวอักษรอังกฤษ ตัวเลข จุด ขีดกลาง หรือขีดล่าง 3–50 ตัว" rules={[{ required: true, pattern: /^[a-zA-Z0-9][a-zA-Z0-9._-]{2,49}$/, message: "Username ต้องมี 3–50 ตัว และขึ้นต้นด้วยตัวอักษรอังกฤษหรือตัวเลข" }]}><Input autoComplete="off" maxLength={50} placeholder="เช่น staff01" /></Form.Item></Col>
          <Col xs={24} sm={12}><Form.Item name="full_name" label="ชื่อ–นามสกุล" preserve={false} rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อ–นามสกุล" }]}><Input autoComplete="off" maxLength={100} /></Form.Item></Col>
          <Col xs={24} sm={12}><Form.Item name="phone" label="เบอร์โทรศัพท์" preserve={false} rules={[{ required: true, message: "กรุณากรอกเบอร์โทรศัพท์" }, { validator: (_, value: string) => !value || /^(?:0[0-9]{8,9}|\+66[0-9]{8,9})$/.test(value.replace(/[\s-]/g, "")) ? Promise.resolve() : Promise.reject(new Error("เบอร์โทรศัพท์ไม่ถูกต้อง")) }]}><Input type="tel" autoComplete="off" maxLength={20} /></Form.Item></Col>
          <Col span={24}><Form.Item name="password" label="รหัสผ่าน" preserve={false} rules={[{ required: true, message: "กรุณากรอกรหัสผ่าน" }, { min: 12, message: "รหัสผ่านต้องมีอย่างน้อย 12 ตัวอักษร" }, { validator: (_, value: string) => value && new TextEncoder().encode(value).length > 72 ? Promise.reject(new Error("รหัสผ่านยาวเกินไป")) : Promise.resolve() }]}><Input.Password autoComplete="new-password" /></Form.Item></Col>
          <Col span={24}><Form.Item name="confirm_password" label="ยืนยันรหัสผ่าน" preserve={false} dependencies={["password"]} rules={[{ required: true, message: "กรุณายืนยันรหัสผ่าน" }, ({ getFieldValue }) => ({ validator: (_, value) => !value || value === getFieldValue("password") ? Promise.resolve() : Promise.reject(new Error("รหัสผ่านไม่ตรงกัน")) })]}><Input.Password autoComplete="new-password" /></Form.Item></Col>
        </>}
        <Col span={24}><Form.Item label="สิทธิ์ในร้าน"><Tag color="blue">พนักงาน (Staff)</Tag></Form.Item></Col>
      </Row>
    </Form>
  </Drawer>;
}
