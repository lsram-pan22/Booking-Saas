import {
  Card,
  Col,
  Row,
  Statistic,
  Typography,
  Table,
  Empty,
} from "antd";

import {
  ShopOutlined,
  UserOutlined,
  CalendarOutlined,
  DollarOutlined,
} from "@ant-design/icons";

const { Title, Text } = Typography;

export default function Dashboard() {
  const columns = [
    {
      title: "Business Name",
      dataIndex: "business_name",
      key: "business_name",
    },
    {
      title: "Owner",
      dataIndex: "owner",
      key: "owner",
    },
    {
      title: "Package",
      dataIndex: "package",
      key: "package",
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
    },
    {
      title: "Created At",
      dataIndex: "created_at",
      key: "created_at",
    },
  ];

  return (
    <>
      <Title level={3}>Dashboard</Title>
      <Text type="secondary">
        Welcome to Booking SaaS Administration
      </Text>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Businesses"
              value={0}
              prefix={<ShopOutlined />}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Users"
              value={0}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Bookings"
              value={0}
              prefix={<CalendarOutlined />}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Revenue"
              value={0}
              prefix={<DollarOutlined />}
              suffix="฿"
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="Recent Businesses"
        style={{ marginTop: 24 }}
      >
        <Table
          columns={columns}
          dataSource={[]}
          rowKey="id"
          pagination={false}
          locale={{
            emptyText: <Empty description="No businesses found" />,
          }}
        />
      </Card>
    </>
  );
}