import { useState } from "react";
import { App, Button, Card, Form, Input, Space, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { changePassword } from "../../../api/tenant/password.api";
import { clearAuthSession } from "../../../utils/auth";
import type { ChangePasswordForm } from "../../../types/shared/password.type";

export default function ChangePassword() {
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  const navigate = useNavigate();
  const submit = async (values: ChangePasswordForm) => {
    setSaving(true);
    try {
      await changePassword(values.current_password, values.password);
      clearAuthSession(); message.success("ตั้งรหัสผ่านใหม่แล้ว กรุณาเข้าสู่ระบบอีกครั้ง"); navigate("/tenant/login", { replace: true });
    } catch (error) { message.error(error instanceof Error ? error.message : "เปลี่ยนรหัสไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <main style={{ minHeight: "100dvh", padding: 24, background: "#F3F5F2" }}><Card title="ตั้งรหัสผ่านใหม่" style={{ maxWidth: 480, margin: "48px auto" }}>
    <Space orientation="vertical" size="large" style={{ width: "100%" }}>
      <Typography.Text>หากได้รับรหัสชั่วคราว กรุณาตั้งรหัสส่วนตัวก่อนใช้งานต่อ</Typography.Text>
      <Form<ChangePasswordForm> layout="vertical" onFinish={submit} disabled={saving}>
        <Form.Item name="current_password" label="รหัสผ่านปัจจุบัน / รหัสชั่วคราว" rules={[{ required: true }]}><Input.Password autoComplete="current-password" /></Form.Item>
        <Form.Item name="password" label="รหัสผ่านใหม่" dependencies={["current_password"]} rules={[{ required: true }, { min: 12, message: "อย่างน้อย 12 ตัวอักษร" }, ({ getFieldValue }) => ({ validator: (_, value: string) => !value || (value !== getFieldValue("current_password") && new TextEncoder().encode(value).length <= 72) ? Promise.resolve() : Promise.reject(new Error("รหัสใหม่ต้องต่างจากรหัสเดิม และไม่เกิน 72 ไบต์")) })]}><Input.Password autoComplete="new-password" /></Form.Item>
        <Form.Item name="confirm_password" label="ยืนยันรหัสผ่านใหม่" dependencies={["password"]} rules={[{ required: true }, ({ getFieldValue }) => ({ validator: (_, value) => value === getFieldValue("password") ? Promise.resolve() : Promise.reject(new Error("รหัสผ่านไม่ตรงกัน")) })]}><Input.Password autoComplete="new-password" /></Form.Item>
        <Button type="primary" block htmlType="submit" loading={saving}>บันทึกรหัสผ่านใหม่</Button>
      </Form>
      <Button disabled={saving} onClick={() => { clearAuthSession(); navigate("/tenant/login", { replace: true }); }}>ออกจากระบบ</Button>
    </Space>
  </Card></main>;
}
