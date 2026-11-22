import { Col, Form, Input } from "antd";

export default function ShopHoursFields() {
  return (
    <>
      <Col xs={24} sm={12}>
        <Form.Item
          name="opening_time"
          label="เวลาเปิดร้าน"
          rules={[
            { required: true, message: "กรุณาระบุเวลาเปิด" },
            { pattern: /^([01]\d|2[0-3]):[0-5]\d$/, message: "ใช้รูปแบบ HH:mm เช่น 09:00" },
          ]}
        >
          <Input placeholder="09:00" maxLength={5} />
        </Form.Item>
      </Col>
      <Col xs={24} sm={12}>
        <Form.Item
          name="closing_time"
          label="เวลาปิดร้าน"
          dependencies={["opening_time"]}
          extra="เวลาเดียวกันทุกวัน ปิดเที่ยงคืนใช้ 24:00"
          rules={[
            { required: true, message: "กรุณาระบุเวลาปิด" },
            { pattern: /^(([01]\d|2[0-3]):[0-5]\d|24:00)$/, message: "ใช้รูปแบบ HH:mm เช่น 18:00" },
            ({ getFieldValue }) => ({
              validator: (_, value: string) => !value || value > getFieldValue("opening_time")
                ? Promise.resolve()
                : Promise.reject(new Error("เวลาปิดต้องหลังเวลาเปิดในวันเดียวกัน")),
            }),
          ]}
        >
          <Input placeholder="18:00" maxLength={5} />
        </Form.Item>
      </Col>
    </>
  );
}
