import { useState } from "react";
import { App, Button, Card, Col, Flex, Form, Input, Row, Space, Typography } from "antd";
import { CalendarOutlined, LockOutlined, UserOutlined } from "@ant-design/icons";
import { Link, useNavigate } from "react-router-dom";
import { loginTenant } from "../../../api/tenant/account.api";
import { saveAuthSession } from "../../../utils/auth";
import { getTenantLandingPath } from "../../../utils/tenant-access";
import type { TenantLoginRequest } from "../../../types/tenant/account.type";

export default function Login() {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const submit = async (values: TenantLoginRequest) => {
    setLoading(true);
    try {
      const session = await loginTenant({ username: values.username.trim().toLowerCase(), password: values.password });
      saveAuthSession(session);
      navigate(getTenantLandingPath(session), { replace: true });
    } catch (error) { message.error(error instanceof Error ? error.message : "เชื่อมต่อระบบไม่สำเร็จ"); }
    finally { setLoading(false); }
  };
  return <main style={{ height: "100dvh", background: "#F3F5F2", overflowY: "auto" }}>
    <Row justify="center" align="middle" style={{ minHeight: "100%", padding: 24, boxSizing: "border-box" }}>
      <Col xs={24} sm={18} md={12} lg={9} xl={7}>
        <Card style={{ borderRadius: 20 }}>
          <Flex vertical align="center" gap={8} style={{ marginBottom: 24 }}>
            <CalendarOutlined style={{ fontSize: 36, color: "#65725F" }} />
            <Typography.Title level={3} style={{ margin: 0 }}>เข้าสู่ระบบจัดการร้านค้า</Typography.Title>
            <Typography.Text type="secondary">Booking SaaS สำหรับเจ้าของธุรกิจและทีมงาน</Typography.Text>
          </Flex>
          <Form<TenantLoginRequest> layout="vertical" size="large" onFinish={submit} disabled={loading} requiredMark={false}>
            <Form.Item name="username" label="Username หรือ Email" rules={[{ required: true, whitespace: true, message: "กรุณากรอก Username หรืออีเมล" }]}><Input prefix={<UserOutlined />} autoComplete="username" placeholder="Username หรือ you@example.com" maxLength={254} /></Form.Item>
            <Form.Item name="password" label="Password" rules={[{ required: true, message: "กรุณากรอกรหัสผ่าน" }]}><Input.Password prefix={<LockOutlined />} autoComplete="current-password" placeholder="รหัสผ่านของคุณ" /></Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>เข้าสู่ระบบ</Button>
          </Form>
          <Space style={{ marginTop: 24, width: "100%", justifyContent: "center" }}><Typography.Text type="secondary">ยังไม่มีบัญชี?</Typography.Text><Link to="/tenant/register">สมัครสมาชิก</Link></Space>
        </Card>
      </Col>
    </Row>
  </main>;
}
