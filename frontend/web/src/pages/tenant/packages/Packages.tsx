import { Alert, Button, Card, Col, Descriptions, Empty, Row, Space, Spin, Tag, Typography } from "antd";
import { useOutletContext } from "react-router-dom";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";

export default function Packages() {
  const { profile, error } = useOutletContext<TenantLayoutContext>();
  if (error) return <Alert type="error" showIcon title={error} action={<Button onClick={() => window.location.reload()}>ลองใหม่</Button>} />;
  if (!profile) return <Card loading><Spin /></Card>;
  const subscription = profile.subscription;
  return <Space orientation="vertical" size="large" style={{ width: "100%" }}>
    <Typography.Title level={3} style={{ margin: 0 }}>แพ็กเกจของฉัน</Typography.Title>
    {!subscription ? <Card><Empty description="ยังไม่มีแพ็กเกจที่เปิดใช้งาน" /></Card> : <Row gutter={[16, 16]}>
      <Col xs={24} xl={8}><Card title="แพ็กเกจปัจจุบัน">
        <Space orientation="vertical" size="middle">
          <Tag color="green">{subscription.package_name}</Tag>
          <Typography.Title level={2} style={{ margin: 0 }}>{(subscription.monthly_price_satang / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" })}<Typography.Text type="secondary"> / เดือน</Typography.Text></Typography.Title>
          <Typography.Text>เริ่มใช้งาน {new Date(subscription.starts_at).toLocaleDateString("th-TH", { timeZone: "Asia/Bangkok" })}</Typography.Text>
          <Typography.Text type="secondary">{subscription.ends_at ? `สิ้นสุด ${new Date(subscription.ends_at).toLocaleDateString("th-TH", { timeZone: "Asia/Bangkok" })}` : "ไม่มีวันหมดอายุ"}</Typography.Text>
        </Space>
      </Card></Col>
      <Col xs={24} xl={16}><Card title="สิทธิ์ที่ได้รับ"><Descriptions bordered column={1} items={[
        { key: "shops", label: "จำนวนร้านค้าสูงสุด", children: subscription.shop_limit },
        { key: "users", label: "บัญชีผู้ใช้รวมเจ้าของร้าน", children: subscription.user_limit },
        { key: "bookings", label: "จำนวนการจอง", children: "ไม่จำกัด" },
        { key: "staff", label: "จัดการพนักงาน", children: subscription.staff_management_level === "full" ? "เต็มรูปแบบ" : "พื้นฐาน" },
        { key: "reports", label: "รายงาน", children: subscription.report_level === "advanced" ? "ขั้นสูง" : "พื้นฐาน" },
        { key: "excel", label: "Export Excel", children: subscription.export_excel ? "ใช้งานได้" : "ไม่รวมในแพ็กเกจ" },
        { key: "api", label: "API / Webhook", children: subscription.api_webhook ? "ใช้งานได้" : "ไม่รวมในแพ็กเกจ" },
      ]} /></Card></Col>
    </Row>}
  </Space>;
}
