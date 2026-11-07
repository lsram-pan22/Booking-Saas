import { useEffect, useState } from "react";
import { Alert, Button, Card, Col, Flex, Row, Table, Tag, Typography } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import { useParams } from "react-router-dom";
import { getServices } from "../../../api/tenant/services.api";
import ServiceDrawer from "../../../components/services/tenant/ServiceDrawer";
import type { ServiceList, ShopService, ShopServicesProps } from "../../../types/tenant/service.type";

function ShopServices({ shopID }: ShopServicesProps) {
  const [data, setData] = useState<ServiceList>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ShopService>();
  useEffect(() => {
    const controller = new AbortController();
    getServices(shopID, controller.signal).then(result => { if (!controller.signal.aborted) setData(result); }).catch(e => {
      if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "โหลดข้อมูลไม่สำเร็จ");
    }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [shopID, retry]);
  const canManage = data?.role === "owner" && ["active", "inactive"].includes(data.shop_status);
  const saved = (service: ShopService) => setData(previous => previous ? { ...previous, items: previous.items.some(item => item.id === service.id) ? previous.items.map(item => item.id === service.id ? service : item) : [...previous.items, service] } : previous);
  return <Row gutter={[24, 24]}>
    <Col span={24}><Flex justify="space-between" align="center" gap={16} wrap>
      <div><Typography.Title level={3} style={{ margin: 0 }}>บริการของร้านค้า</Typography.Title><Typography.Text type="secondary">{data?.shop_name}</Typography.Text></div>
      {canManage && <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(undefined); setOpen(true); }}>เพิ่มบริการ</Button>}
    </Flex></Col>
    <Col span={24}>{error ? <Alert type="error" showIcon title={error} action={<Button loading={loading} onClick={() => { setLoading(true); setError(""); setRetry(value => value + 1); }}>ลองใหม่</Button>} /> : <Card>
      {data?.role === "staff" && <Typography.Paragraph type="secondary">แสดงบริการที่เปิดใช้งานของร้านค้า</Typography.Paragraph>}
      <Table<ShopService> rowKey="id" loading={loading} dataSource={data?.items || []} scroll={{ x: 650 }} locale={{ emptyText: "ยังไม่มีบริการ" }} columns={[
        { title: "บริการ", dataIndex: "name", render: (name: string, service) => <><Typography.Text strong>{name}</Typography.Text>{service.description && <Typography.Paragraph type="secondary" ellipsis={{ rows: 2 }} style={{ marginBottom: 0 }}>{service.description}</Typography.Paragraph>}</> },
        { title: "ระยะเวลา", dataIndex: "duration_minutes", render: (value: number) => `${value} นาที`, width: 120 },
        { title: "ราคา", dataIndex: "price_satang", render: (value: number) => (value / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" }), width: 140 },
        { title: "สถานะ", dataIndex: "status", render: (value: string) => <Tag color={value === "active" ? "green" : "default"}>{value === "active" ? "เปิดใช้งาน" : "ปิดใช้งาน"}</Tag>, width: 120 },
        ...(canManage ? [{ title: "จัดการ", key: "actions", width: 100, render: (_: unknown, service: ShopService) => <Button onClick={() => { setEditing(service); setOpen(true); }}>แก้ไข</Button> }] : []),
      ]} />
    </Card>}</Col>
    {canManage && <ServiceDrawer shopID={shopID} open={open} service={editing} onClose={() => setOpen(false)} onSaved={saved} />}
  </Row>;
}

export default function Services() {
  const { shopID } = useParams();
  return shopID ? <ShopServices key={shopID} shopID={shopID} /> : <Alert type="error" title="ไม่พบร้านค้า" />;
}
