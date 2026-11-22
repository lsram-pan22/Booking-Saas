import type { RegistrationForm } from "../../../types/tenant/account.type";
import { useState } from "react";
import { Link } from "react-router-dom";
import { Alert, App, Button, Card, Col, Descriptions, Divider, Flex, Form, Grid, Input, Modal, Row, Space, Typography } from "antd";
import { registerTenant } from "../../../api/tenant/account.api";

import { CalendarOutlined, CheckCircleOutlined } from "@ant-design/icons";

const registrationPasswordPattern = /^[0-9]*[A-Za-z][0-9]*[A-Za-z][0-9]*[A-Za-z][0-9]*$/;

export default function Register() {
	const { message } = App.useApp();
	const [saving, setSaving] = useState(false);
	const [registered, setRegistered] = useState(false);
  const [form] = Form.useForm<RegistrationForm>();
  const password = Form.useWatch("password", form) ?? "";
  const letterCount = (password.match(/[A-Za-z]/g) ?? []).length;
  const digitCount = (password.match(/[0-9]/g) ?? []).length;
  const passwordIsValid = password.length === 8 && registrationPasswordPattern.test(password);
  const screens = Grid.useBreakpoint();
  const [reviewOpen, setReviewOpen] = useState(false);
  const [review, setReview] = useState<Omit<RegistrationForm, "password" | "confirm_password">>();

  const reviewRegistration = (values: RegistrationForm) => {
    setRegistered(false);
    setReview({
      full_name: values.full_name.trim(),
      phone: values.phone.replace(/[\s-]/g, ""),
      email: values.email.trim(),
    });
    setReviewOpen(true);
  };

  const submitRegistration = async () => {
    if (!review || saving || registered) return;
    setSaving(true);
    try {
      await registerTenant({
        full_name: review.full_name, email: review.email, phone: review.phone,
        password: form.getFieldValue("password"),
      });
      setRegistered(true);
      form.resetFields();
      message.success("สมัครสมาชิกแล้ว");
    } catch (error) {
      message.error(error instanceof Error ? error.message : "ไม่สามารถเชื่อมต่อระบบได้ กรุณาลองใหม่");
    } finally { setSaving(false); }
  };

  return (
    <main style={{ height: "100dvh", boxSizing: "border-box", overflow: "hidden", background: "#F3F5F2", padding: screens.lg ? "20px 24px" : 16 }}>
      <div style={{ maxWidth: 1160, height: "100%", margin: "0 auto", display: "flex", flexDirection: "column", gap: 16 }}>
        <Flex align="center" gap={10} style={{ flexShrink: 0 }}>
          <CalendarOutlined style={{ fontSize: 26, color: "#65725F" }} />
          <Typography.Text strong style={{ fontSize: 20 }}>Booking SaaS</Typography.Text>
        </Flex>
        <Row gutter={{ xs: 0, lg: 40 }} align="middle" wrap={false} style={{ flex: 1, minHeight: 0 }}>
          <Col xs={0} lg={10} style={{ maxHeight: "100%", overflow: "auto" }}>
            <Space orientation="vertical" size={20} style={{ paddingTop: 16 }}>
              <Typography.Text style={{ color: "#65725F", fontWeight: 600 }}>สำหรับเจ้าของธุรกิจ</Typography.Text>
              <Typography.Title level={1} style={{ margin: 0, fontSize: 38, lineHeight: 1.4 }}>เริ่มต้นจัดการธุรกิจ<br />และการจองในที่เดียว</Typography.Title>
              <Typography.Paragraph type="secondary" style={{ fontSize: 16 }}>สร้างบัญชี Tenant สำหรับธุรกิจของคุณ เพื่อเตรียมจัดการบริการ ลูกค้า และทีมงานได้สะดวกขึ้น</Typography.Paragraph>
              <Space orientation="vertical" size={16}>
                {["จัดการตารางจองและบริการ", "รวมข้อมูลลูกค้าไว้ในที่เดียว", "รองรับทีมงานและหลายร้านตามแพ็กเกจ"].map((text) => (
                  <Flex key={text} gap={10} align="center"><CheckCircleOutlined style={{ color: "#65725F" }} /><Typography.Text>{text}</Typography.Text></Flex>
                ))}
              </Space>
            </Space>
          </Col>
          <Col xs={24} lg={14} style={{ minHeight: 0, height: "100%", display: "flex", alignItems: "center" }}>
            <Card style={{ width: "100%", maxHeight: "100%", display: "flex", flexDirection: "column", overflow: "hidden", borderRadius: 20, boxShadow: "0 12px 40px rgba(40,55,35,0.06)" }} styles={{ body: { minHeight: 0, overflowY: "auto", padding: screens.sm ? "20px 24px" : 16 } }}>
              <Typography.Title level={3} style={{ marginTop: 0 }}>สมัครสมาชิก Tenant</Typography.Title>
              <Typography.Paragraph type="secondary">กรอกข้อมูลเพื่อสร้างบัญชี แล้วเพิ่มร้านค้าหลังเข้าสู่ระบบ</Typography.Paragraph>
              <Form<RegistrationForm> form={form} layout="vertical" requiredMark={false} onFinish={reviewRegistration}>
                <Row gutter={16}>
                  <Col span={24}><Divider style={{ margin: "12px 0" }} titlePlacement="start">ข้อมูลบัญชีผู้ดูแล</Divider></Col>
                  <Col xs={24} sm={12}><Form.Item style={{ marginBottom: 12 }} name="full_name" label="ชื่อ–นามสกุล" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อ–นามสกุล" }]}><Input autoComplete="name" maxLength={100} placeholder="ชื่อผู้ดูแลธุรกิจ" /></Form.Item></Col>
                  <Col xs={24} sm={12}><Form.Item style={{ marginBottom: 12 }} name="phone" label="เบอร์โทรศัพท์" rules={[{ required: true, message: "กรุณากรอกเบอร์โทรศัพท์" }, { validator: (_, value: string) => !value || /^(?:0[0-9]{8,9}|\+66[0-9]{8,9})$/.test(value.replace(/[\s-]/g, "")) ? Promise.resolve() : Promise.reject(new Error("กรุณากรอกเบอร์โทรศัพท์ให้ถูกต้อง เช่น 0812345678")) }]}><Input type="tel" autoComplete="tel" maxLength={20} placeholder="0812345678" /></Form.Item></Col>
                  <Col span={24}><Form.Item style={{ marginBottom: 12 }} name="email" label="อีเมล" rules={[{ required: true, message: "กรุณากรอกอีเมล" }, { type: "email", message: "กรุณากรอกอีเมลให้ถูกต้อง" }]}><Input autoComplete="email" type="email" maxLength={254} placeholder="you@example.com" /></Form.Item></Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      style={{ marginBottom: 12 }}
                      name="password"
                      label="รหัสผ่าน"
                      extra={password && !passwordIsValid ? (
                        <Typography.Text type="danger" aria-live="polite">
                          ตัวอักษร {letterCount}/3 · ตัวเลข {digitCount}/5
                        </Typography.Text>
                      ) : undefined}
                      validateTrigger="onChange"
                      rules={[
                        { required: true, message: "กรุณากรอกรหัสผ่าน" },
                        { len: 8, pattern: registrationPasswordPattern, message: "รหัสผ่านยังไม่ตรงตามเงื่อนไข" },
                      ]}
                    >
                      <Input.Password autoComplete="new-password" placeholder="ตั้งรหัสผ่าน" />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}><Form.Item style={{ marginBottom: 12 }} name="confirm_password" label="ยืนยันรหัสผ่าน" dependencies={["password"]} rules={[{ required: true, message: "กรุณายืนยันรหัสผ่าน" }, ({ getFieldValue }) => ({ validator: (_, value) => !value || value === getFieldValue("password") ? Promise.resolve() : Promise.reject(new Error("รหัสผ่านไม่ตรงกัน")) })]}><Input.Password autoComplete="new-password" placeholder="กรอกรหัสผ่านอีกครั้ง" /></Form.Item></Col>
                  <Col span={24}><Button block type="primary" htmlType="submit" disabled={!passwordIsValid}>ตรวจสอบข้อมูลสมัคร</Button></Col>
                </Row>
              </Form>
              <Typography.Paragraph style={{ textAlign: "center", marginTop: 16, marginBottom: 0 }}>มีบัญชีแล้ว? <Link to="/tenant/login">เข้าสู่ระบบ</Link></Typography.Paragraph>
            </Card>
          </Col>
        </Row>
      </div>
      <Modal
        title={registered ? "สมัครสมาชิกแล้ว" : "ตรวจสอบข้อมูลการสมัคร"}
        centered
        width={480}
        open={reviewOpen}
        onCancel={() => { if (!saving) setReviewOpen(false); }}
        closable={!saving}
        keyboard={!saving}
        mask={{ closable: !saving }}
        footer={
          <>
            {!registered && (
              <Typography.Text type="secondary" style={{ display: "block", marginTop: 12, marginBottom: 12, textAlign: "start" }}>
                สมัครแล้วรับแพ็กเกจ Free อัตโนมัติ และเพิ่มร้านค้าได้ภายหลัง
              </Typography.Text>
            )}
            <Flex justify="end" gap={8} wrap>
              {!registered && (
                <Button type="primary" loading={saving} onClick={() => void submitRegistration()}>
                  ยืนยัน
                </Button>
              )}
              <Button disabled={saving} onClick={() => setReviewOpen(false)}>
                {registered ? "ปิด" : "ยกเลิก"}
              </Button>
            </Flex>
          </>
        }
      >
        <Space orientation="vertical" size="large" style={{ width: "100%" }}>
          <Alert 
            type={registered ? "success" : "info"} 
            showIcon 
            title={registered ? "สร้างบัญชีเรียบร้อยแล้ว พร้อมแพ็กเกจ Free" : "กรุณาตรวจสอบข้อมูลก่อนยืนยัน"} 
            description={registered ? "เข้าสู่ระบบเพื่อเริ่มเพิ่มร้านค้าของคุณ" : undefined} 
          />
          <Descriptions 
            column={1} 
            bordered 
            items={[
              { key: "name", label: "ชื่อ–นามสกุล", children: review?.full_name },
              { key: "phone", label: "เบอร์โทรศัพท์", children: review?.phone },
              { key: "email", label: "อีเมล", children: review?.email },
            ]} 
          />
        </Space>
      </Modal>
    </main>
  );
}
