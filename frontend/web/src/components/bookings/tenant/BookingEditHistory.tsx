import { useEffect, useState } from "react";
import { Alert, Button, Divider, Flex, Table, Typography } from "antd";
import { getBookingEditHistory } from "../../../api/tenant/bookings.api";
import type { BookingEdit, BookingEditSnapshot } from "../../../types/tenant/booking.type";

interface BookingEditHistoryProps {
  shopID: string;
  bookingID: string;
  formatTime: (value: string) => string;
}

export default function BookingEditHistory({ shopID, bookingID, formatTime }: BookingEditHistoryProps) {
  const [history, setHistory] = useState<BookingEdit[]>();
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    getBookingEditHistory(shopID, bookingID, controller.signal)
      .then(result => {
        if (!controller.signal.aborted) setHistory(result);
      })
      .catch(error => {
        if (!controller.signal.aborted) {
          setError(error instanceof Error ? error.message : "โหลดประวัติการแก้ไขไม่สำเร็จ");
        }
      });
    return () => controller.abort();
  }, [shopID, bookingID, retry]);

  const fields: { key: keyof BookingEditSnapshot; label: string; format?: (value: string | number) => string }[] = [
    { key: "customer_name", label: "ลูกค้า" },
    { key: "customer_phone", label: "เบอร์โทรศัพท์" },
    { key: "service_name", label: "บริการ" },
    { key: "starts_at", label: "เวลาเริ่ม", format: value => formatTime(String(value)) },
    { key: "ends_at", label: "เวลาสิ้นสุด", format: value => formatTime(String(value)) },
    { key: "duration_minutes", label: "ระยะเวลา", format: value => `${value} นาที` },
    { key: "price_satang", label: "ราคา", format: value => (Number(value) / 100).toLocaleString("th-TH", { style: "currency", currency: "THB" }) },
    { key: "note", label: "หมายเหตุ" },
  ];

  return (
    <>
      <Divider titlePlacement="start">ประวัติการแก้ไข</Divider>
      {error ? (
        <Alert type="error" title={error} action={
          <Button size="small" onClick={() => { setError(""); setRetry(value => value + 1); }}>ลองใหม่</Button>
        } />
      ) : (
        <Flex vertical gap={12}>
          {history?.length === 50 && <Typography.Text type="secondary">แสดง 50 รายการล่าสุด</Typography.Text>}
          <Table<BookingEdit>
            rowKey="id"
            size="small"
            dataSource={history ?? []}
            loading={!history}
            pagination={false}
            scroll={{ x: 420 }}
            locale={{ emptyText: "ยังไม่มีประวัติการแก้ไข" }}
            columns={[
              {
                title: "แก้ไขอะไร",
                key: "changes",
                render: (_, edit) => (
                  <Flex vertical gap={8}>
                    {fields.filter(({ key }) => edit.before[key] !== edit.after[key]).map(({ key, label, format }) => (
                      <div key={key} style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
                        <Typography.Text strong>{label}</Typography.Text>
                        <div>เดิม {format ? format(edit.before[key]) : edit.before[key] || "—"}</div>
                        <div>ใหม่ {format ? format(edit.after[key]) : edit.after[key] || "—"}</div>
                      </div>
                    ))}
                    {edit.before.service_id !== edit.after.service_id && edit.before.service_name === edit.after.service_name && (
                      <Typography.Text>เปลี่ยนบริการเป็นรายการชื่อเดียวกัน</Typography.Text>
                    )}
                  </Flex>
                ),
              },
              { title: "แก้โดย", dataIndex: "editor_name", width: 110 },
              {
                title: "วันที่และเวลา",
                dataIndex: "created_at",
                width: 130,
                render: (value: string) => `${formatTime(value)} น.`,
              },
            ]}
          />
        </Flex>
      )}
    </>
  );
}
