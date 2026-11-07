
import { Select, Space } from "antd";
import { GlobalOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";

export default function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const changeLanguage = (language: string) => {
    i18n.changeLanguage(language);
    localStorage.setItem("app_language", language);
  };

  return (
    <Space align="center" size={8}>
      <GlobalOutlined style={{ fontSize: 18, color: "#6750E8" }} />

      <Select
        value={i18n.resolvedLanguage}
        onChange={changeLanguage}
        style={{ width: 130 }}
        variant="outlined"
        options={[
          {
            label: "🇹🇭 ไทย",
            value: "th",
          },
          {
            label: "🇬🇧 English",
            value: "en",
          },
        ]}
      />
    </Space>
  );
}
