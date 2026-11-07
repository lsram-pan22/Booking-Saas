import { useCallback, useEffect, useState } from "react";
import { Alert, App, Button, Card, Drawer, Flex, Form, Input, InputNumber, Select, Space, Switch, Table, Tag, Typography } from "antd";
import { listPackages, savePackage } from "../../../api/super-admin/packages.api";
import type { Package, PackageForm } from "../../../types/super-admin/package.type";


const defaults: PackageForm = {
  code: "", name: "", description: "", monthly_price: 0, shop_limit: 1, user_limit: 2,
  staff_management: "basic", reports: "basic", export_excel: false, 
  api_webhook: false, active: true, sort_order: 0,
};
const price = (satang: number) => (satang / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" });

export default function Packages() {
  const { message } = App.useApp();
  const [form] = Form.useForm<PackageForm>();
  const [items, setItems] = useState<Package[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Package>();
  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try { setItems(await listPackages()); }
    catch (e) { setError(e instanceof Error ? e.message : "โหลดแพ็กเกจไม่สำเร็จ"); }
    finally { setLoading(false); }
  }, []);
  useEffect(() => {
    let cancelled = false;
    listPackages().then((data) => {
      if (!cancelled) setItems(data);
    }).catch((e: unknown) => {
      if (!cancelled) setError(e instanceof Error ? e.message : "โหลดแพ็กเกจไม่สำเร็จ");
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => { cancelled = true; };
  }, []);

  const edit = (item?: Package) => {
    setEditing(item);
    form.resetFields();
    form.setFieldsValue(item ? { ...item, monthly_price: item.monthly_price_satang / 100 } : defaults);
    setOpen(true);
  };
  const submit = async (values: PackageForm) => {
    setSaving(true);
    try {
      const { monthly_price, ...fields } = values;
      await savePackage({ ...fields, code: fields.code.trim(), name: fields.name.trim(), description: fields.description || "", monthly_price_satang: Math.round(monthly_price * 100) }, editing?.id);
      message.success("บันทึกแพ็กเกจแล้ว");
      setOpen(false);
      await load();
    } catch (e) { message.error(e instanceof Error ? e.message : "บันทึกไม่สำเร็จ"); }
    finally { setSaving(false); }
  };

  return <>
    <Card title="จัดการแพ็กเกจ" extra={<Button type="primary" onClick={() => edit()}>เพิ่มแพ็กเกจ</Button>}>
      <Space orientation="vertical" size="middle" style={{ width: "100%" }}>
        <Typography.Text type="secondary">ราคาเป็นรายเดือน • จองไม่จำกัด • จำนวนบัญชีผู้ใช้รวมเจ้าของร้าน • ทุกแพ็กเกจจัดการบริการ ลูกค้า และปฏิทินการจองได้</Typography.Text>
        {error && <Alert type="error" showIcon title={error} action={<Button onClick={() => void load()}>ลองใหม่</Button>} />}
        <Table<Package> rowKey="id" dataSource={items} loading={loading} scroll={{ x: 1150 }} pagination={{ pageSize: 10 }} columns={[
          { title: "แพ็กเกจ", dataIndex: "name", render: (_, p) => <><Typography.Text strong>{p.name}</Typography.Text><br /><Typography.Text type="secondary">{p.code}</Typography.Text></> },
          { title: "ราคา / เดือน", dataIndex: "monthly_price_satang", render: price },
          { title: "ร้านค้า", dataIndex: "shop_limit" },
          { title: "บัญชีผู้ใช้", dataIndex: "user_limit" },
          { title: "ฟีเจอร์", render: (_, p) => <Flex wrap gap={4}><Tag>พนักงาน: {p.staff_management === "basic" ? "พื้นฐาน" : "เต็มรูปแบบ"}</Tag><Tag>รายงาน: {p.reports === "basic" ? "พื้นฐาน" : "ขั้นสูง"}</Tag>{p.export_excel && <Tag>Excel</Tag>}{p.api_webhook && <Tag>API / Webhook</Tag>}</Flex> },
          { title: "สถานะ", dataIndex: "active", render: (active: boolean) => <Tag color={active ? "green" : "default"}>{active ? "เปิดรับสมัคร" : "ปิดรับสมัคร"}</Tag> },
          { title: "", render: (_, p) => <Button onClick={() => edit(p)}>แก้ไข</Button> },
        ]} />
      </Space>
    </Card>
    <Drawer
      title={editing ? "แก้ไขแพ็กเกจ" : "เพิ่มแพ็กเกจ"}
      placement="right"
      size="min(520px, 100vw)"
      open={open}
      onClose={() => { if (!saving) setOpen(false); }}
      mask={{ closable: !saving }}
      closable={!saving}
      keyboard={!saving}
      forceRender
      footer={<Flex justify="end" gap={8}><Button onClick={() => setOpen(false)} disabled={saving}>ยกเลิก</Button><Button type="primary" htmlType="submit" form="package-form" loading={saving}>บันทึก</Button></Flex>}
    >
      <Form id="package-form" form={form} layout="vertical" onFinish={submit} initialValues={defaults} disabled={saving}>
        <Form.Item name="code" label="รหัสแพ็กเกจ" rules={[{ required: true, pattern: /^[a-z][a-z0-9_-]{0,49}$/, message: "ใช้ตัวอักษรอังกฤษเล็ก ตัวเลข _ หรือ - และขึ้นต้นด้วยตัวอักษร" }]}><Input maxLength={50} /></Form.Item>
        <Form.Item name="name" label="ชื่อแพ็กเกจ" rules={[{ required: true, whitespace: true }]}><Input maxLength={100} /></Form.Item>
        <Form.Item name="description" label="รายละเอียด"><Input.TextArea maxLength={2000} rows={2} /></Form.Item>
        <Form.Item name="monthly_price" label="ราคา / เดือน (บาท)" rules={[{ required: true }]}><InputNumber min={0} max={1000000} precision={2} style={{ width: "100%" }} /></Form.Item>
        <Flex gap={16}>
          <Form.Item name="shop_limit" label="จำนวนร้านค้า" rules={[{ required: true }]}><InputNumber min={1} max={100000} precision={0} /></Form.Item>
          <Form.Item name="user_limit" label="บัญชีผู้ใช้ (รวม Owner)" rules={[{ required: true }]}><InputNumber min={1} max={100000} precision={0} /></Form.Item>
        </Flex>
        <Form.Item name="staff_management" label="จัดการพนักงาน" rules={[{ required: true }]}><Select options={[{ value: "basic", label: "พื้นฐาน" }, { value: "full", label: "เต็มรูปแบบ" }]} /></Form.Item>
        <Form.Item name="reports" label="รายงาน" rules={[{ required: true }]}><Select options={[{ value: "basic", label: "พื้นฐาน" }, { value: "advanced", label: "ขั้นสูง" }]} /></Form.Item>
        <Form.Item name="export_excel" label="Export Excel" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="api_webhook" label="API / Webhook" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="sort_order" label="ลำดับแสดงผล" rules={[{ required: true }]}><InputNumber min={0} max={100000} precision={0} /></Form.Item>
        <Form.Item name="active" label="เปิดให้ Owner สมัคร" valuePropName="checked"><Switch /></Form.Item>
        <Alert type="info" showIcon title="ใช้การปิดรับสมัครแทนการลบแพ็กเกจ เพื่อเก็บข้อมูลสำหรับการสมัครสมาชิก" style={{ marginBottom: 16 }} />
      </Form>
    </Drawer>
  </>;
}
