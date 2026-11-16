import { useState } from "react";
import { Alert, App, Button, Drawer, Flex, Form, Input } from "antd";
import type { PasswordForm, ResetPasswordDrawerProps } from "../../../types/shared/password.type";

export default function ResetPasswordDrawer({ name, open, onClose, onReset }: ResetPasswordDrawerProps) {
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  const submit = async (values: PasswordForm) => {
    setSaving(true);
    try { await onReset(values.password); message.success("รีเซ็ตรหัสแล้ว กรุณาแจ้งรหัสชั่วคราวให้ผู้ใช้โดยตรง"); onClose(); }
    catch (error) { message.error(error instanceof Error ? error.message : "รีเซ็ตรหัสไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Drawer title={`รีเซ็ตรหัสผ่าน: ${name}`} open={open} placement="right" size="min(480px, 100vw)" onClose={onClose} closable={!saving} keyboard={!saving} mask={{ closable: !saving }} destroyOnHidden>
    <Alert showIcon type="warning" title="ผู้ใช้จะออกจากระบบทุกอุปกรณ์ และต้องตั้งรหัสใหม่หลังเข้าสู่ระบบด้วยรหัสชั่วคราว" style={{ marginBottom: 24 }} />
    <Form<PasswordForm> layout="vertical" onFinish={submit} disabled={saving}>
      <Form.Item name="password" label="รหัสผ่านชั่วคราว" rules={[{ required: true }, { min: 12, message: "อย่างน้อย 12 ตัวอักษร" }, { validator: (_, value: string) => !value || new TextEncoder().encode(value).length <= 72 ? Promise.resolve() : Promise.reject(new Error("ไม่เกิน 72 ไบต์")) }]}><Input.Password autoComplete="new-password" /></Form.Item>
      <Form.Item name="confirm_password" label="ยืนยันรหัสผ่านชั่วคราว" dependencies={["password"]} rules={[{ required: true }, ({ getFieldValue }) => ({ validator: (_, value) => value === getFieldValue("password") ? Promise.resolve() : Promise.reject(new Error("รหัสผ่านไม่ตรงกัน")) })]}><Input.Password autoComplete="new-password" /></Form.Item>
      <Flex justify="end" gap={8}><Button onClick={onClose}>ยกเลิก</Button><Button type="primary" danger htmlType="submit" loading={saving}>ยืนยันรีเซ็ตรหัสผ่าน</Button></Flex>
    </Form>
  </Drawer>;
}
