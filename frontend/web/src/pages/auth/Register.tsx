
import {Button, Col, Form, Input, Row, Space, Typography, Checkbox } from "antd";
import { CalendarOutlined, UserOutlined, MailOutlined, LockOutlined, PhoneOutlined, ArrowRightOutlined } from "@ant-design/icons";
import type { RegisterForm } from "../../types/auth";
import { useTranslation } from "react-i18next";

const { Title, Text } = Typography;

export default function Register() {
  const [form] = Form.useForm<RegisterForm>();
  const { t } = useTranslation();

  const onFinish = async (values: RegisterForm) => {
    console.log("Register:", values);

    // TODO: เชื่อมต่อ Register API
  };

  return (
    <Row>
      {/* Register Form */}
      <Col
        xs={24}
        md={12}
        style={{
          display: "flex",
          flexDirection: "column",
        }}
      >
        <div
          style={{
            width: "100%",
            maxWidth: 440,
            margin: "0 auto",
          }}
        >
          <Space orientation="vertical" size={4}>
            <Title level={2}>{t("common.sign_up")}</Title>

            <Text type="secondary">
              {t("auth.register_description")}
            </Text>
          </Space>

          <Form
            form={form}
            layout="vertical"
            size="large"
            onFinish={onFinish}
            style={{ marginTop: 28 }}
          >
            <Form.Item
              label={t("common.full_name")}
              name="full_name"
              rules={[
                { required: true, message: "Please enter your name" },
              ]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="Your full name"
              />
            </Form.Item>

            <Form.Item
              label={t("common.email")}
              name="email"
              rules={[
                { required: true, message: "Please enter your email" },
                { type: "email", message: "Invalid email address" },
              ]}
            >
              <Input
                prefix={<MailOutlined />}
                placeholder="name@example.com"
              />
            </Form.Item>

            <Form.Item
              label={t("common.phone")}
              name="phone"
              rules={[
                { required: true, message: "Please enter your phone number" },
                {
                  pattern: /^[0-9]{9,10}$/,
                  message: "Please enter a valid phone number",
                },
              ]}
            >
              <Input
                prefix={<PhoneOutlined />}
                placeholder="0812345678"
                maxLength={10}
              />
            </Form.Item>

            <Form.Item
              label={t("common.password")}
              name="password"
              rules={[
                { required: true, message: "Please enter your password" },
                { min: 8, message: "Password must be at least 8 characters" },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="At least 8 characters"
              />
            </Form.Item>

            <Form.Item
              label={t("common.confirm_password")}
              name="confirm_password"
              dependencies={["password"]}
              rules={[
                { required: true, message: "Please confirm your password" },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue("password") === value) {
                      return Promise.resolve();
                    }

                    return Promise.reject(
                      new Error("Passwords do not match")
                    );
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="Confirm your password"
              />
            </Form.Item>

            <Form.Item
              name="accept_terms"
              valuePropName="checked"
              rules={[
                {
                  validator: (_, value) =>
                    value
                      ? Promise.resolve()
                      : Promise.reject(
                          new Error("Please accept the terms and conditions")
                        ),
                },
              ]}
            >
              <Checkbox>
                {t("auth.terms")}
              </Checkbox>
            </Form.Item>

            <Form.Item style={{ marginTop: 24 }}>
              <Button
                type="primary"
                htmlType="submit"
                block
                icon={<ArrowRightOutlined />}
                iconPlacement="end"
              >
                {t("common.sign_up")}
              </Button>
            </Form.Item>
          </Form>

          <Row justify="center">
            <Text type="secondary">
              {t("auth.already_have_account")}{" "}
              <Typography.Link href="/login">{t("common.sign_in")}</Typography.Link>
            </Text>
          </Row>
        </div>

        <Text
          type="secondary"
          style={{ textAlign: "center", fontSize: 12 }}
        >
          © 2026 Booking SaaS. All rights reserved.
        </Text>
      </Col>

      {/* Banner */}
      <Col
        xs={0}
        md={12}
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: 48,
          background: "#6750E8",
        }}
      >
        <Space orientation="vertical" size="large">
          <CalendarOutlined
            style={{ fontSize: 72, color: "#FFFFFF" }}
          />

          <Title style={{ color: "#FFFFFF", fontSize: 40 }}>
            Start your journey.
            <br />
            Grow your business.
          </Title>

          <Text style={{ color: "#FFFFFF", fontSize: 16 }}>
            Organize your appointments, manage your team
            and deliver better customer experiences.
          </Text>
        </Space>
      </Col>
    </Row>
  );
}
