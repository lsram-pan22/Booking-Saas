import { Col, Form, TimePicker } from "antd";
import dayjs from "dayjs";
import type { Dayjs } from "dayjs";

interface BookingStartTimeFieldProps {
  openingTime: string | null;
  closingTime: string | null;
}

export default function BookingStartTimeField({ openingTime, closingTime }: BookingStartTimeFieldProps) {
  return (
    <Col xs={24} sm={12}>
      <Form.Item
        name="time"
        label="เวลาเริ่ม"
        extra={openingTime && closingTime ? `ร้านเปิด ${openingTime}–${closingTime} น.` : undefined}
        rules={[{ required: true, message: "กรุณาเลือกเวลา" }]}
        getValueProps={(value?: string) => ({ value: value ? dayjs(`2000-01-01T${value}:00`) : null })}
        normalize={(value: Dayjs | null) => value?.format("HH:mm") ?? ""}
      >
        <TimePicker
          format="HH:mm"
          use12Hours={false}
          needConfirm={false}
          placeholder="เช่น 18:00"
          style={{ width: "100%" }}
        />
      </Form.Item>
    </Col>
  );
}
