import { Space, Typography } from "antd";
import { CalendarOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";

const { Title } = Typography;

export default function AppBrand() {
  const { t } = useTranslation();

  return (
    <Space size="middle">
      <CalendarOutlined
        style={{ fontSize: 28, color: "#6750E8" }}
      />

      <Title level={4} style={{ margin: 0 }}>
        {t("common.app_name")}
      </Title>
    </Space>
  );
}