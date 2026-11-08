import { Alert, Button, Card, Col, Empty, Flex, Row, Space, Spin, Tag, Typography } from "antd";
import { Link, Navigate, useOutletContext } from "react-router-dom";
import { isTenantAuthenticated } from "../../../utils/auth";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";
import CreateShopDrawer from "../../../components/shops/tenant/CreateShopDrawer";
import { shopStatusLabels } from "../../../types/tenant/shop.type";

export default function Dashboard() {
  const { profile, error, addShop } = useOutletContext<TenantLayoutContext>();
  const [creating, setCreating] = useState(false);
  const shopLimit = profile?.subscription?.shop_limit ?? 0;
  const ownedShopCount = profile?.owned_shop_count ?? 0;
  if (!isTenantAuthenticated()) return <Navigate to="/tenant/login" replace />;
  return <main>
    <Space orientation="vertical" size="large" style={{ width: "100%" }}>
      <Flex justify="space-between" align="center" wrap gap={16}>
        <div><Typography.Title level={3} style={{ margin: 0 }}>ร้านค้าของฉัน {profile && <Typography.Text type="secondary" style={{ fontSize: 20 }}>({ownedShopCount}/{shopLimit})</Typography.Text>}</Typography.Title><Typography.Text type="secondary">{profile?.user.full_name}</Typography.Text></div>
        {profile && <Button type="primary" icon={<PlusOutlined />} disabled={!profile.subscription || ownedShopCount >= shopLimit} onClick={() => setCreating(true)}>เพิ่มร้านค้า</Button>}
      </Flex>
      {error && <Alert type="error" showIcon title={error} action={<Button onClick={() => window.location.reload()}>ลองใหม่</Button>} />}
      {!profile && !error && <Flex justify="center"><Spin description="กำลังตรวจสอบสิทธิ์และโหลดร้านค้า"><div style={{ width: 220, height: 80 }} /></Spin></Flex>}
      {profile && !profile.subscription && <Alert type="warning" showIcon title="ต้องมีแพ็กเกจที่เปิดใช้งานก่อนเพิ่มร้านค้า" />}
      {profile && profile.tenants.length === 0 && <Card><Empty description="คุณยังไม่มีร้านค้า เริ่มต้นด้วยการเพิ่มร้านค้าของคุณ" /></Card>}
      <Row gutter={[16, 16]}>
        {profile?.tenants.map((tenant) => <Col key={tenant.id} xs={24} md={12} xl={8}>
          <Card title={tenant.name} extra={<Tag>{tenant.role === "owner" ? "เจ้าของร้าน" : "พนักงาน"}</Tag>}>
            <Space orientation="vertical" style={{ width: "100%" }}>
              <Tag color={tenant.status === "active" ? "green" : tenant.status === "pending" ? "gold" : "default"}>{shopStatusLabels[tenant.status] || tenant.status}</Tag>
              <Typography.Text type="secondary">{tenant.status === "pending" ? "ร้านค้าอยู่ระหว่างรอเปิดใช้งาน" : tenant.status === "inactive" ? "ร้านค้าปิดใช้งาน เจ้าของร้านเปิดกลับได้ในข้อมูลร้านค้า" : "ร้านค้าพร้อมใช้งาน"}</Typography.Text>
              <Link to={tenant.status === "active" ? `/tenant/shops/${tenant.id}/bookings` : `/tenant/shops/${tenant.id}`}><Button type="primary">จัดการร้านค้า</Button></Link>
            </Space>
          </Card>
        </Col>)}
      </Row>
    </Space>
    <CreateShopDrawer shopLimit={shopLimit} open={creating} onClose={() => setCreating(false)} onCreated={addShop} />
  </main>;
}
import { useState } from "react";
import { PlusOutlined } from "@ant-design/icons";
