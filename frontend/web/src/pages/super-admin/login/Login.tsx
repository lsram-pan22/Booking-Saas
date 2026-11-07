
import { useState } from "react";
import {
  Button,
  Card,
  Flex,
  Form,
  Input,
  Typography,
  App,
} from "antd";
import {
  LockOutlined,
  UserOutlined,
  SafetyCertificateOutlined,
} from "@ant-design/icons";
import { Navigate, useNavigate } from "react-router-dom";
import { loginSuperAdmin } from "../../../api/super-admin/auth.api";
import type { LoginRequest } from "../../../types/shared/auth.type";
import {
  saveAuthSession,
  isAuthenticated,
} from "../../../utils/auth";

const { Title, Text } = Typography;

export default function Login() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);

  if (isAuthenticated("super_admin")) {
    return <Navigate to="/super-admin/dashboard" replace />;
  }

  const onFinish = async (values: LoginRequest) => {
    setLoading(true);

    try {
      const result = await loginSuperAdmin(values);

      saveAuthSession(result.data);

      message.success("Login successful");

      navigate("/super-admin/dashboard", {
        replace: true,
      });
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : "Unable to connect to server";

      message.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Flex
      justify="center"
      align="center"
      style={{
        minHeight: "100vh",
        padding: 24,
        background: "#F3F5F2",
      }}
    >
      <Card
        style={{
          width: "100%",
          maxWidth: 420,
          borderRadius: 16,
          boxShadow: "0 8px 30px rgba(0,0,0,0.06)",
        }}
        styles={{ body: { padding: 32 } }}
      >
        <Flex
          vertical
          align="center"
          gap={8}
          style={{ marginBottom: 32 }}
        >
          <SafetyCertificateOutlined
            style={{
              fontSize: 42,
              color: "#7F867B",
              marginBottom: 8,
            }}
          />

          <Title level={3} style={{ margin: 0 }}>
            Booking SaaS
          </Title>

          <Text type="secondary">
            Super Admin Management
          </Text>
        </Flex>

        <Form<LoginRequest>
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
          requiredMark={false}
        >
          <Form.Item
            label="Username"
            name="username"
            rules={[
              {
                required: true,
                message: "Please enter your username",
              },
            ]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="Enter your username"
            />
          </Form.Item>

          <Form.Item
            label="Password"
            name="password"
            rules={[
              {
                required: true,
                message: "Please enter your password",
              },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Enter your password"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 12 }}>
            <Button
              htmlType="submit"
              type="primary"
              block
              loading={loading}
              style={{
                background: "#7F867B",
                borderColor: "#7F867B",
              }}
            >
              Sign In
            </Button>
          </Form.Item>
        </Form>

        <Text
          type="secondary"
          style={{
            display: "block",
            textAlign: "center",
            fontSize: 12,
          }}
        >
          Authorized access only
        </Text>
      </Card>
    </Flex>
  );
}
