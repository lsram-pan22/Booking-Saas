import { Alert, Button, Card, Space, Switch, Table, Tag, Typography } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import type { ShopMember, ShopMembersCardProps } from "../../../types/tenant/shop.type";

const statusLabel = (status: string) => ({ active: "เปิดใช้งาน", inactive: "ปิดใช้งาน", suspended: "ระงับ" })[status] || status;

export default function ShopMembersCard({ data, error, canAdd, onAdd, canManage, updatingMemberID, onStatusChange }: ShopMembersCardProps) {
  return <Card title={`ผู้ใช้งาน${data ? ` (${data.active_count}/${data.user_limit ?? "—"})` : ""}`} extra={<Button type="primary" icon={<PlusOutlined />} disabled={!canAdd || !data || data.user_limit === null || data.active_count >= data.user_limit || Boolean(error) || Boolean(updatingMemberID)} onClick={onAdd}>เพิ่มผู้ใช้งาน</Button>}>
    <Space orientation="vertical" size="middle" style={{ width: "100%" }}>
      <Typography.Text type="secondary">นับสมาชิกที่เปิดใช้งาน รวมเจ้าของร้าน</Typography.Text>
      {error ? <Alert showIcon type="error" title={error} action={<Button onClick={() => window.location.reload()}>ลองใหม่</Button>} /> : <>
        {data?.user_limit === null && <Alert type="warning" showIcon title="ไม่พบแพ็กเกจที่เปิดใช้งาน จึงยังไม่สามารถระบุโควตาผู้ใช้งานได้" />}
        <Table<ShopMember> rowKey="id" dataSource={data?.members || []} loading={!data} scroll={{ x: 800 }} pagination={{ pageSize: 10, hideOnSinglePage: true }} columns={[
          { title: "ชื่อ–นามสกุล", dataIndex: "full_name" },
          { title: "Username", dataIndex: "username" },
          { title: "อีเมล", dataIndex: "email" },
          { title: "เบอร์โทรศัพท์", dataIndex: "phone", render: (phone: string) => phone || "—" },
          { title: "สิทธิ์", dataIndex: "role", render: (role: string) => <Tag color={role === "owner" ? "gold" : "blue"}>{role === "owner" ? "เจ้าของร้าน" : "พนักงาน"}</Tag> },
          { title: "สถานะสมาชิก", dataIndex: "status", render: (status: string) => <Tag color={status === "active" ? "green" : "default"}>{statusLabel(status)}</Tag> },
          { title: "เปิด/ปิด", key: "toggle", render: (_, member) => <Switch
            aria-label={`เปิดหรือปิดสมาชิก ${member.full_name}`}
            checked={member.status === "active"}
            loading={updatingMemberID === member.id}
            checkedChildren="เปิด" unCheckedChildren="ปิด"
            disabled={!canManage || member.role !== "staff" || member.status === "suspended" || Boolean(updatingMemberID) || (member.status !== "active" && (!data?.user_limit || data.active_count >= data.user_limit || member.user_status !== "active"))}
            onChange={(active) => onStatusChange(member, active)}
          /> },
          { title: "สถานะบัญชี", dataIndex: "user_status", render: (status: string) => statusLabel(status) },
        ]} />
      </>}
    </Space>
  </Card>;
}
