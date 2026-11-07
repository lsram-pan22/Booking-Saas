import { Alert, Card, Empty, Spin, Typography } from "antd";
import { useOutletContext, useParams } from "react-router-dom";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";

export default function ShopDashboard() {
  const { profile, error } = useOutletContext<TenantLayoutContext>();
  const { shopID } = useParams();
  if (error) return <Alert type="error" showIcon title={error} />;
  if (!profile) return <Spin />;
  const shop = profile.tenants.find((item) => item.id === shopID && item.status === "active");
  if (!shop) return <Alert type="warning" showIcon title="ยังไม่มีร้านค้าที่คุณเข้าใช้งานได้" description="กรุณาติดต่อเจ้าของร้านเพื่อตรวจสอบสิทธิ์และสถานะร้านค้า" />;
  return <>
    <Typography.Title level={3} style={{ marginTop: 0 }}>Dashboard — {shop.name}</Typography.Title>
    <Card><Empty description="ยังไม่มีข้อมูล Dashboard" /></Card>
  </>;
}
