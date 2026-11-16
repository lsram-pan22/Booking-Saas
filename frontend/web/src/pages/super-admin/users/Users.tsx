import { useEffect, useState } from "react";
import { Alert, Button, Card, Input, Space, Table, Typography } from "antd";
import { listResettableOwners, resetOwnerPassword } from "../../../api/tenant/password.api";
import ResetPasswordDrawer from "../../../components/shops/tenant/ResetPasswordDrawer";
import type { ResettableOwner } from "../../../types/shared/password.type";

export default function Users() {
  const [items, setItems] = useState<ResettableOwner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [resetting, setResetting] = useState<ResettableOwner>();
  useEffect(() => {
    let active = true;
    listResettableOwners(search).then(data => { if (active) { setItems(data); setError(""); } }).catch(e => { if (active) setError(e instanceof Error ? e.message : "โหลดบัญชีไม่สำเร็จ"); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [search]);
  return <Card title="บัญชีเจ้าของร้าน">
    <Space orientation="vertical" size="large" style={{ width: "100%" }}>
      <Typography.Text type="secondary">ค้นหาบัญชีเจ้าของร้านเพื่อรีเซ็ตรหัสผ่าน แสดงสูงสุด 100 บัญชีต่อการค้นหา</Typography.Text>
      <Input.Search placeholder="ค้นหาชื่อ อีเมล หรือ Username" allowClear onSearch={value => setSearch(value.trim())} />
      {error && <Alert type="error" showIcon title={error} />}
      <Table<ResettableOwner> rowKey="id" dataSource={items} loading={loading} scroll={{ x: 650 }} columns={[
        { title: "ชื่อ", dataIndex: "full_name" }, { title: "Username", dataIndex: "username" }, { title: "อีเมล", dataIndex: "email" },
        { title: "จัดการ", key: "actions", render: (_, owner) => <Button disabled={Boolean(error)} onClick={() => setResetting(owner)}>รีเซ็ตรหัสผ่าน</Button> },
      ]} />
    </Space>
    {resetting && <ResetPasswordDrawer name={resetting.full_name} open onClose={() => setResetting(undefined)} onReset={password => resetOwnerPassword(resetting.id, password)} />}
  </Card>;
}
