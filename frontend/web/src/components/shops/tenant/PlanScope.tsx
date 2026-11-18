import { useEffect, useState } from "react";
import { Alert, App, Button, Card, Checkbox, Space, Table, Tag, Typography } from "antd";
import { getPlanOverview, selectFreePlanScope } from "../../../api/tenant/plan.api";
import type { PlanOverview } from "../../../types/tenant/plan.type";
import type { TenantSubscription } from "../../../types/tenant/account.type";

export default function PlanScope({ subscription }: { subscription: TenantSubscription }) {
  const [data, setData] = useState<PlanOverview>();
  const [shops, setShops] = useState<string[]>([]);
  const [members, setMembers] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const { message } = App.useApp();
  useEffect(() => {
    let active = true;
    getPlanOverview().then(result => {
      if (!active) return;
      const current = result.history.find(s => s.status === "active" && new Date(s.starts_at).getTime() <= Date.now() && (!s.ends_at || new Date(s.ends_at).getTime() > Date.now()));
      if (current && current.id !== subscription.id) { window.location.reload(); return; }
      setData(result); setShops(result.shops.filter(s => s.enabled).map(s => s.id));
      setMembers(result.shops.flatMap(s => s.members.filter(m => m.enabled && m.role === "staff").map(m => m.id)));
    }).catch(e => { if (active) setError(e instanceof Error ? e.message : "โหลดสิทธิ์ไม่สำเร็จ"); });
    return () => { active = false; };
  }, [subscription.id]);
  const save = async () => {
    setSaving(true);
    try {
      const eligible = data?.shops.filter(s => shops.includes(s.id)).flatMap(s => s.members.map(m => m.id)) || [];
      await selectFreePlanScope({ shop_ids: shops, member_ids: members.filter(id => eligible.includes(id)) });
      message.success("บันทึกสิทธิ์ Free แล้ว"); window.location.reload();
    } catch (e) { message.error(e instanceof Error ? e.message : "บันทึกสิทธิ์ไม่สำเร็จ"); setSaving(false); }
  };
  return <>
    {error && <Alert type="error" showIcon title={error} />}
    {subscription.package_code === "free" && <Card title={`เลือกร้านที่ใช้สิทธิ์ Free (${shops.length}/${subscription.shop_limit})`} loading={!data && !error}>
      <Space orientation="vertical" size="large" style={{ width: "100%" }}>
        <Alert type={subscription.selection_required ? "warning" : "info"} showIcon title={subscription.selection_required ? "แพ็กเกจเดิมหมดอายุและข้อมูลเกินโควตา Free กรุณาเลือกสิทธิ์ที่จะใช้ต่อ" : "เลือกสิทธิ์ใช้งานตามแพ็กเกจ Free"} description="ข้อมูลร้าน พนักงาน และนัดเดิมยังอยู่ทั้งหมด ส่วนที่ไม่ได้เลือกยังดูประวัติและจัดการสถานะนัดเดิมได้ตามสิทธิ์ แต่รับงานใหม่หรือแก้นัดไม่ได้ จำนวนผู้ใช้ต่อร้านรวมเจ้าของด้วย" />
        {data?.shops.map(shop => {
          const selected = shops.includes(shop.id);
          const used = 1 + shop.members.filter(m => m.role === "staff" && members.includes(m.id)).length;
          return <Card key={shop.id} size="small" title={<Checkbox checked={selected} disabled={saving || (!selected && shops.length >= subscription.shop_limit)} onChange={e => {
            setShops(current => e.target.checked ? [...current, shop.id] : current.filter(id => id !== shop.id));
            if (!e.target.checked) setMembers(current => current.filter(id => !shop.members.some(m => m.id === id)));
          }}>{shop.name}</Checkbox>}>
            <Space orientation="vertical"><Typography.Text>ผู้ใช้ {selected ? used : 0}/{subscription.user_limit} คน</Typography.Text>
              {shop.members.map(member => member.role === "owner" ? <Typography.Text key={member.id}>{member.name} — เจ้าของร้าน (รวมในโควตา)</Typography.Text> : <Checkbox key={member.id} checked={selected && members.includes(member.id)} disabled={saving || !selected || (!members.includes(member.id) && used >= subscription.user_limit)} onChange={e => setMembers(current => e.target.checked ? [...current, member.id] : current.filter(id => id !== member.id))}>{member.name}</Checkbox>)}
            </Space>
          </Card>;
        })}
        <Button type="primary" loading={saving} disabled={!data || Boolean(error) || shops.length > subscription.shop_limit || (Boolean(data.shops.length) && !shops.length)} onClick={() => void save()}>ยืนยันสิทธิ์ Free</Button>
      </Space>
    </Card>}
    <Card title="ประวัติแพ็กเกจ (ล่าสุด 100 รายการ)" loading={!data && !error} style={{ marginTop: 16 }}>
      <Table<TenantSubscription> rowKey="id" dataSource={data?.history || []} scroll={{ x: 600 }} columns={[
        { title: "แพ็กเกจ", dataIndex: "package_name" },
        { title: "สถานะ", dataIndex: "status", render: (status: string) => <Tag color={status === "active" ? "green" : "default"}>{status === "active" ? "ใช้งานอยู่" : status === "expired" ? "หมดอายุ" : status}</Tag> },
        { title: "เริ่ม", dataIndex: "starts_at", render: (date: string) => new Date(date).toLocaleDateString("th-TH", { timeZone: "Asia/Bangkok" }) },
        { title: "สิ้นสุด", dataIndex: "ends_at", render: (date: string | null) => date ? new Date(date).toLocaleDateString("th-TH", { timeZone: "Asia/Bangkok" }) : "ไม่มีวันหมดอายุ" },
      ]} />
    </Card>
  </>;
}
