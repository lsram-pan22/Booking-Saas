
import { Button, Checkbox, Col, Form, Input, Row, Space, Typography } from "antd"
import { CalendarOutlined, LockOutlined, MailOutlined, ArrowRightOutlined } from "@ant-design/icons"
import type { LoginForm } from "../../types/auth";
import { useTranslation } from "react-i18next";

const { Title, Text } = Typography;

export default function Login() {
  const [form] = Form.useForm<LoginForm>();
  const { t } = useTranslation();

  const onFinish = async (values: LoginForm) => {
    console.log("Login:", values);
  };

  return (
    <Row>
      {/* Login Form */}
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
            maxWidth: 420,
            margin: "0 auto",
          }}
        >
          <Space orientation="vertical" size={8}>
            <Title level={2}>{t("auth.welcome_back")}</Title>
            <Text type="secondary">
              {t("auth.login_description")}
            </Text>
          </Space>

          <Form
            form={form}
            layout="vertical"
            size="large"
            initialValues={{ remember: true }}
            onFinish={onFinish}
            style={{ marginTop: 32 }}
          >
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
              label={t("common.password")}
              name="password"
              rules={[
                { required: true, message: "Please enter your password" },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="Enter your password"
              />
            </Form.Item>

            <Row justify="space-between" align="middle">
              <Col>
                <Form.Item
                  name="remember"
                  valuePropName="checked"
                  noStyle
                >
                  <Checkbox>{t("common.remember_me")}</Checkbox>
                </Form.Item>
              </Col>

              <Col>
                <Typography.Link href="/forgot-password">
                  {t("common.forgot_password")}
                </Typography.Link>
              </Col>
            </Row>

            <Form.Item style={{ marginTop: 24 }}>
              <Button
                type="primary"
                htmlType="submit"
                block
                icon={<ArrowRightOutlined />}
                iconPlacement="end"
              >
                {t("common.sign_in")}
              </Button>
            </Form.Item>
          </Form>

          <Row justify="center">
            <Text type="secondary">
              {t("auth.dont_have_account")}{" "}
              <Typography.Link href="/register">
                {t("common.sign_up")}
              </Typography.Link>
            </Text>
          </Row>
        </div>

        <Text type="secondary" style={{ textAlign: "center", fontSize: 12 }}>
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
            Manage your business.
            <br />
            Simplify your bookings.
          </Title>

          <Text style={{ color: "#FFFFFF", fontSize: 16 }}>
            Everything you need to manage appointments,
            staff and customers in one place.
          </Text>
        </Space>
      </Col>
    </Row>
  );
}
