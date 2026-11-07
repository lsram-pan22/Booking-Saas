import { useState } from "react";
import { Alert, App, Button, Card, Col, Form, Input, Row, Typography } from "antd";
import { useOutletContext } from "react-router-dom";
import { updateTenantProfile } from "../../../api/tenant/account.api";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";
import type { TenantSettingsForm } from "../../../types/tenant/account.type";

export default function Settings() {
  const { profile, error, updateUser } = useOutletContext<TenantLayoutContext>();
  const { message } = App.useApp();
  const [saving, setSaving] = useState(false);
  const submit = async (values: TenantSettingsForm) => {
    setSaving(true);
    try { const user = await updateTenantProfile(values); updateUser(user); message.success("บันทึกข้อมูลแล้ว"); }
    catch (e) { message.error(e instanceof Error ? e.message : "บันทึกข้อมูลไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  if (error) return <Alert type="error" showIcon title={error} action={<Button onClick={() => window.location.reload()}>ลองใหม่</Button>} />;
  return <>
    <Typography.Title level={3} style={{ marginTop: 0 }}>ตั้งค่าบัญชี</Typography.Title>
    <Row><Col xs={24} lg={18} xl={14}><Card title="ข้อมูลส่วนตัว" loading={!profile}>
      {profile && <Form<TenantSettingsForm> layout="vertical" initialValues={profile.user} onFinish={submit} disabled={saving}>
        <Row gutter={16}>
          <Col xs={24} sm={12}><Form.Item name="full_name" label="ชื่อ–นามสกุล" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อ–นามสกุล" }]}><Input autoComplete="name" maxLength={100} /></Form.Item></Col>
          <Col xs={24} sm={12}><Form.Item name="phone" label="เบอร์โทรศัพท์" rules={[{ required: true, message: "กรุณากรอกเบอร์โทรศัพท์" }, { validator: (_, value: string) => !value || /^(?:0[0-9]{8,9}|\+66[0-9]{8,9})$/.test(value.replace(/[\s-]/g, "")) ? Promise.resolve() : Promise.reject(new Error("เบอร์โทรศัพท์ไม่ถูกต้อง")) }]}><Input type="tel" autoComplete="tel" maxLength={20} /></Form.Item></Col>
          <Col span={24}><Form.Item label="อีเมลเข้าสู่ระบบ" extra="การเปลี่ยนอีเมลยังไม่เปิดให้ใช้งาน"><Input value={profile.user.email} readOnly /></Form.Item></Col>
          <Col span={24}><Button type="primary" htmlType="submit" loading={saving}>บันทึกการเปลี่ยนแปลง</Button></Col>
        </Row>
      </Form>}
    </Card></Col></Row>
  </>;
}
