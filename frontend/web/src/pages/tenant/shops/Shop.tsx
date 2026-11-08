import { useEffect, useState } from "react";
import { Alert, App, Button, Card, Col, Descriptions, Drawer, Flex, Form, Input, Row, Select, Space, Spin, Switch, Tag, Typography } from "antd";
import { ArrowLeftOutlined, EditOutlined } from "@ant-design/icons";
import { Link, useOutletContext, useParams } from "react-router-dom";
import { getShop, getShopMembers, setShopMemberStatus, updateShop } from "../../../api/tenant/shops.api";
import { businessTypes, shopStatusLabels } from "../../../types/tenant/shop.type";
import type { EditShopForm, ShopDetails, ShopDetailsProps, ShopMember, ShopMembers } from "../../../types/tenant/shop.type";
import ShopMembersCard from "../../../components/shops/tenant/ShopMembersCard";
import AddMemberDrawer from "../../../components/shops/tenant/AddMemberDrawer";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";

function ShopDetailsView({ shopID }: ShopDetailsProps) {
  const { updateShop: updateShopInLayout } = useOutletContext<TenantLayoutContext>();
  const { message } = App.useApp();
  const [form] = Form.useForm<EditShopForm>();
  const [shop, setShop] = useState<ShopDetails>();
  const [error, setError] = useState("");
  const [members, setMembers] = useState<ShopMembers>();
  const [membersError, setMembersError] = useState("");
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [addingMember, setAddingMember] = useState(false);
  const [updatingMemberID, setUpdatingMemberID] = useState<string>();
  const changeMemberStatus = async (member: ShopMember, active: boolean) => {
    if (updatingMemberID) return;
    setUpdatingMemberID(member.id);
    try {
      const result = await setShopMemberStatus(shopID, member.id, { status: active ? "active" : "inactive" });
      setMembers((current) => current ? { ...current, active_count: result.active_count, user_limit: result.user_limit, members: current.members.map((item) => item.id === result.member.id ? result.member : item) } : current);
      message.success(active ? "เปิดใช้งานสมาชิกแล้ว" : "ปิดใช้งานสมาชิกแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "เปลี่ยนสถานะไม่สำเร็จ"); }
    finally { setUpdatingMemberID(undefined); }
  };
  const memberAdded = (member: ShopMember) => {
    setMembers((current) => current ? { ...current, active_count: current.active_count + (member.status === "active" ? 1 : 0), members: [...current.members, member] } : current);
  };
  useEffect(() => {
    const controller = new AbortController();
    getShop(shopID, controller.signal).then(async (data) => {
      if (controller.signal.aborted) return;
      setShop(data);
      if (data.role === "owner") {
        try {
          const memberData = await getShopMembers(shopID, controller.signal);
          if (!controller.signal.aborted) setMembers(memberData);
        } catch (e) {
          if (!controller.signal.aborted) setMembersError(e instanceof Error ? e.message : "โหลดผู้ใช้งานไม่สำเร็จ");
        }
      }
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "โหลดร้านค้าไม่สำเร็จ");
    });
    return () => controller.abort();
  }, [shopID]);
  const submit = async (values: EditShopForm) => {
    setSaving(true);
    try {
      const data = await updateShop(shopID, { name: values.name.trim(), business_type: values.business_type, status: values.active ? "active" : "inactive", opening_time: values.opening_time, closing_time: values.closing_time });
      setShop(data); updateShopInLayout(data); setEditing(false); message.success("บันทึกข้อมูลร้านค้าแล้ว");
    } catch (e) { message.error(e instanceof Error ? e.message : "บันทึกไม่สำเร็จ"); }
    finally { setSaving(false); }
  };
  return <Space orientation="vertical" size="large" style={{ width: "100%" }}>
    <Link to="/tenant/dashboard"><ArrowLeftOutlined /> กลับไปร้านค้าของฉัน</Link>
    {error ? <Alert type="error" showIcon title={error} action={<Button onClick={() => window.location.reload()}>ลองใหม่</Button>} /> : !shop ? <Flex justify="center"><Spin /></Flex> : <>
      <Flex justify="space-between" align="center" wrap gap={16}>
        <div><Typography.Title level={3} style={{ margin: 0 }}>{shop.name}</Typography.Title><Typography.Text type="secondary">จัดการข้อมูลร้านค้า</Typography.Text></div>
        {shop.role === "owner" && ["active", "inactive"].includes(shop.status) && <Button icon={<EditOutlined />} onClick={() => { form.setFieldsValue({ name: shop.name, business_type: shop.business_type, active: shop.status === "active", opening_time: shop.opening_time || "09:00", closing_time: shop.closing_time || "18:00" }); setEditing(true); }}>แก้ไขข้อมูลร้านค้า</Button>}
      </Flex>
      <Row gutter={[16, 16]}><Col xs={24}><Card title="ข้อมูลร้านค้า"><Descriptions bordered column={1} items={[
        { key: "name", label: "ชื่อร้านค้า", children: shop.name },
        { key: "type", label: "ประเภทธุรกิจ", children: businessTypes.find((type) => type.value === shop.business_type)?.label || shop.business_type },
        { key: "hours", label: "เวลาเปิด–ปิดทุกวัน", children: shop.opening_time && shop.closing_time ? `${shop.opening_time}–${shop.closing_time} น.` : "ยังไม่ได้ตั้งค่า" },
        { key: "status", label: "สถานะ", children: <Tag color={shop.status === "active" ? "green" : shop.status === "pending" ? "gold" : "default"}>{shopStatusLabels[shop.status] || shop.status}</Tag> },
        { key: "role", label: "สิทธิ์ของคุณ", children: shop.role === "owner" ? "เจ้าของร้าน" : "พนักงาน" },
        ...(shop.role === "owner" ? [{ key: "users", label: "จำนวนผู้ใช้งาน", children: members ? `${members.active_count}/${members.user_limit ?? "—"} คน (รวมเจ้าของร้าน)` : membersError ? "โหลดข้อมูลไม่สำเร็จ" : "กำลังโหลด…" }] : []),
      ]} /></Card></Col></Row>
      {shop.role === "owner" && <Row><Col span={24}><ShopMembersCard data={members} error={membersError} canAdd={shop.status === "active"} onAdd={() => setAddingMember(true)} canManage={["active", "inactive"].includes(shop.status)} updatingMemberID={updatingMemberID} onStatusChange={(member, active) => { void changeMemberStatus(member, active); }} /></Col></Row>}
      {shop.role !== "owner" && <Alert type="info" showIcon title="คุณดูข้อมูลร้านค้าได้ การแก้ไขข้อมูลสงวนสิทธิ์ให้เจ้าของร้าน" />}
    </>}
    <Drawer title="แก้ไขข้อมูลร้านค้า" placement="right" size="min(480px, 100vw)" open={editing} forceRender onClose={() => { if (!saving) setEditing(false); }} closable={!saving} keyboard={!saving} mask={{ closable: !saving }} footer={<Flex justify="end" gap={8}><Button disabled={saving} onClick={() => setEditing(false)}>ยกเลิก</Button><Button type="primary" htmlType="submit" form="edit-shop-form" loading={saving}>บันทึก</Button></Flex>}>
      <Form<EditShopForm> id="edit-shop-form" form={form} layout="vertical" onFinish={submit} disabled={saving}>
        <Row gutter={16}>
          <Col span={24}><Form.Item name="name" label="ชื่อร้านค้า" rules={[{ required: true, whitespace: true, message: "กรุณากรอกชื่อร้านค้า" }]}><Input maxLength={150} /></Form.Item></Col>
          <Col span={24}><Form.Item name="business_type" label="ประเภทธุรกิจ" rules={[{ required: true, message: "กรุณาเลือกประเภทธุรกิจ" }]}><Select options={businessTypes} /></Form.Item></Col>
          <Col xs={24} sm={12}><Form.Item name="opening_time" label="เวลาเปิดร้าน" rules={[{ required: true, message: "กรุณาระบุเวลาเปิด" }, { pattern: /^([01]\d|2[0-3]):[0-5]\d$/, message: "ใช้รูปแบบ HH:mm เช่น 09:00" }]}><Input placeholder="09:00" maxLength={5} /></Form.Item></Col>
          <Col xs={24} sm={12}><Form.Item name="closing_time" label="เวลาปิดร้าน" dependencies={["opening_time"]} extra="เวลาเดียวกันทุกวัน ปิดเที่ยงคืนใช้ 24:00" rules={[{ required: true, message: "กรุณาระบุเวลาปิด" }, { pattern: /^(([01]\d|2[0-3]):[0-5]\d|24:00)$/, message: "ใช้รูปแบบ HH:mm เช่น 18:00" }, ({ getFieldValue }) => ({ validator: (_, value: string) => !value || value > getFieldValue("opening_time") ? Promise.resolve() : Promise.reject(new Error("เวลาปิดต้องหลังเวลาเปิดในวันเดียวกัน")) })]}><Input placeholder="18:00" maxLength={5} /></Form.Item></Col>
          <Col span={24}><Form.Item name="active" label="เปิดใช้งานร้านค้า" valuePropName="checked" extra="ปิดใช้งานแล้วเจ้าของร้านยังเข้ามาเปิดกลับได้ ร้านยังนับรวมในโควตาแพ็กเกจ"><Switch checkedChildren="เปิด" unCheckedChildren="ปิด" /></Form.Item></Col>
        </Row>
      </Form>
    </Drawer>
    <AddMemberDrawer shopID={shopID} open={addingMember} onClose={() => setAddingMember(false)} onAdded={memberAdded} />
  </Space>;
}

export default function Shop() {
  const { shopID } = useParams();
  return shopID ? <ShopDetailsView key={shopID} shopID={shopID} /> : <Alert type="error" title="ไม่พบรหัสร้านค้า" />;
}
